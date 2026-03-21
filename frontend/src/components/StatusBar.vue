<script setup lang="ts">
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import type { ReplayState } from '@/composables/replay.ts'
import type { LayerVisibility } from '@/components/LayerControl.vue'
import { useSettings } from '@/composables/settings'
import { formatTimeNs } from '@/utils/time'
import { computed, inject, onMounted, onUnmounted, ref, type Ref } from 'vue'

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

const visionPort = ref(10006)
const trackedPort = ref(10010)
const refereePort = ref(10003)
const grSimAddress = ref('127.0.0.1')
const grSimPort = ref(20011)
const autoBallPlacementEnabled = ref(false)
const autoCenterAfterGoalEnabled = ref(false)
const settingsDrawerRef = ref<HTMLDetailsElement | null>(null)

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
  const stageMap: { [key: number]: string } = {
    0: 'NORMAL_FIRST_HALF_PRE',
    1: 'NORMAL_FIRST_HALF',
    2: 'NORMAL_HALF_TIME',
    3: 'NORMAL_SECOND_HALF_PRE',
    4: 'NORMAL_SECOND_HALF',
    5: 'EXTRA_TIME_BREAK',
    6: 'EXTRA_FIRST_HALF_PRE',
    7: 'EXTRA_FIRST_HALF',
    8: 'EXTRA_HALF_TIME',
    9: 'EXTRA_SECOND_HALF_PRE',
    10: 'EXTRA_SECOND_HALF',
    11: 'PENALTY_SHOOTOUT_BREAK',
    12: 'PENALTY_SHOOTOUT',
    13: 'POST_GAME',
  }
  return stageMap[props.referee.stage] || `STAGE_${props.referee.stage}`
})

const commandText = computed(() => {
  if (!props.referee) return '--'
  const commandMap: { [key: number]: string } = {
    0: 'HALT',
    1: 'STOP',
    2: 'NORMAL_START',
    3: 'FORCE_START',
    4: 'PREPARE_KICKOFF_YELLOW',
    5: 'PREPARE_KICKOFF_BLUE',
    6: 'PREPARE_PENALTY_YELLOW',
    7: 'PREPARE_PENALTY_BLUE',
    8: 'DIRECT_FREE_YELLOW',
    9: 'DIRECT_FREE_BLUE',
    10: 'INDIRECT_FREE_YELLOW',
    11: 'INDIRECT_FREE_BLUE',
    12: 'TIMEOUT_YELLOW',
    13: 'TIMEOUT_BLUE',
    14: 'GOAL_YELLOW',
    15: 'GOAL_BLUE',
    16: 'BALL_PLACEMENT_YELLOW',
    17: 'BALL_PLACEMENT_BLUE',
  }
  return commandMap[props.referee.command] || `COMMAND_${props.referee.command}`
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
  grSimAddress.value = config.value.grSimAddress || '127.0.0.1'
  grSimPort.value = config.value.grSimPort || 20011
  autoBallPlacementEnabled.value = config.value.autoBallPlacementEnabled || false
  autoCenterAfterGoalEnabled.value = config.value.autoCenterAfterGoalEnabled || false
}

async function onSaveConfig() {
  await updateConfig({
    visionPort: visionPort.value,
    trackedPort: trackedPort.value,
    refereePort: refereePort.value,
    grSimAddress: grSimAddress.value,
    grSimPort: grSimPort.value,
    autoBallPlacementEnabled: autoBallPlacementEnabled.value,
    autoCenterAfterGoalEnabled: autoCenterAfterGoalEnabled.value,
  })
}

function onReplayFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    emit('replay-file-selected', file)
  }
}

const replayServerPath = ref('')

function onReplayPathLoad() {
  const path = replayServerPath.value.trim()
  if (path) {
    emit('replay-path-selected', path)
  }
}

function closeSettingsDrawer() {
  settingsDrawerRef.value?.removeAttribute('open')
}

function onDocumentPointerDown(event: MouseEvent) {
  const drawer = settingsDrawerRef.value
  if (!drawer?.open) return
  const target = event.target as Node | null
  if (target && drawer.contains(target)) return
  closeSettingsDrawer()
}

function onDocumentKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeSettingsDrawer()
  }
}

