package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"news/internal/json"
)

// API endpoint test configuration
type APITest struct {
	Name     string `json:"name"`
	Method   string `json:"method"`
	URL      string `json:"url"`
	Expected int    `json:"expected_status"`
}

// Test result structure
type TestResult struct {
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	Method     string        `json:"method"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
}

// Statistics for latency analysis
type LatencyStats struct {
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
	Average time.Duration `json:"average"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	Total   int           `json:"total_requests"`
	Success int           `json:"successful_requests"`
	Failed  int           `json:"failed_requests"`
}

func main() {
	fmt.Println("🚀 News API Latency Test Suite")
	fmt.Println("==============================")

	baseURL := "http://localhost:8081"

	// Check if API is running
	if !isAPIRunning(baseURL) {
		fmt.Printf("❌ API is not running on %s\n", baseURL)
		fmt.Println("Please start the API with: make dev")
		os.Exit(1)
	}

	// Define API endpoints to test
	apiTests := []APITest{
		// Health and Basic endpoints
		{"Health Check", "GET", baseURL + "/health", 200},
		{"API Info", "GET", baseURL + "/api", 200},

		// Public endpoints (no auth required)
		{"Get Articles", "GET", baseURL + "/api/articles", 200},
		{"Get Categories", "GET", baseURL + "/api/categories", 200},
		{"Get Tags", "GET", baseURL + "/api/tags", 200},
		{"Search Articles", "GET", baseURL + "/api/search?q=test", 200},

		// Media endpoints
		{"Get Media Files", "GET", baseURL + "/api/media", 200},
		{"Media Stats", "GET", baseURL + "/api/media/stats", 200},

		// Analytics endpoints
		{"Analytics Dashboard", "GET", baseURL + "/api/analytics", 200},

		// User endpoints (may require auth but should respond)
		{"Get Users", "GET", baseURL + "/api/users", 401}, // Expected 401 without auth

		// Settings endpoints
		{"Get Settings", "GET", baseURL + "/api/settings", 200},

		// Translation endpoints
		{"Translation Progress", "GET", baseURL + "/admin/translations/progress", 401}, // Expected 401 without auth

		// WebSocket status endpoints
		{"Notification Stats", "GET", baseURL + "/ws/stats", 200},

		// Documentation endpoints
		{"Swagger UI", "GET", baseURL + "/docs/swagger-ui", 200},
		{"Swagger JSON", "GET", baseURL + "/swagger/doc.json", 200},
	}

	fmt.Printf("🔍 Testing %d endpoints...\n\n", len(apiTests))

	// Run tests
	results := runLatencyTests(apiTests)

	// Display results
	displayResults(results)

	// Calculate and display statistics
	stats := calculateStats(results)
	displayStats(stats)

	// Save results to file
	saveResults(results, stats)
}

func isAPIRunning(baseURL string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("Debug: Error connecting to API: %v\n", err)
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: Error closing response body: %v\n", closeErr)
		}
	}()
	return resp.StatusCode == 200
}

func runLatencyTests(tests []APITest) []TestResult {
	var results []TestResult
	client := &http.Client{Timeout: 30 * time.Second}

	for i, test := range tests {
		fmt.Printf("Testing %d/%d: %s", i+1, len(tests), test.Name)

		start := time.Now()
		resp, err := client.Get(test.URL) // Using GET for simplicity
		latency := time.Since(start)

		result := TestResult{
			Name:    test.Name,
			URL:     test.URL,
			Method:  test.Method,
			Latency: latency,
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.StatusCode = 0
			fmt.Printf(" ❌ (Error: %v)\n", err)
		} else {
			result.StatusCode = resp.StatusCode
			result.Success = (resp.StatusCode == test.Expected)

			if result.Success {
				fmt.Printf(" ✅ (%dms)\n", latency.Milliseconds())
			} else {
				fmt.Printf(" ⚠️  (Got %d, expected %d - %dms)\n",
					resp.StatusCode, test.Expected, latency.Milliseconds())
			}
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Printf("Warning: Error closing response body: %v\n", closeErr)
			}
		}

		results = append(results, result)
		time.Sleep(50 * time.Millisecond) // Small delay between requests
	}

	return results
}

