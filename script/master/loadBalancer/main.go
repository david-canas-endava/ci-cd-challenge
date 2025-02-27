package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	servers = []string{"192.168.56.21", "192.168.56.22"}
	cpuLoad = make(map[string]float64)
	mu      sync.Mutex
)

// Fetch CPU usage from a remote server via SSH
func getCPUUsage(host string) (float64, error) {
	command := fmt.Sprintf(`sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no vagrant@%s mpstat | awk '/all/ {print 100 - $NF}'`, host)
	cmd := exec.Command("sh", "-c", command)
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

func monitorCPUUsage() {
	for {
		tempLoad := make(map[string]float64)

		for _, server := range servers {
			usage, err := getCPUUsage(server)
			if err == nil {
				tempLoad[server] = usage
			} else {
				fmt.Printf("Error getting CPU usage for %s: %v\n", server, err)
			}
		}

		// Lock and update global cpuLoad map
		mu.Lock()
		cpuLoad = tempLoad
		mu.Unlock()

		// Wait 5 minutes before updating again
		time.Sleep(5 * time.Minute)
	}
}
func getLeastLoadedServer() string {
	mu.Lock()
	defer mu.Unlock()

	if len(cpuLoad) == 0 {
		return servers[0] // Fallback to first server if no data
	}

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
	// Start CPU monitoring goroutine
	go monitorCPUUsage()

	http.HandleFunc("/", handleRequest)
	fmt.Println("Load balancer running on port 9000...")
	http.ListenAndServe(":9000", nil)
}