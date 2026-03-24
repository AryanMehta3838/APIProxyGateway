# API Gateway Proxy

Minimal bootstrap for the gateway service.

## Run

### Gateway

```bash
go run ./cmd/gateway -config configs/gateway.dev.yaml
```

Then open `http://localhost:8080/healthz` in a browser and confirm it returns `ok`.
You can also open `http://localhost:8080/readyz` and confirm it returns `ready`.

Press `Ctrl+C` to stop the server and confirm it exits cleanly.

### Echo upstream fixture (for local demos and future proxy tests)

In a second terminal:

```bash
go run ./examples/echo-upstream
```

By default it listens on `:9091`. Try:

```bash
curl -sS "http://127.0.0.1:9091/hello?name=aryan"
```

You should see JSON with `method`, `path`, and `query` echoing the request. Use `-addr` to bind elsewhere. Stop with `Ctrl+C` for graceful shutdown.

### Proxy happy path

With both the gateway and echo upstream running:

```bash
curl -sS "http://127.0.0.1:8080/api/echo/hello?name=aryan"
```

The gateway should proxy the request to the upstream and return the upstream JSON response.
With the current dev config, the gateway strips `/api/echo` before forwarding, so the upstream response should show `"path":"/hello"`.

To see request ID propagation:

```bash
curl -i "http://127.0.0.1:8080/api/echo/hello?name=aryan"
curl -i -H "X-Request-ID: demo-request-id" "http://127.0.0.1:8080/api/echo/hello?name=aryan"
```

The gateway should return `X-Request-ID` on the response. When you provide one, it should be preserved.

Rate limiting is available per route through the `rate_limit` config block. Health endpoints are not rate limited.

## Validate

```bash
go build ./...
go test ./...
```

Example invalid-config check:

```bash
cp configs/gateway.dev.yaml /tmp/gateway.invalid.yaml
perl -0pi -e 's/port: 8080/port: 0/' /tmp/gateway.invalid.yaml
go run ./cmd/gateway -config /tmp/gateway.invalid.yaml
```

The process should exit immediately with a clear startup error mentioning `server.port`.
