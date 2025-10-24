import { Input } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { useState } from 'react'

const { Search } = Input

interface SearchInputProps {
  placeholder?: string
  onSearch: (value: string) => void
  defaultValue?: string
  allowClear?: boolean
  enterButton?: boolean | string
  loading?: boolean
  maxLength?: number
  size?: 'small' | 'middle' | 'large'
}

const SearchInput = ({
  placeholder = '请输入搜索关键词',
  onSearch,
  defaultValue,
  allowClear = true,
  enterButton = true,
  loading,
  maxLength,
  size = 'middle',
}: SearchInputProps) => {
  const [value, setValue] = useState(defaultValue || '')

  const handleSearch = (searchValue: string) => {
    onSearch(searchValue.trim())
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setValue(e.target.value)
  }

  return (
    <Search
      placeholder={placeholder}
      value={value}
      onChange={handleChange}
      onSearch={handleSearch}
      allowClear={allowClear}
      enterButton={enterButton}
      loading={loading}
      maxLength={maxLength}
      size={size}
      prefix={<SearchOutlined />}
      style={{ width: '100%' }}
    />
  )
}

export default SearchInput
