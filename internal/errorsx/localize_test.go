package errorsx

import "testing"

func TestLocalizeMessage(t *testing.T) {
	for input, want := range map[string]string{
		"invalid email or password": "邮箱或密码错误",
		"permission denied":         "没有权限执行此操作",
		"internal server error":     "系统内部错误，请稍后重试",
	} {
		if got := LocalizeMessage(input); got != want {
			t.Fatalf("LocalizeMessage(%q) = %q, want %q", input, got, want)
		}
	}
}
