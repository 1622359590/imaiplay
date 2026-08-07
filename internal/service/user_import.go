package service

import (
	"context"
	"net/mail"
	"strings"

	"github.com/1622359590/imaiplay/internal/errorsx"
)

type UserImportError struct {
	Row    int    `json:"row"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

type UserImportResult struct {
	Total     int               `json:"total"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
	Errors    []UserImportError `json:"errors"`
}

func (service *UserService) Import(ctx context.Context, rows []UserImportRow) (UserImportResult, error) {
	if _, err := tenantAdminID(ctx); err != nil {
		return UserImportResult{}, err
	}
	result := UserImportResult{Total: len(rows), Errors: make([]UserImportError, 0)}
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		email := strings.ToLower(strings.TrimSpace(row.Email))
		phone := normalizePhone(row.Phone)
		role, reason := importRole(row.Role)
		if reason == "" {
			reason = validateUserImportRow(name, email, phone, row.Password)
		}
		if reason == "" {
			if _, err := service.CreateWithPhone(ctx, email, phone, row.Password, name, role); err != nil {
				reason = errorsx.LocalizeMessage(err.Error())
			}
		}
		if reason == "" {
			result.Succeeded++
			continue
		}
		result.Failed++
		result.Errors = append(result.Errors, UserImportError{
			Row: row.Row, Name: name, Email: email, Phone: phone, Role: role, Reason: reason,
		})
	}
	return result, nil
}

func importRole(value string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "学员", "learner":
		return "learner", ""
	case "讲师", "instructor":
		return "instructor", ""
	default:
		return strings.TrimSpace(value), "批量导入仅支持学员和讲师"
	}
}

func validateUserImportRow(name, email, phone, password string) string {
	if name == "" {
		return "请输入姓名"
	}
	if email == "" {
		return "请输入邮箱"
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "邮箱格式不正确"
	}
	if phone != "" && !validPhone(phone) {
		return "手机号格式不正确"
	}
	if len(password) < 8 {
		return "密码至少需要 8 位"
	}
	return ""
}