onMounted(async () => {
  document.addEventListener('mousedown', onDocumentPointerDown)
  document.addEventListener('keydown', onDocumentKeyDown)
  await fetchConfig()
  syncInputsFromConfig()
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocumentPointerDown)
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

      <details ref="settingsDrawerRef" class="settings-drawer">
        <summary>
          <span class="settings-summary-title">Settings</span>
        </summary>
        <div class="settings-panel">
          <header class="settings-panel-header">
            <h3 class="settings-panel-title">Runtime Controls</h3>
            <p class="settings-panel-subtitle">Tune visualization, replay, and network endpoints.</p>
          </header>
          <div class="settings-content settings-overview">
            <div class="settings-group settings-card">
              <span class="settings-group-title">Layer Visibility</span>
              <div class="settings-toggle-list">
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.ball" @change="emit('toggle-layer', 'ball')" /><span>Ball</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.referee" @change="emit('toggle-layer', 'referee')" /><span>Referee</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.fieldLines" @change="emit('toggle-layer', 'fieldLines')" /><span>Field</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.fouls" @change="emit('toggle-layer', 'fouls')" /><span>Fouls</span></label>
              </div>
            </div>

            <div class="settings-group settings-card">
              <span class="settings-group-title">Replay Source</span>
              <div class="file-picker">
                <label class="setting-field-label">
                  <span>Replay Log</span>
                  <input class="input-control" type="file" accept=".log,.gz,.log.gz" :disabled="settingsLoading || props.replayLoading" @change="onReplayFileSelected" />
                </label>
                <div class="drop-hint">Select .log / .log.gz file</div>
              </div>
              <div class="file-picker">
                <label class="setting-field-label">
                  <span>Server Path</span>
                  <input class="input-control" type="text" placeholder="/path/to/file.log" v-model="replayServerPath" :disabled="props.replayLoading" @keydown.enter="onReplayPathLoad" />
                </label>
                <button class="btn-load-path" :disabled="props.replayLoading || !replayServerPath.trim()" @click="onReplayPathLoad">Load</button>
              </div>
            </div>

            <div class="settings-group settings-card settings-card-wide">
              <span class="settings-group-title">Network Endpoints</span>
              <div class="setting-field-grid">
                <label class="setting-field-label">
                  <span>Vision Port</span>
                  <input class="input-control" type="number" v-model.number="visionPort" min="1" max="65535" :disabled="settingsLoading" />
                </label>
                <label class="setting-field-label">
                  <span>Tracked Port</span>
                  <input class="input-control" type="number" v-model.number="trackedPort" min="1" max="65535" :disabled="settingsLoading" />
                </label>
                <label class="setting-field-label">
                  <span>Referee Port</span>
                  <input class="input-control" type="number" v-model.number="refereePort" min="1" max="65535" :disabled="settingsLoading" />
                </label>
                <label class="setting-field-label">
                  <span>grSim Address</span>
                  <input class="input-control" type="text" v-model="grSimAddress" :disabled="settingsLoading" />
                </label>
                <label class="setting-field-label">
                  <span>grSim Port</span>
                  <input class="input-control" type="number" v-model.number="grSimPort" min="1" max="65535" :disabled="settingsLoading" />
                </label>
              </div>
            </div>

            <div class="settings-group settings-card">
              <span class="settings-group-title">Automation</span>
              <div class="settings-toggle-list">
                <label class="settings-toggle-item"><input type="checkbox" v-model="autoBallPlacementEnabled" /><span>Auto Ball Placement</span></label>
                <label class="settings-toggle-item"><input type="checkbox" v-model="autoCenterAfterGoalEnabled" /><span>Auto Center After Goal</span></label>
              </div>
              <div class="config-buttons">
                <button class="replay-button ghost" @click="syncInputsFromConfig" :disabled="settingsLoading">Reset</button>
                <button class="replay-button primary" @click="onSaveConfig" :disabled="settingsLoading">{{ settingsLoading ? 'Saving...' : 'Save' }}</button>
              </div>
              <div v-if="error" class="error-text">{{ error }}</div>
              <div v-else-if="successMessage" class="success-text">{{ successMessage }}</div>
            </div>
          </div>
        </div>
      </details>
    </div>

  </div>
</template>

<style scoped>
#status-bar {
  --bg-deep: rgba(8, 14, 22, 0.92);
  --bg-soft: rgba(20, 32, 46, 0.88);
  --line: #2b3d55;
  --line-strong: #4d6f98;
  --text-muted: #9bb3cf;
  --accent: #79c0ff;
  display: flex;
  flex-direction: column;
  gap: 0.6em;
  padding: 0.75em 1em;
  box-sizing: border-box;
  background: linear-gradient(180deg, var(--bg-soft), var(--bg-deep));
  border-top: 1px solid var(--line);
  color: white;
  width: 100%;
  position: absolute;
  bottom: 0;
  left: 0;
  z-index: 100;
  font-family: monospace;
}

.main-row {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1em;
  flex-wrap: wrap;
}

