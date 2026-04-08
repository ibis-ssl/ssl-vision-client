<script setup lang="ts">
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import type { ReplayState } from '@/composables/replay.ts'
import type { LayerVisibility } from '@/components/LayerControl.vue'
import { useSettings } from '@/composables/settings'
import { usePortStatus } from '@/composables/portStatus'
import { useToast, type ToastCategory } from '@/composables/toast'
import { useNotificationSettings } from '@/composables/notificationSettings'
import { formatTimeNs } from '@/utils/time'
import { REFEREE_COMMAND_LABELS } from '@/utils/referee'
import { Referee_Stage } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import { computed, inject, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

const STAGE_MAP: Record<number, string> = {
  [Referee_Stage.NORMAL_FIRST_HALF_PRE]:    'NORMAL_FIRST_HALF_PRE',
  [Referee_Stage.NORMAL_FIRST_HALF]:        'NORMAL_FIRST_HALF',
  [Referee_Stage.NORMAL_HALF_TIME]:         'NORMAL_HALF_TIME',
  [Referee_Stage.NORMAL_SECOND_HALF_PRE]:   'NORMAL_SECOND_HALF_PRE',
  [Referee_Stage.NORMAL_SECOND_HALF]:       'NORMAL_SECOND_HALF',
  [Referee_Stage.EXTRA_TIME_BREAK]:         'EXTRA_TIME_BREAK',
  [Referee_Stage.EXTRA_FIRST_HALF_PRE]:     'EXTRA_FIRST_HALF_PRE',
  [Referee_Stage.EXTRA_FIRST_HALF]:         'EXTRA_FIRST_HALF',
  [Referee_Stage.EXTRA_HALF_TIME]:          'EXTRA_HALF_TIME',
  [Referee_Stage.EXTRA_SECOND_HALF_PRE]:    'EXTRA_SECOND_HALF_PRE',
  [Referee_Stage.EXTRA_SECOND_HALF]:        'EXTRA_SECOND_HALF',
  [Referee_Stage.PENALTY_SHOOTOUT_BREAK]:   'PENALTY_SHOOTOUT_BREAK',
  [Referee_Stage.PENALTY_SHOOTOUT]:         'PENALTY_SHOOTOUT',
  [Referee_Stage.POST_GAME]:                'POST_GAME',
}

interface Props {
  referee: Referee | null | undefined
  activeSource: string
  sources: Record<string, string>
  visionConnected: boolean
  refereeConnected: boolean
  grsimConnected: boolean
  replayState: ReplayState
  replayLoading: boolean
  layerVisibility: LayerVisibility
}

const props = defineProps<Props>()

interface Emits {
  (e: 'update:activeSource', source: string): void
  (e: 'update:mode', mode: 'live' | 'replay'): void
  (e: 'replay-file-selected', file: File): void
  (e: 'replay-path-selected', path: string): void
  (e: 'toggle-layer', layerName: keyof LayerVisibility): void
}

const emit = defineEmits<Emits>()
const { config, loading: settingsLoading, error, successMessage, fetchConfig, updateConfig } = useSettings()
const { portStatus, isPortActive } = usePortStatus()
const { addToast } = useToast()
const { settings: notifSettings, requestSystemPermission } = useNotificationSettings()

const VISION_PORTS = [10006, 10020]
const TRACKER_PORTS = [11010, 10010]
const REFEREE_PORTS = [11003, 10003]

const visionPort = ref(10006)
const trackedPort = ref(10010)
const refereePort = ref(10003)

const NOTIFICATION_CATEGORY_LABELS: Record<ToastCategory, string> = {
  referee: 'Referee Commands',
  autoref: 'Autoref Game Events',
  system: 'System Events',
  settings: 'Settings Feedback',
}

async function onSystemNotifyToggle(category: ToastCategory): Promise<void> {
  if (notifSettings.value.categories[category].systemNotification) {
    await requestSystemPermission(category)
  }
}
const simAddress = ref('127.0.0.1')
const autoBallPlacementEnabled = ref(false)
const autoCenterAfterGoalEnabled = ref(false)
const isSettingsOpen = ref(false)

const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject', { value: null } as Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>)

const stageText = computed(() => {
  if (!props.referee) return '--'
  return STAGE_MAP[props.referee.stage] ?? `STAGE_${props.referee.stage}`
})

const commandText = computed(() => {
  if (!props.referee) return '--'
  return REFEREE_COMMAND_LABELS[props.referee.command] || `COMMAND_${props.referee.command}`
})

const timeText = computed(() => {
  if (!props.referee || props.referee.stageTimeLeft === undefined) return '--:--'
  // stageTimeLeft はマイクロ秒、formatTimeNs はナノ秒を受け取る
  return formatTimeNs(Number(props.referee.stageTimeLeft) * 1000)
})

const blueScore = computed(() => props.referee?.blue?.score ?? 0)
const yellowScore = computed(() => props.referee?.yellow?.score ?? 0)
const blueName = computed(() => props.referee?.blue?.name || '')
const yellowName = computed(() => props.referee?.yellow?.name || '')

const selectedText = computed(() => {
  if (!selectedObject.value) return '選択なし'
  if (selectedObject.value.type === 'ball') return 'ボール'
  if (selectedObject.value.type === 'robot') {
    const team = selectedObject.value.team === 'YELLOW' ? 'イエロー' : 'ブルー'
    return `${team} ロボット #${selectedObject.value.id}`
  }
  return '選択なし'
})

const sourceOptions = computed(() => Object.entries(props.sources))
const replayMode = computed(() => props.replayState.mode === 'replay')

function changeSource(source: string) {
  emit('update:activeSource', source)
}

function toggleMode() {
  emit('update:mode', replayMode.value ? 'live' : 'replay')
}

function syncInputsFromConfig() {
  visionPort.value = config.value.visionPort
  trackedPort.value = config.value.trackedPort
  refereePort.value = config.value.refereePort
  simAddress.value = config.value.simAddress || '127.0.0.1'
  autoBallPlacementEnabled.value = config.value.autoBallPlacementEnabled || false
  autoCenterAfterGoalEnabled.value = config.value.autoCenterAfterGoalEnabled || false
}

async function onSaveConfig() {
  await updateConfig({
    visionPort: visionPort.value,
    trackedPort: trackedPort.value,
    refereePort: refereePort.value,
    simAddress: simAddress.value,
    autoBallPlacementEnabled: autoBallPlacementEnabled.value,
    autoCenterAfterGoalEnabled: autoCenterAfterGoalEnabled.value,
  })
  if (successMessage.value) {
    addToast({ category: 'settings', level: 'success', title: successMessage.value, duration: 3000 })
  } else if (error.value) {
    addToast({ category: 'settings', level: 'error', title: error.value, duration: 5000 })
  }
}

function tryAutoSwitch(
  portRef: { value: number },
  ports: number[],
  statuses: { port: number; active: boolean }[],
  serviceName: string,
  messages: string[],
): boolean {
  const currentActive = statuses.find(s => s.port === portRef.value)?.active ?? false
  if (currentActive) return false

  const alternative = ports.find(p => p !== portRef.value)
  if (alternative === undefined) return false

  const altActive = statuses.find(s => s.port === alternative)?.active ?? false
  if (!altActive) return false

  portRef.value = alternative
  messages.push(`${serviceName} ポートを ${alternative} に自動切替しました`)
  return true
}

watch(portStatus, (newStatus) => {
  if (!newStatus) return

  const messages: string[] = []
  const changed = [
    tryAutoSwitch(visionPort, VISION_PORTS, newStatus.vision, 'Vision', messages),
    tryAutoSwitch(trackedPort, TRACKER_PORTS, newStatus.tracker, 'Tracker', messages),
    tryAutoSwitch(refereePort, REFEREE_PORTS, newStatus.referee, 'Referee', messages),
  ].some(Boolean)

  if (changed) {
    onSaveConfig()
    addToast({
      category: 'system',
      level: 'info',
      title: 'ポート自動切替',
      message: messages.join(' / '),
      duration: 5000,
    })
  }
})

function onReplayFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    emit('replay-file-selected', file)
    closeSettingsDrawer()
  }
}

