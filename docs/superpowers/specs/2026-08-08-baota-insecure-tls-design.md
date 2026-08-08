# BaoTa Self-Signed TLS Design

## Goal

Allow the ImaiPlay backend to call a same-host BaoTa panel that uses a self-signed HTTPS certificate, while preserving strict certificate verification by default.

## Configuration

Add `BAOTA_TLS_INSECURE_SKIP_VERIFY`, parsed as a boolean and defaulting to `false`. The Docker Compose service passes this value to the backend container, and `.env.example` documents that it is only intended for a trusted same-host Docker-to-BaoTa connection.

## Architecture

The application configuration exposes the boolean as `BaotaTLSInsecureSkipVerify`. When BaoTa integration is enabled, server startup passes the value into the BaoTa client. The client builds its own HTTP transport only when insecure verification is explicitly enabled; otherwise it continues using the existing strict default client.

The exception is scoped to BaoTa API requests. Storage, media, portal, and all other HTTPS clients remain unchanged.

## Security and Observability

- Default behavior remains strict TLS verification.
- Enabling the exception requires the exact environment value `true`.
- Server startup emits a warning when the exception is enabled.
- The option does not disable encryption; it disables only certificate identity and trust verification for BaoTa API requests.
- Operators should prefer a valid BaoTa panel certificate when the panel is not confined to the same host.

## Testing

- Configuration tests prove the value defaults to `false` and loads `true` from the environment.
- BaoTa client tests use a self-signed TLS test server to prove strict mode rejects the certificate and explicit insecure mode accepts it.
- Server wiring, Docker Compose configuration, and example environment documentation are covered by targeted checks and the full Go test suite.

