package context

import (
	stdcontext "context"
	"testing"
)

func TestTenantContextRoundTrip(t *testing.T) {
	ctx := WithTenant(stdcontext.Background(), "tenant1", SourceSubdomain)

	code, source := TenantFromContext(ctx)
	if code != "tenant1" || source != SourceSubdomain {
		t.Fatalf("TenantFromContext() = (%q, %q), want (%q, %q)",
			code, source, "tenant1", SourceSubdomain)
	}
}

func TestTenantFromContextDefaultsToUnknown(t *testing.T) {
	code, source := TenantFromContext(stdcontext.Background())
	if code != UnknownTenant || source != SourceUnknown {
		t.Fatalf("TenantFromContext() = (%q, %q), want (%q, %q)",
			code, source, UnknownTenant, SourceUnknown)
	}
}