const replayServerPath = ref('')

function onReplayPathLoad() {
  const path = replayServerPath.value.trim()
  if (path) {
    emit('replay-path-selected', path)
    closeSettingsDrawer()
  }
}

function closeSettingsDrawer() {
  isSettingsOpen.value = false
}

function toggleSettings() {
  isSettingsOpen.value = !isSettingsOpen.value
}

function onDocumentKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeSettingsDrawer()
  }
}

onMounted(async () => {
  document.addEventListener('keydown', onDocumentKeyDown)
  await fetchConfig()
  syncInputsFromConfig()
})

onUnmounted(() => {
  document.removeEventListener('keydown', onDocumentKeyDown)
})
</script>

<template>
  <div id="status-bar">
    <div class="main-row">
      <div class="info-section connection-status">
        <span class="label">接続:</span>
        <span class="connection-indicator" :class="{ connected: visionConnected }" title="Vision">V</span>
        <span class="connection-indicator" :class="{ connected: refereeConnected }" title="Referee">R</span>
        <span class="connection-indicator" :class="{ connected: grsimConnected }" title="grSim">G</span>
      </div>
      <button class="mode-badge" :class="{ replay: replayMode }" @click="toggleMode()">
        {{ replayMode ? 'REPLAY' : 'LIVE' }}
      </button>

      <div class="info-section source-selector">
        <span class="label">ソース:</span>
        <select :value="props.activeSource" @change="changeSource(($event.target as HTMLSelectElement).value)" class="source-select">
          <option v-for="[key, name] in sourceOptions" :key="key" :value="key">
            {{ name }}
          </option>
        </select>
      </div>

      <div class="info-section"><span class="value">{{ stageText }}</span></div>
      <div class="info-section"><span class="value command">{{ commandText }}</span></div>

      <div class="info-section score">
        <span class="team blue">
          <span v-if="blueName" class="team-name">{{ blueName }}</span>
          <span class="team-score">{{ blueScore }}</span>
        </span>
        <span class="separator">-</span>
        <span class="team yellow">
          <span v-if="yellowName" class="team-name">{{ yellowName }}</span>
          <span class="team-score">{{ yellowScore }}</span>
        </span>
      </div>

      <div class="info-section"><span class="value time">{{ timeText }}</span></div>
      <div class="info-section selection">
        <span class="label">選択:</span>
        <span class="value selection-text" :class="{ selected: selectedObject }">{{ selectedText }}</span>
      </div>

      <button class="settings-button" :class="{ active: isSettingsOpen }" @click="toggleSettings">
        Settings
      </button>
    </div>

  </div>

  <Teleport to="body">
    <Transition name="settings-backdrop">
      <div v-if="isSettingsOpen" class="settings-backdrop" @click="closeSettingsDrawer" />
    </Transition>
    <Transition name="settings-drawer">
      <div v-if="isSettingsOpen" class="settings-drawer" role="dialog" aria-label="Settings">
        <div class="settings-drawer-inner">
          <header class="settings-panel-header">
            <div class="settings-panel-header-row">
              <h3 class="settings-panel-title">Settings</h3>
              <button class="settings-close-button" @click="closeSettingsDrawer" aria-label="閉じる">✕</button>
            </div>
            <p class="settings-panel-subtitle">表示・リプレイ・ネットワーク設定</p>
          </header>
          <div class="settings-content settings-overview">

            <div class="settings-section">
              <div class="settings-section-header">表示</div>
              <div class="settings-section-body">
                <div class="settings-group settings-card">
                  <span class="settings-group-title">Layer Visibility</span>
                  <div class="settings-toggle-list">
                    <label class="settings-toggle-item" data-tooltip="ボールの描画を切り替え"><input type="checkbox" :checked="props.layerVisibility.ball" @change="emit('toggle-layer', 'ball')" /><span>Ball</span></label>
                    <label class="settings-toggle-item" data-tooltip="レフェリー情報（配置指示等）の表示を切り替え"><input type="checkbox" :checked="props.layerVisibility.referee" @change="emit('toggle-layer', 'referee')" /><span>Referee</span></label>
                    <label class="settings-toggle-item" data-tooltip="フィールドラインの描画を切り替え"><input type="checkbox" :checked="props.layerVisibility.fieldLines" @change="emit('toggle-layer', 'fieldLines')" /><span>Field</span></label>
                    <label class="settings-toggle-item" data-tooltip="ファール/ゲームイベントの表示を切り替え"><input type="checkbox" :checked="props.layerVisibility.fouls" @change="emit('toggle-layer', 'fouls')" /><span>Fouls</span></label>
                  </div>
                  <span class="settings-group-title tracker-overlay-title">Tracker Overlay</span>
                  <div class="settings-toggle-list">
                    <label class="settings-toggle-item" data-tooltip="ロボット・ボールの速度ベクトルを表示"><input type="checkbox" :checked="props.layerVisibility.velocity" @change="emit('toggle-layer', 'velocity')" /><span>Velocity</span></label>
                    <label class="settings-toggle-item" data-tooltip="キックされたボールの予測軌道を表示"><input type="checkbox" :checked="props.layerVisibility.kickedBall" @change="emit('toggle-layer', 'kickedBall')" /><span>Kicked Ball</span></label>
                    <label class="settings-toggle-item" data-tooltip="Trackerの検出信頼度に応じた透明度表示"><input type="checkbox" :checked="props.layerVisibility.trackerVisibility" @change="emit('toggle-layer', 'trackerVisibility')" /><span>Visibility Opacity</span></label>
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-section">
              <div class="settings-section-header">データソース</div>
              <div class="settings-section-body">
                <div class="settings-group settings-card">
                  <span class="settings-group-title">Replay Source</span>
                  <div class="file-picker">
                    <label class="setting-field-label" data-tooltip="ローカルのログファイルを選択して再生">
                      <span>Replay Log</span>
                      <input class="input-control" type="file" accept=".log,.gz,.log.gz" :disabled="settingsLoading || props.replayLoading" @change="onReplayFileSelected" />
                    </label>
                    <div class="drop-hint">Select .log / .log.gz file</div>
                  </div>
                  <div class="file-picker">
                    <label class="setting-field-label" data-tooltip="サーバー上のログファイルパスを指定して再生">
                      <span>Server Path</span>
                      <input class="input-control" type="text" placeholder="/path/to/file.log" v-model="replayServerPath" :disabled="props.replayLoading" @keydown.enter="onReplayPathLoad" />
                    </label>
                    <button class="btn-load-path" :disabled="props.replayLoading || !replayServerPath.trim()" @click="onReplayPathLoad">Load</button>
                  </div>
                </div>

                <div class="settings-group settings-card">
                  <span class="settings-group-title">Network Endpoints</span>
                  <div class="setting-field-grid">
                    <div class="setting-field-label" data-tooltip="SSL-Visionのマルチキャストポート">
                      <span>Vision Port</span>
                      <div class="port-radio-group">
                        <label v-for="p in VISION_PORTS" :key="p" class="port-radio" :class="{ 'port-inactive': !isPortActive('vision', p) }">
                          <input type="radio" v-model.number="visionPort" :value="p" :disabled="settingsLoading" />
                          <span class="port-dot" :class="{ active: isPortActive('vision', p) }"></span>
                          <span>{{ p }}</span>
                        </label>
                      </div>
                    </div>
                    <div class="setting-field-label" data-tooltip="Tracker（ssl-vision-tracker）のポート">
                      <span>Tracked Port</span>
                      <div class="port-radio-group">
                        <label v-for="p in TRACKER_PORTS" :key="p" class="port-radio" :class="{ 'port-inactive': !isPortActive('tracker', p) }">
                          <input type="radio" v-model.number="trackedPort" :value="p" :disabled="settingsLoading" />
                          <span class="port-dot" :class="{ active: isPortActive('tracker', p) }"></span>
                          <span>{{ p }}</span>
                        </label>
                      </div>
                    </div>
                    <div class="setting-field-label" data-tooltip="Game Controllerのレフェリーポート">
                      <span>Referee Port</span>
                      <div class="port-radio-group">
                        <label v-for="p in REFEREE_PORTS" :key="p" class="port-radio" :class="{ 'port-inactive': !isPortActive('referee', p) }">
                          <input type="radio" v-model.number="refereePort" :value="p" :disabled="settingsLoading" />
                          <span class="port-dot" :class="{ active: isPortActive('referee', p) }"></span>
                          <span>{{ p }}</span>
                        </label>
                      </div>
                    </div>
                    <label class="setting-field-label" data-tooltip="シミュレータの送信先IPアドレス（SSL Simulation Protocol）">
                      <span>Sim Address</span>
                      <input class="input-control" type="text" v-model="simAddress" :disabled="settingsLoading" />
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-section">
              <div class="settings-section-header">動作・通知</div>
              <div class="settings-section-body">
                <div class="settings-group settings-card">
                  <span class="settings-group-title">Automation</span>
                  <div class="settings-toggle-list">
                    <label class="settings-toggle-item" data-tooltip="ファール後にボールを自動で指定位置に配置"><input type="checkbox" v-model="autoBallPlacementEnabled" /><span>Auto Ball Placement</span></label>
                    <label class="settings-toggle-item" data-tooltip="ゴール後にボールを自動でセンターに配置"><input type="checkbox" v-model="autoCenterAfterGoalEnabled" /><span>Auto Center After Goal</span></label>
                  </div>
                  <div class="config-buttons">
                    <button class="replay-button ghost" @click="syncInputsFromConfig" :disabled="settingsLoading">Reset</button>
                    <button class="replay-button primary" @click="onSaveConfig" :disabled="settingsLoading">{{ settingsLoading ? 'Saving...' : 'Save' }}</button>
                  </div>
                  <div v-if="error" class="error-text">{{ error }}</div>
                  <div v-else-if="successMessage" class="success-text">{{ successMessage }}</div>
                </div>

                <div class="settings-group settings-card">
                  <span class="settings-group-title">Notifications</span>
                  <div class="settings-toggle-list">
                    <label class="settings-toggle-item" data-tooltip="すべてのトースト通知を有効化/無効化">
                      <input type="checkbox" v-model="notifSettings.globalEnabled" />
                      <span>Enable All</span>
                    </label>
                  </div>
                  <span class="settings-group-title notif-sub-title">Per Category</span>
                  <div class="settings-toggle-list">
                    <label
                      v-for="(cfg, key) in notifSettings.categories"
                      :key="key"
                      class="settings-toggle-item"
                      data-tooltip="カテゴリ別にトースト通知を制御"
                    >
                      <input type="checkbox" v-model="cfg.enabled" :disabled="!notifSettings.globalEnabled" />
                      <span>{{ NOTIFICATION_CATEGORY_LABELS[key] ?? key }}</span>
                    </label>
                  </div>
                  <span class="settings-group-title notif-sub-title">System Notifications</span>
                  <div class="settings-toggle-list">
                    <label
                      v-for="(cfg, key) in notifSettings.categories"
                      :key="key"
                      class="settings-toggle-item"
                      data-tooltip="OSのシステム通知として表示（ブラウザ権限が必要）"
                    >
                      <input
                        type="checkbox"
                        v-model="cfg.systemNotification"
                        :disabled="!notifSettings.globalEnabled || !cfg.enabled"
                        @change="onSystemNotifyToggle(key as ToastCategory)"
                      />
                      <span>{{ NOTIFICATION_CATEGORY_LABELS[key] ?? key }}</span>
                    </label>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
