package config

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	for _, key := range []string{"MCP_TRANSPORT", "MCP_HTTP_ADDR", "MCP_TUNNEL_ADDR", "MCP_BEARER_TOKEN", "MCP_TUNNEL_TOKEN", "MCP_ALLOWED_HOSTS", "BOOSTY_READ_ONLY", "BOOSTY_TOKEN_FILE"} {
		t.Setenv(key, "")
	}
	if _, err := Load(); err == nil {
		t.Fatal("HTTP without token accepted")
	}
	t.Setenv("MCP_TRANSPORT", "stdio")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MCP_TRANSPORT", "http")
	t.Setenv("MCP_BEARER_TOKEN", strings.Repeat("a", 32))
	t.Setenv("MCP_TUNNEL_ADDR", "127.0.0.1:8081")
	for _, token := range []string{"", strings.Repeat("a", 32)} {
		t.Setenv("MCP_TUNNEL_TOKEN", token)
		if _, err := Load(); err == nil {
			t.Fatal("invalid tunnel auth accepted")
		}
	}
	t.Setenv("MCP_TUNNEL_TOKEN", strings.Repeat("b", 32))
	t.Setenv("BOOSTY_READ_ONLY", "true")
	c, err := Load()
	if err != nil || !c.ReadOnly {
		t.Fatal("valid config rejected")
	}
	t.Setenv("BOOSTY_READ_ONLY", "typo")
	if _, err := Load(); err == nil {
		t.Fatal("invalid boolean accepted")
	}
}
