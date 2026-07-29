package errorsx

import "strings"

var messageTranslations = map[string]string{
	"internal server error":                       "系统内部错误，请稍后重试",
	"invalid request":                             "请求参数不正确",
	"invalid email or password":                   "邮箱或密码错误",
	"identifier and password are required":        "请输入邮箱或手机号和密码",
	"missing or invalid token":                    "登录状态无效，请重新登录",
	"permission denied":                           "没有权限执行此操作",
	"tenant not found":                            "租户不存在",
	"tenant is unavailable":                       "租户暂不可用",
	"invalid phone":                               "手机号格式不正确",
	"invalid role":                                "用户角色不正确",
	"email already exists":                        "邮箱已存在",
	"phone already exists":                        "手机号已存在",
	"invalid or expired verification code":        "验证码无效或已过期",
	"too many verification attempts":              "验证码尝试次数过多",
	"invalid verification code":                   "验证码错误",
	"password must be at least 8 characters":      "密码至少需要 8 位",
	"not enrolled in this course":                 "你尚未加入该课程",
	"user is disabled":                            "该用户已停用",
	"learner is disabled":                         "该学员已停用",
	"learner already enrolled":                    "该学员已经加入课程",
	"category name is required":                   "请输入分类名称",
	"invalid content type":                        "内容类型不正确",
	"invalid parent category":                     "父级分类不正确",
	"storage connection failed":                   "存储连接失败",
	"unsupported file type or size exceeds limit": "文件类型不支持或文件大小超过限制",
	"create tenant failed":                        "创建租户失败",
	"tenant code already exists":                  "租户编码已存在",
	"custom domain already exists":                "自定义域名已存在",
	"custom domain is invalid":                    "自定义域名不正确",
	"superadmin already initialized":              "总管理员已经初始化",
}

// LocalizeMessage converts known API error messages to user-facing Chinese.
// Unknown messages are preserved so internal diagnostics are not discarded.
func LocalizeMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if translated, ok := messageTranslations[strings.ToLower(trimmed)]; ok {
		return translated
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.Contains(lower, "network error"), strings.Contains(lower, "connection refused"):
		return "网络异常，请检查服务是否可用"
	case strings.Contains(lower, "timeout"):
		return "请求超时，请稍后重试"
	case strings.Contains(lower, "record not found"), strings.HasSuffix(lower, " not found"):
		return "请求的数据不存在"
	case strings.Contains(lower, "foreign key constraint"):
		return "数据关联校验失败"
	case strings.Contains(lower, "unique constraint"), strings.Contains(lower, "duplicate key"):
		return "数据已存在，不能重复创建"
	}
	if isEnglishMessage(trimmed) {
		return "请求失败，请稍后重试"
	}
	return message
}

func isEnglishMessage(message string) bool {
	letter := false
	for _, char := range message {
		if char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' {
			letter = true
			continue
		}
		if char >= '\u4e00' && char <= '\u9fff' {
			return false
		}
	}
	return letter
}
