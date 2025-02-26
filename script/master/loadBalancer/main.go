package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var servers = []string{"192.168.56.21", "192.168.56.22"}

func getCPUUsage(host string) (float64, error) {
	cmd := exec.Command(`sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no vagrant@%s "awk 'NR==1 {t=\$2+\$4+\$5; if (t == 0) print 0; else print (\$2+\$4) / t * 100}' /proc/stat"`,host)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}
	return usage, nil
}

func getLeastLoadedServer() string {
	var wg sync.WaitGroup
	cpuLoad := make(map[string]float64)
	mu := sync.Mutex{}

	for _, server := range servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			usage, err := getCPUUsage(s)
			if err == nil {
				mu.Lock()
				cpuLoad[s] = usage
				mu.Unlock()
			}
		}(server)
	}

	wg.Wait()

	// Choose the server with the lowest CPU load
	leastLoaded := servers[0]
	for _, server := range servers {
		if cpuLoad[server] < cpuLoad[leastLoaded] {
			leastLoaded = server
		}
	}
	return leastLoaded
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    target := getLeastLoadedServer()
    if target == "" {
        http.Error(w, "No available servers", http.StatusServiceUnavailable)
        return
    }

    // Construct full target URL
    targetURL := "http://" + target + ":3000" + r.URL.Path

    // Create a new request
    req, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        http.Error(w, "Error creating request", http.StatusInternalServerError)
        return
    }

    // Copy headers from original request to preserve content type, cookies, etc.
    req.Header = r.Header

    // Perform the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Error forwarding request", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    // Copy headers and status code from backend response
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    w.WriteHeader(resp.StatusCode)

    // Copy response body to client
    io.Copy(w, resp.Body)
}


func main() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("Load balancer running on port 9000...")
	http.ListenAndServe(":9000", nil)
}
