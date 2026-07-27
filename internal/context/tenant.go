package context

import stdcontext "context"

const (
	SourceSubdomain    = "subdomain"
	SourceHeaderID     = "header_id"
	SourceHeaderCode   = "header_code"
	SourceUnknown      = "unknown"
	SourceCustomDomain = "custom_domain"
	UnknownTenant      = "unknown"
)

type tenantKey struct{}

type tenant struct {
	code   string
	source string
}

func WithTenant(ctx stdcontext.Context, code, source string) stdcontext.Context {
	return stdcontext.WithValue(ctx, tenantKey{}, tenant{code: code, source: source})
}

func TenantFromContext(ctx stdcontext.Context) (code string, source string) {
	value, ok := ctx.Value(tenantKey{}).(tenant)
	if !ok {
		return UnknownTenant, SourceUnknown
	}
	return value.code, value.source
}
