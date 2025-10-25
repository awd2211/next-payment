/**
 * 操作确认和撤销 Hook
 * 提供操作确认、撤销、重做等功能
 */
import { Modal, message } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import { useState, useCallback, useRef } from 'react'

export interface ConfirmOptions {
  title?: string
  content?: string
  okText?: string
  cancelText?: string
  danger?: boolean
  okType?: 'primary' | 'default' | 'dashed' | 'link' | 'text'
}

/**
 * 确认对话框 Hook
 */
export function useConfirm() {
  const [loading, setLoading] = useState(false)

  const confirm = useCallback(
    (action: () => Promise<any>, options: ConfirmOptions = {}): Promise<any> => {
      return new Promise((resolve) => {
        Modal.confirm({
          title: options.title || '确认操作',
          icon: ExclamationCircleOutlined({ style: {} }) as any,
          content: options.content || '确定要执行此操作吗?',
          okText: options.okText || '确定',
          cancelText: options.cancelText || '取消',
          okType: options.okType || (options.danger ? 'danger' : 'primary'),
          okButtonProps: options.danger ? { danger: true } : undefined,
          onOk: async () => {
            setLoading(true)
            try {
              const result = await action()
              resolve(result)
            } catch (error: any) {
              message.error(error.message || '操作失败')
              resolve(null)
            } finally {
              setLoading(false)
            }
          },
          onCancel: () => {
            resolve(null)
          },
        })
      })
    },
    []
  )

  return { confirm, loading }
}

/**
 * 撤销/重做 Hook
 */
interface HistoryState<T> {
  past: T[]
  present: T
  future: T[]
}

export function useHistory<T>(initialState: T, maxHistory: number = 50) {
  const [state, setState] = useState<HistoryState<T>>({
    past: [],
    present: initialState,
    future: [],
  })

  /**
   * 设置新状态
   */
  const set = useCallback(
    (newState: T | ((prev: T) => T)) => {
      setState((currentState) => {
        const newPresent = typeof newState === 'function'
          ? (newState as (prev: T) => T)(currentState.present)
          : newState

        // 限制历史记录数量
        const newPast = [...currentState.past, currentState.present].slice(-maxHistory)

        return {
          past: newPast,
          present: newPresent,
          future: [], // 清空future
        }
      })
    },
    [maxHistory]
  )

  /**
   * 撤销
   */
  const undo = useCallback(() => {
    setState((currentState) => {
      if (currentState.past.length === 0) {
        message.warning('没有可撤销的操作')
        return currentState
      }

      const newPast = [...currentState.past]
      const newPresent = newPast.pop()!
      const newFuture = [currentState.present, ...currentState.future]

      message.success('已撤销')

      return {
        past: newPast,
        present: newPresent,
        future: newFuture,
      }
    })
  }, [])

  /**
   * 重做
   */
  const redo = useCallback(() => {
    setState((currentState) => {
      if (currentState.future.length === 0) {
        message.warning('没有可重做的操作')
        return currentState
      }

      const newFuture = [...currentState.future]
      const newPresent = newFuture.shift()!
      const newPast = [...currentState.past, currentState.present]

      message.success('已重做')

      return {
        past: newPast,
        present: newPresent,
        future: newFuture,
      }
    })
  }, [])

  /**
   * 重置
   */
  const reset = useCallback(
    (newState?: T) => {
      setState({
        past: [],
        present: newState !== undefined ? newState : initialState,
        future: [],
      })
    },
    [initialState]
  )

  /**
   * 清空历史
   */
  const clear = useCallback(() => {
    setState((currentState) => ({
      past: [],
      present: currentState.present,
      future: [],
    }))
  }, [])

  return {
    state: state.present,
    set,
    undo,
    redo,
    reset,
    clear,
    canUndo: state.past.length > 0,
    canRedo: state.future.length > 0,
    historySize: state.past.length,
  }
}

/**
 * 操作重试 Hook
 */
export function useRetry() {
  const [retrying, setRetrying] = useState(false)
  const retryCountRef = useRef(0)

  const retry = useCallback(
    async <T = any,>(
      action: () => Promise<T>,
      options: {
        maxRetries?: number
        retryDelay?: number
        onRetry?: (attempt: number) => void
        onSuccess?: (result: T) => void
        onError?: (error: any) => void
      } = {}
    ): Promise<T | null> => {
      const { maxRetries = 3, retryDelay = 1000, onRetry, onSuccess, onError } = options

      setRetrying(true)
      retryCountRef.current = 0

      const attemptAction = async (): Promise<T | null> => {
        try {
          const result = await action()
          message.success('操作成功')
          onSuccess?.(result)
          return result
        } catch (error: any) {
          retryCountRef.current++

          if (retryCountRef.current < maxRetries) {
            message.warning(`操作失败,正在重试 (${retryCountRef.current}/${maxRetries})...`)
            onRetry?.(retryCountRef.current)

            await new Promise((resolve) => setTimeout(resolve, retryDelay))
            return attemptAction()
          } else {
            message.error(`操作失败,已重试 ${maxRetries} 次`)
            onError?.(error)
            return null
          }
        } finally {
          if (retryCountRef.current >= maxRetries) {
            setRetrying(false)
          }
        }
      }

      return attemptAction()
    },
    []
  )

  return { retry, retrying }
}

/**
 * 批量操作确认 Hook
 */
export function useBatchConfirm() {
  const confirm = useCallback(
    (
      selectedItems: any[],
      action: (items: any[]) => Promise<any>,
      options: {
        title?: string
        itemName?: string
        warningMessage?: string
      } = {}
    ): Promise<any> => {
      const { title, itemName = '项', warningMessage } = options

      if (selectedItems.length === 0) {
        message.warning('请先选择数据')
        return Promise.resolve(null)
      }

      return new Promise((resolve) => {
        const contentText = `您已选择 ${selectedItems.length} ${itemName},确定要执行此操作吗?${
          warningMessage ? `\n\n警告: ${warningMessage}` : ''
        }`

        Modal.confirm({
          title: title || '批量操作确认',
          icon: ExclamationCircleOutlined({ style: {} }) as any,
          content: contentText,
          okText: '确定',
          cancelText: '取消',
          okButtonProps: { danger: true },
          onOk: async () => {
            try {
              const result = await action(selectedItems)
              message.success('批量操作成功')
              resolve(result)
            } catch (error: any) {
              message.error(error.message || '批量操作失败')
              resolve(null)
            }
          },
          onCancel: () => {
            resolve(null)
          },
        })
      })
    },
    []
  )

  return { confirm }
}
