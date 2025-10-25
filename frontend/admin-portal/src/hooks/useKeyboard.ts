/**
 * 键盘快捷键 Hook
 * 支持组合键和全局快捷键
 */
import { useEffect, useCallback, useRef } from 'react'

export type KeyModifier = 'ctrl' | 'shift' | 'alt' | 'meta'

export interface KeyboardShortcut {
  /**
   * 主键
   */
  key: string

  /**
   * 修饰键
   */
  modifiers?: KeyModifier[]

  /**
   * 回调函数
   */
  callback: (event: KeyboardEvent) => void

  /**
   * 描述
   */
  description?: string

  /**
   * 是否阻止默认行为
   */
  preventDefault?: boolean

  /**
   * 是否停止事件传播
   */
  stopPropagation?: boolean

  /**
   * 是否禁用
   */
  disabled?: boolean
}

/**
 * 单个快捷键 Hook
 */
export function useKeyPress(
  targetKey: string,
  callback: (event: KeyboardEvent) => void,
  options: {
    modifiers?: KeyModifier[]
    preventDefault?: boolean
    stopPropagation?: boolean
    disabled?: boolean
  } = {}
) {
  const { modifiers = [], preventDefault = true, stopPropagation = false, disabled = false } = options

  const callbackRef = useRef(callback)
  callbackRef.current = callback

  useEffect(() => {
    if (disabled) return

    const handleKeyDown = (event: KeyboardEvent) => {
      // 检查修饰键
      const hasCtrl = modifiers.includes('ctrl') ? event.ctrlKey : !event.ctrlKey
      const hasShift = modifiers.includes('shift') ? event.shiftKey : !event.shiftKey
      const hasAlt = modifiers.includes('alt') ? event.altKey : !event.altKey
      const hasMeta = modifiers.includes('meta') ? event.metaKey : !event.metaKey

      // 检查主键
      const isTargetKey = event.key.toLowerCase() === targetKey.toLowerCase()

      if (isTargetKey && hasCtrl && hasShift && hasAlt && hasMeta) {
        if (preventDefault) {
          event.preventDefault()
        }
        if (stopPropagation) {
          event.stopPropagation()
        }
        callbackRef.current(event)
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [targetKey, modifiers, preventDefault, stopPropagation, disabled])
}

/**
 * 多个快捷键 Hook
 */
export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
  const shortcutsRef = useRef(shortcuts)
  shortcutsRef.current = shortcuts

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      shortcutsRef.current.forEach((shortcut) => {
        if (shortcut.disabled) return

        const { key, modifiers = [], callback, preventDefault = true, stopPropagation = false } = shortcut

        // 检查修饰键
        const hasCtrl = modifiers.includes('ctrl') === event.ctrlKey
        const hasShift = modifiers.includes('shift') === event.shiftKey
        const hasAlt = modifiers.includes('alt') === event.altKey
        const hasMeta = modifiers.includes('meta') === event.metaKey

        // 检查是否所有非指定的修饰键都未按下
        const noExtraModifiers =
          (!event.ctrlKey || modifiers.includes('ctrl')) &&
          (!event.shiftKey || modifiers.includes('shift')) &&
          (!event.altKey || modifiers.includes('alt')) &&
          (!event.metaKey || modifiers.includes('meta'))

        // 检查主键
        const isTargetKey = event.key.toLowerCase() === key.toLowerCase()

        if (isTargetKey && hasCtrl && hasShift && hasAlt && hasMeta && noExtraModifiers) {
          if (preventDefault) {
            event.preventDefault()
          }
          if (stopPropagation) {
            event.stopPropagation()
          }
          callback(event)
        }
      })
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [])

  return { shortcuts: shortcutsRef.current }
}

/**
 * 热键帮助 Hook
 * 显示所有可用的快捷键
 */
export function useHotkeysHelp(shortcuts: KeyboardShortcut[]) {
  const formatShortcut = useCallback((shortcut: KeyboardShortcut): string => {
    const parts: string[] = []

    if (shortcut.modifiers?.includes('ctrl')) parts.push('Ctrl')
    if (shortcut.modifiers?.includes('shift')) parts.push('Shift')
    if (shortcut.modifiers?.includes('alt')) parts.push('Alt')
    if (shortcut.modifiers?.includes('meta')) parts.push('Meta')

    parts.push(shortcut.key.toUpperCase())

    return parts.join(' + ')
  }, [])

  const getHelpText = useCallback(() => {
    return shortcuts
      .filter((s) => !s.disabled && s.description)
      .map((s) => ({
        keys: formatShortcut(s),
        description: s.description || '',
      }))
  }, [shortcuts, formatShortcut])

  return { getHelpText, formatShortcut }
}

/**
 * 常用快捷键预设
 */
export const commonShortcuts = {
  /**
   * 搜索
   */
  search: (callback: () => void): KeyboardShortcut => ({
    key: 'k',
    modifiers: ['ctrl'],
    callback,
    description: '打开搜索',
  }),

  /**
   * 刷新
   */
  refresh: (callback: () => void): KeyboardShortcut => ({
    key: 'r',
    modifiers: ['ctrl'],
    callback,
    description: '刷新页面',
  }),

  /**
   * 保存
   */
  save: (callback: () => void): KeyboardShortcut => ({
    key: 's',
    modifiers: ['ctrl'],
    callback,
    description: '保存',
  }),

  /**
   * 撤销
   */
  undo: (callback: () => void): KeyboardShortcut => ({
    key: 'z',
    modifiers: ['ctrl'],
    callback,
    description: '撤销',
  }),

  /**
   * 重做
   */
  redo: (callback: () => void): KeyboardShortcut => ({
    key: 'y',
    modifiers: ['ctrl'],
    callback,
    description: '重做',
  }),

  /**
   * 全选
   */
  selectAll: (callback: () => void): KeyboardShortcut => ({
    key: 'a',
    modifiers: ['ctrl'],
    callback,
    description: '全选',
  }),

  /**
   * 删除
   */
  delete: (callback: () => void): KeyboardShortcut => ({
    key: 'Delete',
    modifiers: [],
    callback,
    description: '删除',
  }),

  /**
   * 关闭
   */
  close: (callback: () => void): KeyboardShortcut => ({
    key: 'Escape',
    modifiers: [],
    callback,
    description: '关闭/取消',
  }),

  /**
   * 帮助
   */
  help: (callback: () => void): KeyboardShortcut => ({
    key: '?',
    modifiers: ['shift'],
    callback,
    description: '显示帮助',
  }),

  /**
   * 新建
   */
  create: (callback: () => void): KeyboardShortcut => ({
    key: 'n',
    modifiers: ['ctrl'],
    callback,
    description: '新建',
  }),
}

/**
 * 输入框焦点管理 Hook
 */
export function useInputFocus(inputRef: React.RefObject<HTMLInputElement>, shortcut: string = '/') {
  useKeyPress(
    shortcut,
    () => {
      inputRef.current?.focus()
    },
    { preventDefault: true }
  )
}

/**
 * 序列键 Hook (如 g g)
 */
export function useKeySequence(
  sequence: string[],
  callback: () => void,
  options: {
    timeout?: number
    disabled?: boolean
  } = {}
) {
  const { timeout = 1000, disabled = false } = options
  const pressedKeys = useRef<string[]>([])
  const timeoutRef = useRef<NodeJS.Timeout>()

  useEffect(() => {
    if (disabled) return

    const handleKeyDown = (event: KeyboardEvent) => {
      // 添加按键
      pressedKeys.current.push(event.key.toLowerCase())

      // 清除之前的超时
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }

      // 检查序列
      const currentSequence = pressedKeys.current.join(' ')
      const targetSequence = sequence.join(' ')

      if (currentSequence === targetSequence) {
        event.preventDefault()
        callback()
        pressedKeys.current = []
      } else if (targetSequence.startsWith(currentSequence)) {
        // 部分匹配,设置超时
        timeoutRef.current = setTimeout(() => {
          pressedKeys.current = []
        }, timeout)
      } else {
        // 不匹配,重置
        pressedKeys.current = []
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [sequence, callback, timeout, disabled])
}
