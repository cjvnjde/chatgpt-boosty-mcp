package boosty

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

func (t Token) valid() bool {
	return tokenText(t.AccessToken) && tokenText(t.RefreshToken) && tokenText(t.DeviceID)
}
func tokenText(s string) bool {
	if s == "" || len(s) > 16384 {
		return false
	}
	for _, b := range []byte(s) {
		if b < 33 || b > 126 {
			return false
		}
	}
	return true
}
func number(v any) int64 {
	s := ""
	switch n := v.(type) {
	case json.Number:
		s = string(n)
	case string:
		s = n
	}
	f, _ := strconv.ParseFloat(s, 64)
	if f > 1e12 {
		f /= 1000
	}
	if f < 0 || f > 1e11 {
		return 0
	}
	return int64(f)
}

// ParseCookies accepts only the session cookies; all other cookies are discarded.
func ParseCookies(header string) (Token, error) {
	header = strings.TrimSpace(header)
	if len(header) >= 7 && strings.EqualFold(header[:7], "cookie:") {
		header = header[7:]
	}
	cookies := map[string]string{}
	for _, part := range strings.Split(header, ";") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok {
			cookies[k] = v
		}
	}
	raw := cookies["auth"]
	if !strings.HasPrefix(raw, "{") {
		decoded, err := url.PathUnescape(raw)
		if err != nil {
			return Token{}, errors.New("invalid auth cookie encoding")
		}
		raw = decoded
	}
	var session map[string]any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if dec.Decode(&session) != nil {
		return Token{}, errors.New("auth cookie missing or invalid")
	}
	a, _ := session["accessToken"].(string)
	r, _ := session["refreshToken"].(string)
	t := Token{AccessToken: a, RefreshToken: r, DeviceID: cookies["_clientId"], ExpiresAt: number(session["expiresAt"])}
	if !t.valid() {
		return Token{}, errors.New("auth cookie must contain valid accessToken and refreshToken, and _clientId is required")
	}
	return t, nil
}
func (c *Client) Import(ctx context.Context, t Token) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !t.valid() {
		return errors.New("invalid Boosty credentials")
	}
	unlock, err := c.lockStore(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	if writeJSON(c.TokenFile, t) != nil {
		return errors.New("cannot save Boosty credentials; check volume permissions")
	}
	c.pending = nil
	return nil
}
func (c *Client) loadToken() (Token, error) {
	b, err := os.ReadFile(c.TokenFile)
	if os.IsNotExist(err) {
		return Token{}, nil
	}
	if err != nil {
		return Token{}, errors.New("cannot read Boosty credential file")
	}
	var t Token
	if json.Unmarshal(b, &t) != nil || !t.valid() {
		return Token{}, errors.New("invalid Boosty credential file; run login again")
	}
	return t, nil
}
func (c *Client) credentials(ctx context.Context, rejected string) (Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	unlock, err := c.lockStore(ctx)
	if err != nil {
		return Token{}, err
	}
	defer unlock()
	if c.pending != nil {
		if writeJSON(c.TokenFile, *c.pending) != nil {
			return Token{}, errors.New("cannot persist rotated Boosty credentials; fix volume permissions before restarting")
		}
		c.pending = nil
	}
	t, err := c.loadToken()
	if err != nil {
		return Token{}, err
	}
	if t.AccessToken == "" {
		return t, nil
	}
	expired := t.ExpiresAt != 0 && t.ExpiresAt <= time.Now().Unix()+60
	if !expired && (rejected == "" || rejected != t.AccessToken) {
		return t, nil
	}
	result, err := c.request(ctx, Request{Method: "POST", Path: "/oauth/token/", Form: url.Values{"device_id": {t.DeviceID}, "device_os": {"web"}, "grant_type": {"refresh_token"}, "refresh_token": {t.RefreshToken}}}, t.AccessToken)
	if err != nil {
		return Token{}, err
	}
	obj, ok := result.(map[string]any)
	if !ok {
		return Token{}, errors.New("invalid Boosty refresh response")
	}
	a, _ := obj["access_token"].(string)
	r, _ := obj["refresh_token"].(string)
	next := Token{AccessToken: a, RefreshToken: r, DeviceID: t.DeviceID}
	if id, ok := obj["device_id"].(string); ok && id != "" {
		next.DeviceID = id
	}
	if !next.valid() {
		return Token{}, errors.New("Boosty refresh response missing valid credentials")
	}
	if seconds := number(obj["expires_in"]); seconds > 0 {
		next.ExpiresAt = time.Now().Unix() + seconds
	}
	c.pending = &next
	if writeJSON(c.TokenFile, next) != nil {
		return Token{}, errors.New("cannot persist rotated Boosty credentials; fix volume permissions before restarting")
	}
	c.pending = nil
	return next, nil
}
