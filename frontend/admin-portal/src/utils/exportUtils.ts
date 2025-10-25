/**
 * 数据导出工具
 * 支持导出为 CSV 和 Excel 格式
 */
import * as XLSX from 'xlsx'
import { message } from 'antd'

/**
 * 导出为 CSV
 */
export function exportToCSV<T = any>(
  data: T[],
  columns: { title: string; dataIndex: string; render?: (value: any, record: T) => any }[],
  filename: string = 'export.csv'
) {
  try {
    // 创建 CSV 头部
    const headers = columns.map((col) => col.title).join(',')

    // 创建 CSV 内容
    const rows = data.map((record) => {
      return columns
        .map((col) => {
          const value = record[col.dataIndex as keyof T]
          // 如果有 render 函数,使用 render 后的值
          const displayValue = col.render ? col.render(value, record) : value
          // 处理特殊字符和逗号
          const stringValue = String(displayValue || '')
          return stringValue.includes(',') || stringValue.includes('"')
            ? `"${stringValue.replace(/"/g, '""')}"`
            : stringValue
        })
        .join(',')
    })

    // 组合 CSV 内容
    const csvContent = [headers, ...rows].join('\n')

    // 添加 BOM 以支持中文
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' })

    // 下载文件
    downloadBlob(blob, filename)
    message.success('导出成功')
  } catch (error) {
    console.error('Export to CSV failed:', error)
    message.error('导出失败')
  }
}

/**
 * 导出为 Excel
 */
export function exportToExcel<T = any>(
  data: T[],
  columns: { title: string; dataIndex: string; render?: (value: any, record: T) => any }[],
  filename: string = 'export.xlsx',
  sheetName: string = 'Sheet1'
) {
  try {
    // 准备表头
    const headers = columns.map((col) => col.title)

    // 准备数据
    const rows = data.map((record) => {
      return columns.map((col) => {
        const value = record[col.dataIndex as keyof T]
        // 如果有 render 函数,使用 render 后的值
        const displayValue = col.render ? col.render(value, record) : value
        return displayValue
      })
    })

    // 组合数据
    const sheetData = [headers, ...rows]

    // 创建工作簿
    const wb = XLSX.utils.book_new()
    const ws = XLSX.utils.aoa_to_sheet(sheetData)

    // 设置列宽
    const colWidths = columns.map(() => ({ wch: 15 }))
    ws['!cols'] = colWidths

    // 添加工作表
    XLSX.utils.book_append_sheet(wb, ws, sheetName)

    // 导出文件
    XLSX.writeFile(wb, filename)
    message.success('导出成功')
  } catch (error) {
    console.error('Export to Excel failed:', error)
    message.error('导出失败')
  }
}

/**
 * 下载 Blob 文件
 */
function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

/**
 * 导出 JSON 数据
 */
export function exportToJSON<T = any>(data: T[], filename: string = 'export.json') {
  try {
    const jsonContent = JSON.stringify(data, null, 2)
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' })
    downloadBlob(blob, filename)
    message.success('导出成功')
  } catch (error) {
    console.error('Export to JSON failed:', error)
    message.error('导出失败')
  }
}

/**
 * 批量导出多个工作表到一个 Excel 文件
 */
export function exportMultipleSheets(
  sheets: Array<{
    name: string
    data: any[]
    columns: { title: string; dataIndex: string; render?: (value: any, record: any) => any }[]
  }>,
  filename: string = 'export.xlsx'
) {
  try {
    const wb = XLSX.utils.book_new()

    sheets.forEach((sheet) => {
      // 准备表头
      const headers = sheet.columns.map((col) => col.title)

      // 准备数据
      const rows = sheet.data.map((record) => {
        return sheet.columns.map((col) => {
          const value = record[col.dataIndex]
          const displayValue = col.render ? col.render(value, record) : value
          return displayValue
        })
      })

      // 组合数据
      const sheetData = [headers, ...rows]

      // 创建工作表
      const ws = XLSX.utils.aoa_to_sheet(sheetData)

      // 设置列宽
      const colWidths = sheet.columns.map(() => ({ wch: 15 }))
      ws['!cols'] = colWidths

      // 添加工作表
      XLSX.utils.book_append_sheet(wb, ws, sheet.name)
    })

    // 导出文件
    XLSX.writeFile(wb, filename)
    message.success('导出成功')
  } catch (error) {
    console.error('Export multiple sheets failed:', error)
    message.error('导出失败')
  }
}
