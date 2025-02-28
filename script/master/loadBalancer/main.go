package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	servers          = []string{"192.168.56.21", "192.168.56.22"}
	availableServers = []string{}
	cpuLoad          = make(map[string]float64)
	mu               sync.Mutex
)

// Check if a server is available via SSH
func isServerAvailable(host string) bool {
	command := fmt.Sprintf(`sudo sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no vagrant@%s "echo OK"`, host)
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	return err == nil
}

// Update available servers every 5 minutes
func monitorServerAvailability() {
	for {
		tempAvailable := []string{}

		for _, server := range servers {
			if isServerAvailable(server) {
				tempAvailable = append(tempAvailable, server)
			}
		}

		mu.Lock()
		availableServers = tempAvailable
		mu.Unlock()

		fmt.Println("Available servers:", availableServers)

		time.Sleep(1 * time.Minute) // Check every 5 minutes
	}
}

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

		mu.Lock()
		activeServers := availableServers // Copy available servers to avoid locking too long
		mu.Unlock()

		for _, server := range activeServers {
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

		time.Sleep(1 * time.Minute)
	}
}


func getRandomServerByLoad() string {
	mu.Lock()
	defer mu.Unlock()

	if len(availableServers) == 0 {
		fmt.Println("No available servers!")
		return ""
	}

	var totalWeight float64
	weights := make(map[string]float64)

	for _, server := range availableServers {
		weight := 100 - cpuLoad[server]
		if weight < 0 {
			weight = 0 // Prevent negative weights
		}
		weights[server] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return availableServers[rand.Intn(len(availableServers))] // All servers overloaded, pick random
	}

	// Random selection based on weights
	rnd := rand.Float64() * totalWeight
	accumulated := 0.0

	for _, server := range availableServers {
		accumulated += weights[server]
		if rnd <= accumulated {
			return server
		}
	}

	return availableServers[0] // Fallback case
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	target := getRandomServerByLoad()

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
	go monitorServerAvailability()
	go monitorCPUUsage()

	http.HandleFunc("/", handleRequest)
	fmt.Println("Load balancer running on port 9000...")
	http.ListenAndServe(":9000", nil)
}
