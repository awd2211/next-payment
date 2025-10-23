/**
 * 数据验证工具函数
 */

/**
 * 验证邮箱
 */
export const isEmail = (email: string): boolean => {
  const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return regex.test(email)
}

/**
 * 验证手机号（中国）
 */
export const isPhone = (phone: string): boolean => {
  const regex = /^1[3-9]\d{9}$/
  return regex.test(phone)
}

/**
 * 验证URL
 */
export const isUrl = (url: string): boolean => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * 验证IP地址
 */
export const isIP = (ip: string): boolean => {
  const regex = /^(\d{1,3}\.){3}\d{1,3}$/
  if (!regex.test(ip)) return false
  return ip.split('.').every(part => {
    const num = parseInt(part, 10)
    return num >= 0 && num <= 255
  })
}

/**
 * 验证密码强度
 * @param password 密码
 * @returns 强度等级 0-4
 */
export const getPasswordStrength = (password: string): number => {
  if (!password) return 0
  
  let strength = 0
  
  // 长度
  if (password.length >= 8) strength++
  if (password.length >= 12) strength++
  
  // 包含小写字母
  if (/[a-z]/.test(password)) strength++
  
  // 包含大写字母
  if (/[A-Z]/.test(password)) strength++
  
  // 包含数字
  if (/[0-9]/.test(password)) strength++
  
  // 包含特殊字符
  if (/[^a-zA-Z0-9]/.test(password)) strength++
  
  // 返回0-4的强度等级
  return Math.min(strength, 4)
}

/**
 * 验证银行卡号（Luhn算法）
 */
export const isValidBankCard = (cardNo: string): boolean => {
  const cleanCardNo = cardNo.replace(/\s/g, '')
  if (!/^\d{13,19}$/.test(cleanCardNo)) return false
  
  let sum = 0
  let isEven = false
  
  for (let i = cleanCardNo.length - 1; i >= 0; i--) {
    let digit = parseInt(cleanCardNo.charAt(i), 10)
    
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
 * 验证身份证号（中国）
 */
export const isValidIdCard = (idCard: string): boolean => {
  const regex = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/
  if (!regex.test(idCard)) return false
  
  // 验证校验码
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']
  
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(idCard.charAt(i), 10) * weights[i]
  }
  
  const checkCode = checkCodes[sum % 11]
  return checkCode === idCard.charAt(17).toUpperCase()
}

/**
 * 验证金额
 * @param amount 金额
 * @param min 最小值
 * @param max 最大值
 */
export const isValidAmount = (
  amount: number,
  min = 0.01,
  max = 999999999.99
): boolean => {
  return amount >= min && amount <= max
}

/**
 * 验证货币代码
 */
export const isValidCurrency = (currency: string): boolean => {
  const validCurrencies = [
    'USD', 'EUR', 'GBP', 'JPY', 'CNY', 'HKD', 'AUD', 'CAD', 'SGD', 'KRW',
    'BTC', 'ETH', 'USDT'
  ]
  return validCurrencies.includes(currency.toUpperCase())
}

/**
 * 验证用户名（字母开头，字母数字下划线，4-20位）
 */
export const isValidUsername = (username: string): boolean => {
  const regex = /^[a-zA-Z][a-zA-Z0-9_]{3,19}$/
  return regex.test(username)
}

/**
 * 验证是否为空
 */
export const isEmpty = (value: any): boolean => {
  if (value === null || value === undefined) return true
  if (typeof value === 'string') return value.trim() === ''
  if (Array.isArray(value)) return value.length === 0
  if (typeof value === 'object') return Object.keys(value).length === 0
  return false
}

