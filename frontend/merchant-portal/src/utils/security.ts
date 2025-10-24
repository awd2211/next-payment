/**
 * 安全工具集 - CSP、XSS防护、输入验证
 */

/**
 * 内容安全策略配置
 */
export const cspConfig = {
  'default-src': ["'self'"],
  'script-src': [
    "'self'",
    "'unsafe-inline'", // 允许内联脚本(生产环境应移除)
    'https://js.stripe.com',
  ],
  'style-src': ["'self'", "'unsafe-inline'", 'https://fonts.googleapis.com'],
  'img-src': ["'self'", 'data:', 'https:'],
  'font-src': ["'self'", 'https://fonts.gstatic.com'],
  'connect-src': [
    "'self'",
    'http://localhost:40080',
    'http://localhost:40002',
    'http://localhost:40003',
    'ws://localhost:5174',
    'https://api.stripe.com',
  ],
  'frame-src': ["'self'", 'https://js.stripe.com', 'https://hooks.stripe.com'],
  'object-src': ["'none'"],
  'base-uri': ["'self'"],
  'form-action': ["'self'"],
}

/**
 * 生成CSP头部字符串
 */
export const generateCSPHeader = (): string => {
  return Object.entries(cspConfig)
    .map(([key, values]) => `${key} ${values.join(' ')}`)
    .join('; ')
}

/**
 * XSS防护 - HTML转义
 */
export const escapeHTML = (str: string): string => {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

/**
 * XSS防护 - 移除危险标签
 */
export const sanitizeHTML = (html: string): string => {
  const dangerousTags = /<script|<iframe|<object|<embed|<link|javascript:/gi
  return html.replace(dangerousTags, '')
}

/**
 * 验证URL是否安全(防止open redirect)
 */
export const isValidURL = (url: string): boolean => {
  try {
    const urlObj = new URL(url, window.location.origin)
    // 只允许http/https协议
    if (!['http:', 'https:'].includes(urlObj.protocol)) {
      return false
    }
    // 只允许当前域名或指定的白名单域名
    const allowedHosts = [
      window.location.host,
      'localhost:5174',
      'localhost:40080',
      'js.stripe.com',
      'api.stripe.com',
    ]
    return allowedHosts.includes(urlObj.host)
  } catch {
    return false
  }
}

/**
 * 验证输入长度
 */
export const validateLength = (
  value: string,
  min: number,
  max: number
): { valid: boolean; message?: string } => {
  if (value.length < min) {
    return { valid: false, message: `长度不能少于${min}个字符` }
  }
  if (value.length > max) {
    return { valid: false, message: `长度不能超过${max}个字符` }
  }
  return { valid: true }
}

/**
 * 验证邮箱格式
 */
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

/**
 * 验证手机号格式(支持多国)
 */
export const validatePhone = (phone: string, countryCode = 'CN'): boolean => {
  const patterns: Record<string, RegExp> = {
    CN: /^1[3-9]\d{9}$/, // 中国大陆
    US: /^\+?1?\d{10}$/, // 美国
    HK: /^[5-9]\d{7}$/, // 香港
  }
  return patterns[countryCode]?.test(phone) ?? false
}

/**
 * 密码强度验证
 */
export const validatePasswordStrength = (
  password: string
): { strength: 'weak' | 'medium' | 'strong'; message: string } => {
  if (password.length < 8) {
    return { strength: 'weak', message: '密码长度不足8位' }
  }

  let score = 0
  if (/[a-z]/.test(password)) score++ // 小写字母
  if (/[A-Z]/.test(password)) score++ // 大写字母
  if (/\d/.test(password)) score++ // 数字
  if (/[^a-zA-Z0-9]/.test(password)) score++ // 特殊字符

  if (score < 2) {
    return { strength: 'weak', message: '密码过于简单' }
  }
  if (score === 2 || score === 3) {
    return { strength: 'medium', message: '密码强度中等' }
  }
  return { strength: 'strong', message: '密码强度强' }
}

/**
 * 安全的JSON解析
 */
export const safeJSONParse = <T = any>(str: string, defaultValue: T): T => {
  try {
    return JSON.parse(str) as T
  } catch {
    return defaultValue
  }
}

/**
 * 防止点击劫持 - 检查是否在iframe中
 */
export const preventClickjacking = (): void => {
  if (window.self !== window.top) {
    window.top!.location = window.self.location
  }
}

/**
 * 生成随机Token(CSRF保护)
 */
export const generateCSRFToken = (): string => {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('')
}
