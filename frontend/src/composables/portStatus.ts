import { ref, onUnmounted } from 'vue'
import { ReconnectingWebSocket } from '@/helpers/websocket'

export interface PortStatus {
  port: number
  active: boolean
}

export interface ServicePortStatus {
  vision: PortStatus[]
  tracker: PortStatus[]
  referee: PortStatus[]
}

const status = ref<ServicePortStatus>({
  vision: [],
  tracker: [],
  referee: [],
})

let ws: ReconnectingWebSocket | null = null
let refCount = 0

function ensureConnected() {
  if (ws) return
  ws = new ReconnectingWebSocket('/api/port-status')
  ws.registerTextConsumer((data: string) => {
    try {
      status.value = JSON.parse(data) as ServicePortStatus
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

export function usePortStatus() {
  refCount++
  ensureConnected()

  onUnmounted(() => {
    refCount--
    maybeDisconnect()
  })

  function isPortActive(service: keyof ServicePortStatus, port: number): boolean {
    const ports = status.value[service]
    const entry = ports.find(p => p.port === port)
    return entry?.active ?? false
  }

  return { portStatus: status, isPortActive }
}
