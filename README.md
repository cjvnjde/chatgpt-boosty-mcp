# Boosty MCP (Go)

Personal, single-account MCP server for Boosty, implemented directly in Go with the official MCP Go SDK and `net/http`. Neither `boosty-api` nor the TypeScript `boosty-mcp` is a dependency.

Supports Streamable HTTP at `/mcp`, stdio, terminal cookie import, persistent rotating credentials, and separate authenticated direct/tunnel listeners. Linux is the supported native target. Docker/Dokploy configuration and the complete tool catalog follow in the next milestone.

```sh
go build -o bin/boosty-mcp ./cmd/boosty-mcp
./bin/boosty-mcp login
./bin/boosty-mcp status
MCP_TRANSPORT=stdio ./bin/boosty-mcp serve
```

HTTP requires `MCP_BEARER_TOKEN` (32+ characters), with a default listener at `127.0.0.1:8080`. Credentials default to `data/token.json`; override with `BOOSTY_TOKEN_FILE`. Log in on your own device, copy the Boosty Cookie header from browser DevTools, and paste it into the terminal prompt. Close that private browser window without logging out so it does not compete to refresh the same session. Credentials are never accepted as MCP tool arguments.

Boosty's internal API is undocumented. Endpoint contracts were inspected in the user-supplied [API reference implementation](https://github.com/beekamai/boosty-api) and [MCP feature reference](https://github.com/krasnovmv/boosty-mcp); they were not installed or executed. Public endpoints can be used without login; private content still requires the connected account's access rights.
