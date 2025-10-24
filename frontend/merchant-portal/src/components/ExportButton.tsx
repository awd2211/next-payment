import { Button, Dropdown, message } from 'antd'
import type { MenuProps } from 'antd'
import { DownloadOutlined, FileExcelOutlined, FilePdfOutlined, FileTextOutlined } from '@ant-design/icons'
import { useState } from 'react'

interface ExportButtonProps {
  data: any[]
  filename?: string
  onExport?: (format: 'csv' | 'excel' | 'pdf') => Promise<void> | void
  loading?: boolean
}

const ExportButton = ({ data, filename = 'export', onExport, loading: externalLoading }: ExportButtonProps) => {
  const [loading, setLoading] = useState(false)

  const exportToCSV = () => {
    if (data.length === 0) {
      message.warning('没有数据可导出')
      return
    }

    // 转换为CSV
    const headers = Object.keys(data[0])
    const csvContent = [
      headers.join(','),
      ...data.map(row => headers.map(header => JSON.stringify(row[header] || '')).join(',')),
    ].join('\n')

    // 下载文件
    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `${filename}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    message.success('导出成功')
  }

  const handleExport = async (format: 'csv' | 'excel' | 'pdf') => {
    if (onExport) {
      setLoading(true)
      try {
        await onExport(format)
        message.success(`导出${format.toUpperCase()}成功`)
      } catch (error) {
        message.error('导出失败')
      } finally {
        setLoading(false)
      }
    } else {
      // 默认导出CSV
      if (format === 'csv') {
        exportToCSV()
      } else {
        message.info(`${format.toUpperCase()}导出功能开发中`)
      }
    }
  }

  const items: MenuProps['items'] = [
    {
      key: 'csv',
      label: 'CSV 格式',
      icon: <FileTextOutlined />,
      onClick: () => handleExport('csv'),
    },
    {
      key: 'excel',
      label: 'Excel 格式',
      icon: <FileExcelOutlined />,
      onClick: () => handleExport('excel'),
    },
    {
      key: 'pdf',
      label: 'PDF 格式',
      icon: <FilePdfOutlined />,
      onClick: () => handleExport('pdf'),
    },
  ]

  return (
    <Dropdown menu={{ items }} placement="bottomRight">
      <Button
        icon={<DownloadOutlined />}
        loading={loading || externalLoading}
      >
        导出
      </Button>
    </Dropdown>
  )
}

export default ExportButton
