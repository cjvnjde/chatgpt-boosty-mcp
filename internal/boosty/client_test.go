package boosty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCookieImport(t *testing.T) {
	for _, raw := range []string{`{"accessToken":"access","refreshToken":"refresh","expiresAt":2000000000000}`, url.PathEscape(`{"accessToken":"access","refreshToken":"refresh","expiresAt":"2000000000000"}`)} {
		token, err := ParseCookies("Cookie: unrelated=secret; auth=" + raw + "; _clientId=device")
		if err != nil || token.ExpiresAt != 2000000000 || token.DeviceID != "device" {
			t.Fatalf("cookie parsing failed: %v", err)
		}
		c := New(filepath.Join(t.TempDir(), "token.json"))
		if err = c.Import(context.Background(), token); err != nil {
			t.Fatal(err)
		}
		stat, _ := os.Stat(c.TokenFile)
		if stat.Mode().Perm() != 0600 {
			t.Fatal("credentials not private")
		}
	}
	for _, raw := range []string{"auth=secret", `auth={"accessToken":"bad\nsecret","refreshToken":"refresh"}; _clientId=device`} {
		if _, err := ParseCookies(raw); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("bad cookie accepted or leaked")
		}
	}
}

// Browsers can escape quotes and commas while leaving the JSON braces literal.
func TestPartiallyEncodedCookie(t *testing.T) {
	header := `_clientId=test-device; auth={%22accessToken%22:%22test-access%22%2C%22refreshToken%22:%22test-refresh%22%2C%22expiresAt%22:2000000000000%2C%22isNewUser%22:false}; last_acc={%22name%22:%22Example%22}`
	token, err := ParseCookies(header)
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "test-access" || token.RefreshToken != "test-refresh" || token.DeviceID != "test-device" || token.ExpiresAt != 2000000000 {
		t.Fatal("incorrect parsed credentials")
	}
}

func TestDecodedCookiePreservesPercentAndPlus(t *testing.T) {
	token, err := ParseCookies(`auth={"accessToken":"test%22+access","refreshToken":"test%25+refresh"}; _clientId=test-device`)
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "test%22+access" || token.RefreshToken != "test%25+refresh" {
		t.Fatal("decoded token was altered")
	}
}

func TestMalformedCookieRejected(t *testing.T) {
	for _, raw := range []string{`{%ZZ}`, `{"accessToken":"test","refreshToken":"test"}trailing`, `null`, `[]`} {
		if _, err := ParseCookies("auth=" + raw + "; _clientId=test-device"); err == nil {
			t.Fatal("malformed cookie accepted")
		}
	}
}

func TestConcurrentRefreshAndReload(t *testing.T) {
	var refreshes atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/" {
			refreshes.Add(1)
			r.ParseForm()
			if r.Form.Get("device_id") != "device" || r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "refresh" {
				t.Error("invalid refresh form")
			}
			fmt.Fprint(w, `{"access_token":"new","refresh_token":"rotated","expires_in":3600}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer new" {
			w.WriteHeader(401)
			return
		}
		fmt.Fprint(w, `{"data":[],"extra":{"offset":"next"}}`)
	}))
	defer upstream.Close()
	file := filepath.Join(t.TempDir(), "token.json")
	first := New(file)
	first.BaseURL = upstream.URL
	if err := first.Import(context.Background(), Token{"old", "refresh", "device", time.Now().Add(-time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 15 {
		wg.Go(func() {
			c := first
			if i%2 == 0 {
				c = New(file)
				c.BaseURL = upstream.URL
			}
			if _, err := c.Do(context.Background(), Request{Method: "GET", Path: "/v1/user/current", RequireAuth: true}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if refreshes.Load() != 1 {
		t.Fatalf("refreshes = %d", refreshes.Load())
	}
	token, err := first.loadToken()
	if err != nil || token.RefreshToken != "rotated" {
		t.Fatal("rotation not persisted")
	}
}
func Test401RefreshAndNoUnsafeRetry(t *testing.T) {
	var attempts, refreshes int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/" {
			refreshes++
			fmt.Fprint(w, `{"access_token":"new","refresh_token":"rotated","expires_in":3600}`)
			return
		}
		attempts++
		if r.Header.Get("Authorization") == "Bearer old" {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(429)
		fmt.Fprint(w, "secret")
	}))
	defer upstream.Close()
	c := New(filepath.Join(t.TempDir(), "token.json"))
	c.BaseURL = upstream.URL
	c.Import(context.Background(), Token{"old", "refresh", "device", 0})
	_, err := c.Do(context.Background(), Request{Method: "POST", Path: "/v1/dialog/", RequireAuth: true})
	if err == nil || attempts != 2 || refreshes != 1 || strings.Contains(err.Error(), "secret") {
		t.Fatalf("bad retry/error: %v attempts=%d", err, attempts)
	}
}
func TestAnonymousAndRedirect(t *testing.T) {
	var leaked atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Store(true) }))
	defer target.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/anon" {
			if r.Header.Get("Authorization") != "" {
				t.Error("anonymous request authenticated")
			}
			json.NewEncoder(w).Encode(map[string]any{"ok": true})
			return
		}
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer upstream.Close()
	c := New(filepath.Join(t.TempDir(), "token.json"))
	c.BaseURL = upstream.URL
	c.Import(context.Background(), Token{"access", "refresh", "device", 0})
	if _, err := c.Do(context.Background(), Request{Method: "GET", Path: "/v1/anon", Anonymous: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Do(context.Background(), Request{Method: "GET", Path: "/v1/redirect"}); err == nil || leaked.Load() {
		t.Fatal("redirect followed")
	}
}