#status-bar {
  display: flex;
  flex-direction: column;
  gap: 0.6em;
  padding: 0.75em 1em;
  box-sizing: border-box;
  background: var(--md-sys-color-surface-container);
  backdrop-filter: var(--app-glass-blur);
  border-top: 1px solid var(--md-sys-color-outline-variant);
  color: var(--md-sys-color-on-surface);
  width: 100%;
  position: absolute;
  bottom: 0;
  left: 0;
  z-index: 100;
}

.main-row {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.85em;
  flex-wrap: wrap;
}

.info-section {
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.info-section.score {
  gap: 0.75em;
  font-size: 1.1em;
  font-weight: 600;
}

.label {
  color: var(--md-sys-color-on-surface-variant);
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.value {
  color: var(--md-sys-color-on-surface);
  font-weight: 600;
  font-family: var(--md-sys-typescale-mono-font);
}

.value.command {
  color: var(--app-color-command);
}

.value.time {
  color: var(--app-color-live);
  font-size: 1.05em;
}

.team {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.2em 0.65em;
  border-radius: var(--md-sys-shape-corner-small);
  gap: 0.1em;
}

.team.blue {
  background-color: var(--app-color-team-blue-container);
  color: var(--app-color-team-blue);
  border: 1px solid var(--app-color-team-blue-border);
}

.team.yellow {
  background-color: var(--app-color-team-yellow-container);
  color: var(--app-color-team-yellow);
  border: 1px solid var(--app-color-team-yellow-border);
}

.team-name {
  font-size: 0.68rem;
  opacity: 0.85;
  font-weight: 500;
}

.team-score {
  font-size: 1.2em;
  font-weight: 700;
  font-family: var(--md-sys-typescale-mono-font);
}

.separator {
  color: var(--md-sys-color-outline);
  font-weight: 500;
}

.info-section.selection,
.source-selector {
  background: var(--md-sys-color-surface-container-low);
  padding: 0.3em 0.75em;
  border-radius: var(--md-sys-shape-corner-full);
  border: 1px solid var(--md-sys-color-outline-variant);
}

.selection-text {
  color: var(--md-sys-color-on-surface-variant);
  font-size: 0.88em;
}

.selection-text.selected {
  color: var(--app-color-success);
  font-weight: 600;
}

.connection-status {
  gap: 0.3em;
}

.connection-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.6em;
  height: 1.6em;
  padding: 0 0.3em;
  border-radius: var(--md-sys-shape-corner-extra-small);
  background-color: var(--md-sys-color-error-container);
  color: var(--md-sys-color-on-error-container);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.connection-indicator.connected {
  background-color: var(--app-color-success-container);
  color: var(--app-color-success);
  border: 1px solid var(--app-color-success-border);
}

.source-select {
  background-color: transparent;
  color: var(--md-sys-color-on-surface);
  border: none;
  border-radius: var(--md-sys-shape-corner-full);
  padding: 0.1em 0.4em;
  font-size: 0.88em;
  cursor: pointer;
}

.mode-badge {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  border-radius: var(--md-sys-shape-corner-full);
  padding: 0.3em 0.8em;
  cursor: pointer;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard),
              color var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
  border: none;
}

.mode-badge:not(.replay) {
  background: var(--app-color-success-container);
  color: var(--app-color-live);
  border: 1px solid var(--app-color-success-border);
}

.mode-badge:not(.replay):hover {
  background: color-mix(in srgb, var(--app-color-success) 25%, transparent);
}

.mode-badge.replay {
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
}

.mode-badge.replay:hover {
  background: var(--md-sys-color-primary);
  color: var(--md-sys-color-on-primary);
}

/* Settings ボタン（ステータスバー内） */
.settings-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--md-sys-color-on-surface);
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-corner-full);
  background: var(--md-sys-color-surface-container-low);
  padding: 0.3em 0.8em;
  font-size: 0.88rem;
  font-weight: 600;
  font-family: var(--md-sys-typescale-body-font);
  transition:
    background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard),
    color var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.settings-button:hover {
  background: var(--md-sys-color-surface-container-high);
}

