/**
 * 表单验证工具函数
 * 提供常用的验证规则
 */

/**
 * 邮箱验证
 */
export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

/**
 * 手机号验证 (中国大陆)
 */
export function validatePhone(phone: string): boolean {
  const phoneRegex = /^1[3-9]\d{9}$/
  return phoneRegex.test(phone)
}

/**
 * 密码强度验证
 * 至少8位,包含大小写字母和数字
 */
export function validatePasswordStrength(password: string): {
  valid: boolean
  strength: 'weak' | 'medium' | 'strong'
  messages: string[]
} {
  const messages: string[] = []
  let strength: 'weak' | 'medium' | 'strong' = 'weak'

  if (password.length < 8) {
    messages.push('密码至少8个字符')
  }

  if (!/[a-z]/.test(password)) {
    messages.push('密码需包含小写字母')
  }

  if (!/[A-Z]/.test(password)) {
    messages.push('密码需包含大写字母')
  }

  if (!/[0-9]/.test(password)) {
    messages.push('密码需包含数字')
  }

  if (!/[^a-zA-Z0-9]/.test(password)) {
    messages.push('建议包含特殊字符')
  }

  // 计算强度
  if (messages.length === 0) {
    strength = 'strong'
  } else if (messages.length <= 2) {
    strength = 'medium'
  }

  return {
    valid: messages.length === 0,
    strength,
    messages,
  }
}

/**
 * URL 验证
 */
export function validateURL(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * IP 地址验证
 */
export function validateIP(ip: string): boolean {
  const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
  return ipRegex.test(ip)
}

/**
 * 身份证号验证 (中国大陆)
 */
export function validateIDCard(idCard: string): boolean {
  const idCardRegex = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/
  return idCardRegex.test(idCard)
}

/**
 * 银行卡号验证 (Luhn算法)
 */
export function validateBankCard(cardNo: string): boolean {
  if (!/^\d+$/.test(cardNo)) {
    return false
  }

  // Luhn算法
  let sum = 0
  let isEven = false

  for (let i = cardNo.length - 1; i >= 0; i--) {
    let digit = parseInt(cardNo[i])

    if (isEven) {
      digit *= 2
      if (digit > 9) {
        digit -= 9
      }
    }

    sum += digit
    isEven = !isEven
  }

  return sum % 10 === 0
}

/**
 * 金额验证
 */
export function validateAmount(amount: number | string, min: number = 0, max: number = Number.MAX_SAFE_INTEGER): boolean {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount

  if (isNaN(value)) {
    return false
  }

  return value >= min && value <= max
}

/**
 * 用户名验证
 * 4-20位,字母数字下划线
 */
export function validateUsername(username: string): boolean {
  const usernameRegex = /^[a-zA-Z0-9_]{4,20}$/
  return usernameRegex.test(username)
}

/**
 * 中文姓名验证
 */
export function validateChineseName(name: string): boolean {
  const nameRegex = /^[\u4e00-\u9fa5]{2,10}$/
  return nameRegex.test(name)
}

/**
 * Ant Design Form 验证规则生成器
 */
export const formRules = {
  required: (message: string = '此字段为必填项') => ({
    required: true,
    message,
  }),

  email: (message: string = '请输入有效的邮箱地址') => ({
    type: 'email' as const,
    message,
  }),

  phone: (message: string = '请输入有效的手机号') => ({
    pattern: /^1[3-9]\d{9}$/,
    message,
  }),

  url: (message: string = '请输入有效的URL') => ({
    type: 'url' as const,
    message,
  }),

  minLength: (min: number, message?: string) => ({
    min,
    message: message || `至少${min}个字符`,
  }),

  maxLength: (max: number, message?: string) => ({
    max,
    message: message || `最多${max}个字符`,
  }),

  range: (min: number, max: number, message?: string) => ({
    min,
    max,
    message: message || `长度在${min}-${max}之间`,
  }),

  numberRange: (min: number, max: number, message?: string) => ({
    type: 'number' as const,
    min,
    max,
    message: message || `值应在${min}-${max}之间`,
  }),

  password: (message: string = '密码至少8位,包含大小写字母和数字') => ({
    validator: (_: any, value: string) => {
      const result = validatePasswordStrength(value)
      if (result.valid || result.strength === 'medium') {
        return Promise.resolve()
      }
      return Promise.reject(new Error(result.messages.join(', ')))
    },
  }),

  confirmPassword: (passwordField: string = 'password', message: string = '两次输入的密码不一致') => ({
    validator: (_: any, value: string, callback: any) => {
      const form = callback?.form
      if (!value || form?.getFieldValue(passwordField) === value) {
        return Promise.resolve()
      }
      return Promise.reject(new Error(message))
    },
  }),

  username: (message: string = '用户名4-20位,仅支持字母数字下划线') => ({
    pattern: /^[a-zA-Z0-9_]{4,20}$/,
    message,
  }),

  amount: (min: number = 0, max: number = 999999999, message?: string) => ({
    validator: (_: any, value: number) => {
      if (!value || (value >= min && value <= max)) {
        return Promise.resolve()
      }
      return Promise.reject(new Error(message || `金额应在${min}-${max}之间`))
    },
  }),

  bankCard: (message: string = '请输入有效的银行卡号') => ({
    validator: (_: any, value: string) => {
      if (!value || validateBankCard(value)) {
        return Promise.resolve()
      }
      return Promise.reject(new Error(message))
    },
  }),

  idCard: (message: string = '请输入有效的身份证号') => ({
    validator: (_: any, value: string) => {
      if (!value || validateIDCard(value)) {
        return Promise.resolve()
      }
      return Promise.reject(new Error(message))
    },
  }),
}
