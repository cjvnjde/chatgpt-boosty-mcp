package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/boosty"
	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/config"
	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
func run() error {
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command == "health" {
		client := http.Client{Timeout: 2 * time.Second}
		addr := os.Getenv("MCP_HTTP_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8080"
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return errors.New("invalid MCP_HTTP_ADDR")
		}
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		resp, err := client.Get("http://" + net.JoinHostPort(host, port) + "/healthz")
		if err != nil {
			return errors.New("health check failed")
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New("unhealthy")
		}
		return nil
	}
	if command == "login" {
		return loginCommand(os.Args[2:])
	}
	if command == "status" {
		return status()
	}
	if command != "serve" {
		return errors.New("usage: boosty-mcp [serve|login|status|health]")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	c := boosty.New(cfg.TokenFile)
	server := mcpserver.New(c, cfg.ReadOnly)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.Transport == "stdio" {
		return server.Run(ctx, &mcp.StdioTransport{})
	}
	var servers []*http.Server
	var listeners []net.Listener
	// Bind every listener before serving; a port conflict must fail startup.
	for _, endpoint := range []struct{ addr, token string }{{cfg.Addr, cfg.Bearer}, {cfg.TunnelAddr, cfg.TunnelToken}} {
		if endpoint.addr == "" {
			continue
		}
		ln, err := net.Listen("tcp", endpoint.addr)
		if err != nil {
			for _, l := range listeners {
				l.Close()
			}
			return err
		}
		listeners = append(listeners, ln)
		servers = append(servers, &http.Server{Handler: mcpserver.Handler(server, endpoint.token, cfg.Hosts), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 90 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10})
	}
	errs := make(chan error, len(servers))
	for i, s := range servers {
		go func() { errs <- s.Serve(listeners[i]) }()
		slog.Info("MCP listening", "address", listeners[i].Addr().String())
	}
	select {
	case <-ctx.Done():
	case err = <-errs:
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, s := range servers {
		if e := s.Shutdown(shutdown); e != nil {
			s.Close()
		}
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func status() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := boosty.New(config.TokenFile()).Do(ctx, boosty.Request{Method: "GET", Path: "/v1/user/current", RequireAuth: true})
	if err == nil {
		slog.Info("Boosty session is authenticated")
	}
	return err
}
