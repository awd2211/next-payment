/**
 * 格式化工具函数
 */

/**
 * 格式化金额 (分转元，保留2位小数)
 * @param amount 金额（分）
 * @param currency 货币符号
 * @returns 格式化后的金额字符串
 */
export const formatAmount = (amount: number, currency = '¥'): string => {
  if (amount === null || amount === undefined) return `${currency}0.00`
  return `${currency}${(amount / 100).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
}

/**
 * 格式化金额为万元
 * @param amount 金额（分）
 * @returns 格式化后的金额字符串
 */
export const formatAmountInWan = (amount: number): string => {
  if (amount === null || amount === undefined) return '0万'
  return `${(amount / 1000000).toFixed(1)}万`
}

/**
 * 格式化百分比
 * @param value 数值
 * @param decimals 小数位数
 * @returns 格式化后的百分比字符串
 */
export const formatPercent = (value: number, decimals = 1): string => {
  if (value === null || value === undefined) return '0%'
  return `${value.toFixed(decimals)}%`
}

/**
 * 格式化数字（千分位）
 * @param num 数字
 * @returns 格式化后的字符串
 */
export const formatNumber = (num: number): string => {
  if (num === null || num === undefined) return '0'
  return num.toLocaleString('zh-CN')
}

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @returns 格式化后的文件大小
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 隐藏敏感信息（手机号、邮箱等）
 * @param str 原始字符串
 * @param start 保留开始位数
 * @param end 保留结束位数
 * @returns 脱敏后的字符串
 */
export const maskString = (str: string, start = 3, end = 4): string => {
  if (!str || str.length <= start + end) return str
  return `${str.slice(0, start)}${'*'.repeat(str.length - start - end)}${str.slice(-end)}`
}

/**
 * 格式化银行卡号（每4位一个空格）
 * @param cardNo 银行卡号
 * @returns 格式化后的卡号
 */
export const formatBankCard = (cardNo: string): string => {
  if (!cardNo) return ''
  return cardNo.replace(/\s/g, '').replace(/(\d{4})(?=\d)/g, '$1 ')
}

/**
 * 截断字符串
 * @param str 原始字符串
 * @param maxLength 最大长度
 * @returns 截断后的字符串
 */
export const truncate = (str: string, maxLength = 50): string => {
  if (!str || str.length <= maxLength) return str
  return `${str.slice(0, maxLength)}...`
}

/**
 * 首字母大写
 * @param str 字符串
 * @returns 首字母大写的字符串
 */
export const capitalize = (str: string): string => {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}

/**
 * 驼峰转下划线
 * @param str 驼峰字符串
 * @returns 下划线字符串
 */
export const camelToSnake = (str: string): string => {
  return str.replace(/([A-Z])/g, '_$1').toLowerCase()
}

/**
 * 下划线转驼峰
 * @param str 下划线字符串
 * @returns 驼峰字符串
 */
export const snakeToCamel = (str: string): string => {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase())
}





