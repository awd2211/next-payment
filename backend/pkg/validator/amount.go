package validator

import (
	"errors"
	"math"
)

// AmountValidator 金额验证器
type AmountValidator struct{}

// NewAmountValidator 创建金额验证器
func NewAmountValidator() *AmountValidator {
	return &AmountValidator{}
}

// Validate 验证金额
func (v *AmountValidator) Validate(amount float64) error {
	if amount < 0 {
		return errors.New("金额不能为负数")
	}
	if amount == 0 {
		return errors.New("金额不能为零")
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return errors.New("金额必须是有效数字")
	}
	return nil
}

// ValidatePositive 验证金额为正数（可以为0）
func (v *AmountValidator) ValidatePositive(amount float64) error {
	if amount < 0 {
		return errors.New("金额不能为负数")
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return errors.New("金额必须是有效数字")
	}
	return nil
}

// ValidateRange 验证金额在指定范围内
func (v *AmountValidator) ValidateRange(amount, min, max float64) error {
	if err := v.Validate(amount); err != nil {
		return err
	}
	if amount < min {
		return errors.New("金额低于最小值")
	}
	if amount > max {
		return errors.New("金额超过最大值")
	}
	return nil
}

// ValidatePrecision 验证金额精度
func (v *AmountValidator) ValidatePrecision(amount float64, decimalPlaces int) error {
	multiplier := math.Pow(10, float64(decimalPlaces))
	rounded := math.Round(amount*multiplier) / multiplier
	if amount != rounded {
		return errors.New("金额精度超过允许的小数位数")
	}
	return nil
}

// FormatAmount 格式化金额为指定精度
func (v *AmountValidator) FormatAmount(amount float64, decimalPlaces int) float64 {
	multiplier := math.Pow(10, float64(decimalPlaces))
	return math.Round(amount*multiplier) / multiplier
}

// ValidateAmountForCurrency 验证金额是否符合货币规则
func (v *AmountValidator) ValidateAmountForCurrency(amount float64, currency string) error {
	if err := v.Validate(amount); err != nil {
		return err
	}

	// 根据不同货币验证精度
	switch currency {
	case "JPY", "KRW", "VND": // 无小数位
		if err := v.ValidatePrecision(amount, 0); err != nil {
			return errors.New("该货币不支持小数位")
		}
	case "BHD", "JOD", "KWD", "OMR", "TND": // 3位小数
		if err := v.ValidatePrecision(amount, 3); err != nil {
			return errors.New("该货币最多支持3位小数")
		}
	default: // 大多数货币2位小数
		if err := v.ValidatePrecision(amount, 2); err != nil {
			return errors.New("该货币最多支持2位小数")
		}
	}

	return nil
}
