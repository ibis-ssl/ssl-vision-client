import { useLocalStorage } from '@vueuse/core'

export type ToastCategory = 'referee' | 'autoref' | 'system' | 'settings'

export interface NotificationCategoryConfig {
  enabled: boolean
  systemNotification: boolean
}

export interface NotificationSettings {
  globalEnabled: boolean
  categories: Record<ToastCategory, NotificationCategoryConfig>
}

const STORAGE_KEY = 'ssl-vision-notification-settings'

const DEFAULT_SETTINGS: NotificationSettings = {
  globalEnabled: true,
  categories: {
    referee: { enabled: true, systemNotification: false },
    autoref: { enabled: true, systemNotification: false },
    system: { enabled: true, systemNotification: false },
    settings: { enabled: true, systemNotification: false },
  },
}

const settings = useLocalStorage<NotificationSettings>(STORAGE_KEY, DEFAULT_SETTINGS, {
  mergeDefaults: (stored: NotificationSettings, defaults: NotificationSettings) => ({
    globalEnabled: stored.globalEnabled ?? defaults.globalEnabled,
    categories: {
      referee: { ...defaults.categories.referee, ...stored.categories?.referee },
      autoref: { ...defaults.categories.autoref, ...stored.categories?.autoref },
      system: { ...defaults.categories.system, ...stored.categories?.system },
      settings: { ...defaults.categories.settings, ...stored.categories?.settings },
    },
  }),
})

export function useNotificationSettings() {
  function isEnabled(category: ToastCategory): boolean {
    return settings.value.globalEnabled && settings.value.categories[category].enabled
  }

  function shouldSystemNotify(category: ToastCategory): boolean {
    return isEnabled(category) && settings.value.categories[category].systemNotification
  }

  async function requestSystemPermission(category: ToastCategory): Promise<void> {
    if (!('Notification' in window)) return
    if (Notification.permission === 'granted') return
    const result = await Notification.requestPermission()
    if (result !== 'granted') {
      settings.value.categories[category].systemNotification = false
    }
  }

  return { settings, isEnabled, shouldSystemNotify, requestSystemPermission }
}
