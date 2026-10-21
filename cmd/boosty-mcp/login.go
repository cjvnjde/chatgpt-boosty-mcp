package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/boosty"
	"github.com/cjvnjde/chatgpt-boosty-mcp/internal/config"
	"golang.org/x/sys/unix"
)

// Login is deliberately local-only: session cookies never pass through an MCP tool.
func loginCommand(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: boosty-mcp login (paste Cookie header on stdin)")
	}
	fmt.Fprintln(os.Stderr, "Log into boosty.to in a private browser window on your own device. Copy the Cookie request header from DevTools > Network, then close that window WITHOUT logging out.")
	fmt.Fprintln(os.Stderr, "Paste the Cookie header here (hidden on a terminal), then press Enter:")
	fd := int(os.Stdin.Fd())
	old, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err == nil {
		hidden := *old
		hidden.Lflag &^= unix.ECHO
		if unix.IoctlSetTermios(fd, unix.TCSETS, &hidden) != nil {
			return errors.New("cannot disable terminal echo; import through a private stdin pipe instead")
		}
		defer unix.IoctlSetTermios(fd, unix.TCSETS, old)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 128<<10)
	if !scanner.Scan() {
		return errors.New("no Cookie header supplied or input too long")
	}
	token, err := boosty.ParseCookies(scanner.Text())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := boosty.New(config.TokenFile()).Import(ctx, token); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "\nSession saved. The running server reloads it automatically. Run boosty-mcp status to verify authentication.")
	return nil
}
