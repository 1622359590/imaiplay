package service

import (
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

func TestTenantAccessibleLifecycle(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name, status string
		ends         *time.Time
		want         bool
	}{
		{name: "active", status: "active", want: true},
		{name: "suspended", status: "suspended", want: false},
		{name: "deleted", status: "deleted", want: false},
		{name: "trial current", status: "trial", ends: ptrTime(now.Add(time.Hour)), want: true},
		{name: "trial expired", status: "trial", ends: ptrTime(now.Add(-time.Hour)), want: false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			got, _ := TenantAccessible(&domain.Tenant{Status: 1, LifecycleStatus: item.status, TrialEndsAt: item.ends}, now)
			if got != item.want {
				t.Fatalf("accessible = %v, want %v", got, item.want)
			}
		})
	}
}

func TestTenantAccessibleRejectsDisabledTenantRegardlessOfLifecycle(t *testing.T) {
	now := time.Now().UTC()
	trialEndsAt := now.Add(time.Hour)
	for _, lifecycleStatus := range []string{"", "active", "trial"} {
		t.Run(lifecycleStatus, func(t *testing.T) {
			accessible, _ := TenantAccessible(&domain.Tenant{
				Status:          0,
				LifecycleStatus: lifecycleStatus,
				TrialEndsAt:     &trialEndsAt,
			}, now)
			if accessible {
				t.Fatalf("disabled tenant with lifecycle status %q was accessible", lifecycleStatus)
			}
		})
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
