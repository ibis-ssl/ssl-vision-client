<script setup lang="ts">
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import { computed, inject, type Ref } from 'vue'

interface Props {
  referee: Referee
  activeSource: string
  sources: Record<string, string>
  visionConnected: boolean
  refereeConnected: boolean
  grsimConnected: boolean
}

const props = defineProps<Props>()

interface Emits {
  (e: 'update:activeSource', source: string): void
}

const emit = defineEmits<Emits>()

// 選択状態を取得
const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject', { value: null } as Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>)

// ゲームステージの文字列変換
const stageText = computed(() => {
  const stageMap: { [key: number]: string } = {
    0: 'NORMAL_FIRST_HALF',
    1: 'NORMAL_SECOND_HALF',
    2: 'EXTRA_FIRST_HALF',
    3: 'EXTRA_SECOND_HALF',
    4: 'PENALTY_SHOOTOUT',
    5: 'PENALTY_SHOOTOUT_BREAK',
    6: 'NORMAL_FIRST_HALF_PRE',
    7: 'NORMAL_SECOND_HALF_PRE',
    8: 'EXTRA_FIRST_HALF_PRE',
    9: 'EXTRA_SECOND_HALF_PRE',
    10: 'PENALTY_SHOOTOUT_POST',
    11: 'POST_GAME',
    12: 'EXTRA_TIME_BREAK',
  }
  return stageMap[props.referee.stage] || `STAGE_${props.referee.stage}`
})

// コマンドの文字列変換
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

// 時間の変換（マイクロ秒から秒へ）
const timeText = computed(() => {
  if (props.referee.stageTimeLeft === undefined) return '--:--'
  const seconds = Math.floor(Number(props.referee.stageTimeLeft) / 1000000)
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${minutes}:${secs.toString().padStart(2, '0')}`
})

// スコア
const blueScore = computed(() => props.referee.blue?.score ?? 0)
const yellowScore = computed(() => props.referee.yellow?.score ?? 0)

// チーム名
const blueName = computed(() => props.referee.blue?.name || '')
const yellowName = computed(() => props.referee.yellow?.name || '')

// 選択状態の表示テキスト
const selectedText = computed(() => {
  if (!selectedObject.value) return '選択なし'

  if (selectedObject.value.type === 'ball') {
    return 'ボール'
  }

  if (selectedObject.value.type === 'robot') {
    const team = selectedObject.value.team === 'YELLOW' ? 'イエロー' : 'ブルー'
    return `${team} ロボット #${selectedObject.value.id}`
  }

  return '選択なし'
})

// ソース切替
const sourceOptions = computed(() => Object.entries(props.sources))

function changeSource(source: string) {
  emit('update:activeSource', source)
}
</script>

<template>
  <div id="status-bar">
    <!-- 接続状態 -->
    <div class="info-section connection-status">
      <span class="label">接続:</span>
      <span class="connection-indicator" :class="{ connected: visionConnected }" title="Vision">V</span>
      <span class="connection-indicator" :class="{ connected: refereeConnected }" title="Referee">R</span>
      <span class="connection-indicator" :class="{ connected: grsimConnected }" title="grSim">G</span>
    </div>

    <!-- ソース切替 -->
    <div class="info-section source-selector">
      <span class="label">ソース:</span>
      <select :value="props.activeSource" @change="changeSource(($event.target as HTMLSelectElement).value)" class="source-select">
        <option v-for="[key, name] in sourceOptions" :key="key" :value="key">
          {{ name }}
        </option>
      </select>
    </div>

    <div class="info-section">
      <span class="value">{{ stageText }}</span>
    </div>
    <div class="info-section">
      <span class="value command">{{ commandText }}</span>
    </div>
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
    <div class="info-section">
      <span class="value time">{{ timeText }}</span>
    </div>
    <div class="info-section selection">
      <span class="label">選択:</span>
      <span class="value selection-text" :class="{ selected: selectedObject }">
        {{ selectedText }}
      </span>
    </div>
  </div>
</template>

<style scoped>
#status-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2em;
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

.info-section.selection {
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

/* 接続状態インジケータ */
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
  transition: background-color 0.3s;
}

.connection-indicator.connected {
  background-color: #44ff44;
  color: black;
}

/* ソース切替 */
.source-selector {
  background-color: rgba(50, 50, 50, 0.8);
  padding: 0.3em 0.8em;
  border-radius: 4px;
  border: 1px solid #444;
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

.source-select:hover {
  border-color: #888;
}

.source-select:focus {
  outline: none;
  border-color: #00ff88;
}
</style>