.settings-button.active {
  background: var(--md-sys-color-secondary-container);
  color: var(--md-sys-color-on-secondary-container);
  border-color: var(--md-sys-color-secondary);
}

/* バックドロップ */
.settings-backdrop {
  position: fixed;
  inset: 0;
  z-index: 490;
  background: rgba(0, 0, 0, 0.32);
}

/* 右サイドドロワー本体 */
.settings-drawer {
  position: fixed;
  top: 0;
  right: 0;
  height: 100dvh;
  width: 400px;
  max-width: 100vw;
  z-index: 500;
  background: var(--md-sys-color-surface-container-low);
  backdrop-filter: var(--app-glass-blur);
  border-left: 1px solid var(--app-glass-border);
  box-shadow: var(--md-sys-elevation-4);
  display: flex;
  flex-direction: column;
}

.settings-drawer-inner {
  flex: 1;
  overflow-y: auto;
  padding: 1.25em 1em 2em;
  display: flex;
  flex-direction: column;
}

.settings-panel-header {
  margin-bottom: 1em;
}

.settings-panel-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.settings-panel-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--md-sys-color-on-surface);
}

.settings-close-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2em;
  height: 2em;
  border-radius: var(--md-sys-shape-corner-full);
  border: none;
  background: transparent;
  color: var(--md-sys-color-on-surface-variant);
  cursor: pointer;
  font-size: 0.9rem;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.settings-close-button:hover {
  background: var(--md-sys-color-surface-container-high);
  color: var(--md-sys-color-on-surface);
}

