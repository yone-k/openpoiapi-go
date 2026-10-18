package openpoiapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Version is the version of this SDK. It is sent in the default User-Agent.
const Version = "0.1.0"

// DefaultBaseURL is the base URL of the OpenPOI API.
const DefaultBaseURL = "https://api.openpoiapi.com"

// Client calls the OpenPOI API. Create one with NewClient. A Client is safe
// for concurrent use.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client in NewClient.
type Option func(*Client) error

// NewClient returns a Client configured by opts. Without options it calls
// DefaultBaseURL with an HTTP client that times out after 30 seconds and the
// User-Agent "openpoiapi-go/<Version>". Options are applied in order and the
// first error is returned.
func NewClient(opts ...Option) (*Client, error) {
	baseURL, err := parseBaseURL(DefaultBaseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "openpoiapi-go/" + Version,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithHTTPClient makes the Client send requests with hc, for example to
// change the timeout or the transport. hc must not be nil.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		if hc == nil {
			return errors.New("openpoiapi: http client is nil")
		}
		c.httpClient = hc
		return nil
	}
}

// WithBaseURL makes the Client call rawURL instead of DefaultBaseURL.
// rawURL must be an absolute http or https URL without a query or fragment.
// Endpoint paths such as /v1/search are appended to its path.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) error {
		baseURL, err := parseBaseURL(rawURL)
		if err != nil {
			return err
		}
		c.baseURL = baseURL
		return nil
	}
}

// WithUserAgent replaces the default User-Agent header with ua. ua must not
// be empty.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		if ua == "" {
			return errors.New("openpoiapi: user agent is empty")
		}
		c.userAgent = ua
		return nil
	}
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("openpoiapi: invalid base URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("openpoiapi: invalid base URL %q: scheme must be http or https", rawURL)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("openpoiapi: invalid base URL %q: host is empty", rawURL)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("openpoiapi: invalid base URL %q: query and fragment are not allowed", rawURL)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawPath = ""
	return u, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := *c.baseURL
	u.Path = c.baseURL.Path + path
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("openpoiapi: create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openpoiapi: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return newAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("openpoiapi: decode response: %w", err)
	}
	return nil
}

func newAPIError(resp *http.Response) *APIError {
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	// The decode error is ignored rather than returned: a non-JSON error body
	// (for example from a proxy) should still surface as an APIError with the
	// status code.
	_ = json.NewDecoder(resp.Body).Decode(&body)
	message := body.Error
	if message == "" {
		message = body.Message
	}
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}
	return &APIError{StatusCode: resp.StatusCode, Message: message}
}

func setString(query url.Values, key, value string) {
	if value == "" {
		return
	}
	query.Set(key, value)
}

func setInt(query url.Values, key string, value int) {
	if value == 0 {
		return
	}
	query.Set(key, strconv.Itoa(value))
}
