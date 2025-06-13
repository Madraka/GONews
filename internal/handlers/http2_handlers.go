package handlers

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// HTTP2StatusResponse represents HTTP/2 server status
type HTTP2StatusResponse struct {
	Protocol         string                 `json:"protocol"`
	HTTP2Enabled     bool                   `json:"http2_enabled"`
	TLSEnabled       bool                   `json:"tls_enabled"`
	H2CEnabled       bool                   `json:"h2c_enabled"`
	ServerPush       bool                   `json:"server_push_available"`
	Multiplexing     bool                   `json:"multiplexing_available"`
	ConnectionInfo   ConnectionInfo         `json:"connection_info"`
	PerformanceStats PerformanceStats       `json:"performance_stats"`
	Features         map[string]interface{} `json:"features"`
}

// ConnectionInfo represents connection details
type ConnectionInfo struct {
	RemoteAddr  string `json:"remote_addr"`
	LocalAddr   string `json:"local_addr"`
	TLSVersion  string `json:"tls_version,omitempty"`
	CipherSuite string `json:"cipher_suite,omitempty"`
	Compression string `json:"compression,omitempty"`
	KeepAlive   bool   `json:"keep_alive"`
}

// PerformanceStats represents performance metrics
type PerformanceStats struct {
	Goroutines       int    `json:"goroutines"`
	MemoryUsage      string `json:"memory_usage"`
	ConnectionsCount int    `json:"active_connections"`
}

// GetHTTP2Status returns HTTP/2 server status and capabilities
// @Summary Get HTTP/2 server status
// @Description Returns detailed information about HTTP/2 server capabilities and current connection
// @Tags System
// @Produce json
// @Success 200 {object} HTTP2StatusResponse
// @Router /debug/http2/status [get]
func GetHTTP2Status(c *gin.Context) {
	req := c.Request

	// Detect protocol version
	protocol := req.Proto
	isHTTP2 := protocol == "HTTP/2.0"
	isH2C := isHTTP2 && req.TLS == nil
	isTLS := req.TLS != nil

	// Check server push capability
	pusher := c.Writer.Pusher()
	hasServerPush := pusher != nil

	// Get connection info
	connInfo := ConnectionInfo{
		RemoteAddr: req.RemoteAddr,
		LocalAddr:  req.Host,
		KeepAlive:  req.Header.Get("Connection") != "close",
	}

	// TLS information if available
	if req.TLS != nil {
		connInfo.TLSVersion = getTLSVersion(req.TLS.Version)
		connInfo.CipherSuite = getCipherSuiteName(req.TLS.CipherSuite)
	}

	// Check compression
	if req.Header.Get("Accept-Encoding") != "" {
		connInfo.Compression = req.Header.Get("Accept-Encoding")
	}

	// Performance stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	perfStats := PerformanceStats{
		Goroutines:  runtime.NumGoroutine(),
		MemoryUsage: formatBytes(m.Alloc),
		// Note: Getting active connections count would require additional tracking
		ConnectionsCount: 1, // Current connection
	}

	// Feature detection
	features := map[string]interface{}{
		"server_push_hints":     req.Header.Get("Link") != "",
		"header_compression":    isHTTP2, // HTTP/2 has HPACK compression
		"stream_multiplexing":   isHTTP2,
		"binary_framing":        isHTTP2,
		"flow_control":          isHTTP2,
		"stream_prioritization": isHTTP2,
	}

	response := HTTP2StatusResponse{
		Protocol:         protocol,
		HTTP2Enabled:     isHTTP2,
		TLSEnabled:       isTLS,
		H2CEnabled:       isH2C,
		ServerPush:       hasServerPush,
		Multiplexing:     isHTTP2,
		ConnectionInfo:   connInfo,
		PerformanceStats: perfStats,
		Features:         features,
	}

	c.JSON(http.StatusOK, response)
}

// TestHTTP2Push demonstrates HTTP/2 server push capability
// @Summary Test HTTP/2 server push
// @Description Demonstrates HTTP/2 server push by pushing static resources
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /debug/http2/push-test [get]
func TestHTTP2Push(c *gin.Context) {
	pusher := c.Writer.Pusher()
	if pusher == nil {
		c.JSON(http.StatusOK, gin.H{
			"server_push": false,
			"message":     "Server push not available (not HTTP/2 or client doesn't support it)",
			"protocol":    c.Request.Proto,
		})
		return
	}

	// Try to push some resources
	pushTargets := []string{
		"/health",
		"/metrics",
	}

	pushedResources := []string{}
	failedPushes := []string{}

	for _, target := range pushTargets {
		if err := pusher.Push(target, nil); err != nil {
			failedPushes = append(failedPushes, target)
		} else {
			pushedResources = append(pushedResources, target)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"server_push":      true,
		"protocol":         c.Request.Proto,
		"pushed_resources": pushedResources,
		"failed_pushes":    failedPushes,
		"push_count":       len(pushedResources),
		"message":          "HTTP/2 server push test completed",
	})
}

// Helper functions

func getTLSVersion(version uint16) string {
	switch version {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

func getCipherSuiteName(cipherSuite uint16) string {
	// Map of common cipher suites
	cipherSuites := map[uint16]string{
		0x1301: "TLS_AES_128_GCM_SHA256",
		0x1302: "TLS_AES_256_GCM_SHA384",
		0x1303: "TLS_CHACHA20_POLY1305_SHA256",
		0xc02f: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		0xc030: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		0xcca8: "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
	}

	if name, exists := cipherSuites[cipherSuite]; exists {
		return name
	}
	return "Unknown"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
