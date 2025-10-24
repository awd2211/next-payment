import { Space, DatePicker, Button } from 'antd'
import dayjs, { Dayjs } from 'dayjs'
import { useState } from 'react'

const { RangePicker } = DatePicker

interface DateRangeFilterProps {
  onChange?: (dates: [Dayjs | null, Dayjs | null] | null) => void
  defaultValue?: [Dayjs | null, Dayjs | null] | null
  showQuickButtons?: boolean
}

const DateRangeFilter = ({
  onChange,
  defaultValue,
  showQuickButtons = true,
}: DateRangeFilterProps) => {
  const [dates, setDates] = useState<[Dayjs | null, Dayjs | null] | null>(defaultValue || null)

  const handleChange = (newDates: [Dayjs | null, Dayjs | null] | null) => {
    setDates(newDates)
    onChange?.(newDates)
  }

  const setToday = () => {
    const today = dayjs()
    handleChange([today.startOf('day'), today.endOf('day')])
  }

  const setYesterday = () => {
    const yesterday = dayjs().subtract(1, 'day')
    handleChange([yesterday.startOf('day'), yesterday.endOf('day')])
  }

  const setThisWeek = () => {
    const today = dayjs()
    handleChange([today.startOf('week'), today.endOf('week')])
  }

  const setThisMonth = () => {
    const today = dayjs()
    handleChange([today.startOf('month'), today.endOf('month')])
  }

  const setLast7Days = () => {
    const today = dayjs()
    handleChange([today.subtract(7, 'days').startOf('day'), today.endOf('day')])
  }

  const setLast30Days = () => {
    const today = dayjs()
    handleChange([today.subtract(30, 'days').startOf('day'), today.endOf('day')])
  }

  const clearDates = () => {
    handleChange(null)
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <RangePicker
        value={dates}
        onChange={handleChange}
        style={{ width: '100%' }}
        format="YYYY-MM-DD"
      />
      {showQuickButtons && (
        <Space wrap>
          <Button size="small" onClick={setToday}>
            今天
          </Button>
          <Button size="small" onClick={setYesterday}>
            昨天
          </Button>
          <Button size="small" onClick={setThisWeek}>
            本周
          </Button>
          <Button size="small" onClick={setThisMonth}>
            本月
          </Button>
          <Button size="small" onClick={setLast7Days}>
            最近7天
          </Button>
          <Button size="small" onClick={setLast30Days}>
            最近30天
          </Button>
          <Button size="small" onClick={clearDates}>
            清空
          </Button>
        </Space>
      )}
    </Space>
  )
}

export default DateRangeFilter
