type ResponseLike = {
  code?: unknown
  message?: unknown
  response?: {
    status?: unknown
    data?: { message?: unknown; error?: unknown }
  }
}

const DEFAULT_ERROR_MESSAGE = '请求失败，请稍后重试'

const businessMessages: Record<string, string> = {
  'invalid email or password': '邮箱或密码错误',
  'permission denied': '您没有权限执行此操作',
  'authentication required': '请先登录后再操作',
  'missing or invalid token': '登录信息已失效，请重新登录',
  'invalid refresh token': '登录信息已失效，请重新登录',
  'tenant is suspended': '当前企业已暂停使用，请联系管理员',
  'tenant trial expired': '当前企业试用期已结束，请联系管理员',
  'tenant is unavailable': '当前企业暂时不可用，请联系管理员',
  'tenant not found': '未找到对应企业',
  'user is disabled': '当前账号已被停用，请联系管理员',
  'learner is disabled': '当前学员账号已被停用，请联系管理员',
  'email already exists': '该邮箱已被使用',
  'phone already exists': '该手机号已被使用',
  'course not found': '课程不存在或已被删除',
  'resource not found': '资源不存在或已被删除',
  'course material not found': '课程资料不存在或已被删除',
  'resource is in use': '该资源正在使用中，暂时无法删除',
  'course category already exists': '课程分类已存在',
  'course category is referenced': '该分类仍被课程使用，暂时无法删除',
  'course material already exists': '该课程资料已存在',
  'custom domain already exists': '该域名已被使用',
  'learner already enrolled': '该学员已分配此课程',
  'not enrolled in this course': '您尚未获得该课程的学习权限',
  'invalid or expired verification code': '验证码错误或已失效',
  'invalid verification code': '验证码错误',
  'too many verification attempts': '验证码尝试次数过多，请稍后再试',
  'please wait before requesting another code': '操作过于频繁，请稍后再获取验证码',
  'password must be at least 8 characters': '密码不能少于 8 个字符',
  'unsupported file type or size exceeds limit': '文件类型不支持或文件大小超过限制',
  'storage connection failed': '存储服务连接失败，请稍后重试',
  'storage connection unavailable': '存储服务暂时不可用，请稍后重试',
  'playback is not configured': '视频播放服务尚未配置，请联系管理员',
  'portal service is unavailable': '学习门户暂时不可用，请稍后重试',
  'invalid request': '请求参数有误，请检查后重试',
  'account_exists_multiple_tenants': '该账号属于多个企业，请先选择企业',
}

const statusMessages: Record<number, string> = {
  400: '请求参数有误，请检查后重试',
  401: '登录信息已失效，请重新登录',
  403: '您没有权限执行此操作',
  404: '请求的内容不存在',
  408: '请求超时，请稍后重试',
  409: '数据已存在或当前状态发生冲突',
  413: '上传文件过大，请选择较小的文件',
  422: '提交的内容有误，请检查后重试',
  429: '操作过于频繁，请稍后再试',
  500: '服务器暂时异常，请稍后重试',
  502: '服务暂时不可用，请稍后重试',
  503: '服务暂时不可用，请稍后重试',
  504: '服务响应超时，请稍后重试',
}

function hasChinese(value: string): boolean {
  return /[\u3400-\u9fff]/.test(value)
}

function rawResponseMessage(error: unknown): string | undefined {
  const data = (error as ResponseLike | null)?.response?.data
  const value = data?.message ?? data?.error
  return typeof value === 'string' && value.trim() ? value.trim() : undefined
}

export function responseStatus(error: unknown): number | undefined {
  const value = (error as ResponseLike | null)?.response?.status
  return typeof value === 'number' ? value : undefined
}

export function responseMessage(error: unknown): string | undefined {
  return rawResponseMessage(error)
}

export function userFacingErrorMessage(
  error: unknown,
  fallback = DEFAULT_ERROR_MESSAGE,
): string {
  const candidate = rawResponseMessage(error)
  if (candidate && hasChinese(candidate)) return candidate

  const normalized = candidate?.toLowerCase()
  if (normalized && businessMessages[normalized]) return businessMessages[normalized]

  const errorLike = error as ResponseLike | null
  const code = typeof errorLike?.code === 'string' ? errorLike.code : undefined
  const message = typeof errorLike?.message === 'string' ? errorLike.message.trim() : undefined
  const normalizedErrorMessage = message?.toLowerCase()

  if (normalizedErrorMessage && businessMessages[normalizedErrorMessage]) {
    return businessMessages[normalizedErrorMessage]
  }

  if (code === 'ECONNABORTED' || /timeout/i.test(message ?? '')) {
    return '请求超时，请稍后重试'
  }
  if (code === 'ERR_NETWORK' || message === 'Network Error') {
    return '网络连接失败，请检查网络后重试'
  }

  const status = responseStatus(error)
  if (status && statusMessages[status]) return statusMessages[status]

  if (message && hasChinese(message)) return message
  return hasChinese(fallback) ? fallback : DEFAULT_ERROR_MESSAGE
}
