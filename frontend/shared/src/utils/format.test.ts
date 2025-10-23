import { describe, it, expect } from 'vitest'
import { formatAmount, formatNumber, formatDate, formatPhone, formatBankCard } from './format'

describe('formatAmount', () => {
  it('应该正确格式化金额（分转元）', () => {
    expect(formatAmount(12345)).toBe('¥123.45')
    expect(formatAmount(100)).toBe('¥1.00')
    expect(formatAmount(0)).toBe('¥0.00')
  })

  it('应该处理负数', () => {
    expect(formatAmount(-12345)).toBe('-¥123.45')
  })

  it('应该处理不同货币', () => {
    expect(formatAmount(12345, 'USD')).toBe('$123.45')
    expect(formatAmount(12345, 'EUR')).toBe('€123.45')
  })
})

describe('formatNumber', () => {
  it('应该正确添加千分位分隔符', () => {
    expect(formatNumber(1234567)).toBe('1,234,567')
    expect(formatNumber(123)).toBe('123')
    expect(formatNumber(0)).toBe('0')
  })

  it('应该处理小数', () => {
    expect(formatNumber(1234.56)).toBe('1,234.56')
  })
})

describe('formatDate', () => {
  it('应该正确格式化日期', () => {
    const date = new Date('2024-03-15T10:30:00')
    expect(formatDate(date, 'YYYY-MM-DD')).toBe('2024-03-15')
    expect(formatDate(date, 'YYYY-MM-DD HH:mm:ss')).toContain('2024-03-15 10:30:00')
  })

  it('应该处理字符串日期', () => {
    expect(formatDate('2024-03-15', 'YYYY-MM-DD')).toBe('2024-03-15')
  })
})

describe('formatPhone', () => {
  it('应该正确脱敏手机号', () => {
    expect(formatPhone('13812345678')).toBe('138****5678')
  })

  it('应该处理无效输入', () => {
    expect(formatPhone('')).toBe('')
    expect(formatPhone('123')).toBe('123')
  })
})

describe('formatBankCard', () => {
  it('应该正确脱敏银行卡号', () => {
    expect(formatBankCard('6222021234567890123')).toBe('6222 **** **** **** 0123')
  })

  it('应该处理无效输入', () => {
    expect(formatBankCard('')).toBe('')
    expect(formatBankCard('123')).toBe('123')
  })
})
