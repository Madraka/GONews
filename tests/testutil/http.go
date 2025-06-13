package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"news/internal/json"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestServer wraps httptest.Server for API testing
type TestServer struct {
	*httptest.Server
	Client *http.Client
}

// NewTestServer creates a new test server
func NewTestServer(handler http.Handler) *TestServer {
	server := httptest.NewServer(handler)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &TestServer{
		Server: server,
		Client: client,
	}
}

// POST makes a POST request to the test server
func (ts *TestServer) POST(t *testing.T, path string, body interface{}, headers ...map[string]string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", ts.URL+path, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply additional headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	resp, err := ts.Client.Do(req)
	require.NoError(t, err)

	return resp
}

// GET makes a GET request to the test server
func (ts *TestServer) GET(t *testing.T, path string, headers ...map[string]string) *http.Response {
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	require.NoError(t, err)

	// Apply headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	resp, err := ts.Client.Do(req)
	require.NoError(t, err)

	return resp
}

// PUT makes a PUT request to the test server
func (ts *TestServer) PUT(t *testing.T, path string, body interface{}, headers ...map[string]string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("PUT", ts.URL+path, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply additional headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	resp, err := ts.Client.Do(req)
	require.NoError(t, err)

	return resp
}

// DELETE makes a DELETE request to the test server
func (ts *TestServer) DELETE(t *testing.T, path string, headers ...map[string]string) *http.Response {
	req, err := http.NewRequest("DELETE", ts.URL+path, nil)
	require.NoError(t, err)

	// Apply headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	resp, err := ts.Client.Do(req)
	require.NoError(t, err)

	return resp
}

// SetupGinTestMode sets Gin to test mode
func SetupGinTestMode() {
	gin.SetMode(gin.TestMode)
}

// AuthHeader creates authorization header
func AuthHeader(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// ParseJSONResponse parses JSON response into target struct
func ParseJSONResponse(t *testing.T, resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Skip JSON parsing if body is empty (for 404 responses)
	if len(body) == 0 {
		t.Logf("Empty response body for status %d", resp.StatusCode)
		return fmt.Errorf("empty response body")
	}

	err = json.Unmarshal(body, target)
	require.NoError(t, err, "Failed to parse JSON response: %s", string(body))
	return nil
}

// GetResponseBody reads and returns response body as string
func GetResponseBody(t *testing.T, resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Warning: Failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}
