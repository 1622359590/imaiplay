package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
)

func TestUserImportPartiallySucceedsWithoutExposingPasswords(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	users := NewUserService(userRepo)
	admin := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin",
	)

	result, err := users.Import(admin, []UserImportRow{
		{Row: 2, Name: " 张三 ", Email: " ZHANG@example.com ", Password: "password1"},
		{Row: 3, Name: "李四", Email: "li@example.com", Role: "讲师", Password: "password2"},
		{Row: 4, Name: "重复", Email: "zhang@example.com", Password: "password3"},
		{Row: 5, Name: "弱密码", Email: "weak@example.com", Password: "short"},
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Total != 4 || result.Succeeded != 2 || result.Failed != 2 || len(result.Errors) != 2 {
		t.Fatalf("Import() = %#v", result)
	}
	if result.Errors[0].Row != 4 || result.Errors[0].Reason != "邮箱已存在" ||
		result.Errors[1].Row != 5 || result.Errors[1].Reason != "密码至少需要 8 位" {
		t.Fatalf("errors = %#v", result.Errors)
	}
	encoded, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	for _, password := range []string{"password1", "password2", "password3", "short"} {
		if strings.Contains(string(encoded), password) {
			t.Fatalf("response exposes password %q: %s", password, encoded)
		}
	}
	created, total, listErr := users.List(admin, 0, 20)
	if listErr != nil || total != 2 || len(created) != 2 {
		t.Fatalf("List() = %#v, %d, %v", created, total, listErr)
	}
	if created[0].Name != "张三" || created[0].Email != "zhang@example.com" || created[0].Role != "learner" ||
		created[1].Role != "instructor" {
		t.Fatalf("created users = %#v", created)
	}
}

func TestUserImportRequiresTenantAdmin(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	users := NewUserService(userRepo)
	row := []UserImportRow{{Row: 2, Name: "学员", Email: "learner@example.com", Password: "password1"}}
	for _, role := range []string{"learner", "superadmin"} {
		ctx := usercontext.WithUser(context.Background(), role, "tenant-1", role+"@example.com", role)
		if _, err := users.Import(ctx, row); errorCode(err) != 40300 {
			t.Fatalf("Import(%s) error = %#v, want forbidden", role, err)
		}
	}
}

func TestUserImportValidatesFieldsAndRoles(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	users := NewUserService(userRepo)
	admin := usercontext.WithUser(context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin")
	tests := []struct {
		name   string
		row    UserImportRow
		reason string
	}{
		{name: "name", row: UserImportRow{Email: "name@example.com", Password: "password1"}, reason: "请输入姓名"},
		{name: "email", row: UserImportRow{Name: "邮箱", Email: "not-an-email", Password: "password1"}, reason: "邮箱格式不正确"},
		{name: "phone", row: UserImportRow{Name: "手机", Email: "phone@example.com", Phone: "123", Password: "password1"}, reason: "手机号格式不正确"},
		{name: "tenant admin", row: UserImportRow{Name: "站长", Email: "admin2@example.com", Role: "站长", Password: "password1"}, reason: "批量导入仅支持学员和讲师"},
		{name: "superadmin", row: UserImportRow{Name: "总管理员", Email: "root@example.com", Role: "superadmin", Password: "password1"}, reason: "批量导入仅支持学员和讲师"},
		{name: "unknown", row: UserImportRow{Name: "未知", Email: "unknown@example.com", Role: "unknown", Password: "password1"}, reason: "批量导入仅支持学员和讲师"},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.row.Row = index + 2
			result, err := users.Import(admin, []UserImportRow{test.row})
			if err != nil || result.Succeeded != 0 || result.Failed != 1 ||
				len(result.Errors) != 1 || result.Errors[0].Reason != test.reason {
				t.Fatalf("Import() = %#v, %v", result, err)
			}
		})
	}
}

func TestUserImportScopesDuplicatesToCurrentTenant(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	users := NewUserService(userRepo)
	otherAdmin := usercontext.WithUser(context.Background(), "other-admin", "tenant-2", "other@example.com", "tenant_admin")
	if _, err := users.Create(otherAdmin, "shared@example.com", "password1", "其他租户", "learner"); err != nil {
		t.Fatal(err)
	}
	admin := usercontext.WithUser(context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin")
	result, err := users.Import(admin, []UserImportRow{{
		Row: 2, Name: "当前租户", Email: "shared@example.com", Password: "password1",
	}})
	if err != nil || result.Succeeded != 1 || result.Failed != 0 {
		t.Fatalf("Import() = %#v, %v", result, err)
	}
}
