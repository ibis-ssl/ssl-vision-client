<script setup lang="ts">
import type { GameEvent, GameEvent_Type } from '@/proto/gc/ssl_gc_game_event_pb.ts'
import type { Team } from '@/proto/gc/ssl_gc_common_pb.ts'
import { computed, inject, onMounted, onUnmounted, ref, type Ref } from 'vue'

const props = defineProps<{
  gameEvents: GameEvent[]
}>()

// フィールド回転の状態を取得
const rotateField = inject<Ref<boolean>>('rotate-field')!

// テキストの回転用transform関数
function textTransform(x: number, y: number) {
  if (rotateField.value) {
    return `rotate(-90,${x},${y})`
  }
  return ''
}

// 表示期間（ミリ秒）
const DISPLAY_DURATION = 5000

// 現在時刻を定期的に更新
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  timer = setInterval(() => {
    now.value = Date.now()
  }, 100)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})

// 表示すべきイベントをフィルタリング
const visibleEvents = computed(() => {
  return props.gameEvents.filter((event) => {
    if (!event.createdTimestamp) return false
    const eventTime = Number(event.createdTimestamp) / 1000 // マイクロ秒からミリ秒へ
    const elapsed = now.value - eventTime
    return elapsed >= 0 && elapsed < DISPLAY_DURATION
  })
})

// ファールタイプごとの日本語名とアイコン
const foulInfo: Record<number, { name: string; icon: string }> = {
  22: { name: 'CRASH', icon: '💥' }, // BOT_CRASH_UNIQUE
  21: { name: 'CRASH', icon: '💥' }, // BOT_CRASH_DRAWN
  24: { name: 'PUSH', icon: '👊' }, // BOT_PUSHED_BOT
  31: { name: 'エリア内', icon: '🚫' }, // DEFENDER_IN_DEFENSE_AREA
  19: { name: '接近', icon: '⚠️' }, // ATTACKER_TOO_CLOSE_TO_DEFENSE_AREA
  17: { name: 'ドリブル', icon: '🏃' }, // BOT_DRIBBLED_BALL_TOO_FAR
  18: { name: '速度超過', icon: '⚡' }, // BOT_KICKED_BALL_TOO_FAST
  28: { name: 'STOP超過', icon: '🚦' }, // BOT_TOO_FAST_IN_STOP
  29: { name: 'キック点', icon: '📍' }, // DEFENDER_TOO_CLOSE_TO_KICK_POINT
}

// チームカラーを取得
function getTeamColor(team?: Team): string {
  if (team === undefined) return '#888888'
  // Team enum: UNKNOWN=0, YELLOW=1, BLUE=2
  return team === 1 ? '#ffcc00' : team === 2 ? '#0066cc' : '#888888'
}

// イベントから位置情報を取得
function getEventLocation(event: GameEvent): { x: number; y: number } | null {
  if (!event.event || !event.event.value) return null

  const eventData = event.event.value
  if ('location' in eventData && eventData.location) {
    // GameEventのlocationはすでにメートル単位 [m] で来ている
    return {
      x: eventData.location.x ?? 0,
      y: -(eventData.location.y ?? 0), // Y軸を反転
    }
  }

  return null
}

// イベントからチームを取得
function getEventTeam(event: GameEvent): Team | undefined {
  if (!event.event || !event.event.value) return undefined

  const eventData = event.event.value
  if ('byTeam' in eventData) {
    return eventData.byTeam
  }

  return undefined
}

// イベントからボット番号を取得
function getEventBot(event: GameEvent): number | undefined {
  if (!event.event || !event.event.value) return undefined

  const eventData = event.event.value
  if ('byBot' in eventData) {
    return eventData.byBot
  }

  return undefined
}

// イベントから補足情報を取得
function getEventDetails(event: GameEvent): string | undefined {
  if (!event.event || !event.event.value) return undefined

  const eventData = event.event.value

  // 速度超過系
  if ('initialBallSpeed' in eventData && eventData.initialBallSpeed !== undefined) {
    return `${eventData.initialBallSpeed.toFixed(1)}m/s`
  }
  if ('speed' in eventData && eventData.speed !== undefined) {
    return `${eventData.speed.toFixed(1)}m/s`
  }

  // 距離系（既にメートル単位 [m]）
  if ('distance' in eventData && eventData.distance !== undefined) {
    return `${eventData.distance.toFixed(2)}m`
  }

  // ドリブル距離（start/endは既にメートル単位 [m]）
  if ('start' in eventData && 'end' in eventData && eventData.start && eventData.end) {
    const dx = (eventData.end.x ?? 0) - (eventData.start.x ?? 0)
    const dy = (eventData.end.y ?? 0) - (eventData.start.y ?? 0)
    const dist = Math.sqrt(dx * dx + dy * dy)
    return `${dist.toFixed(2)}m`
  }

  // 押し出し距離（既にメートル単位 [m]）
  if ('pushedDistance' in eventData && eventData.pushedDistance !== undefined) {
    return `${eventData.pushedDistance.toFixed(2)}m`
  }

  // 衝突速度
  if ('crashSpeed' in eventData && eventData.crashSpeed !== undefined) {
    return `${eventData.crashSpeed.toFixed(1)}m/s`
  }

  return undefined
}