func displayResults(results []TestResult) {
	fmt.Println("\n📊 Detailed Results:")
	fmt.Println("=====================")
	fmt.Printf("%-30s %-8s %-12s %-8s %s\n", "Endpoint", "Status", "Latency", "Success", "URL")
	fmt.Println(strings.Repeat("-", 100))

	for _, result := range results {
		status := fmt.Sprintf("%d", result.StatusCode)
		if result.StatusCode == 0 {
			status = "ERROR"
		}

		successIcon := "✅"
		if !result.Success {
			successIcon = "❌"
		}

		fmt.Printf("%-30s %-8s %-12s %-8s %s\n",
			truncateString(result.Name, 29),
			status,
			fmt.Sprintf("%dms", result.Latency.Milliseconds()),
			successIcon,
			result.URL,
		)
	}
}

func calculateStats(results []TestResult) LatencyStats {
	var latencies []time.Duration
	successCount := 0

	for _, result := range results {
		latencies = append(latencies, result.Latency)
		if result.Success {
			successCount++
		}
	}

	if len(latencies) == 0 {
		return LatencyStats{}
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}

	stats := LatencyStats{
		Min:     latencies[0],
		Max:     latencies[len(latencies)-1],
		Average: total / time.Duration(len(latencies)),
		Total:   len(results),
		Success: successCount,
		Failed:  len(results) - successCount,
	}

	// Calculate percentiles
	if len(latencies) > 0 {
		p95Index := int(float64(len(latencies)) * 0.95)
		if p95Index >= len(latencies) {
			p95Index = len(latencies) - 1
		}
		stats.P95 = latencies[p95Index]

		p99Index := int(float64(len(latencies)) * 0.99)
		if p99Index >= len(latencies) {
			p99Index = len(latencies) - 1
		}
		stats.P99 = latencies[p99Index]
	}

	return stats
}

func displayStats(stats LatencyStats) {
	fmt.Println("\n📈 Latency Statistics:")
	fmt.Println("======================")
	fmt.Printf("Total Requests:     %d\n", stats.Total)
	fmt.Printf("Successful:         %d (%.1f%%)\n", stats.Success,
		float64(stats.Success)/float64(stats.Total)*100)
	fmt.Printf("Failed:             %d (%.1f%%)\n", stats.Failed,
		float64(stats.Failed)/float64(stats.Total)*100)
	fmt.Println()
	fmt.Printf("Min Latency:        %dms\n", stats.Min.Milliseconds())
	fmt.Printf("Max Latency:        %dms\n", stats.Max.Milliseconds())
	fmt.Printf("Average Latency:    %dms\n", stats.Average.Milliseconds())
	fmt.Printf("95th Percentile:    %dms\n", stats.P95.Milliseconds())
	fmt.Printf("99th Percentile:    %dms\n", stats.P99.Milliseconds())

	// Performance assessment
	fmt.Println("\n🎯 Performance Assessment:")
	fmt.Println("===========================")
	avgMs := stats.Average.Milliseconds()
	if avgMs < 50 {
		fmt.Println("🟢 Excellent: Average latency under 50ms")
	} else if avgMs < 100 {
		fmt.Println("🟡 Good: Average latency under 100ms")
	} else if avgMs < 200 {
		fmt.Println("🟠 Fair: Average latency under 200ms")
	} else {
		fmt.Println("🔴 Poor: Average latency over 200ms - needs optimization")
	}

	if stats.P95.Milliseconds() > 500 {
		fmt.Println("⚠️  Warning: 95th percentile latency is high (>500ms)")
	}
}

func saveResults(results []TestResult, stats LatencyStats) {
	report := struct {
		Timestamp time.Time    `json:"timestamp"`
		Results   []TestResult `json:"results"`
		Stats     LatencyStats `json:"statistics"`
	}{
		Timestamp: time.Now(),
		Results:   results,
		Stats:     stats,
	}

	filename := fmt.Sprintf("latency_test_report_%s.json",
		time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("⚠️  Could not save report: %v\n", err)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: Error closing report file: %v\n", closeErr)
		}
	}()

	// Use MarshalIndent for pretty printing
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Could not encode report: %v\n", err)
		return
	}

	if _, err := file.Write(jsonData); err != nil {
		fmt.Printf("⚠️  Could not write report: %v\n", err)
		return
	}

	fmt.Printf("\n💾 Report saved to: %s\n", filename)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
