package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/boosty"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connect(t *testing.T, c *boosty.Client, readonly bool) *mcp.ClientSession {
	t.Helper()
	a, b := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := New(c, readonly).Connect(ctx, a, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}
func TestRequestContracts(t *testing.T) {
	tests := []struct {
		name, method, path, query, form string
		args                            map[string]any
		jsonBody                        map[string]any
		anonymous                       bool
	}{
		{name: "list_comment_replies", method: "GET", path: "/v1/blog/demo/post/post1/comment/", query: "limit=20&parent_id=42", args: map[string]any{"blog": "demo", "post_id": "post1", "parent_id": 42}},
		{name: "list_media", method: "GET", path: "/v1/blog/demo/media_album/", query: "limit=20&limit_by=media&type=all", args: map[string]any{"blog": "demo"}},
		{name: "get_dialog_with_user", method: "GET", path: "/v1/dialog", query: "user_id=42", args: map[string]any{"user_id": "42"}},
		{name: "send_message", method: "POST", path: "/v1/dialog/42/message", form: `data=[{"type":"text","content":"hello"}]`, args: map[string]any{"dialog_id": "42", "data": []any{map[string]any{"type": "text", "content": "hello"}}}},
		{name: "update_post", method: "PUT", path: "/v1/blog/demo/post/post1", form: "deny_comments=false&price=0&title=", args: map[string]any{"blog": "demo", "post_id": "post1", "price": 0, "deny_comments": false, "title": ""}},
		{name: "update_post_deferred_access", method: "PUT", path: "/v1/blog/demo/post/post1/deferred_access", args: map[string]any{"blog": "demo", "post_id": "post1", "price": 0, "isCommentsDenied": false, "applyTime": 2000000000}, jsonBody: map[string]any{"price": float64(0), "isCommentsDenied": false, "applyTime": float64(2000000000)}},
		{name: "like_comment", method: "POST", path: "/v1/blog/demo/post/post1/comment/42/reaction", form: "reaction=like", args: map[string]any{"blog": "demo", "post_id": "post1", "comment_id": "42"}},
		{name: "like_post", method: "POST", path: "/v1/blog/demo/post/post1/like", args: map[string]any{"blog": "demo", "post_id": "post1"}},
		{name: "vote_poll", method: "POST", path: "/v1/poll/42/vote", form: "answer=1%2C2&is_anonymous=false", args: map[string]any{"poll_id": "42", "answer": "1,2", "is_anonymous": false}},
		{name: "mark_notifications_read", method: "PUT", path: "/v1/notification/standalone/read/", form: "event_id=1%2C2", args: map[string]any{"event_id": "1,2"}},
		{name: "update_target", method: "PUT", path: "/v1/target/42", form: "target_id=42&description=hello&target_sum=5", args: map[string]any{"target_id": "42", "description": "hello", "target_sum": 5}},
		{name: "list_visit_sources", method: "GET", path: "/v1/blog/stat/demo/visits", query: "limit=20&order=gt&period=month&type=utm", args: map[string]any{"blog": "demo", "period": "month", "type": "utm"}},
		{name: "get_paid_sales_by_date", method: "GET", path: "/v1/blog/demo/bundle/paid/by_date", query: "from=1&to=2", args: map[string]any{"blog": "demo", "sale_kind": "bundle", "from": 1, "to": 2}},
		{name: "list_payment_cards", method: "GET", path: "/v1/payment/bind/list", query: "payment_partner=advpay", args: map[string]any{}},
		{name: "get_default_currency", method: "GET", path: "/v1/payment/default_currency", args: map[string]any{}, anonymous: true},
		{name: "build_report", method: "GET", path: "/v1/blog/stat/demo/csv/donations", query: "from=1&to=2", args: map[string]any{"blog": "demo", "report_type": "donations", "from": 1, "to": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("got %s %s", r.Method, r.URL.Path)
				}
				wantQ, _ := url.ParseQuery(tt.query)
				if !reflect.DeepEqual(r.URL.Query(), wantQ) {
					t.Errorf("query %v want %v", r.URL.Query(), wantQ)
				}
				if tt.jsonBody != nil {
					var got map[string]any
					json.NewDecoder(r.Body).Decode(&got)
					if !reflect.DeepEqual(got, tt.jsonBody) || r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("JSON %v", got)
					}
				} else {
					r.ParseForm()
					wantF, _ := url.ParseQuery(tt.form)
					// JSON object key order is immaterial within a form field.
					if tt.name == "send_message" {
						var got, want any
						json.Unmarshal([]byte(r.PostForm.Get("data")), &got)
						json.Unmarshal([]byte(wantF.Get("data")), &want)
						if !reflect.DeepEqual(got, want) {
							t.Errorf("content blocks %v", got)
						}
					} else if !reflect.DeepEqual(r.PostForm, wantF) {
						t.Errorf("form %v want %v", r.PostForm, wantF)
					}
				}
				if (r.Header.Get("Authorization") == "") != tt.anonymous {
					t.Error("incorrect auth")
				}
				fmt.Fprint(w, `{"data":[],"extra":{"offset":"opaque-next"},"unknownFutureField":true}`)
			}))
			defer upstream.Close()
			c := boosty.New(filepath.Join(t.TempDir(), "token.json"))
			c.BaseURL = upstream.URL
			if err := c.Import(context.Background(), boosty.Token{AccessToken: "access", RefreshToken: "refresh", DeviceID: "device"}); err != nil {
				t.Fatal(err)
			}
			cs := connect(t, c, false)
			result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: tt.name, Arguments: tt.args})
			if err != nil || result.IsError || !called {
				t.Fatalf("call failed: %v %v", result, err)
			}
			b, _ := json.Marshal(result.StructuredContent)
			if !strings.Contains(string(b), "opaque-next") || !strings.Contains(string(b), "unknownFutureField") {
				t.Fatal("response data lost")
			}
		})
	}
}
func TestValidationAndReadOnly(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); fmt.Fprint(w, `{}`) }))
	defer upstream.Close()
	c := boosty.New(filepath.Join(t.TempDir(), "token.json"))
	c.BaseURL = upstream.URL
	c.Import(context.Background(), boosty.Token{AccessToken: "a", RefreshToken: "r", DeviceID: "d"})
	cs := connect(t, c, false)
	for _, tt := range []struct {
		name string
		args map[string]any
	}{
		{"update_post", map[string]any{"blog": "demo", "post_id": "post1"}},
		{"update_target", map[string]any{"target_id": "42", "description": "hello"}},
		{"update_dialog_settings", map[string]any{"can_send_all": false}},
		{"mark_notifications_read", map[string]any{"event_id": ""}},
		{"vote_poll", map[string]any{"poll_id": "42", "answer": ""}},
		{"end_sessions", map[string]any{"including_current": false}},
		{"get_stats_metrics", map[string]any{"blog": "demo", "from": 20, "to": 10}},
		{"get_paid_sales_days", map[string]any{"blog": "demo", "sale_kind": "post", "date": 2000000000000}},
		{"list_media", map[string]any{"blog": "demo", "type": "ok_video"}},
	} {
		result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: tt.name, Arguments: tt.args})
		if err == nil && !result.IsError {
			t.Errorf("accepted invalid %s", tt.name)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("invalid input reached upstream")
	}
	ro := connect(t, c, true)
	list, err := ro.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 54 {
		t.Fatalf("read-only tools: %d", len(list.Tools))
	}
	for _, tool := range list.Tools {
		if !tool.Annotations.ReadOnlyHint {
			t.Fatal("write exposed in read-only mode")
		}
	}
	result, err := ro.CallTool(context.Background(), &mcp.CallToolParams{Name: "build_report", Arguments: map[string]any{"blog": "demo", "report_type": "donations"}})
	if err == nil && !result.IsError {
		t.Fatal("report generation exposed in read-only mode")
	}
}

