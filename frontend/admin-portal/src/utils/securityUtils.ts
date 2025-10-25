/**
 * 安全工具函数
 * XSS 防护、CSRF 防护、敏感数据处理
 */

/**
 * XSS 防护 - HTML 实体编码
 */
export function escapeHtml(unsafe: string): string {
  return unsafe
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

/**
 * XSS 防护 - 移除危险标签和属性
 */
export function sanitizeHtml(dirty: string): string {
  const div = document.createElement('div')
  div.textContent = dirty
  return div.innerHTML
}

/**
 * XSS 防护 - URL 参数编码
 */
export function encodeUrlParam(param: string): string {
  return encodeURIComponent(param)
}

/**
 * 生成 CSRF Token
 */
export function generateCSRFToken(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * 获取或创建 CSRF Token
 */
export function getCSRFToken(): string {
  let token = sessionStorage.getItem('csrf_token')
  if (!token) {
    token = generateCSRFToken()
    sessionStorage.setItem('csrf_token', token)
  }
  return token
}

/**
 * 验证 CSRF Token (前端校验)
 */
export function validateCSRFToken(token: string): boolean {
  const storedToken = sessionStorage.getItem('csrf_token')
  return !!storedToken && storedToken === token
}

/**
 * 敏感数据加密 (简单混淆,非安全加密)
 * 注意: 前端加密只能防止明文传输,真正的加密应在后端进行
 */
export function obfuscate(text: string, key: string = 'default-key'): string {
  let result = ''
  for (let i = 0; i < text.length; i++) {
    const charCode = text.charCodeAt(i) ^ key.charCodeAt(i % key.length)
    result += String.fromCharCode(charCode)
  }
  return btoa(result) // Base64 编码
}

/**
 * 敏感数据解密
 */
export function deobfuscate(encoded: string, key: string = 'default-key'): string {
  try {
    const decoded = atob(encoded)
    let result = ''
    for (let i = 0; i < decoded.length; i++) {
      const charCode = decoded.charCodeAt(i) ^ key.charCodeAt(i % key.length)
      result += String.fromCharCode(charCode)
    }
    return result
  } catch (error) {
    console.error('Deobfuscation failed:', error)
    return ''
  }
}

/**
 * 安全的本地存储 (带混淆)
 */
export const secureStorage = {
  /**
   * 存储数据
   */
  set(key: string, value: any, obfuscateValue: boolean = false): void {
    try {
      const stringValue = typeof value === 'string' ? value : JSON.stringify(value)
      const finalValue = obfuscateValue ? obfuscate(stringValue) : stringValue
      localStorage.setItem(key, finalValue)
    } catch (error) {
      console.error('SecureStorage set error:', error)
    }
  },

  /**
   * 获取数据
   */
  get<T = any>(key: string, isObfuscated: boolean = false): T | null {
    try {
      const value = localStorage.getItem(key)
      if (!value) return null

      const decodedValue = isObfuscated ? deobfuscate(value) : value

      try {
        return JSON.parse(decodedValue) as T
      } catch {
        return decodedValue as T
      }
    } catch (error) {
      console.error('SecureStorage get error:', error)
      return null
    }
  },

  /**
   * 删除数据
   */
  remove(key: string): void {
    localStorage.removeItem(key)
  },

  /**
   * 清空所有数据
   */
  clear(): void {
    localStorage.clear()
  },
}

/**
 * 密码强度检查
 */
export function checkPasswordStrength(password: string): {
  score: number // 0-4
  level: 'very-weak' | 'weak' | 'medium' | 'strong' | 'very-strong'
  feedback: string[]
} {
  let score = 0
  const feedback: string[] = []

  // 长度检查
  if (password.length < 8) {
    feedback.push('密码至少需要8个字符')
  } else if (password.length >= 12) {
    score++
  }

  // 包含小写字母
  if (/[a-z]/.test(password)) {
    score++
  } else {
    feedback.push('密码应包含小写字母')
  }

  // 包含大写字母
  if (/[A-Z]/.test(password)) {
    score++
  } else {
    feedback.push('密码应包含大写字母')
  }

  // 包含数字
  if (/[0-9]/.test(password)) {
    score++
  } else {
    feedback.push('密码应包含数字')
  }

  // 包含特殊字符
  if (/[^a-zA-Z0-9]/.test(password)) {
    score++
  } else {
    feedback.push('建议包含特殊字符')
  }

  // 检查常见弱密码
  const commonPasswords = ['123456', 'password', '12345678', 'qwerty', '123456789', 'abc123']
  if (commonPasswords.some((p) => password.toLowerCase().includes(p))) {
    score = Math.max(0, score - 2)
    feedback.push('密码过于简单,请使用更复杂的组合')
  }

  const levels = ['very-weak', 'weak', 'medium', 'strong', 'very-strong'] as const
  const level = levels[Math.min(score, 4)]

  return { score, level, feedback }
}

/**
 * 安全的字符串比较 (防止时序攻击)
 */
export function secureCompare(a: string, b: string): boolean {
  if (a.length !== b.length) {
    return false
  }

  let result = 0
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i)
  }

  return result === 0
}