.settings-panel-subtitle {
  margin: 0.25em 0 0;
  color: var(--md-sys-color-on-surface-variant);
  font-size: 0.78rem;
}

/* アニメーション: バックドロップ */
.settings-backdrop-enter-active {
  transition: opacity var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-standard);
}
.settings-backdrop-leave-active {
  transition: opacity var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}
.settings-backdrop-enter-from,
.settings-backdrop-leave-to {
  opacity: 0;
}

/* アニメーション: ドロワー（右からスライドイン） */
.settings-drawer-enter-active {
  transition: transform var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-emphasized-decelerate);
}
.settings-drawer-leave-active {
  transition: transform var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-emphasized-accelerate);
}
.settings-drawer-enter-from,
.settings-drawer-leave-to {
  transform: translateX(100%);
}

.settings-content {
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-corner-large);
  background: var(--md-sys-color-surface-container-lowest);
  padding: 0.75em;
}

.settings-overview {
  display: flex;
  flex-direction: column;
  gap: 0.75em;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: 0.5em;
}

.settings-section-header {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--md-sys-color-on-surface-variant);
  padding: 0 0.25em;
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.settings-section-header::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--md-sys-color-outline-variant);
}

.settings-section-body {
  display: flex;
  flex-direction: column;
  gap: 0.65em;
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 0.6em;
}

