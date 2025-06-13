package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run http2-test.go <url>")
	}

	url := os.Args[1]

	fmt.Printf("🧪 Testing HTTP/2 connection to: %s\n", url)
	fmt.Println("=" + fmt.Sprintf("%*s", len(url)+35, "="))

	// Test HTTP/2 with different configurations
	testConfigs := []struct {
		name   string
		client *http.Client
	}{
		{
			name:   "HTTP/2 with TLS",
			client: createHTTP2TLSClient(),
		},
		{
			name:   "HTTP/2 Cleartext (H2C)",
			client: createHTTP2CleartextClient(),
		},
		{
			name:   "Standard HTTP/1.1",
			client: createHTTP1Client(),
		},
	}

	for _, config := range testConfigs {
		fmt.Printf("\n🔍 Testing: %s\n", config.name)
		fmt.Println(strings.Repeat("-", 40))

		start := time.Now()
		resp, err := config.client.Get(url)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Printf("Warning: Error closing response body: %v\n", closeErr)
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ Error reading response: %v\n", err)
			continue
		}

		fmt.Printf("✅ Status: %s\n", resp.Status)
		fmt.Printf("✅ Protocol: %s\n", resp.Proto)
		fmt.Printf("✅ Duration: %v\n", duration)
		fmt.Printf("✅ Content Length: %d bytes\n", len(body))

		// Show important headers
		for _, header := range []string{"Server", "Content-Type", "Content-Encoding"} {
			if value := resp.Header.Get(header); value != "" {
				fmt.Printf("✅ %s: %s\n", header, value)
			}
		}

		// Test server push if available
		if resp.Proto == "HTTP/2.0" {
			fmt.Printf("🚀 HTTP/2 Features:\n")
			fmt.Printf("   • Multiplexing: ✅ Available\n")
			fmt.Printf("   • Server Push: %s\n", checkServerPush(resp))
			fmt.Printf("   • Header Compression: ✅ Available\n")
		}
	}

	// Concurrent request test for HTTP/2 multiplexing
	fmt.Printf("\n🚀 Testing HTTP/2 Multiplexing (10 concurrent requests)\n")
	fmt.Println(strings.Repeat("-", 50))
	testConcurrentRequests(url)
}

func createHTTP2TLSClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For self-signed certificates
			},
		},
		Timeout: 30 * time.Second,
	}
}

func createHTTP2CleartextClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
		Timeout: 30 * time.Second,
	}
}

func createHTTP1Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 30 * time.Second,
	}
}

func checkServerPush(resp *http.Response) string {
	if resp.Header.Get("Link") != "" {
		return "✅ Hints Available"
	}
	return "❌ Not Available"
}

func testConcurrentRequests(url string) {
	client := createHTTP2TLSClient()

	// If HTTPS fails, try HTTP/2 cleartext
	if resp, err := client.Get(url); err != nil || resp.StatusCode != 200 {
		client = createHTTP2CleartextClient()
	}

	start := time.Now()
	results := make(chan time.Duration, 10)

	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		go func(id int) {
			requestStart := time.Now()
			resp, err := client.Get(fmt.Sprintf("%s/health", url))
			requestDuration := time.Since(requestStart)

			if err != nil {
				fmt.Printf("❌ Request %d failed: %v\n", id+1, err)
				results <- 0
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("Warning: Failed to close response body: %v\n", err)
				}
			}()

			fmt.Printf("✅ Request %d completed in %v (%s)\n", id+1, requestDuration, resp.Proto)
			results <- requestDuration
		}(i)
	}

	// Collect results
	var total time.Duration
	successful := 0
	for i := 0; i < 10; i++ {
		duration := <-results
		if duration > 0 {
			total += duration
			successful++
		}
	}

	totalTime := time.Since(start)
	fmt.Printf("\n📊 Concurrent Test Results:\n")
	fmt.Printf("✅ Total time: %v\n", totalTime)
	fmt.Printf("✅ Successful requests: %d/10\n", successful)
	if successful > 0 {
		fmt.Printf("✅ Average response time: %v\n", total/time.Duration(successful))
		fmt.Printf("✅ Requests per second: %.2f\n", float64(successful)/totalTime.Seconds())
	}
}
