package server

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HTTP2Config holds HTTP/2 server configuration
type HTTP2Config struct {
	Port              string
	CertFile          string
	KeyFile           string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int
	H2CEnabled        bool // HTTP/2 Cleartext for development
}

// DefaultHTTP2Config returns default HTTP/2 configuration
func DefaultHTTP2Config() *HTTP2Config {
	readTimeout := time.Duration(getEnvInt("SERVER_READ_TIMEOUT_SECONDS", 30)) * time.Second
	writeTimeout := time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT_SECONDS", 30)) * time.Second
	idleTimeout := time.Duration(getEnvInt("SERVER_IDLE_TIMEOUT_SECONDS", 120)) * time.Second
	readHeaderTimeout := time.Duration(getEnvInt("SERVER_READ_HEADER_TIMEOUT_SECONDS", 5)) * time.Second
	maxHeaderBytes := getEnvInt("SERVER_MAX_HEADER_BYTES", 4194304) // 4MB default

	return &HTTP2Config{
		Port:              getEnvOrDefault("PORT", "8080"),
		CertFile:          getEnvOrDefault("TLS_CERT_FILE", ""),
		KeyFile:           getEnvOrDefault("TLS_KEY_FILE", ""),
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		H2CEnabled:        getEnvOrDefault("HTTP2_H2C_ENABLED", "true") == "true",
	}
}

// CreateHTTP2Server creates an optimized HTTP/2 server
func CreateHTTP2Server(config *HTTP2Config, handler *gin.Engine) (*http.Server, error) {
	// Parse TLS configuration from environment
	tlsMinVersion := tls.VersionTLS12
	if getEnvOrDefault("TLS_MIN_VERSION", "1.2") == "1.3" {
		tlsMinVersion = tls.VersionTLS13
	}

	tlsMaxVersion := tls.VersionTLS13
	if getEnvOrDefault("TLS_MAX_VERSION", "1.3") == "1.2" {
		tlsMaxVersion = tls.VersionTLS12
	}

	// Configure TLS for HTTP/2 with environment-based settings
	tlsConfig := &tls.Config{
		MinVersion:               uint16(tlsMinVersion),
		MaxVersion:               uint16(tlsMaxVersion),
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		NextProtos: []string{"h2", "http/1.1"},
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:              ":" + config.Port,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
		TLSConfig:         tlsConfig,
	}

	// Configure HTTP/2 server parameters from environment
	http2Server := &http2.Server{
		MaxConcurrentStreams:         uint32(getEnvInt("HTTP2_MAX_CONCURRENT_STREAMS", 250)),
		MaxReadFrameSize:             uint32(getEnvInt("HTTP2_MAX_FRAME_SIZE", 16384)),
		PermitProhibitedCipherSuites: false,
		IdleTimeout:                  time.Duration(getEnvInt("HTTP2_IDLE_TIMEOUT_SECONDS", 300)) * time.Second,
		MaxUploadBufferPerConnection: int32(getEnvInt("HTTP2_WRITE_BUFFER_SIZE", 32768)),
		MaxUploadBufferPerStream:     int32(getEnvInt("HTTP2_READ_BUFFER_SIZE", 32768)),
	}

	// Determine if we should use HTTPS, H2C, or HTTP/1.1
	tlsEnabled := getEnvOrDefault("HTTP2_TLS_ENABLED", "false") == "true"

	if config.CertFile != "" && config.KeyFile != "" && tlsEnabled {
		// Production mode with TLS
		srv.Handler = handler

		// Configure HTTP/2 over TLS with custom settings
		if err := http2.ConfigureServer(srv, http2Server); err != nil {
			return nil, err
		}
	} else if config.H2CEnabled {
		// Development mode with H2C (HTTP/2 cleartext)
		srv.Handler = h2c.NewHandler(handler, http2Server)
		// Remove TLS config for cleartext
		srv.TLSConfig = nil
	} else {
		// Fallback to HTTP/1.1
		srv.Handler = handler
		srv.TLSConfig = nil
	}

	return srv, nil
}

// OptimizedTCPListener wraps net.TCPListener with HTTP/2 optimizations
type OptimizedTCPListener struct {
	*net.TCPListener
}

// Accept accepts connections with TCP optimizations for HTTP/2
func (l *OptimizedTCPListener) Accept() (net.Conn, error) {
	conn, err := l.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	// HTTP/2 specific TCP optimizations
	if err := conn.SetKeepAlive(true); err != nil {
		// Log warning but don't fail - not critical
	}
	if err := conn.SetKeepAlivePeriod(30 * time.Second); err != nil {
		// Log warning but don't fail - not critical
	}
	if err := conn.SetNoDelay(true); err != nil {
		// Log warning but don't fail - not critical
	}

	// Set buffer sizes optimized for HTTP/2
	if err := conn.SetReadBuffer(64 * 1024); err == nil {
		if err := conn.SetWriteBuffer(64 * 1024); err != nil {
			// Log warning but don't fail - not critical
		}
	}

	return conn, nil
}

// CreateOptimizedListener creates a TCP listener optimized for HTTP/2
func CreateOptimizedListener(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	if tcpListener, ok := listener.(*net.TCPListener); ok {
		return &OptimizedTCPListener{TCPListener: tcpListener}, nil
	}

	return listener, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer or returns default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
