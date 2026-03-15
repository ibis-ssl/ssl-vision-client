<script setup lang="ts">
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import type { ReplayState } from '@/composables/replay.ts'
import type { LayerVisibility } from '@/components/LayerControl.vue'
import { useSettings } from '@/composables/settings'
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
  (e: 'toggle-layer', layerName: keyof LayerVisibility): void
  (e: 'replay-play'): void
  (e: 'replay-pause'): void
  (e: 'replay-seek', positionNs: number): void
  (e: 'replay-step', delta: number): void
  (e: 'replay-rate', rate: number): void
}

const emit = defineEmits<Emits>()
const { config, loading: settingsLoading, error, successMessage, fetchConfig, updateConfig } = useSettings()
type SettingsTab = 'layer' | 'replay' | 'network' | 'automation'

const visionPort = ref(10006)
const trackedPort = ref(10010)
const refereePort = ref(10003)
const grSimAddress = ref('127.0.0.1')
const grSimPort = ref(20011)
const autoBallPlacementEnabled = ref(false)
const settingsDrawerRef = ref<HTMLDetailsElement | null>(null)
const activeSettingsTab = ref<SettingsTab>('layer')
const settingsTabs: Array<{ id: SettingsTab; label: string }> = [
  { id: 'layer', label: 'Layer' },
  { id: 'replay', label: 'Replay' },
  { id: 'network', label: 'Network' },
  { id: 'automation', label: 'Automation' },
]

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
const replayDuration = computed(() => Math.max(props.replayState.durationNs, 0))
const replayPosition = computed(() => Math.min(props.replayState.positionNs, replayDuration.value))

const replayAnnotations = computed(() => {
  const duration = replayDuration.value
  if (duration <= 0) return []
  return props.replayState.annotations.map((a) => ({
    ...a,
    left: (a.timestampNs / duration) * 100,
  }))
})

function changeSource(source: string) {
  emit('update:activeSource', source)
}