.settings-card {
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-corner-medium);
  background: var(--md-sys-color-surface-container);
  padding: 0.7em 0.75em;
}


.settings-group-title {
  color: var(--md-sys-color-primary);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  margin: 0;
}

.tracker-overlay-title {
  margin-top: 0.5em;
  color: var(--md-sys-color-tertiary);
}

.notif-sub-title {
  margin-top: 0.5em;
  color: var(--md-sys-color-tertiary);
}

.settings-toggle-list {
  display: grid;
  gap: 0.4em;
}

.settings-toggle-item {
  display: inline-flex;
  align-items: center;
  gap: 0.5em;
  font-size: 0.875rem;
  color: var(--md-sys-color-on-surface);
  cursor: pointer;
}

.settings-toggle-item input[type="checkbox"] {
  accent-color: var(--md-sys-color-primary);
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.setting-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65em;
}

.setting-field-label {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.3em;
  font-size: 0.875rem;
  color: var(--md-sys-color-on-surface);
}

.file-picker {
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-corner-small);
  background: var(--md-sys-color-surface-container-high);
  padding: 0.55em 0.65em;
  display: grid;
  gap: 0.4em;
}

.input-control {
  background: var(--md-sys-color-surface-container-lowest);
  color: var(--md-sys-color-on-surface);
  border: 1px solid var(--md-sys-color-outline);
  border-radius: var(--md-sys-shape-corner-extra-small);
  padding: 0.3em 0.5em;
  width: 100%;
  box-sizing: border-box;
  font-size: 0.875rem;
  transition: border-color var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.input-control:focus {
  outline: none;
  border-color: var(--md-sys-color-primary);
}

.drop-hint {
  font-size: 0.72rem;
  color: var(--md-sys-color-on-surface-variant);
}

.btn-load-path {
  margin-top: 0.35em;
  padding: 0.35em 0.9em;
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
  border: none;
  border-radius: var(--md-sys-shape-corner-full);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 500;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.btn-load-path:hover:not(:disabled) {
  background: var(--md-sys-color-primary);
  color: var(--md-sys-color-on-primary);
}

.btn-load-path:disabled {
  opacity: 0.38;
  cursor: not-allowed;
}

.config-buttons {
  display: flex;
  gap: 0.5em;
  margin-top: 0.2em;
}

.replay-button {
  padding: 0.35em 1em;
  border-radius: var(--md-sys-shape-corner-full);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 500;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard),
              color var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.replay-button.primary {
  background: var(--md-sys-color-primary);
  color: var(--md-sys-color-on-primary);
  border: none;
}

.replay-button.primary:hover:not(:disabled) {
  background: var(--md-sys-color-on-primary-container);
  color: var(--md-sys-color-primary-container);
}

.replay-button.ghost {
  background: transparent;
  color: var(--md-sys-color-primary);
  border: 1px solid var(--md-sys-color-outline);
}

.replay-button.ghost:hover:not(:disabled) {
  background: var(--md-sys-color-primary-hover);
}

.replay-button:disabled {
  opacity: 0.38;
  cursor: not-allowed;
}

.error-text {
  color: var(--md-sys-color-error);
  font-size: 0.82rem;
}

.success-text {
  color: var(--app-color-success);
  font-size: 0.82rem;
}

.port-radio-group {
  display: flex;
  gap: 0.75em;
  flex-wrap: wrap;
  margin-top: 0.2em;
}

.port-radio {
  display: inline-flex;
  align-items: center;
  gap: 0.35em;
  cursor: pointer;
  font-size: 0.875rem;
  font-family: var(--md-sys-typescale-mono-font);
  color: var(--md-sys-color-on-surface);
}


.port-radio input[type="radio"] {
  accent-color: var(--md-sys-color-primary);
  cursor: pointer;
}

.port-radio.port-inactive {
  opacity: 0.38;
}

.port-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--md-sys-color-outline);
  flex-shrink: 0;
}

.port-dot.active {
  background: var(--app-color-success);
  box-shadow: 0 0 5px rgba(129, 201, 149, 0.6);
}

/* ツールチップ (CSS-only) */
[data-tooltip] {
  position: relative;
}

[data-tooltip]::after {
  content: attr(data-tooltip);
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  background: var(--md-sys-color-inverse-surface);
  color: var(--md-sys-color-inverse-on-surface);
  font-size: 0.72rem;
  font-weight: 400;
  padding: 0.35em 0.65em;
  border-radius: var(--md-sys-shape-corner-extra-small);
  white-space: nowrap;
  pointer-events: none;
  opacity: 0;
  transition: opacity var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
  z-index: 300;
  font-family: var(--md-sys-typescale-body-font);
}

[data-tooltip]:hover::after {
  opacity: 1;
}

@media (max-width: 600px) {
  .settings-drawer {
    width: 100vw;
    border-left: none;
  }
  .setting-field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
