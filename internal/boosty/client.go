// Package boosty implements the Boosty HTTP contract directly using net/http.
package boosty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const APIURL = "https://api.boosty.to"
const maxResponse = 16 << 20

type Client struct {
	mu        sync.Mutex // Go memory synchronization complements the cross-process file lock.
	BaseURL   string
	HTTP      *http.Client
	TokenFile string
	// pending is accessed only while holding the credential file lock. Retaining
	// it allows persistence to be retried without spending a rotated token twice.
	pending *Token
}

func New(tokenFile string) *Client {
	return &Client{BaseURL: APIURL, TokenFile: tokenFile, HTTP: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

type Request struct {
	Method, Path           string
	Query, Form            url.Values
	JSON                   any
	RequireAuth, Anonymous bool
}

type APIError struct{ Status int }

func (e *APIError) Error() string {
	switch e.Status {
	case 401:
		return "Boosty authentication expired or rejected; import a fresh session with login"
	case 403:
		return "Boosty denied access; this operation requires the appropriate subscription or blog ownership"
	case 429:
		return "Boosty rate limit reached; wait before trying again"
	default:
		return fmt.Sprintf("Boosty returned HTTP %d", e.Status)
	}
}

func (c *Client) Do(ctx context.Context, r Request) (any, error) {
	if !strings.HasPrefix(r.Path, "/v1/") || strings.ContainsAny(r.Path, "?#") {
		return nil, errors.New("invalid API path")
	}
	var token Token
	var err error
	if !r.Anonymous {
		token, err = c.credentials(ctx, "")
		if err != nil {
			return nil, err
		}
	}
	if r.RequireAuth && token.AccessToken == "" {
		return nil, errors.New("Boosty login required; run boosty-mcp login in the server terminal")
	}
	result, err := c.request(ctx, r, token.AccessToken)
	var apiErr *APIError
	if !r.Anonymous && errors.As(err, &apiErr) && apiErr.Status == 401 && token.RefreshToken != "" {
		token, err = c.credentials(ctx, token.AccessToken)
		if err != nil {
			return nil, err
		}
		return c.request(ctx, r, token.AccessToken)
	}
	return result, err
}

func (c *Client) request(ctx context.Context, r Request, access string) (any, error) {
	endpoint := c.BaseURL + r.Path
	if len(r.Query) > 0 {
		endpoint += "?" + r.Query.Encode()
	}
	var body io.Reader
	if r.Form != nil {
		body = strings.NewReader(r.Form.Encode())
	}
	if r.JSON != nil {
		b, err := json.Marshal(r.JSON)
		if err != nil {
			return nil, errors.New("invalid JSON request")
		}
		body = strings.NewReader(string(b))
	}
	req, err := http.NewRequestWithContext(ctx, r.Method, endpoint, body)
	if err != nil {
		return nil, errors.New("cannot construct Boosty request")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "boosty-mcp/1.0")
	req.Header.Set("Origin", "https://boosty.to")
	req.Header.Set("Referer", "https://boosty.to/")
	if access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	if r.Form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if r.JSON != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("Boosty network request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{resp.StatusCode}
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil {
		return nil, errors.New("cannot read Boosty response")
	}
	if len(b) > maxResponse {
		return nil, errors.New("Boosty response too large; use a smaller page")
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return map[string]any{"success": true}, nil
	}
	var result any
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.UseNumber()
	if dec.Decode(&result) != nil || !json.Valid(b) {
		return nil, errors.New("Boosty returned invalid JSON")
	}
	return result, nil
}
