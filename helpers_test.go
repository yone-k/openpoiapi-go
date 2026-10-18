package openpoiapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

func ptr[T any](v T) *T {
	return &v
}

func dump(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
}

type stubServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
}

func newStubServer(t *testing.T, status int, body string) *stubServer {
	t.Helper()
	s := &stubServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Header: r.Header.Clone(),
		})
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *stubServer) Requests() []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]recordedRequest(nil), s.requests...)
}

func (s *stubServer) LastRequest(t *testing.T) recordedRequest {
	t.Helper()
	requests := s.Requests()
	if len(requests) == 0 {
		t.Fatalf("server received no request")
	}
	return requests[len(requests)-1]
}

func newTestClient(t *testing.T, baseURL string, opts ...openpoiapi.Option) *openpoiapi.Client {
	t.Helper()
	client, err := openpoiapi.NewClient(append([]openpoiapi.Option{openpoiapi.WithBaseURL(baseURL)}, opts...)...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
