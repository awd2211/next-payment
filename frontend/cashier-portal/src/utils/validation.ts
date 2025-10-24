// 验证信用卡号（Luhn算法）
export const validateCardNumber = (cardNumber: string): boolean => {
  const cleaned = cardNumber.replace(/\s+/g, '')
  if (!/^\d{13,19}$/.test(cleaned)) return false

  let sum = 0
  let isEven = false

  for (let i = cleaned.length - 1; i >= 0; i--) {
    let digit = parseInt(cleaned.charAt(i), 10)

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

// 验证有效期
export const validateExpiryDate = (expiryDate: string): boolean => {
  const match = expiryDate.match(/^(\d{2})\/(\d{2})$/)
  if (!match) return false

  const month = parseInt(match[1], 10)
  const year = parseInt('20' + match[2], 10)

  if (month < 1 || month > 12) return false

  const now = new Date()
  const expiry = new Date(year, month - 1)

  return expiry > now
}

// 验证CVV
export const validateCVV = (cvv: string): boolean => {
  return /^\d{3,4}$/.test(cvv)
}

// 格式化卡号（每4位加空格）
export const formatCardNumber = (value: string): string => {
  const cleaned = value.replace(/\s+/g, '')
  const match = cleaned.match(/.{1,4}/g)
  return match ? match.join(' ') : cleaned
}

// 格式化有效期
export const formatExpiryDate = (value: string): string => {
  const cleaned = value.replace(/\D/g, '')
  if (cleaned.length >= 2) {
    return cleaned.slice(0, 2) + '/' + cleaned.slice(2, 4)
  }
  return cleaned
}

// 格式化金额
export const formatAmount = (amount: number, currency: string): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
  }).format(amount / 100)
}

// 获取卡类型
export const getCardType = (cardNumber: string): string => {
  const cleaned = cardNumber.replace(/\s+/g, '')

  if (/^4/.test(cleaned)) return 'visa'
  if (/^5[1-5]/.test(cleaned)) return 'mastercard'
  if (/^3[47]/.test(cleaned)) return 'amex'
  if (/^6(?:011|5)/.test(cleaned)) return 'discover'

  return 'unknown'
}
