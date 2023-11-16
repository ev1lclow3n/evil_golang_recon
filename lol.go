package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	ColorRed    = "\033[91m"
	ColorOrange = "\033[38;5;208m"
	ColorLime   = "\033[92m"
	ColorBlue   = "\033[94m"
	ColorReset  = "\033[0m"
)

func getStatusCodeColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode <= 220:
		return ColorLime
	case statusCode >= 400 && statusCode <= 420:
		return ColorRed
	case statusCode >= 300 && statusCode <= 320:
		return ColorOrange
	case statusCode >= 500 && statusCode <= 550:
		return ColorBlue
	default:
		return ""
	}
}

func checkSubdomainsStatus(subdomains []string) map[string]int {
	statusCodes := make(map[string]int)
	var wg sync.WaitGroup

	for _, subdomain := range subdomains {
		wg.Add(1)
		go func(sd string) {
			defer wg.Done()
			url := "http://" + sd

			resp, err := http.Head(url)
			if err != nil {
				statusCodes[url] = http.StatusNotFound
				return
			}
			defer resp.Body.Close()

			statusCodes[url] = resp.StatusCode
		}(subdomain)
	}

	wg.Wait()
	return statusCodes
}

func scanPorts(subdomain string, wg *sync.WaitGroup) {
	defer wg.Done()

	for port := 20; port <= 9999; port++ {
		target := fmt.Sprintf("%s:%d", subdomain, port)
		conn, err := net.DialTimeout("tcp", target, 1*time.Second)
		if err != nil {
			continue
		}
		defer conn.Close()

		fmt.Printf("%s%s is accessible on port %d%s\n", ColorLime, subdomain, port, ColorReset)
	}
}

func readSubdomainsFromFile(filename string) ([]string, error) {
	var subdomains []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		subdomains = append(subdomains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subdomains, nil
}

func main() {
	// Replace with your subdomains file path
	subdomainsFile := "subdomains.txt"

	subdomains, err := readSubdomainsFromFile(subdomainsFile)
	if err != nil {
		fmt.Println("Error reading subdomains file:", err)
		return
	}

	results := checkSubdomainsStatus(subdomains)

	// Print status codes for each subdomain with colors
	for subdomain, status := range results {
		color := getStatusCodeColor(status)
		resetColor := ColorReset
		if color == "" {
			resetColor = ""
		}
		if status != http.StatusNotFound {
			fmt.Printf("%s%s : %d%s\n", color, subdomain, status, resetColor)
		}
	}

	var wg sync.WaitGroup
	for _, subdomain := range subdomains {
		wg.Add(1)
		go scanPorts(subdomain, &wg)
	}

	wg.Wait()
}