/**
 * 生成随机字符串
 */
export function generateRandomString(length: number = 32): string {
  const array = new Uint8Array(length)
  crypto.getRandomValues(array)
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * 检测潜在的 XSS 攻击
 */
export function detectXSS(input: string): boolean {
  const xssPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /javascript:/gi,
    /on\w+\s*=/gi, // 事件处理器
    /<iframe/gi,
    /<object/gi,
    /<embed/gi,
  ]

  return xssPatterns.some((pattern) => pattern.test(input))
}

/**
 * 检测潜在的 SQL 注入
 */
export function detectSQLInjection(input: string): boolean {
  const sqlPatterns = [
    /(\bor\b|\band\b).*?=/gi,
    /union.*?select/gi,
    /drop\s+table/gi,
    /insert\s+into/gi,
    /delete\s+from/gi,
    /update.*?set/gi,
    /--/gi, // SQL 注释
    /;.*?(drop|delete|insert|update|select)/gi,
  ]

  return sqlPatterns.some((pattern) => pattern.test(input))
}

/**
 * 输入验证和清理
 */
export function validateInput(
  input: string,
  options: {
    maxLength?: number
    allowHtml?: boolean
    checkXSS?: boolean
    checkSQL?: boolean
  } = {}
): { valid: boolean; sanitized: string; errors: string[] } {
  const { maxLength, allowHtml = false, checkXSS = true, checkSQL = true } = options
  const errors: string[] = []
  let sanitized = input

  // 长度检查
  if (maxLength && input.length > maxLength) {
    errors.push(`输入长度不能超过 ${maxLength} 个字符`)
    sanitized = input.substring(0, maxLength)
  }

  // XSS 检测
  if (checkXSS && detectXSS(input)) {
    errors.push('检测到潜在的 XSS 攻击')
  }

  // SQL 注入检测
  if (checkSQL && detectSQLInjection(input)) {
    errors.push('检测到潜在的 SQL 注入攻击')
  }

  // HTML 清理
  if (!allowHtml) {
    sanitized = escapeHtml(sanitized)
  }

  return {
    valid: errors.length === 0,
    sanitized,
    errors,
  }
}

/**
 * 安全的JSON解析
 */
export function safeJSONParse<T = any>(json: string, defaultValue: T | null = null): T | null {
  try {
    return JSON.parse(json) as T
  } catch (error) {
    console.error('JSON parse error:', error)
    return defaultValue
  }
}

/**
 * 清理 URL (移除潜在的危险协议)
 */
export function sanitizeUrl(url: string): string {
  const dangerousProtocols = ['javascript:', 'data:', 'vbscript:']

  for (const protocol of dangerousProtocols) {
    if (url.toLowerCase().trim().startsWith(protocol)) {
      return ''
    }
  }

  return url
}
