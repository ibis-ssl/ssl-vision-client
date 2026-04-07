import { ref } from 'vue'

export interface Config {
  visionPort: number
  trackedPort: number
  refereePort: number
  visionIP?: string
  trackedIP?: string
  refereeIP?: string
  simAddress?: string
  simPort?: number
  autoBallPlacementEnabled?: boolean
  autoCenterAfterGoalEnabled?: boolean
}

const API_BASE = '/api/config'
const config = ref<Config>({
  visionPort: 10006,
  trackedPort: 10010,
  refereePort: 10003,
  autoBallPlacementEnabled: false,
  autoCenterAfterGoalEnabled: false,
})

const loading = ref(false)
const error = ref<string | null>(null)
const successMessage = ref<string | null>(null)
let fetchedOnce = false

export function useSettings() {

  // 設定を取得（force=false の場合、複数コンポーネントから呼ばれても1回のみAPIリクエストを行う）
  const fetchConfig = async (force = false) => {
    if (!force && fetchedOnce) return
    fetchedOnce = true
    loading.value = true
    error.value = null

    try {
      const response = await fetch(API_BASE)
      if (!response.ok) {
        throw new Error('設定の取得に失敗しました')
      }
      const data = await response.json()
      config.value = {
        autoBallPlacementEnabled: false,
        autoCenterAfterGoalEnabled: false,
        ...data,
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : '不明なエラーが発生しました'
      console.error('Failed to fetch config:', e)
    } finally {
      loading.value = false
    }
  }

  // 設定を更新
  const updateConfig = async (newConfig: {
    visionPort: number
    trackedPort: number
    refereePort: number
    simAddress?: string
    simPort?: number
    autoBallPlacementEnabled?: boolean
    autoCenterAfterGoalEnabled?: boolean
  }) => {
    loading.value = true
    error.value = null
    successMessage.value = null

    try {
      const response = await fetch(API_BASE, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newConfig),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || '設定の更新に失敗しました')
      }

      // 設定を再取得
      await fetchConfig(true)
      successMessage.value = '設定を保存し、適用しました'

      // 成功メッセージを3秒後に消す
      setTimeout(() => {
        successMessage.value = null
      }, 3000)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '不明なエラーが発生しました'
      console.error('Failed to update config:', e)
    } finally {
      loading.value = false
    }
  }

  return {
    config,
    loading,
    error,
    successMessage,
    fetchConfig,
    updateConfig,
  }
}
