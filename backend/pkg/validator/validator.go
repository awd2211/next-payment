package validator

// Validator 验证器接口
type Validator struct {
	Amount   *AmountValidator
	Currency *CurrencyValidator
	String   *StringValidator
}

// New 创建验证器实例
func New() *Validator {
	return &Validator{
		Amount:   NewAmountValidator(),
		Currency: NewCurrencyValidator(),
		String:   NewStringValidator(),
	}
}

// 全局验证器实例（可选）
var Default = New()

// 便捷函数

// ValidateEmail 验证邮箱
func ValidateEmail(email string) error {
	return Default.String.ValidateEmail(email)
}

// ValidatePhone 验证手机号
func ValidatePhone(phone string) error {
	return Default.String.ValidatePhone(phone)
}

// ValidateURL 验证URL
func ValidateURL(urlStr string) error {
	return Default.String.ValidateURL(urlStr)
}

// ValidateAmount 验证金额
func ValidateAmount(amount float64) error {
	return Default.Amount.Validate(amount)
}

// ValidateCurrency 验证货币代码
func ValidateCurrency(currency string) error {
	return Default.Currency.Validate(currency)
}

// ValidateCreditCard 验证信用卡号
func ValidateCreditCard(cardNumber string) error {
	return Default.String.ValidateCreditCard(cardNumber)
}
