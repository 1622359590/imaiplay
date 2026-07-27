package context

import (
	stdcontext "context"
	"testing"
)

func TestUserContextRoundTrip(t *testing.T) {
	ctx := WithUser(
		stdcontext.Background(),
		"user-1",
		"tenant-1",
		"admin@example.com",
		"tenant_admin",
	)
	userID, tenantID, email, role, ok := UserFromContext(ctx)
	if !ok || userID != "user-1" || tenantID != "tenant-1" ||
		email != "admin@example.com" || role != "tenant_admin" {
		t.Fatalf("UserFromContext() = %q, %q, %q, %q, %t",
			userID, tenantID, email, role, ok)
	}
}

func TestUserFromEmptyContext(t *testing.T) {
	_, _, _, _, ok := UserFromContext(stdcontext.Background())
	if ok {
		t.Fatal("UserFromContext() ok = true, want false")
	}
}
