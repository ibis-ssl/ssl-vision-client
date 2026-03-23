<script setup lang="ts">
import { GameEvent_Type } from '@/proto/gc/ssl_gc_game_event_pb.ts'
import type { GameEvent } from '@/proto/gc/ssl_gc_game_event_pb.ts'
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
const DISPLAY_DURATION = 10000

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

// ファールタイプごとの日本語名とアイコン（SSLルール準拠）
const foulInfo: Partial<Record<GameEvent_Type, { name: string; icon: string }>> = {
  // ボールアウト
  [GameEvent_Type.BALL_LEFT_FIELD_TOUCH_LINE]:           { name: 'タッチライン通過', icon: '🟡' },
  [GameEvent_Type.BALL_LEFT_FIELD_GOAL_LINE]:            { name: 'ゴールライン通過', icon: '🟡' },
  [GameEvent_Type.AIMLESS_KICK]:                         { name: 'エイムレスキック', icon: '🟡' },
  // 試合停止を伴うファール
  [GameEvent_Type.KEEPER_HELD_BALL]:                     { name: 'キーパーボール保持', icon: '🛑' },
  [GameEvent_Type.ATTACKER_DOUBLE_TOUCHED_BALL]:         { name: 'ダブルタッチ', icon: '🛑' },
  [GameEvent_Type.BOT_DRIBBLED_BALL_TOO_FAR]:            { name: 'オーバードリブル', icon: '🛑' },
  [GameEvent_Type.ATTACKER_TOO_CLOSE_TO_DEFENSE_AREA]:   { name: 'エリア接近', icon: '🛑' },
  [GameEvent_Type.BOT_PUSHED_BOT]:                       { name: 'プッシング', icon: '🛑' },
  [GameEvent_Type.BOT_HELD_BALL_DELIBERATELY]:           { name: 'ホールディング', icon: '🛑' },
  [GameEvent_Type.BOT_TIPPED_OVER]:                      { name: '転倒', icon: '🛑' },
  [GameEvent_Type.DEFENDER_IN_DEFENSE_AREA]:             { name: 'マルチプルDF', icon: '🛑' },
  [GameEvent_Type.BOUNDARY_CROSSING]:                    { name: '境界線横断', icon: '🛑' },
  [GameEvent_Type.BOT_DROPPED_PARTS]:                    { name: '部品脱落', icon: '🛑' },
  // 試合停止を伴わないファール
  [GameEvent_Type.ATTACKER_TOUCHED_BALL_IN_DEFENSE_AREA]: { name: 'エリア内タッチ', icon: '⚠️' },
  [GameEvent_Type.BOT_KICKED_BALL_TOO_FAST]:             { name: 'ボール速度超過', icon: '⚠️' },
  [GameEvent_Type.BOT_CRASH_DRAWN]:                      { name: '衝突(引分)', icon: '⚠️' },
  [GameEvent_Type.BOT_CRASH_UNIQUE]:                     { name: '衝突', icon: '⚠️' },
  // アウトオブプレイ中のファール
  [GameEvent_Type.BOT_INTERFERED_PLACEMENT]:             { name: '配置妨害', icon: '⚠️' },
  [GameEvent_Type.BOT_TOO_FAST_IN_STOP]:                 { name: 'ストップ中速度超過', icon: '⚠️' },
  [GameEvent_Type.DEFENDER_TOO_CLOSE_TO_KICK_POINT]:     { name: 'ボール接近', icon: '⚠️' },
  [GameEvent_Type.EXCESSIVE_BOT_SUBSTITUTION]:           { name: '交代回数超過', icon: '⚠️' },
  // ゴール関連
  [GameEvent_Type.POSSIBLE_GOAL]:                        { name: 'ゴール(確認中)', icon: '⚽' },
  [GameEvent_Type.GOAL]:                                 { name: 'ゴール', icon: '⚽' },
  [GameEvent_Type.INVALID_GOAL]:                         { name: '無効ゴール', icon: '❌' },
  // PK・試合停滞
  [GameEvent_Type.PENALTY_KICK_FAILED]:                  { name: 'PK失敗', icon: '❌' },
  [GameEvent_Type.NO_PROGRESS_IN_GAME]:                  { name: '試合の停滞', icon: '⏸️' },
  [GameEvent_Type.TOO_MANY_ROBOTS]:                      { name: 'ロボット数超過', icon: '🤖' },
  // 非推奨イベント（locationあり）
  [GameEvent_Type.INDIRECT_GOAL]:                        { name: '間接ゴール', icon: '⚽' },
  [GameEvent_Type.CHIPPED_GOAL]:                         { name: 'チップゴール', icon: '⚽' },
  [GameEvent_Type.KICK_TIMEOUT]:                         { name: 'キックTO', icon: '⏱️' },
  [GameEvent_Type.ATTACKER_TOUCHED_OPPONENT_IN_DEFENSE_AREA]:         { name: 'エリア内接触', icon: '⚠️' },
  [GameEvent_Type.ATTACKER_TOUCHED_OPPONENT_IN_DEFENSE_AREA_SKIPPED]: { name: 'エリア内接触 (スキップ)', icon: '⚠️' },
  [GameEvent_Type.BOT_CRASH_UNIQUE_SKIPPED]:             { name: '衝突 (スキップ)', icon: '⚠️' },
  [GameEvent_Type.BOT_PUSHED_BOT_SKIPPED]:               { name: 'プッシング (スキップ)', icon: '⚠️' },
  [GameEvent_Type.DEFENDER_IN_DEFENSE_AREA_PARTIALLY]:   { name: '部分的エリア侵入', icon: '🛑' },
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

  // ballLocationフィールドへのフォールバック（tooManyRobots等）
  if ('ballLocation' in eventData && eventData.ballLocation) {
    return {
      x: eventData.ballLocation.x ?? 0,
      y: -(eventData.ballLocation.y ?? 0),
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
  if ('byBot' in eventData) return eventData.byBot
  if ('kickingBot' in eventData) return eventData.kickingBot // ゴール系

  return undefined
}

// イベントから補足情報を取得
function getEventDetails(event: GameEvent): string | undefined {
  if (!event.event || !event.event.value) return undefined

  const eventData = event.event.value

  if ('initialBallSpeed' in eventData && eventData.initialBallSpeed !== undefined) {
    return `${eventData.initialBallSpeed.toFixed(1)}m/s`
  }
  if ('speed' in eventData && eventData.speed !== undefined) {
    return `${eventData.speed.toFixed(1)}m/s`
  }
  if ('crashSpeed' in eventData && eventData.crashSpeed !== undefined) {
    return `${eventData.crashSpeed.toFixed(1)}m/s`
  }
  if ('pushedDistance' in eventData && eventData.pushedDistance !== undefined) {
    return `${eventData.pushedDistance.toFixed(2)}m`
  }
  if ('start' in eventData && 'end' in eventData && eventData.start && eventData.end) {
    const dx = (eventData.end.x ?? 0) - (eventData.start.x ?? 0)
    const dy = (eventData.end.y ?? 0) - (eventData.start.y ?? 0)
    return `${Math.sqrt(dx * dx + dy * dy).toFixed(2)}m`
  }
  // distanceより先に判定: placementSucceededはdistanceも持つが配置時間を優先する
  if ('timeTaken' in eventData && eventData.timeTaken !== undefined) {
    const precision = 'precision' in eventData && eventData.precision !== undefined
      ? ` 精度${eventData.precision.toFixed(3)}m`
      : ''
    return `${eventData.timeTaken.toFixed(1)}秒${precision}`
  }
  if ('distance' in eventData && eventData.distance !== undefined) {
    return `${eventData.distance.toFixed(2)}m`
  }
  if ('maxBallHeight' in eventData && eventData.maxBallHeight !== undefined) {
    return `高さ${eventData.maxBallHeight.toFixed(2)}m`
  }
  if ('time' in eventData && eventData.time !== undefined) {
    return `${eventData.time.toFixed(1)}秒`
  }
  if ('remainingDistance' in eventData && eventData.remainingDistance !== undefined) {
    return `残り${eventData.remainingDistance.toFixed(2)}m`
  }
  if ('numRobotsOnField' in eventData && eventData.numRobotsOnField !== undefined) {
    return `${eventData.numRobotsOnField}/${eventData.numRobotsAllowed ?? '?'}台`
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