function formatTimeNs(ns: number): string {
  const sec = Math.max(0, Math.floor(ns / 1e9))
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

function onSeek(event: Event) {
  const el = event.target as HTMLInputElement
  emit('replay-seek', Number(el.value))
}

function togglePlayPause() {
  if (props.replayState.playing) {
    emit('replay-pause')
  } else {
    emit('replay-play')
  }
}

function onRateChange(event: Event) {
  const el = event.target as HTMLSelectElement
  emit('replay-rate', Number(el.value))
}

function syncInputsFromConfig() {
  visionPort.value = config.value.visionPort
  trackedPort.value = config.value.trackedPort
  refereePort.value = config.value.refereePort
  grSimAddress.value = config.value.grSimAddress || '127.0.0.1'
  grSimPort.value = config.value.grSimPort || 20011
  autoBallPlacementEnabled.value = config.value.autoBallPlacementEnabled || false
}

async function onSaveConfig() {
  await updateConfig({
    visionPort: visionPort.value,
    trackedPort: trackedPort.value,
    refereePort: refereePort.value,
    grSimAddress: grSimAddress.value,
    grSimPort: grSimPort.value,
    autoBallPlacementEnabled: autoBallPlacementEnabled.value,
  })
}

function onReplayFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    emit('replay-file-selected', file)
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
  <div id="status-bar" :class="{ 'with-replay': replayMode }">
    <div class="main-row">
      <div class="info-section connection-status">
        <span class="label">接続:</span>
        <span class="connection-indicator" :class="{ connected: visionConnected }" title="Vision">V</span>
        <span class="connection-indicator" :class="{ connected: refereeConnected }" title="Referee">R</span>
        <span class="connection-indicator" :class="{ connected: grsimConnected }" title="grSim">G</span>
      </div>

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

          <div class="settings-tabs" role="tablist" aria-label="Settings tabs">
            <button
              v-for="tab in settingsTabs"
              :key="tab.id"
              class="settings-tab"
              :class="{ active: activeSettingsTab === tab.id }"
              role="tab"
              :aria-selected="activeSettingsTab === tab.id"
              :tabindex="activeSettingsTab === tab.id ? 0 : -1"
              type="button"
              @click="activeSettingsTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>

          <div class="settings-content" role="tabpanel">
            <div v-if="activeSettingsTab === 'layer'" class="settings-group">
              <span class="settings-group-title">Layer Visibility</span>
              <div class="settings-toggle-list">
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.ball" @change="emit('toggle-layer', 'ball')" /><span>Ball</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.referee" @change="emit('toggle-layer', 'referee')" /><span>Referee</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.fieldLines" @change="emit('toggle-layer', 'fieldLines')" /><span>Field</span></label>
                <label class="settings-toggle-item"><input type="checkbox" :checked="props.layerVisibility.fouls" @change="emit('toggle-layer', 'fouls')" /><span>Fouls</span></label>
              </div>
            </div>

            <div v-else-if="activeSettingsTab === 'replay'" class="settings-group">
              <span class="settings-group-title">Replay Source</span>
              <div class="settings-toggle-list">
                <label class="settings-toggle-item"><input type="radio" name="data-mode" value="live" :checked="props.replayState.mode === 'live'" @change="emit('update:mode', 'live')" /><span>Live</span></label>
                <label class="settings-toggle-item"><input type="radio" name="data-mode" value="replay" :checked="props.replayState.mode === 'replay'" @change="emit('update:mode', 'replay')" /><span>Replay</span></label>
              </div>
              <div class="file-picker">
                <label class="setting-field-label">
                  <span>Replay Log</span>
                  <input class="input-control" type="file" accept=".log,.gz,.log.gz" :disabled="settingsLoading || props.replayLoading" @change="onReplayFileSelected" />
                </label>
                <div class="drop-hint">Select .log / .log.gz file</div>
              </div>
            </div>

            <div v-else-if="activeSettingsTab === 'network'" class="settings-group">
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

            <div v-else class="settings-group">
              <span class="settings-group-title">Automation</span>
              <div class="settings-toggle-list">
                <label class="settings-toggle-item"><input type="checkbox" v-model="autoBallPlacementEnabled" /><span>Auto Ball Placement</span></label>
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

    <div v-if="replayMode || props.replayLoading" class="replay-row" :class="{ loading: props.replayLoading }">
      <div v-if="props.replayLoading" class="replay-loading-overlay" role="status" aria-live="polite">
        <span class="spinner large" aria-hidden="true" />
        <span class="overlay-text">ログファイルを読み込み中...</span>
      </div>
      <div class="replay-controls" role="group" aria-label="Replay controls">
        <button class="replay-button ghost" title="Previous frame" :disabled="props.replayLoading" @click="emit('replay-step', -1)">◀◀</button>
        <button class="replay-button primary" title="Play or pause" :disabled="props.replayLoading" @click="togglePlayPause">{{ props.replayState.playing ? 'Pause' : 'Play' }}</button>
        <button class="replay-button ghost" title="Next frame" :disabled="props.replayLoading" @click="emit('replay-step', 1)">▶▶</button>
        <label class="rate-select-wrap">
          <span>Rate</span>
          <select class="replay-rate" :value="props.replayState.rate" :disabled="props.replayLoading" @change="onRateChange">
            <option :value="0.25">0.25x</option>
            <option :value="0.5">0.5x</option>
            <option :value="1">1x</option>
            <option :value="1.5">1.5x</option>
            <option :value="2">2x</option>
          </select>
        </label>
      </div>

      <div class="timeline-wrap">
        <div class="timeline-markers">
          <span
            v-for="(annotation, idx) in replayAnnotations"
            :key="`ann-${idx}-${annotation.timestampNs}`"
            class="timeline-marker"
            :style="{ left: `${annotation.left}%` }"
            :title="annotation.label"
          />
        </div>
        <input
          class="timeline-slider"
          type="range"
          min="0"
          :max="replayDuration"
          :value="replayPosition"
          :disabled="props.replayLoading"
          @input="onSeek"
        />
      </div>

      <div class="replay-time">
        <span class="current-time">{{ formatTimeNs(replayPosition) }}</span>
        <span class="time-separator">/</span>
        <span>{{ formatTimeNs(replayDuration) }}</span>
      </div>
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

.replay-row {
  position: relative;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.9em;
  align-items: center;
  background: linear-gradient(180deg, rgba(10, 18, 29, 0.96), rgba(6, 12, 21, 0.96));
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 0.55em 0.65em;
}

.replay-row.loading .replay-controls,
.replay-row.loading .timeline-wrap,
.replay-row.loading .replay-time {
  opacity: 0.35;
}

.replay-loading-overlay {
  position: absolute;
  inset: 0.25em;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.7em;
  border-radius: 8px;
  background: rgba(5, 13, 24, 0.72);
  border: 1px solid #4da0ff;
  backdrop-filter: blur(2px);
}

.overlay-text {
  font-size: 0.98em;
  font-weight: 800;
  letter-spacing: 0.03em;
  color: #d9ecff;
  text-shadow: 0 0 8px rgba(77, 160, 255, 0.45);
}

.replay-controls {
  display: flex;
  align-items: center;
  gap: 0.45em;
  flex-wrap: wrap;
}

.replay-button {
  border: 1px solid transparent;
  color: white;
  border-radius: 999px;
  padding: 0.3em 0.9em;
  cursor: pointer;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.replay-button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.replay-button.ghost {
  background: rgba(12, 20, 32, 0.95);
  border-color: var(--line-strong);
}

.replay-button.primary {
  background: linear-gradient(180deg, #2a89ff, #1f6fd1);
  border-color: #4da0ff;
}

.replay-button:hover {
  filter: brightness(1.08);
}

.replay-rate {
  border: 1px solid var(--line-strong);
  background: rgba(10, 14, 18, 0.92);
  color: white;
  border-radius: 999px;
  padding: 0.28em 0.52em;
}

.rate-select-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  color: var(--text-muted);
  font-size: 0.82em;
  padding: 0.22em 0.35em 0.22em 0.5em;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(12, 20, 32, 0.8);
}

.timeline-wrap {
  position: relative;
  background: rgba(4, 9, 16, 0.95);
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 0.18em 0.45em;
}

.timeline-slider {
  width: 100%;
  accent-color: var(--accent);
}

.timeline-markers {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  pointer-events: none;
}

.timeline-marker {
  position: absolute;
  width: 2px;
  height: 10px;
  background: #ffd700;
  transform: translateX(-1px) translateY(-50%);
}

.replay-time {
  min-width: 110px;
  text-align: right;
  color: #d7e7fb;
  font-variant-numeric: tabular-nums;
  background: rgba(12, 20, 32, 0.8);
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 0.24em 0.55em;
}

.current-time {
  color: #9ad0ff;
  font-weight: 700;
}

.time-separator {
  color: #7f99b9;
  margin: 0 0.25em;
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

.settings-tabs {
  display: flex;
  gap: 0.35em;
  margin-bottom: 0.65em;
  overflow-x: auto;
  padding-bottom: 0.1em;
}

.settings-tab {
  border: 1px solid #334861;
  background: rgba(10, 18, 30, 0.85);
  color: #95b2d3;
  border-radius: 999px;
  padding: 0.28em 0.75em;
  font-size: 0.78em;
  font-weight: 700;
  letter-spacing: 0.02em;
  cursor: pointer;
  white-space: nowrap;
}

.settings-tab.active {
  color: #e8f4ff;
  border-color: #62b2ff;
  background: linear-gradient(180deg, rgba(55, 130, 216, 0.95), rgba(30, 90, 160, 0.95));
  box-shadow: 0 0 0 1px rgba(98, 178, 255, 0.35);
}

.settings-content {
  border: 1px solid #2d4058;
  border-radius: 10px;
  background: rgba(8, 15, 25, 0.78);
  padding: 0.7em;
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 0.6em;
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

.spinner {
  width: 0.9em;
  height: 0.9em;
  border: 2px solid rgba(184, 218, 255, 0.35);
  border-top-color: #b8daff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner.large {
  width: 1.2em;
  height: 1.2em;
  border-width: 3px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
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
  .replay-row {
    grid-template-columns: 1fr;
    gap: 0.5em;
  }
  .replay-time {
    text-align: left;
  }
  .settings-panel {
    right: auto;
    left: 0;
    width: min(560px, calc(100vw - 2em));
  }
  .setting-field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
