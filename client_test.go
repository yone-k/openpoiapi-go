package openpoiapi_test

import (
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

func TestNewClient(t *testing.T) {
	t.Run("オプションなしでクライアントを作れる", func(t *testing.T) {
		client, err := openpoiapi.NewClient()
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client == nil {
			t.Fatalf("NewClient() = nil, want client")
		}
	})

	t.Run("WithHTTPClientにnilを渡すとエラーを返す", func(t *testing.T) {
		client, err := openpoiapi.NewClient(openpoiapi.WithHTTPClient(nil))
		if err == nil {
			t.Fatalf("NewClient() error = nil, want error")
		}
		if client != nil {
			t.Errorf("NewClient() client = %v, want nil", client)
		}
	})

	invalidBaseURLs := []string{
		"",
		"ftp://example.com",
		"/relative",
		"http://",
		"http://example.com?x=1",
		"http://example.com#f",
		"://bad",
	}
	for _, raw := range invalidBaseURLs {
		t.Run("WithBaseURLに不正なURL「"+raw+"」を渡すとエラーを返す", func(t *testing.T) {
			client, err := openpoiapi.NewClient(openpoiapi.WithBaseURL(raw))
			if err == nil {
				t.Fatalf("NewClient() error = nil, want error")
			}
			if client != nil {
				t.Errorf("NewClient() client = %v, want nil", client)
			}
		})
	}

	validBaseURLs := []string{
		"http://127.0.0.1:8080/prefix/",
		"https://api.openpoiapi.com",
	}
	for _, raw := range validBaseURLs {
		t.Run("WithBaseURLに正しいURL「"+raw+"」を渡すとクライアントを作れる", func(t *testing.T) {
			client, err := openpoiapi.NewClient(openpoiapi.WithBaseURL(raw))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if client == nil {
				t.Fatalf("NewClient() = nil, want client")
			}
		})
	}

	t.Run("WithUserAgentに空文字を渡すとエラーを返す", func(t *testing.T) {
		client, err := openpoiapi.NewClient(openpoiapi.WithUserAgent(""))
		if err == nil {
			t.Fatalf("NewClient() error = nil, want error")
		}
		if client != nil {
			t.Errorf("NewClient() client = %v, want nil", client)
		}
	})
}