.info-section {
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.info-section.score {
  gap: 1em;
  font-size: 1.2em;
  font-weight: bold;
}

.label {
  color: var(--text-muted);
  font-size: 0.82em;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.value {
  color: white;
  font-weight: bold;
}

.value.command {
  color: #ffd700;
}

.value.time {
  color: #00ff00;
  font-size: 1.1em;
}

.team {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.2em 0.6em;
  border-radius: 4px;
  gap: 0.1em;
}

.team.blue {
  background-color: #0066cc;
  color: white;
}

.team.yellow {
  background-color: #ffcc00;
  color: black;
}

.team-name {
  font-size: 0.7em;
  opacity: 0.9;
}

.team-score {
  font-size: 1.2em;
  font-weight: bold;
}

.separator {
  color: #666;
  font-weight: bold;
}

.info-section.selection,
.source-selector {
  background: rgba(18, 27, 40, 0.9);
  padding: 0.32em 0.7em;
  border-radius: 999px;
  border: 1px solid var(--line);
}

.selection-text {
  color: #888;
  font-size: 0.95em;
}

.selection-text.selected {
  color: #00ff88;
  font-weight: bold;
}

.connection-status {
  gap: 0.3em;
}

.connection-indicator {
  display: inline-block;
  width: 1.5em;
  height: 1.5em;
  line-height: 1.5em;
  text-align: center;
  border-radius: 3px;
  background-color: #ff4444;
  color: white;
  font-size: 0.8em;
  font-weight: bold;
}

.connection-indicator.connected {
  background-color: #44ff44;
  color: black;
}

.source-select {
  background-color: rgba(10, 14, 18, 0.9);
  color: white;
  border: 1px solid var(--line-strong);
  border-radius: 999px;
  padding: 0.2em 0.5em;
  font-family: monospace;
  font-size: 0.9em;
  cursor: pointer;
}

.mode-badge {
  font-size: 0.72em;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  border-radius: 999px;
  padding: 0.2em 0.65em;
  cursor: pointer;
  font-family: monospace;
  transition: filter 0.15s;
  background: transparent;
}

.mode-badge:hover {
  filter: brightness(1.2);
}

.mode-badge:not(.replay) {
  border: 1px solid #44ff44;
  color: #44ff44;
}

.mode-badge.replay {
  border: 1px solid #79c0ff;
  color: #79c0ff;
}

.settings-drawer {
  position: relative;
}

.settings-drawer summary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #d7e7fb;
  user-select: none;
  list-style: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(18, 27, 40, 0.9);
  padding: 0.32em 0.75em;
}

.settings-drawer summary::-webkit-details-marker {
  display: none;
}

.settings-summary-title {
  font-weight: 700;
}

.settings-panel {
  margin-top: 0.45em;
  position: absolute;
  right: 0;
  bottom: calc(100% + 0.55em);
  z-index: 220;
  width: min(760px, calc(100vw - 2em));
  background: linear-gradient(160deg, rgba(7, 14, 24, 0.98), rgba(9, 18, 31, 0.98));
  border: 1px solid #3e5b83;
  border-radius: 12px;
  padding: 0.8em 0.85em 0.9em;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.52), inset 0 0 0 1px rgba(121, 192, 255, 0.08);
}

.settings-panel-header {
  margin-bottom: 0.65em;
}

.settings-panel-title {
  margin: 0;
  font-size: 0.96em;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  color: #d6ebff;
}

.settings-panel-subtitle {
  margin: 0.2em 0 0;
  color: #8cb0d6;
  font-size: 0.78em;
}

.settings-content {
  border: 1px solid #2d4058;
  border-radius: 10px;
  background: rgba(8, 15, 25, 0.78);
  padding: 0.7em;
}

.settings-overview {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65em;
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 0.6em;
}

.settings-card {
  border: 1px solid #344b68;
  border-radius: 8px;
  background: rgba(12, 20, 32, 0.6);
  padding: 0.6em;
}

.settings-card-wide {
  grid-column: 1 / -1;
}

.settings-group-title {
  color: #89c9ff;
  font-size: 0.82em;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  margin: 0;
}

.settings-toggle-list {
  display: grid;
  gap: 0.45em;
}

.settings-toggle-item {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  font-size: 0.9em;
  color: #d4e7ff;
}

.setting-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.6em;
}

.setting-field-label {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.28em;
  font-size: 0.9em;
  color: #d4e7ff;
}

.file-picker {
  border: 1px solid #3b5677;
  border-radius: 8px;
  background: rgba(12, 20, 32, 0.78);
  padding: 0.55em 0.6em;
  display: grid;
  gap: 0.4em;
}

.input-control {
  background: #0d1218;
  color: white;
  border: 1px solid var(--line-strong);
  border-radius: 6px;
  padding: 0.28em 0.45em;
  width: 100%;
  box-sizing: border-box;
}

.drop-hint {
  font-size: 0.78em;
  color: var(--text-muted);
}

.btn-load-path {
  margin-top: 0.4em;
  padding: 0.3em 0.8em;
  background: #1e3a5f;
  color: white;
  border: 1px solid #3b5677;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85em;
}

.btn-load-path:hover:not(:disabled) {
  background: #2a5080;
}

.btn-load-path:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.config-buttons {
  display: flex;
  gap: 0.5em;
  margin-top: 0.1em;
}

.error-text {
  color: #ff7777;
  font-size: 0.85em;
}

.success-text {
  color: #66ff99;
  font-size: 0.85em;
}

@media (max-width: 900px) {
  .settings-panel {
    right: auto;
    left: 0;
    width: min(560px, calc(100vw - 2em));
  }
  .settings-overview {
    grid-template-columns: 1fr;
  }
  .settings-card-wide {
    grid-column: auto;
  }
  .setting-field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
