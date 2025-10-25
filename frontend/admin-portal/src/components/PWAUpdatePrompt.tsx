import { useEffect } from 'react'
import { Button, notification } from 'antd'
// @ts-ignore - virtual module
import { useRegisterSW } from 'virtual:pwa-register/react'
import { ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const PWAUpdatePrompt = () => {
  const { t } = useTranslation()
  const [notificationApi, contextHolder] = notification.useNotification()

  const {
    offlineReady: [offlineReady, setOfflineReady],
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegistered(r: any) {
      console.log('SW Registered: ' + r)
    },
    onRegisterError(error: any) {
      console.log('SW registration error', error)
    },
  })

  const close = () => {
    setOfflineReady(false)
    setNeedRefresh(false)
    notificationApi.destroy('pwa-update')
  }

  const handleUpdate = () => {
    updateServiceWorker(true)
    close()
  }

  useEffect(() => {
    if (offlineReady) {
      notificationApi.success({
        key: 'pwa-update',
        message: t('pwa.offlineReady', '应用已就绪'),
        description: t('pwa.offlineReadyDesc', '应用已缓存,可离线使用'),
        duration: 3,
      })
    }
  }, [offlineReady, notificationApi, t])

  useEffect(() => {
    if (needRefresh) {
      notificationApi.info({
        key: 'pwa-update',
        message: t('pwa.newVersionAvailable', '新版本可用'),
        description: t('pwa.newVersionDesc', '检测到新版本,点击更新按钮刷新应用'),
        duration: 0,
        btn: (
          <Button type="primary" icon={<ReloadOutlined />} onClick={handleUpdate}>
            {t('pwa.update', '更新')}
          </Button>
        ),
        onClose: close,
      })
    }
  }, [needRefresh, notificationApi, t])

  return <>{contextHolder}</>
}

export default PWAUpdatePrompt