// フェードアウト用の透明度を計算
function getOpacity(event: GameEvent): number {
  if (!event.createdTimestamp) return 1
  const eventTime = Number(event.createdTimestamp) / 1000
  const elapsed = now.value - eventTime
  const remaining = DISPLAY_DURATION - elapsed

  // 最後の1秒でフェードアウト
  if (remaining < 1000) {
    return remaining / 1000
  }

  return 1
}
</script>

<template>
  <g v-for="event in visibleEvents" :key="event.id">
    <g v-if="getEventLocation(event)" :opacity="getOpacity(event)">
      <!-- 背景円 -->
      <circle
        :cx="getEventLocation(event)!.x"
        :cy="getEventLocation(event)!.y"
        r="0.04"
        :fill="getTeamColor(getEventTeam(event))"
        stroke="black"
        stroke-width="0.002"
        opacity="0.8"
      />

      <!-- アイコン（絵文字） -->
      <text
        :x="getEventLocation(event)!.x"
        :y="getEventLocation(event)!.y"
        :transform="textTransform(getEventLocation(event)!.x, getEventLocation(event)!.y)"
        text-anchor="middle"
        dominant-baseline="central"
        font-size="0.12"
        fill="white"
        style="pointer-events: none; user-select: none"
      >
        {{ foulInfo[event.type]?.icon || '⚠️' }}
      </text>

      <!-- ファール名 -->
      <text
        :x="getEventLocation(event)!.x + (rotateField ? 0.14 : 0)"
        :y="getEventLocation(event)!.y + (rotateField ? 0 : 0.15)"
        :transform="textTransform(getEventLocation(event)!.x + (rotateField ? 0.14 : 0), getEventLocation(event)!.y + (rotateField ? 0 : 0.15))"
        text-anchor="middle"
        dominant-baseline="central"
        font-size="0.10"
        font-weight="bold"
        fill="white"
        stroke="black"
        stroke-width="0.004"
        style="paint-order: stroke; pointer-events: none; user-select: none"
      >
        {{ foulInfo[event.type]?.name || 'ファール' }}
      </text>

      <!-- ボット番号（存在する場合） -->
      <text
        v-if="getEventBot(event) !== undefined"
        :x="getEventLocation(event)!.x + (rotateField ? 0.24 : 0)"
        :y="getEventLocation(event)!.y + (rotateField ? 0 : 0.28)"
        :transform="textTransform(getEventLocation(event)!.x + (rotateField ? 0.24 : 0), getEventLocation(event)!.y + (rotateField ? 0 : 0.28))"
        text-anchor="middle"
        dominant-baseline="central"
        font-size="0.08"
        fill="white"
        stroke="black"
        stroke-width="0.003"
        style="paint-order: stroke; pointer-events: none; user-select: none"
      >
        #{{ getEventBot(event) }}
      </text>

      <!-- 補足情報（速度、距離など） -->
      <text
        v-if="getEventDetails(event)"
        :x="getEventLocation(event)!.x + (rotateField ? (getEventBot(event) !== undefined ? 0.34 : 0.24) : 0)"
        :y="getEventLocation(event)!.y + (rotateField ? 0 : (getEventBot(event) !== undefined ? 0.40 : 0.28))"
        :transform="textTransform(getEventLocation(event)!.x + (rotateField ? (getEventBot(event) !== undefined ? 0.34 : 0.24) : 0), getEventLocation(event)!.y + (rotateField ? 0 : (getEventBot(event) !== undefined ? 0.40 : 0.28)))"
        text-anchor="middle"
        dominant-baseline="central"
        font-size="0.08"
        font-weight="bold"
        fill="#ffff00"
        stroke="black"
        stroke-width="0.003"
        style="paint-order: stroke; pointer-events: none; user-select: none"
      >
        {{ getEventDetails(event) }}
      </text>
    </g>
  </g>
</template>
