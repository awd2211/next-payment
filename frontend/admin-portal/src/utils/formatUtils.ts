/**
 * 格式化工具函数
 * 统一数据格式化逻辑
 */
import dayjs from 'dayjs'

/**
 * 格式化金额 (分转元)
 */
export function formatAmount(amount: number | string, currency: string = 'USD', decimals: number = 2): string {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount
  const formattedValue = (value / 100).toFixed(decimals)

  // 货币符号映射
  const currencySymbols: Record<string, string> = {
    USD: '$',
    EUR: '€',
    GBP: '£',
    JPY: '¥',
    CNY: '¥',
    HKD: 'HK$',
    SGD: 'S$',
  }

  const symbol = currencySymbols[currency.toUpperCase()] || currency
  return `${symbol}${formattedValue}`
}

/**
 * 格式化百分比
 */
export function formatPercentage(value: number, decimals: number = 2): string {
  return `${value.toFixed(decimals)}%`
}

/**
 * 格式化数字 (添加千分位)
 */
export function formatNumber(value: number | string, decimals?: number): string {
  const num = typeof value === 'string' ? parseFloat(value) : value

  if (isNaN(num)) {
    return '0'
  }

  const formatted = decimals !== undefined ? num.toFixed(decimals) : num.toString()
  return formatted.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 格式化日期时间
 */
export function formatDateTime(date: string | Date | dayjs.Dayjs, format: string = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化日期
 */
export function formatDate(date: string | Date | dayjs.Dayjs): string {
  return formatDateTime(date, 'YYYY-MM-DD')
}

/**
 * 格式化时间
 */
export function formatTime(date: string | Date | dayjs.Dayjs): string {
  return formatDateTime(date, 'HH:mm:ss')
}

/**
 * 格式化相对时间 (如: 3分钟前)
 */
export function formatRelativeTime(date: string | Date | dayjs.Dayjs): string {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 格式化手机号 (脱敏)
 */
export function formatPhone(phone: string, mask: boolean = false): string {
  if (!phone) return '-'

  if (mask && phone.length >= 11) {
    return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
  }

  return phone
}

/**
 * 格式化邮箱 (脱敏)
 */
export function formatEmail(email: string, mask: boolean = false): string {
  if (!email) return '-'

  if (mask) {
    const [username, domain] = email.split('@')
    if (username.length <= 2) {
      return email
    }
    const maskedUsername = username[0] + '*'.repeat(username.length - 2) + username[username.length - 1]
    return `${maskedUsername}@${domain}`
  }

  return email
}

/**
 * 格式化身份证号 (脱敏)
 */
export function formatIDCard(idCard: string, mask: boolean = true): string {
  if (!idCard) return '-'

  if (mask && idCard.length >= 18) {
    return idCard.replace(/(\d{6})\d{8}(\d{4})/, '$1********$2')
  }

  return idCard
}

/**
 * 格式化银行卡号 (脱敏)
 */
export function formatBankCard(cardNo: string, mask: boolean = true): string {
  if (!cardNo) return '-'

  if (mask && cardNo.length >= 16) {
    return cardNo.replace(/(\d{4})\d+(\d{4})/, '$1 **** **** $2')
  }

  // 每4位添加空格
  return cardNo.replace(/(\d{4})/g, '$1 ').trim()
}

/**
 * 格式化支付状态
 */
export function formatPaymentStatus(status: string): { text: string; color: string } {
  const statusMap: Record<string, { text: string; color: string }> = {
    pending: { text: '待支付', color: 'orange' },
    processing: { text: '处理中', color: 'blue' },
    success: { text: '成功', color: 'green' },
    failed: { text: '失败', color: 'red' },
    refunded: { text: '已退款', color: 'default' },
    cancelled: { text: '已取消', color: 'default' },
  }

  return statusMap[status] || { text: status, color: 'default' }
}

/**
 * 格式化商户状态
 */
export function formatMerchantStatus(status: string): { text: string; color: string } {
  const statusMap: Record<string, { text: string; color: string }> = {
    pending: { text: '待审核', color: 'orange' },
    active: { text: '正常', color: 'green' },
    suspended: { text: '已冻结', color: 'red' },
    rejected: { text: '已拒绝', color: 'default' },
  }

  return statusMap[status] || { text: status, color: 'default' }
}

/**
 * 格式化 KYC 状态
 */
export function formatKYCStatus(status: string): { text: string; color: string } {
  const statusMap: Record<string, { text: string; color: string }> = {
    pending: { text: '待审核', color: 'orange' },
    verified: { text: '已认证', color: 'green' },
    rejected: { text: '已拒绝', color: 'red' },
  }

  return statusMap[status] || { text: status, color: 'default' }
}

/**
 * 截断文本
 */
export function truncateText(text: string, maxLength: number = 50, ellipsis: string = '...'): string {
  if (!text) return '-'

  if (text.length <= maxLength) {
    return text
  }

  return text.substring(0, maxLength) + ellipsis
}

/**
 * 高亮搜索关键词
 */
export function highlightKeyword(text: string, keyword: string): string {
  if (!keyword) return text

  const regex = new RegExp(`(${keyword})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}
