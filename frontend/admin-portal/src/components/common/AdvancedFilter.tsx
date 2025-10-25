/**
 * 高级筛选器组件
 * 支持多种筛选条件和组合
 */
import { useState } from 'react'
import { Card, Form, Row, Col, Input, Select, DatePicker, Button, Space, Collapse } from 'antd'
import { SearchOutlined, ReloadOutlined, DownOutlined, UpOutlined } from '@ant-design/icons'
import type { FormInstance } from 'antd'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

export type FilterFieldType = 'input' | 'select' | 'dateRange' | 'date' | 'number' | 'custom'

export interface FilterField {
  /**
   * 字段名
   */
  name: string

  /**
   * 字段标签
   */
  label: string

  /**
   * 字段类型
   */
  type: FilterFieldType

  /**
   * 占位符
   */
  placeholder?: string

  /**
   * 选项 (用于 select)
   */
  options?: Array<{ label: string; value: any }>

  /**
   * 默认值
   */
  defaultValue?: any

  /**
   * 自定义渲染 (用于 custom 类型)
   */
  render?: (form: FormInstance) => React.ReactNode

  /**
   * 是否可清空
   */
  allowClear?: boolean

  /**
   * 列宽
   */
  span?: number
}

export interface AdvancedFilterProps {
  /**
   * 筛选字段配置
   */
  fields: FilterField[]

  /**
   * 搜索回调
   */
  onSearch: (values: Record<string, any>) => void

  /**
   * 重置回调
   */
  onReset?: () => void

  /**
   * 是否默认展开
   */
  defaultExpanded?: boolean

  /**
   * 是否显示展开/收起按钮
   */
  showExpandButton?: boolean

  /**
   * 收起时显示的字段数量
   */
  collapsedRowCount?: number

  /**
   * 表单实例 (可选,用于外部控制)
   */
  form?: FormInstance

  /**
   * 每行显示的字段数量
   */
  colCount?: number
}

const AdvancedFilter: React.FC<AdvancedFilterProps> = ({
  fields,
  onSearch,
  onReset,
  defaultExpanded = false,
  showExpandButton = true,
  collapsedRowCount = 1,
  form: externalForm,
  colCount = 4,
}) => {
  const [form] = Form.useForm(externalForm)
  const [expanded, setExpanded] = useState(defaultExpanded)
  const [loading, setLoading] = useState(false)

  /**
   * 处理搜索
   */
  const handleSearch = async () => {
    setLoading(true)
    try {
      const values = await form.validateFields()

      // 处理日期范围
      Object.keys(values).forEach((key) => {
        const field = fields.find((f) => f.name === key)
        if (field?.type === 'dateRange' && values[key]) {
          const [start, end] = values[key]
          values[`${key}_start`] = start ? dayjs(start).startOf('day').toISOString() : undefined
          values[`${key}_end`] = end ? dayjs(end).endOf('day').toISOString() : undefined
          delete values[key]
        } else if (field?.type === 'date' && values[key]) {
          values[key] = dayjs(values[key]).toISOString()
        }
      })

      // 过滤空值
      const filteredValues = Object.fromEntries(
        Object.entries(values).filter(([_, v]) => v !== undefined && v !== null && v !== '')
      )

      onSearch(filteredValues)
    } catch (error) {
      console.error('Validation failed:', error)
    } finally {
      setLoading(false)
    }
  }

  /**
   * 处理重置
   */
  const handleReset = () => {
    form.resetFields()
    onReset?.()
    onSearch({})
  }

  /**
   * 渲染表单项
   */
  const renderFormItem = (field: FilterField) => {
    const commonProps = {
      placeholder: field.placeholder || `请输入${field.label}`,
      allowClear: field.allowClear !== false,
    }

    switch (field.type) {
      case 'input':
        return <Input {...commonProps} />

      case 'number':
        return <Input type="number" {...commonProps} />

      case 'select':
        return (
          <Select {...commonProps} options={field.options} placeholder={field.placeholder || `请选择${field.label}`} />
        )

      case 'dateRange':
        return <RangePicker style={{ width: '100%' }} />

      case 'date':
        return <DatePicker style={{ width: '100%' }} />

      case 'custom':
        return field.render?.(form)

      default:
        return <Input {...commonProps} />
    }
  }

  // 计算显示的字段
  const span = 24 / colCount
  const visibleFields = expanded ? fields : fields.slice(0, collapsedRowCount * colCount)

  return (
    <Card bodyStyle={{ paddingBottom: 0 }}>
      <Form form={form} layout="vertical" onFinish={handleSearch}>
        <Row gutter={16}>
          {visibleFields.map((field) => (
            <Col span={field.span || span} key={field.name}>
              <Form.Item label={field.label} name={field.name} initialValue={field.defaultValue}>
                {renderFormItem(field)}
              </Form.Item>
            </Col>
          ))}

          {/* 操作按钮 */}
          <Col span={span} style={{ display: 'flex', alignItems: 'flex-end' }}>
            <Form.Item style={{ marginBottom: 24, width: '100%' }}>
              <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
                <Button type="primary" htmlType="submit" icon={<SearchOutlined />} loading={loading}>
                  搜索
                </Button>
                <Button icon={<ReloadOutlined />} onClick={handleReset}>
                  重置
                </Button>
                {showExpandButton && fields.length > collapsedRowCount * colCount && (
                  <Button type="link" onClick={() => setExpanded(!expanded)}>
                    {expanded ? (
                      <>
                        收起 <UpOutlined />
                      </>
                    ) : (
                      <>
                        展开 <DownOutlined />
                      </>
                    )}
                  </Button>
                )}
              </Space>
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Card>
  )
}

export default AdvancedFilter
