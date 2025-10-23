import { describe, it, expect } from 'vitest'
import { isEmail, isPhone, isBankCard, isIDCard, isURL } from './validate'

describe('isEmail', () => {
  it('应该验证有效邮箱', () => {
    expect(isEmail('user@example.com')).toBe(true)
    expect(isEmail('test.user+tag@domain.co.uk')).toBe(true)
  })

  it('应该拒绝无效邮箱', () => {
    expect(isEmail('invalid')).toBe(false)
    expect(isEmail('user@')).toBe(false)
    expect(isEmail('@example.com')).toBe(false)
    expect(isEmail('')).toBe(false)
  })
})

describe('isPhone', () => {
  it('应该验证有效手机号', () => {
    expect(isPhone('13812345678')).toBe(true)
    expect(isPhone('15987654321')).toBe(true)
  })

  it('应该拒绝无效手机号', () => {
    expect(isPhone('12345678901')).toBe(false)
    expect(isPhone('1381234567')).toBe(false)
    expect(isPhone('')).toBe(false)
  })
})

describe('isBankCard', () => {
  it('应该验证有效银行卡号（Luhn算法）', () => {
    expect(isBankCard('6222021234567890123')).toBe(true)
    // 添加更多真实的测试卡号
  })

  it('应该拒绝无效银行卡号', () => {
    expect(isBankCard('1234567890')).toBe(false)
    expect(isBankCard('')).toBe(false)
  })
})

describe('isIDCard', () => {
  it('应该验证有效身份证号', () => {
    expect(isIDCard('110101199003071234')).toBe(true)
  })

  it('应该拒绝无效身份证号', () => {
    expect(isIDCard('12345678901234567')).toBe(false)
    expect(isIDCard('')).toBe(false)
  })
})

describe('isURL', () => {
  it('应该验证有效URL', () => {
    expect(isURL('https://www.example.com')).toBe(true)
    expect(isURL('http://localhost:3000')).toBe(true)
  })

  it('应该拒绝无效URL', () => {
    expect(isURL('not-a-url')).toBe(false)
    expect(isURL('')).toBe(false)
  })
})
