import { readonly, ref } from 'vue'
import { type ToastCategory, useNotificationSettings } from '@/composables/notificationSettings'

export type { ToastCategory }
export type ToastLevel = 'info' | 'success' | 'warning' | 'error'

export interface Toast {
  id: number
  category: ToastCategory
  level: ToastLevel
  title: string
  message?: string
}

export interface AddToastOptions {
  category: ToastCategory
  level?: ToastLevel
  title: string
  message?: string
  duration?: number
}

const MAX_VISIBLE = 5
const DEFAULT_DURATION = 4000

let nextId = 0
const toasts = ref<Toast[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()

function clearTimer(id: number): void {
  const timer = timers.get(id)
  if (timer) {
    clearTimeout(timer)
    timers.delete(id)
  }
}

function removeToast(id: number): void {
  clearTimer(id)
  const idx = toasts.value.findIndex(t => t.id === id)
  if (idx !== -1) toasts.value.splice(idx, 1)
}

export function useToast() {
  const { isEnabled, shouldSystemNotify } = useNotificationSettings()

  function addToast(options: AddToastOptions): void {
    if (!isEnabled(options.category)) return

    const id = nextId++
    const duration = options.duration ?? DEFAULT_DURATION

    toasts.value.unshift({
      id,
      category: options.category,
      level: options.level ?? 'info',
      title: options.title,
      message: options.message,
    })

    if (toasts.value.length > MAX_VISIBLE) {
      const [evicted] = toasts.value.splice(MAX_VISIBLE)
      if (evicted) clearTimer(evicted.id)
    }

    if (duration > 0) {
      timers.set(id, setTimeout(() => removeToast(id), duration))
    }

    if (shouldSystemNotify(options.category) && 'Notification' in window && Notification.permission === 'granted') {
      new Notification(options.title, { body: options.message, tag: `ssl-vision-${options.category}` })
    }
  }

  function clearAll(): void {
    timers.forEach(timer => clearTimeout(timer))
    timers.clear()
    toasts.value = []
  }

  return {
    toasts: readonly(toasts),
    addToast,
    removeToast,
    clearAll,
  }
}
