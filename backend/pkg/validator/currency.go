package validator

import "errors"

// CurrencyValidator 货币验证器
type CurrencyValidator struct {
	validCurrencies map[string]bool
}

// NewCurrencyValidator 创建货币验证器
func NewCurrencyValidator() *CurrencyValidator {
	return &CurrencyValidator{
		validCurrencies: supportedCurrencies(),
	}
}

// Validate 验证货币代码
func (v *CurrencyValidator) Validate(currency string) error {
	if currency == "" {
		return errors.New("货币代码不能为空")
	}
	if len(currency) != 3 {
		return errors.New("货币代码必须是3个字符")
	}
	if !v.validCurrencies[currency] {
		return errors.New("不支持的货币代码")
	}
	return nil
}

// IsSupported 检查货币是否支持
func (v *CurrencyValidator) IsSupported(currency string) bool {
	return v.validCurrencies[currency]
}

// GetSupportedCurrencies 获取支持的货币列表
func (v *CurrencyValidator) GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(v.validCurrencies))
	for currency := range v.validCurrencies {
		currencies = append(currencies, currency)
	}
	return currencies
}

// supportedCurrencies 返回支持的货币代码（ISO 4217）
func supportedCurrencies() map[string]bool {
	return map[string]bool{
		// 主要货币
		"USD": true, // 美元
		"EUR": true, // 欧元
		"GBP": true, // 英镑
		"CNY": true, // 人民币
		"JPY": true, // 日元
		"KRW": true, // 韩元
		"HKD": true, // 港币
		"TWD": true, // 新台币
		"SGD": true, // 新加坡元
		"AUD": true, // 澳元
		"CAD": true, // 加元
		"CHF": true, // 瑞士法郎

		// 亚洲货币
		"INR": true, // 印度卢比
		"THB": true, // 泰铢
		"MYR": true, // 马来西亚林吉特
		"PHP": true, // 菲律宾比索
		"IDR": true, // 印尼盾
		"VND": true, // 越南盾

		// 其他主要货币
		"RUB": true, // 俄罗斯卢布
		"BRL": true, // 巴西雷亚尔
		"MXN": true, // 墨西哥比索
		"ZAR": true, // 南非兰特
		"SEK": true, // 瑞典克朗
		"NOK": true, // 挪威克朗
		"DKK": true, // 丹麦克朗
		"PLN": true, // 波兰兹罗提
		"TRY": true, // 土耳其里拉
		"AED": true, // 阿联酋迪拉姆
		"SAR": true, // 沙特里亚尔
		"NZD": true, // 新西兰元

		// 加密货币（如果支持）
		"BTC": true, // 比特币
		"ETH": true, // 以太坊
		"USDT": true, // 泰达币
		"USDC": true, // USD Coin
	}
}
