<script setup lang="ts">
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import type { ReplayState } from '@/composables/replay.ts'
import { computed, inject, type Ref } from 'vue'

interface Props {
  referee: Referee
  activeSource: string
  sources: Record<string, string>
  visionConnected: boolean
  refereeConnected: boolean
  grsimConnected: boolean
  replayState: ReplayState
}

const props = defineProps<Props>()

interface Emits {
  (e: 'update:activeSource', source: string): void
  (e: 'replay-play'): void
  (e: 'replay-pause'): void
  (e: 'replay-seek', positionNs: number): void
  (e: 'replay-step', delta: number): void
  (e: 'replay-rate', rate: number): void
}

const emit = defineEmits<Emits>()

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
  if (props.referee.stageTimeLeft === undefined) return '--:--'
  const seconds = Math.floor(Number(props.referee.stageTimeLeft) / 1000000)
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${minutes}:${secs.toString().padStart(2, '0')}`
})

const blueScore = computed(() => props.referee.blue?.score ?? 0)
const yellowScore = computed(() => props.referee.yellow?.score ?? 0)
const blueName = computed(() => props.referee.blue?.name || '')
const yellowName = computed(() => props.referee.yellow?.name || '')

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
    </div>

    <div v-if="replayMode" class="replay-row">
      <div class="replay-controls">
        <button class="replay-button" @click="emit('replay-step', -1)">◀◀</button>
        <button class="replay-button" @click="togglePlayPause">{{ props.replayState.playing ? 'Pause' : 'Play' }}</button>
        <button class="replay-button" @click="emit('replay-step', 1)">▶▶</button>
        <select class="replay-rate" :value="props.replayState.rate" @change="onRateChange">
          <option :value="0.25">0.25x</option>
          <option :value="0.5">0.5x</option>
          <option :value="1">1x</option>
          <option :value="1.5">1.5x</option>
          <option :value="2">2x</option>
        </select>
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
          @input="onSeek"
        />
      </div>

      <div class="replay-time">
        {{ formatTimeNs(replayPosition) }} / {{ formatTimeNs(replayDuration) }}
      </div>
    </div>
  </div>
</template>

<style scoped>
#status-bar {
  display: flex;
  flex-direction: column;
  gap: 0.6em;
  padding: 0.8em 1em;
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.8);
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
  gap: 1.5em;
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
  color: #aaa;
  font-size: 0.9em;
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
  background-color: rgba(50, 50, 50, 0.8);
  padding: 0.3em 0.8em;
  border-radius: 4px;
  border: 1px solid #444;
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
  background-color: rgba(30, 30, 30, 0.9);
  color: white;
  border: 1px solid #666;
  border-radius: 3px;
  padding: 0.2em 0.5em;
  font-family: monospace;
  font-size: 0.9em;
  cursor: pointer;
}

.replay-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.8em;
  align-items: center;
}

.replay-controls {
  display: flex;
  align-items: center;
  gap: 0.4em;
}

.replay-button {
  border: 1px solid #666;
  background: #222;
  color: white;
  border-radius: 4px;
  padding: 0.2em 0.6em;
  cursor: pointer;
}

.replay-rate {
  border: 1px solid #666;
  background: #222;
  color: white;
  border-radius: 4px;
  padding: 0.2em;
}

.timeline-wrap {
  position: relative;
}

.timeline-slider {
  width: 100%;
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
  min-width: 90px;
  text-align: right;
  color: #ddd;
}

@media (max-width: 900px) {
  .replay-row {
    grid-template-columns: 1fr;
    gap: 0.5em;
  }
  .replay-time {
    text-align: left;
  }
}
</style>
