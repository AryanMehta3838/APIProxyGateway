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

### Redis-backed rate limiting (T10)

The dev config enables Redis-backed limiting:

```yaml
redis:
  enabled: true
  addr: localhost:6379
```

Start Redis with Docker:

```bash
docker run --rm -p 6379:6379 redis:7-alpine
```

Then run the gateway and upstream as above, and hit the limited route repeatedly:

```bash
for i in 1 2 3 4; do curl -s -o /dev/null -w "%{http_code}\n" "http://127.0.0.1:8080/api/echo/hello"; done
```

Expected with the current dev config (`requests: 3`): first three requests return `200`, fourth returns `429`.
If Redis is down while Redis limiting is enabled, limited routes return `503`.

You can also run all three services together:

```bash
docker compose -f deployments/docker-compose.yml up
```

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
