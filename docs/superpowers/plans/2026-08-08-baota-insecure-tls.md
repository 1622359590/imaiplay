# BaoTa Self-Signed TLS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `BAOTA_TLS_INSECURE_SKIP_VERIFY=true` allow only the BaoTa API client to connect to a trusted same-host panel with a self-signed certificate.

**Architecture:** Extend application configuration with a default-false boolean and pass it through Docker Compose into the BaoTa client. The client creates a dedicated HTTP transport with `tls.Config.InsecureSkipVerify` only when the flag is explicitly enabled; strict TLS remains the default for BaoTa and all other clients.

**Tech Stack:** Go 1.24, `net/http`, `crypto/tls`, Viper, Docker Compose, Go `testing` and `httptest`.

## Global Constraints

- `BAOTA_TLS_INSECURE_SKIP_VERIFY` defaults to `false`.
- The exception applies only to BaoTa API requests.
- Existing custom `HTTPClient` injection in BaoTa tests remains authoritative.
- Enabling the option emits a startup warning without logging credentials.
- No new third-party dependency is added.

---

### Task 1: Load and expose the opt-in configuration

**Files:**
- Modify: `internal/config/config_test.go`
- Modify: `internal/config/config.go`
- Modify: `docker-compose.yml`
- Modify: `.env.example`
- Modify: `docs/域名配置指南.md`

**Interfaces:**
- Consumes: environment variable `BAOTA_TLS_INSECURE_SKIP_VERIFY`.
- Produces: `Config.BaotaTLSInsecureSkipVerify bool`.

- [ ] **Step 1: Write the failing configuration tests**

Add `BAOTA_TLS_INSECURE_SKIP_VERIFY=true` to `TestLoadEnvironment` and require `BaotaTLSInsecureSkipVerify: true` in its expected `Config`. Add `BAOTA_TLS_INSECURE_SKIP_VERIFY` to `unsetConfigEnvironment`, while leaving `defaultConfig()` with the zero-value `false` expectation.

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
go test ./internal/config -run 'TestLoad(Default|Environment)$' -count=1
```

Expected: compilation fails because `Config` has no `BaotaTLSInsecureSkipVerify` field.

- [ ] **Step 3: Implement the minimal configuration path**

Add the field and Viper mapping:

```go
BaotaTLSInsecureSkipVerify bool
```

```go
v.SetDefault("BAOTA_TLS_INSECURE_SKIP_VERIFY", false)
```

```go
BaotaTLSInsecureSkipVerify: v.GetBool("BAOTA_TLS_INSECURE_SKIP_VERIFY"),
```

Pass it into the container in `docker-compose.yml`:

```yaml
BAOTA_TLS_INSECURE_SKIP_VERIFY: ${BAOTA_TLS_INSECURE_SKIP_VERIFY:-false}
```

Document the opt-in in `.env.example` and `docs/域名配置指南.md`, explicitly limiting it to trusted same-host self-signed BaoTa panels.

- [ ] **Step 4: Run the focused tests and verify GREEN**

Run:

```bash
go test ./internal/config -count=1
docker compose -f docker-compose.yml -f docker-compose.bt.yml config >/dev/null
```

Expected: both commands succeed.

- [ ] **Step 5: Commit the configuration slice**

```bash
git add internal/config/config.go internal/config/config_test.go docker-compose.yml .env.example docs/域名配置指南.md
git commit -m "feat: configure baota self-signed tls opt-in"
```

### Task 2: Scope insecure TLS to the BaoTa client

**Files:**
- Modify: `internal/baota/client_test.go`
- Modify: `internal/baota/client.go`
- Modify: `cmd/server/main.go`

**Interfaces:**
- Consumes: `Config.BaotaTLSInsecureSkipVerify`.
- Produces: `baota.Client.TLSInsecureSkipVerify bool` and a dedicated TLS transport when true.

- [ ] **Step 1: Write the failing self-signed TLS tests**

Use `httptest.NewTLSServer` to return `{}` from `/data`. Add one test proving a default client rejects the self-signed certificate, and a second test constructing:

```go
client := &Client{
    PanelURL:               server.URL,
    APIKey:                "test-key",
    TLSInsecureSkipVerify: true,
}
```

and proving `client.request(context.Background(), "/data", url.Values{"action": {"getData"}}, false, false)` succeeds.

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
go test ./internal/baota -run 'TestClient(SelfSignedTLS|AllowsExplicitInsecureTLS)$' -count=1
```

Expected: compilation fails because `Client` has no `TLSInsecureSkipVerify` field.

- [ ] **Step 3: Implement the minimal scoped transport**

Add the field to `Client`. When `HTTPClient` is nil and the flag is true, clone `http.DefaultTransport`, set a dedicated `tls.Config{InsecureSkipVerify: true}` with a security comment explaining the explicit same-host opt-in, and attach it to the new `http.Client`. When the flag is false, preserve the current strict default client. Do not alter an injected `HTTPClient`.

Wire the value in `cmd/server/main.go`:

```go
TLSInsecureSkipVerify: cfg.BaotaTLSInsecureSkipVerify,
```

Emit a `slog.Warn` at startup when enabled, naming only the panel URL and never the API key.

- [ ] **Step 4: Run focused and full verification**

Run:

```bash
gofmt -w internal/config/config.go internal/config/config_test.go internal/baota/client.go internal/baota/client_test.go cmd/server/main.go
go test ./internal/baota ./internal/config ./cmd/server -count=1
go test ./... -count=1
go vet ./...
docker compose -f docker-compose.yml -f docker-compose.bt.yml config >/dev/null
```

Expected: every command succeeds; strict TLS rejection and explicit opt-in acceptance are both covered.

- [ ] **Step 5: Commit the client behavior**

```bash
git add internal/baota/client.go internal/baota/client_test.go cmd/server/main.go
git commit -m "feat: allow opt-in baota self-signed tls"
```

