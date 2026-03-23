import { onMounted, onUnmounted, ref } from 'vue'

export type ReplayMode = 'live' | 'replay'

export interface ReplayAnnotation {
  timestampNs: number
  label: string
}

export interface ReplayState {
  mode: ReplayMode
  loaded: boolean
  playing: boolean
  rate: number
  positionNs: number
  durationNs: number
  fileName: string
  error: string
  annotations: ReplayAnnotation[]
}

const defaultState: ReplayState = {
  mode: 'live',
  loaded: false,
  playing: false,
  rate: 1,
  positionNs: 0,
  durationNs: 0,
  fileName: '',
  error: '',
  annotations: [],
}

export function useReplay() {
  const state = ref<ReplayState>({ ...defaultState })
  const loading = ref(false)
  const isSeeking = ref(false)
  const error = ref<string | null>(null)
  let timerId: number | undefined

  const fetchState = async () => {
    if (isSeeking.value) return
    try {
      const res = await fetch('/api/replay/state')
      if (!res.ok) return
      const data = await res.json()
      state.value = data
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch replay state'
    }
  }

  const postControl = async (payload: Record<string, unknown>) => {
    try {
      const res = await fetch('/api/replay/control', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || 'Control request failed')
      }
      state.value = await res.json()
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Control request failed'
    }
  }

  const uploadFile = async (file: File) => {
    loading.value = true
    try {
      const formData = new FormData()
      formData.append('file', file)
      const res = await fetch('/api/replay/upload', {
        method: 'POST',
        body: formData,
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || 'Upload failed')
      }
      state.value = await res.json()
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Upload failed'
    } finally {
      loading.value = false
    }
  }

  const loadFromPath = async (path: string) => {
    loading.value = true
    try {
      const res = await fetch('/api/replay/load-path', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || 'Load from path failed')
      }
      state.value = await res.json()
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Load from path failed'
    } finally {
      loading.value = false
    }
  }

  const setMode = async (mode: ReplayMode) => {
    await postControl({ action: 'set_mode', mode })
  }

  const play = async () => postControl({ action: 'play' })
  const pause = async () => postControl({ action: 'pause' })
  const seek = async (positionNs: number) => postControl({ action: 'seek', positionNs })
  const step = async (delta: number) => postControl({ action: 'step', delta })
  const setRate = async (rate: number) => postControl({ action: 'set_rate', rate })

  onMounted(() => {
    fetchState()
    timerId = window.setInterval(fetchState, 300)
  })

  onUnmounted(() => {
    if (timerId) {
      window.clearInterval(timerId)
    }
  })

  return {
    state,
    loading,
    isSeeking,
    error,
    fetchState,
    uploadFile,
    loadFromPath,
    setMode,
    play,
    pause,
    seek,
    step,
    setRate,
  }
}
