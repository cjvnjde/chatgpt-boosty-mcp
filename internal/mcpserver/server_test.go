package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/boosty"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPTools(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blog/example/post/" || r.URL.Query().Get("limit") != "20" {
			t.Errorf("unexpected request %s", r.URL)
		}
		fmt.Fprint(w, `{"data":[],"extra":{"offset":"next"}}`)
	}))
	defer upstream.Close()
	c := boosty.New(filepath.Join(t.TempDir(), "token.json"))
	c.BaseURL = upstream.URL
	s := New(c, false)
	a, b := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := s.Connect(ctx, a, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	list, err := cs.ListTools(ctx, nil)
	if err != nil || len(list.Tools) != len(Endpoints()) {
		t.Fatalf("discovery: %v", err)
	}
	r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_posts", Arguments: map[string]any{"blog": "example"}})
	if err != nil || r.IsError {
		t.Fatalf("call: %v %v", r, err)
	}
	for _, args := range []map[string]any{{"blog": "../user"}, {"blog": "example", "limit": 101}, {"blog": "example", "surprise": "bad"}} {
		r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_posts", Arguments: args})
		if err == nil && !r.IsError {
			t.Fatalf("invalid input accepted: %v", args)
		}
	}
}
func TestHTTPBoundaries(t *testing.T) {
	h := Handler(New(boosty.New(filepath.Join(t.TempDir(), "token.json")), false), strings.Repeat("a", 32), []string{"localhost:8080"})
	for _, tc := range []struct {
		host, token, origin, path string
		want                      int
	}{
		{"evil", "", "", "/mcp", 421}, {"localhost:8080", "", "", "/mcp", 401}, {"localhost:8080", strings.Repeat("b", 32), "", "/mcp", 401}, {"localhost:8080", strings.Repeat("a", 32), "https://evil.test", "/mcp", 403}, {"localhost:8080", "", "", "/healthz", 200},
	} {
		req := httptest.NewRequest("POST", "http://"+tc.host+tc.path, strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer "+tc.token)
		req.Header.Set("Origin", tc.origin)
		if tc.path == "/healthz" {
			req.Method = "GET"
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s got %d want %d", tc.host, w.Code, tc.want)
		}
	}
}
