package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Transport, Addr, TunnelAddr, Bearer, TunnelToken, TokenFile string
	Hosts                                                       []string
	ReadOnly                                                    bool
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func TokenFile() string { return env("BOOSTY_TOKEN_FILE", "data/token.json") }
func Load() (Config, error) {
	c := Config{Transport: env("MCP_TRANSPORT", "http"), Addr: env("MCP_HTTP_ADDR", "127.0.0.1:8080"), TunnelAddr: os.Getenv("MCP_TUNNEL_ADDR"), Bearer: os.Getenv("MCP_BEARER_TOKEN"), TunnelToken: os.Getenv("MCP_TUNNEL_TOKEN"), TokenFile: TokenFile()}
	for _, h := range strings.Split(env("MCP_ALLOWED_HOSTS", "localhost:8080,127.0.0.1:8080,localhost:8081,127.0.0.1:8081"), ",") {
		if h = strings.TrimSpace(h); h != "" {
			c.Hosts = append(c.Hosts, h)
		}
	}
	switch env("BOOSTY_READ_ONLY", "false") {
	case "true":
		c.ReadOnly = true
	case "false":
	default:
		return c, errors.New("BOOSTY_READ_ONLY must be true or false")
	}
	if c.Transport != "http" && c.Transport != "stdio" {
		return c, errors.New("MCP_TRANSPORT must be http or stdio")
	}
	if c.Transport == "http" {
		if len(c.Bearer) < 32 {
			return c, errors.New("MCP_BEARER_TOKEN must be at least 32 characters")
		}
		if c.TunnelAddr != "" && (len(c.TunnelToken) < 32 || c.TunnelToken == c.Bearer || c.TunnelAddr == c.Addr) {
			return c, errors.New("tunnel requires a distinct listener and a distinct MCP_TUNNEL_TOKEN of at least 32 characters")
		}
	}
	return c, nil
}