type authTransport struct{ token string }

func (a authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	clone.Header.Set("Authorization", "Bearer "+a.token)
	return http.DefaultTransport.RoundTrip(clone)
}
func TestStreamableHTTPAndDistinctTokens(t *testing.T) {
	c := boosty.New(filepath.Join(t.TempDir(), "token.json"))
	s := New(c, true)
	direct, tunnel := strings.Repeat("d", 32), strings.Repeat("t", 32)
	for _, key := range []string{direct, tunnel} {
		httpServer := httptest.NewUnstartedServer(nil)
		host := httpServer.Listener.Addr().String()
		httpServer.Config.Handler = Handler(s, key, []string{host})
		httpServer.Start()
		defer httpServer.Close()
		client := mcp.NewClient(&mcp.Implementation{Name: "http-test", Version: "1"}, nil)
		cs, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: httpServer.URL + "/mcp", HTTPClient: &http.Client{Transport: authTransport{key}}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		list, err := cs.ListTools(context.Background(), nil)
		if err != nil || len(list.Tools) != 54 {
			t.Fatalf("HTTP discovery: %v", err)
		}
		cs.Close()
		other := direct
		if key == direct {
			other = tunnel
		}
		req, _ := http.NewRequest("POST", httpServer.URL+"/mcp", strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer "+other)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Fatal("other listener token accepted")
		}
	}
}
func TestHTTPBodyLimit(t *testing.T) {
	token := strings.Repeat("t", 32)
	h := Handler(New(boosty.New(filepath.Join(t.TempDir(), "token.json")), true), token, []string{"localhost"})
	req := httptest.NewRequest("POST", "http://localhost/mcp", strings.NewReader(strings.Repeat("a", (2<<20)+1)))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 413 {
		t.Fatalf("oversize request status %d", w.Code)
	}
}
