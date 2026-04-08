import { ref, onUnmounted } from 'vue'
import { ReconnectingWebSocket } from '@/helpers/websocket'

export interface ServiceStats {
  visionHz: number
  cameraCount: number
  trackerHz: number
  trackerActive: boolean
  refereeHz: number
}

const stats = ref<ServiceStats>({
  visionHz: 0,
  cameraCount: 0,
  trackerHz: 0,
  trackerActive: false,
  refereeHz: 0,
})

let ws: ReconnectingWebSocket | null = null
let refCount = 0

function ensureConnected() {
  if (ws) return
  ws = new ReconnectingWebSocket('/api/service-stats')
  ws.registerTextConsumer((data: string) => {
    try {
      stats.value = JSON.parse(data) as ServiceStats
    } catch {
      // ignore parse errors
    }
  })
  ws.connect()
}

function maybeDisconnect() {
  if (refCount <= 0 && ws) {
    ws.disconnect()
    ws = null
  }
}

export function useServiceStats() {
  refCount++
  ensureConnected()

  onUnmounted(() => {
    refCount--
    maybeDisconnect()
  })

  return { serviceStats: stats }
}
