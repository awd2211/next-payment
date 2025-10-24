import { useState, useCallback } from 'react'

/**
 * 表单字段验证规则
 */
interface ValidationRule {
  required?: boolean
  min?: number
  max?: number
  pattern?: RegExp
  validator?: (value: any) => boolean | Promise<boolean>
  message?: string
}

/**
 * 表单配置
 */
interface FormConfig<T> {
  initialValues: T
  validation?: {
    [K in keyof T]?: ValidationRule[]
  }
  onSubmit?: (values: T) => void | Promise<void>
}

/**
 * 表单状态
 */
interface FormState<T> {
  values: T
  errors: Partial<Record<keyof T, string>>
  touched: Partial<Record<keyof T, boolean>>
  isSubmitting: boolean
  isValidating: boolean
}

/**
 * 表单操作
 */
interface FormActions<T> {
  setFieldValue: (field: keyof T, value: any) => void
  setFieldError: (field: keyof T, error: string) => void
  setFieldTouched: (field: keyof T, touched: boolean) => void
  validateField: (field: keyof T) => Promise<boolean>
  validateForm: () => Promise<boolean>
  handleSubmit: (e?: React.FormEvent) => Promise<void>
  resetForm: () => void
  setValues: (values: Partial<T>) => void
}

/**
 * 表单Hook - 简化表单处理
 *
 * @example
 * const [formState, formActions] = useForm({
 *   initialValues: { email: '', password: '' },
 *   validation: {
 *     email: [
 *       { required: true, message: '请输入邮箱' },
 *       { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: '邮箱格式不正确' }
 *     ],
 *     password: [
 *       { required: true, message: '请输入密码' },
 *       { min: 8, message: '密码至少8位' }
 *     ]
 *   },
 *   onSubmit: async (values) => {
 *     await login(values)
 *   }
 * })
 */
function useForm<T extends Record<string, any>>(
  config: FormConfig<T>
): [FormState<T>, FormActions<T>] {
  const [values, setValues] = useState<T>(config.initialValues)
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({})
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isValidating, setIsValidating] = useState(false)

  /**
   * 验证单个字段
   */
  const validateField = useCallback(
    async (field: keyof T): Promise<boolean> => {
      const value = values[field]
      const rules = config.validation?.[field]

      if (!rules || rules.length === 0) {
        return true
      }

      for (const rule of rules) {
        // 必填验证
        if (rule.required && !value) {
          setErrors((prev) => ({
            ...prev,
            [field]: rule.message || '此字段为必填项',
          }))
          return false
        }

        // 跳过空值的其他验证
        if (!value) {
          continue
        }

        // 最小长度验证
        if (rule.min !== undefined && String(value).length < rule.min) {
          setErrors((prev) => ({
            ...prev,
            [field]: rule.message || `最少${rule.min}个字符`,
          }))
          return false
        }

        // 最大长度验证
        if (rule.max !== undefined && String(value).length > rule.max) {
          setErrors((prev) => ({
            ...prev,
            [field]: rule.message || `最多${rule.max}个字符`,
          }))
          return false
        }

        // 正则验证
        if (rule.pattern && !rule.pattern.test(String(value))) {
          setErrors((prev) => ({
            ...prev,
            [field]: rule.message || '格式不正确',
          }))
          return false
        }

        // 自定义验证
        if (rule.validator) {
          const isValid = await rule.validator(value)
          if (!isValid) {
            setErrors((prev) => ({
              ...prev,
              [field]: rule.message || '验证失败',
            }))
            return false
          }
        }
      }

      // 验证通过,清除错误
      setErrors((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
      return true
    },
    [values, config.validation]
  )

  /**
   * 验证整个表单
   */
  const validateForm = useCallback(async (): Promise<boolean> => {
    setIsValidating(true)
    const fields = Object.keys(values) as Array<keyof T>
    const validationResults = await Promise.all(
      fields.map((field) => validateField(field))
    )
    setIsValidating(false)
    return validationResults.every((result) => result)
  }, [values, validateField])

  /**
   * 设置字段值
   */
  const setFieldValue = useCallback((field: keyof T, value: any) => {
    setValues((prev) => ({ ...prev, [field]: value }))
  }, [])

  /**
   * 设置字段错误
   */
  const setFieldError = useCallback((field: keyof T, error: string) => {
    setErrors((prev) => ({ ...prev, [field]: error }))
  }, [])

  /**
   * 设置字段触摸状态
   */
  const setFieldTouched = useCallback((field: keyof T, isTouched: boolean) => {
    setTouched((prev) => ({ ...prev, [field]: isTouched }))
  }, [])

  /**
   * 处理表单提交
   */
  const handleSubmit = useCallback(
    async (e?: React.FormEvent) => {
      if (e) {
        e.preventDefault()
      }

      // 标记所有字段为已触摸
      const allTouched = Object.keys(values).reduce(
        (acc, key) => ({ ...acc, [key]: true }),
        {} as Record<keyof T, boolean>
      )
      setTouched(allTouched)

      // 验证表单
      const isValid = await validateForm()
      if (!isValid) {
        return
      }

      // 执行提交
      if (config.onSubmit) {
        setIsSubmitting(true)
        try {
          await config.onSubmit(values)
        } finally {
          setIsSubmitting(false)
        }
      }
    },
    [values, validateForm, config]
  )

  /**
   * 重置表单
   */
  const resetForm = useCallback(() => {
    setValues(config.initialValues)
    setErrors({})
    setTouched({})
    setIsSubmitting(false)
    setIsValidating(false)
  }, [config.initialValues])

  /**
   * 批量设置值
   */
  const setFormValues = useCallback((newValues: Partial<T>) => {
    setValues((prev) => ({ ...prev, ...newValues }))
  }, [])

  return [
    {
      values,
      errors,
      touched,
      isSubmitting,
      isValidating,
    },
    {
      setFieldValue,
      setFieldError,
      setFieldTouched,
      validateField,
      validateForm,
      handleSubmit,
      resetForm,
      setValues: setFormValues,
    },
  ]
}

export default useForm
