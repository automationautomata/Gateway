package proxy_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/config"
	"gateway/server/proxy"
)

type MockLogger struct {
	debugCalls []string
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields map[string]any) {
	m.debugCalls = append(m.debugCalls, msg)
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields map[string]any) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields map[string]any) {
	m.warnCalls = append(m.warnCalls, msg)
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields map[string]any) {
	m.errorCalls = append(m.errorCalls, msg)
}

type MockProxyMetric struct {
	incCalls map[string]int
}

func NewMockProxyMetric() *MockProxyMetric {
	return &MockProxyMetric{
		incCalls: make(map[string]int),
	}
}

func (m *MockProxyMetric) Inc(dest string) {
	m.incCalls[dest]++
}

type MockLimiter struct {
	allowed bool
	err     error
}

func (m *MockLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return m.allowed, m.err
}

type MockLimiterMetric struct {
	incCalls []struct {
		allowed bool
		dest    string
	}
}

func (m *MockLimiterMetric) Inc(allowed bool, dest string) {
	m.incCalls = append(m.incCalls, struct {
		allowed bool
		dest    string
	}{allowed, dest})
}

func TestGetProxyWithoutHostRules(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default-response"))
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "any-host.com"
	req.URL.Scheme = "http"
	req.URL.Host = "any-host.com"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != "default-response" {
		t.Errorf("Expected default response, got %q", string(body))
	}

	if metricMock.incCalls[backend.URL] != 1 {
		t.Errorf("Expected default backend to be called once")
	}
}

