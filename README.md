# Boosty MCP (Go)

A personal, single-account MCP server for [Boosty](https://boosty.to), implemented in Go using `net/http` and the official MCP Go SDK. **Neither `boosty-api` nor the TypeScript `boosty-mcp` is installed or used as a dependency.** The structure follows the neighboring MyAnimeList and Proton Calendar MCP projects: Streamable HTTP, stdio, persistent credentials, a non-root container, separate authenticated direct/tunnel listeners, and an optional OpenAI tunnel sidecar.

## Tools

83 tools cover blogs, subscription levels, posts, search, media, comments, reactions, polls, account settings, messages, notifications, targets, income, and creator analytics. The complete names, HTTP routes, required fields, and verification status are in [API coverage](docs/api-coverage.md).

Set `BOOSTY_READ_ONLY=true` to expose only the 54 read tools. All 29 operations with side effects are removed, including report generation (which Boosty implements using GET). Writes have MCP mutation annotations. Many use undocumented, experimental contracts; their tool descriptions identify this explicitly.

Responses are returned under `response`, preserving original fields, content blocks, access flags and pagination metadata. Lists return **one page**, generally 20 entries by default, with a server-imposed maximum of 100. Pass the previous `response.extra.offset` or `response.offset` as the string `offset`; endpoints differ and may supply other pagination fields. Timestamps are Unix **seconds**, not milliseconds. Identifiers used in paths are strings. `blog` and `blog_url` mean a blog slug such as `boosty`, not a URL. Filters use Boosty's snake_case names. Comma-separated ID filters are strings.

Post/comment/message content uses Boosty's native `data` blocks (including Draft.js JSON inside text blocks); no media upload or binary download is implemented. Omitted update fields stay omitted: updating a title does not clear post content, and explicit zero/false/empty values are preserved. `update_dialog_settings` requires all six permission switches; `update_target` requires description and sum.

## Configure and log in

1. Copy `.env.example` to `.env`.
2. Generate **two different** tokens with `openssl rand -hex 32`: `MCP_BEARER_TOKEN` for direct MCP clients, and `MCP_TUNNEL_TOKEN` for the tunnel listener.
3. For a tunnel, fill in `CONTROL_PLANE_API_KEY` and `CONTROL_PLANE_TUNNEL_ID` from your OpenAI tunnel configuration.

The MCP tokens authenticate callers of this server. They are unrelated to your Boosty session or the OpenAI control-plane key. `.env` is read by Docker Compose; the Go binary does not load it automatically.

Public blog/post reads work without login. For private content, subscriptions, messages or creator operations, start the service and run:

```sh
docker compose exec boosty-mcp boosty-mcp login
docker compose exec boosty-mcp boosty-mcp status
```

Login works in a **Dokploy container terminal or SSH terminal**, with no browser on the server:

1. Open a private browser window on your own device and log in at `https://boosty.to`.
2. In DevTools → Network, inspect a logged-in request and copy its **Cookie request header**, containing `auth` and `_clientId`.
3. Close the private window **without logging out**. An open browser can rotate the same refresh token and invalidate the server's copy.
4. Paste the header into `boosty-mcp login`. Terminal input is hidden; only the two required session cookies are retained. Avoid putting cookies in shell arguments, chat, logs or source control.
5. Run `boosty-mcp status` to check the session with Boosty. It prints no tokens or profile data.

The running service loads newly imported credentials on its next request. No restart is needed. Login is a local command, never an MCP tool. It accepts one Cookie header line on stdin, so private piping is also possible.

Credentials live at `/app/data/token.json` in the `boosty-data` volume (native default `data/token.json`). Refresh tokens rotate; refreshed credentials are saved atomically with mode `0600` before a request proceeds. A Linux file lock coordinates refresh and login across processes. Keep **one server replica per volume** and preserve the volume when redeploying. `docker compose down -v` deletes it. Volume backups contain credentials. If persistence fails after rotation, the process retains the new token and retries saving; fix storage before restarting. If a crash loses an unsaved rotation, import a fresh session.

## Dokploy and ChatGPT tunnels

In Dokploy, create a **Compose** service using this repository and `main`. Select `docker-compose.yml` followed by `docker-compose.tunnel.yml`. If the UI accepts one Compose path, select **`docker-compose.dokploy.yml`** (uses Compose `include`, requiring Compose 2.20+). Add the `.env.example` variables in Dokploy's environment settings and deploy. Open the `boosty-mcp` container terminal and run `boosty-mcp login`, then `boosty-mcp status`.

Equivalent local command:

```sh
docker compose -f docker-compose.yml -f docker-compose.tunnel.yml up -d --build
```

The tunnel sidecar connects to `http://boosty-mcp:8081/mcp`, adding its dedicated bearer token to **both discovery and runtime requests**. No domain or inbound published port is required. The base Compose file has no OpenAI dependency; the tunnel is optional. `Dockerfile.tunnel` pins the official OpenAI tunnel client to commit `0f870e50a973fa820d4c409000059e181e8d242b` (v0.0.14), matching the sibling projects.

Create/select the tunnel in OpenAI Platform and associate it with the target ChatGPT workspace. In ChatGPT's custom MCP setup, choose **Tunnel** and select or enter that tunnel ID. The sidecar supplies upstream authentication. Keep it running during discovery and tool calls. See [OpenAI Secure MCP Tunnel](https://developers.openai.com/api/docs/guides/secure-mcp-tunnels) for current setup and permissions.

The server uses UID/GID `10001:10001`, a read-only root filesystem, dropped capabilities and a writable named volume. Docker initializes the volume ownership; bind mounts must be owned by that UID/GID. `/healthz` checks process liveness only, not Boosty login or tunnel connectivity. Use `status`, an MCP call and tunnel logs for those checks.

## Direct MCP clients (bearer token)

For direct local HTTP:

```sh
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
```

Connect to `http://127.0.0.1:8080/mcp` with `Authorization: Bearer <MCP_BEARER_TOKEN>`. No OAuth application or OpenAI key is required for direct clients.

For a public HTTPS domain in Dokploy, route it to service `boosty-mcp`, port **8080**, and add the exact forwarded Host (including any port) to `MCP_ALLOWED_HOSTS`. Preserve the existing internal hosts for health and tunnel access. Keep port **8081** private. Both listeners enforce bearer authentication, a Host allowlist, cross-origin protection and a 2 MiB request limit. This is a single-account bearer-authenticated server, not a multi-user OAuth provider.

Example for clients accepting a remote MCP configuration:

```json
{
  "mcpServers": {
    "boosty": {
      "url": "https://your-mcp-domain.example/mcp",
      "headers": { "Authorization": "Bearer YOUR_MCP_BEARER_TOKEN" }
    }
  }
}
```

Client configuration formats vary; use its Streamable HTTP transport and bearer-header setting.

## Native Go / stdio

Requires Go 1.27.1+ on Linux, matching the Docker toolchain and sibling projects.

```sh
make build
make check
./bin/boosty-mcp login
./bin/boosty-mcp status
MCP_TRANSPORT=stdio ./bin/boosty-mcp serve
```

An MCP client can launch `bin/boosty-mcp serve` with `MCP_TRANSPORT=stdio` and an absolute `BOOSTY_TOKEN_FILE`. Stdio needs no MCP bearer token because access is controlled by the local process. Logs go to stderr.

| Variable | Native default / purpose |
| --- | --- |
| `MCP_TRANSPORT` | `http` or `stdio` |
| `MCP_HTTP_ADDR` | `127.0.0.1:8080` |
| `MCP_BEARER_TOKEN` | Required for HTTP; at least 32 characters |
| `MCP_TUNNEL_ADDR` | Empty disables the second listener |
| `MCP_TUNNEL_TOKEN` | Distinct token of at least 32 characters when enabled |
| `MCP_ALLOWED_HOSTS` | Localhost/127.0.0.1 on ports 8080 and 8081; comma separated |
| `BOOSTY_TOKEN_FILE` | `data/token.json` |
| `BOOSTY_READ_ONLY` | `false`; set `true` to omit write tools |

## Validation and limitations

Tests cover MCP discovery/calls, Streamable HTTP, read-only mode, form versus JSON contracts, required arguments, partial updates, pagination preservation, Host/bearer/origin/body limits, cookie parsing, atomic private storage, concurrent refresh, 401 refresh/retry and redirect rejection. Network errors and upstream error bodies are sanitized; rate limits and ambiguous failures are not automatically retried. CI runs race tests, vet, formatting, Compose validation and a Docker build.

On 2026-10-21, anonymous live checks passed for the public `boosty` blog profile, posts, subscription levels, media album and guest currency. Authenticated operations have **not** been live-tested by this project; tests use local mock servers, and experimental writes need account-specific verification. Boosty's internal API can change without notice. The server does not bypass subscription restrictions. No actual Dokploy deployment or ChatGPT workspace connection is performed by building this repository.

References inspected as source text only: [beekamai/boosty-api](https://github.com/beekamai/boosty-api/tree/123b7914aff32cdd96a62b6c5639e2719450a0d9), [krasnovmv/boosty-mcp](https://github.com/krasnovmv/boosty-mcp/tree/865281f8b7a9296374f0b05f6e3dd7781ade49b0). Dependencies: [official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk). Tunnel configuration: [pinned official configuration reference](https://github.com/openai/tunnel-client/blob/0f870e50a973fa820d4c409000059e181e8d242b/docs/configuration.md).