func TestGetProxyWithHostMatch(t *testing.T) {
	backendDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default"))
	}))
	defer backendDefault.Close()

	backendHost := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("host-specific"))
	}))
	defer backendHost.Close()

	rules := config.ReverseProxyRules{
		Default: backendDefault.URL,
		Hosts: []config.HostRules{
			{
				Host:    "special.example.com",
				Default: &backendHost.URL,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "special.example.com"
	req.URL.Scheme = "http"
	req.URL.Host = "special.example.com"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	if string(body) != "host-specific" {
		t.Errorf("Expected 'host-specific', got %q", string(body))
	}
}

func TestGetProxyWithHostNotFound(t *testing.T) {
	backendDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default"))
	}))
	defer backendDefault.Close()

	backendSpecial := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("special"))
	}))
	defer backendSpecial.Close()

	rules := config.ReverseProxyRules{
		Default: backendDefault.URL,
		Hosts: []config.HostRules{
			{
				Host:    "special.example.com",
				Default: &backendSpecial.URL,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "unknown.example.com"
	req.URL.Scheme = "http"
	req.URL.Host = "unknown.example.com"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	if string(body) != "default" {
		t.Errorf("Expected 'default', got %q", string(body))
	}
}

func TestGetProxyWithPathRuleMatching(t *testing.T) {
	backendDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default"))
	}))
	defer backendDefault.Close()

	backendAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("api"))
	}))
	defer backendAPI.Close()

	rules := config.ReverseProxyRules{
		Default: backendDefault.URL,
		Hosts: []config.HostRules{
			{
				Host: "example.com",
				Pathes: map[string]string{
					"/": backendAPI.URL,
				},
				Default: &backendDefault.URL,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "example.com"
	req.URL.Scheme = "http"
	req.URL.Host = "example.com"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	if string(body) != "api" {
		t.Errorf("Expected 'api', got %q", string(body))
	}
}

func TestGetProxyWithHostDefaultFallback(t *testing.T) {
	backendGlobalDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("global-default"))
	}))
	defer backendGlobalDefault.Close()

	backendHostDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("host-default"))
	}))
	defer backendHostDefault.Close()

	rules := config.ReverseProxyRules{
		Default: backendGlobalDefault.URL,
		Hosts: []config.HostRules{
			{
				Host:    "example.com",
				Default: &backendHostDefault.URL,
				Pathes:  map[string]string{},
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "example.com"
	req.URL.Scheme = "http"
	req.URL.Host = "example.com"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	if string(body) != "host-default" {
		t.Errorf("Expected 'host-default', got %q", string(body))
	}
}

func TestGetProxyWithMultipleHosts(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("host1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("host2"))
	}))
	defer backend2.Close()

	backendDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default"))
	}))
	defer backendDefault.Close()

	rules := config.ReverseProxyRules{
		Default: backendDefault.URL,
		Hosts: []config.HostRules{
			{
				Host:    "host1.com",
				Default: &backend1.URL,
			},
			{
				Host:    "host2.com",
				Default: &backend2.URL,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	// Test host1
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Host = "host1.com"
	req1.URL.Scheme = "http"
	req1.URL.Host = "host1.com"
	w1 := httptest.NewRecorder()

	p.ServeHTTP(w1, req1)

	body1, _ := io.ReadAll(w1.Body)
	if string(body1) != "host1" {
		t.Errorf("Expected 'host1', got %q", string(body1))
	}

	// Test host2
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Host = "host2.com"
	req2.URL.Scheme = "http"
	req2.URL.Host = "host2.com"
	w2 := httptest.NewRecorder()

	p.ServeHTTP(w2, req2)

	body2, _ := io.ReadAll(w2.Body)
	if string(body2) != "host2" {
		t.Errorf("Expected 'host2', got %q", string(body2))
	}

	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Host = "unknown.com"
	req3.URL.Scheme = "http"
	req3.URL.Host = "unknown.com"
	w3 := httptest.NewRecorder()

	p.ServeHTTP(w3, req3)

	body3, _ := io.ReadAll(w3.Body)
	if string(body3) != "default" {
		t.Errorf("Expected 'default', got %q", string(body3))
	}
}

func TestNewHttpReverseProxyWithInvalidHostPath(t *testing.T) {
	invalidBackend := "ht!tp://invalid"
	rules := config.ReverseProxyRules{
		Default: "http://localhost:8080",
		Hosts: []config.HostRules{
			{
				Host: "example.com",
				Pathes: map[string]string{
					"/api": invalidBackend,
				},
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	_, err := proxy.NewHttpReverseProxy(input)
	if err == nil {
		t.Fatal("Expected error for invalid host path backend")
	}
}

func TestNewHttpReverseProxyWithInvalidHostDefault(t *testing.T) {
	invalidBackend := "ht!tp://invalid"
	rules := config.ReverseProxyRules{
		Default: "http://localhost:8080",
		Hosts: []config.HostRules{
			{
				Host:    "example.com",
				Default: &invalidBackend,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	_, err := proxy.NewHttpReverseProxy(input)
	if err == nil {
		t.Fatal("Expected error for invalid host default backend")
	}
}

func TestServeHTTPWithLimiterAllowed(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("allowed"))
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}
	loggerMock := &MockLogger{}
	proxyMetricMock := NewMockProxyMetric()
	limiterMock := &MockLimiter{allowed: true}
	limiterMetricMock := &MockLimiterMetric{}

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: proxyMetricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input,
		proxy.WithLimiter(limiterMock, limiterMetricMock),
	)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	if len(limiterMetricMock.incCalls) != 1 {
		t.Errorf("Expected limiter metric to be called once, got %d", len(limiterMetricMock.incCalls))
	}

	if limiterMetricMock.incCalls[0].allowed != true {
		t.Error("Expected allowed to be true")
	}
}

func TestServeHTTPWithLimiterDenied(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach"))
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}
	loggerMock := &MockLogger{}
	proxyMetricMock := NewMockProxyMetric()
	limiterMock := &MockLimiter{allowed: false}
	limiterMetricMock := &MockLimiterMetric{}

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: proxyMetricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input,
		proxy.WithLimiter(limiterMock, limiterMetricMock),
	)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	if len(limiterMetricMock.incCalls) != 1 {
		t.Errorf("Expected limiter metric to be called once, got %d", len(limiterMetricMock.incCalls))
	}

	if limiterMetricMock.incCalls[0].allowed != false {
		t.Error("Expected allowed to be false")
	}

	if proxyMetricMock.incCalls[backend.URL] != 1 {
		t.Errorf("Expected proxy metric to be called, got %d", proxyMetricMock.incCalls[backend.URL])
	}
}

func TestServeHTTPWithHostBasedRouting(t *testing.T) {
	backendDefault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("default"))
	}))
	defer backendDefault.Close()

	backendHost1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("host1"))
	}))
	defer backendHost1.Close()

	rules := config.ReverseProxyRules{
		Default: backendDefault.URL,
		Hosts: []config.HostRules{
			{
				Host:    "api.example.com",
				Default: &backendHost1.URL,
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Host = "api.example.com"
	req1.URL.Scheme = "http"
	req1.URL.Host = "api.example.com"
	w1 := httptest.NewRecorder()
	p.ServeHTTP(w1, req1)

	body1, _ := io.ReadAll(w1.Body)
	if string(body1) != "host1" {
		t.Errorf("Expected 'host1' for api.example.com, got %q", string(body1))
	}

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Host = "example.com"
	req2.URL.Scheme = "http"
	req2.URL.Host = "example.com"
	w2 := httptest.NewRecorder()
	p.ServeHTTP(w2, req2)

	body2, _ := io.ReadAll(w2.Body)
	if string(body2) != "default" {
		t.Errorf("Expected 'default' for example.com, got %q", string(body2))
	}
}

func TestRequestHeadersPreservation(t *testing.T) {
	var capturedHeaders http.Header
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, err := proxy.NewHttpReverseProxy(input)
	if err != nil {
		t.Fatalf("NewHttpReverseProxy failed: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	if capturedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Error("Custom header not preserved")
	}

	if capturedHeaders.Get("Authorization") != "Bearer token123" {
		t.Error("Authorization header not preserved")
	}
}

func BenchmarkProxy(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, _ := proxy.NewHttpReverseProxy(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "http://example.com/", nil)
		w := httptest.NewRecorder()
		p.ServeHTTP(w, req)
	}
}

func BenchmarkProxyWithLimiter(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
	}

	loggerMock := &MockLogger{}
	proxyMetricMock := NewMockProxyMetric()
	limiterMock := &MockLimiter{allowed: true}
	limiterMetricMock := &MockLimiterMetric{}

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: proxyMetricMock,
	}

	p, _ := proxy.NewHttpReverseProxy(input,
		proxy.WithLimiter(limiterMock, limiterMetricMock),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "http://example.com/", nil)
		w := httptest.NewRecorder()
		p.ServeHTTP(w, req)
	}
}

func BenchmarkProxyWithComplexRouting(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rules := config.ReverseProxyRules{
		Default: backend.URL,
		Hosts: []config.HostRules{
			{
				Host: "host1.example.com",
				Pathes: map[string]string{
					"/api":    backend.URL,
					"/v2":     backend.URL,
					"/users":  backend.URL,
					"/orders": backend.URL,
				},
			},
			{
				Host: "host2.example.com",
				Pathes: map[string]string{
					"/api": backend.URL,
				},
			},
		},
	}

	loggerMock := &MockLogger{}
	metricMock := NewMockProxyMetric()

	input := proxy.HttpProxyInput{
		Rules:       rules,
		Log:         loggerMock,
		ProxyMetric: metricMock,
	}

	p, _ := proxy.NewHttpReverseProxy(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		paths := []string{"/api/data", "/v2/info", "/users/list", "/"}
		pathIdx := i % len(paths)

		req := httptest.NewRequest("GET", paths[pathIdx], nil)
		req.Host = "host1.example.com"
		req.URL.Scheme = "http"
		req.URL.Host = "host1.example.com"
		w := httptest.NewRecorder()
		p.ServeHTTP(w, req)
	}
}
