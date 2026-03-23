<script setup lang="ts">
import type { ReplayState, ReplayAnnotation } from '@/composables/replay.ts'
import { formatTimeNs } from '@/utils/time'
import { computed, onBeforeUnmount, ref, watch } from 'vue'

interface Props {
  replayState: ReplayState
  loading: boolean
}

interface Emits {
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'seek', positionNs: number): void
  (e: 'step', delta: number): void
  (e: 'rate', rate: number): void
  (e: 'drag-state', isDragging: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const replayDuration = computed(() => Math.max(props.replayState.durationNs, 0))
const replayPosition = computed(() => Math.min(props.replayState.positionNs, replayDuration.value))

const replayAnnotations = computed<(ReplayAnnotation & { left: number })[]>(() => {
  const duration = replayDuration.value
  if (duration <= 0) return []
  return props.replayState.annotations.map((a) => ({
    ...a,
    left: (a.timestampNs / duration) * 100,
  }))
})

function togglePlayPause() {
  if (props.replayState.playing) {
    emit('pause')
  } else {
    emit('play')
  }
}

function onRateChange(event: Event) {
  emit('rate', Number((event.target as HTMLSelectElement).value))
}

// カスタムタイムライン
const timelineTrackRef = ref<HTMLDivElement | null>(null)
const isDragging = ref(false)

// 楽観的UI: ドラッグ中はサーバー応答を待たずローカル値を表示
const optimisticPositionNs = ref<number | null>(null)

const displayPosition = computed(() => {
  if (optimisticPositionNs.value !== null) {
    return Math.min(optimisticPositionNs.value, replayDuration.value)
  }
  return replayPosition.value
})

// ドラッグ終了後、次のAPIレスポンスが来たら楽観的値をリセット
watch(
  () => props.replayState.positionNs,
  () => {
    if (!isDragging.value) {
      optimisticPositionNs.value = null
    }
  },
)

// シークのスロットリング（50ms間隔でAPIコールを抑制）
const SEEK_THROTTLE_MS = 50
let pendingSeekPosition: number | null = null
let seekThrottleTimer: number | null = null

function emitThrottledSeek(positionNs: number) {
  pendingSeekPosition = positionNs
  if (seekThrottleTimer !== null) return

  emit('seek', positionNs)

  seekThrottleTimer = window.setTimeout(() => {
    seekThrottleTimer = null
    if (pendingSeekPosition !== null) {
      emit('seek', pendingSeekPosition)
      pendingSeekPosition = null
    }
  }, SEEK_THROTTLE_MS)
}

function flushPendingSeek() {
  if (seekThrottleTimer !== null) {
    clearTimeout(seekThrottleTimer)
    seekThrottleTimer = null
  }
  if (pendingSeekPosition !== null) {
    emit('seek', pendingSeekPosition)
    pendingSeekPosition = null
  }
}

onBeforeUnmount(() => {
  if (seekThrottleTimer !== null) {
    clearTimeout(seekThrottleTimer)
  }
})

function positionFromPointer(clientX: number): number {
  const el = timelineTrackRef.value
  if (!el || replayDuration.value <= 0) return 0
  const rect = el.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width))
  return Math.round(ratio * replayDuration.value)
}

function onTrackPointerDown(event: PointerEvent) {
  if (props.loading) return
  isDragging.value = true
  emit('drag-state', true)
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
  const pos = positionFromPointer(event.clientX)
  optimisticPositionNs.value = pos
  emit('seek', pos)
}

function onTrackPointerMove(event: PointerEvent) {
  if (!isDragging.value) return
  const pos = positionFromPointer(event.clientX)
  optimisticPositionNs.value = pos
  emitThrottledSeek(pos)
}

function onTrackPointerUp(event: PointerEvent) {
  isDragging.value = false
  emit('drag-state', false)
  ;(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId)
  flushPendingSeek()
}

const onTrackPointerCancel = onTrackPointerUp

const progressPercent = computed(() => {
  if (replayDuration.value <= 0) return 0
  return (displayPosition.value / replayDuration.value) * 100
})
</script>

<template>
  <div class="replay-bar" :class="{ loading }">
    <!-- ファイル情報 -->
    <div class="file-info" role="status" :aria-live="loading ? 'polite' : 'off'">
      <span class="replay-badge">REPLAY</span>
      <span v-if="loading" class="spinner" aria-hidden="true" />
      <span v-if="loading" class="loading-label">読み込み中...</span>
      <span v-else-if="replayState.fileName" class="file-name" :title="replayState.fileName">{{ replayState.fileName }}</span>
    </div>

    <!-- トランスポートコントロール -->
    <div class="transport" role="group" aria-label="Replay controls">
      <button class="icon-btn skip" title="Previous frame (←)" :disabled="loading" @click="emit('step', -1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polygon points="19 20 9 12 19 4 19 20" />
          <line x1="5" y1="19" x2="5" y2="5" />
        </svg>
      </button>
      <button class="icon-btn play-pause" :title="replayState.playing ? 'Pause (Space)' : 'Play (Space)'" :disabled="loading" @click="togglePlayPause">
        <!-- Pause -->
        <svg v-if="replayState.playing" viewBox="0 0 24 24" fill="currentColor">
          <rect x="6" y="4" width="4" height="16" rx="1" />
          <rect x="14" y="4" width="4" height="16" rx="1" />
        </svg>
        <!-- Play -->
        <svg v-else viewBox="0 0 24 24" fill="currentColor">
          <polygon points="5 3 19 12 5 21 5 3" />
        </svg>
      </button>
      <button class="icon-btn skip" title="Next frame (→)" :disabled="loading" @click="emit('step', 1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polygon points="5 4 15 12 5 20 5 4" />
          <line x1="19" y1="4" x2="19" y2="20" />
        </svg>
      </button>
    </div>

    <!-- タイムライン -->
    <div class="timeline-section">
      <div
        ref="timelineTrackRef"
        class="timeline-track"
        :class="{ dragging: isDragging, disabled: loading }"
        @pointerdown="onTrackPointerDown"
        @pointermove="onTrackPointerMove"
        @pointerup="onTrackPointerUp"
        @pointercancel="onTrackPointerCancel"
      >
        <div class="timeline-progress" :style="{ width: `${progressPercent}%` }" />
        <div class="timeline-thumb" :style="{ left: `${progressPercent}%` }" />
        <!-- アノテーションマーカー -->
        <span
          v-for="(ann, idx) in replayAnnotations"
          :key="`ann-${idx}-${ann.timestampNs}`"
          class="timeline-marker"
          :style="{ left: `${ann.left}%` }"
          :title="ann.label"
        />
      </div>
      <div class="time-display">
        <span class="current-time">{{ formatTimeNs(displayPosition) }}</span>
        <span class="time-sep">/</span>
        <span>{{ formatTimeNs(replayDuration) }}</span>
      </div>
    </div>

    <!-- レート選択 -->
    <label class="rate-wrap">
      <span class="rate-label">Rate</span>
      <select class="rate-select" :value="replayState.rate" :disabled="loading" @change="onRateChange">
        <option :value="0.25">0.25x</option>
        <option :value="0.5">0.5x</option>
        <option :value="1">1x</option>
        <option :value="1.5">1.5x</option>
        <option :value="2">2x</option>
      </select>
    </label>
  </div>
</template>

<style scoped>
.replay-bar {
  position: relative;
  display: grid;
  grid-template-columns: auto auto 1fr auto;
  align-items: center;
  gap: 1em;
  padding: 0.6em 1em;
  background: var(--md-sys-color-surface-container);
  backdrop-filter: var(--app-glass-blur);
  border-top: 1px solid var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-surface);
  width: 100%;
  box-sizing: border-box;
}

.replay-bar.loading .transport,
.replay-bar.loading .rate-wrap {
  opacity: 0.45;
}

.loading-label {
  font-size: 0.82em;
  color: var(--md-sys-color-primary);
  animation: pulse 1.2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

.file-info {
  display: flex;
  align-items: center;
  gap: 0.5em;
  min-width: 0;
  max-width: 220px;
}

.replay-badge {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--md-sys-color-on-secondary-container);
  background: var(--md-sys-color-secondary-container);
  border-radius: var(--md-sys-shape-corner-full);
  padding: 0.2em 0.65em;
  flex-shrink: 0;
  text-transform: uppercase;
}

.file-name {
  font-size: 0.82em;
  color: var(--md-sys-color-on-surface-variant);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.transport {
  display: flex;
  align-items: center;
  gap: 0.4em;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--md-sys-color-surface-container-high);
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: 50%;
  color: var(--md-sys-color-on-surface);
  cursor: pointer;
  padding: 0;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
  flex-shrink: 0;
}

.icon-btn.skip {
  width: 32px;
  height: 32px;
  min-width: 32px;
}

.icon-btn.skip svg {
  width: 16px;
  height: 16px;
}

.icon-btn.play-pause {
  width: 40px;
  height: 40px;
  min-width: 40px;
  background: var(--md-sys-color-primary-container);
  border-color: transparent;
  color: var(--md-sys-color-on-primary-container);
}

.icon-btn.play-pause svg {
  width: 18px;
  height: 18px;
}

.icon-btn:hover:not(:disabled) {
  background: var(--md-sys-color-surface-container-highest);
}

.icon-btn.play-pause:hover:not(:disabled) {
  background: var(--md-sys-color-primary);
  color: var(--md-sys-color-on-primary);
}

.icon-btn:disabled {
  opacity: 0.38;
  cursor: not-allowed;
}

.timeline-section {
  display: flex;
  flex-direction: column;
  gap: 0.3em;
  min-width: 0;
}

.timeline-track {
  position: relative;
  height: 4px;
  background: var(--md-sys-color-surface-container-highest);
  border-radius: var(--md-sys-shape-corner-full);
  cursor: pointer;
  transition: height var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
  margin: 10px 0;
}

.timeline-track:hover:not(.disabled),
.timeline-track.dragging {
  height: 6px;
}

.timeline-track.disabled {
  cursor: not-allowed;
}

.replay-bar.loading .timeline-track.disabled {
  animation: shimmer 1.5s ease-in-out infinite;
}

@keyframes shimmer {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 0.3; }
}

.timeline-progress {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: var(--md-sys-color-primary);
  border-radius: var(--md-sys-shape-corner-full);
  pointer-events: none;
}

.timeline-thumb {
  position: absolute;
  top: 50%;
  width: 12px;
  height: 12px;
  background: var(--md-sys-color-primary);
  border: 2px solid var(--md-sys-color-on-primary-container);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  transition: width var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard),
              height var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
  box-shadow: var(--md-sys-elevation-1);
}

.timeline-track:hover:not(.disabled) .timeline-thumb,
.timeline-track.dragging .timeline-thumb {
  width: 16px;
  height: 16px;
}

.timeline-marker {
  position: absolute;
  top: 50%;
  width: 2px;
  height: 10px;
  background: var(--app-color-warning);
  transform: translate(-1px, -50%);
  pointer-events: none;
  border-radius: 1px;
  opacity: 0.8;
}

.time-display {
  display: flex;
  align-items: center;
  gap: 0.25em;
  font-size: 0.82em;
  font-family: var(--md-sys-typescale-mono-font);
  color: var(--md-sys-color-on-surface-variant);
  font-variant-numeric: tabular-nums;
}

.current-time {
  color: var(--md-sys-color-primary);
  font-weight: 700;
}

.time-sep {
  color: var(--md-sys-color-outline);
}

.rate-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  color: var(--md-sys-color-on-surface-variant);
  font-size: 0.82em;
  padding: 0.25em 0.4em 0.25em 0.6em;
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-corner-full);
  background: var(--md-sys-color-surface-container-low);
}

.rate-label {
  flex-shrink: 0;
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: 0.04em;
}

.rate-select {
  border: none;
  background: transparent;
  color: var(--md-sys-color-on-surface);
  border-radius: var(--md-sys-shape-corner-full);
  padding: 0.1em 0.3em;
  font-family: var(--md-sys-typescale-mono-font);
  font-size: 0.85em;
  cursor: pointer;
}

.rate-select:disabled {
  opacity: 0.38;
  cursor: not-allowed;
}

.spinner {
  width: 0.9em;
  height: 0.9em;
  border: 2px solid var(--md-sys-color-primary-tint);
  border-top-color: var(--md-sys-color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 900px) {
  .replay-bar {
    grid-template-columns: 1fr;
    gap: 0.5em;
  }

  .file-info {
    max-width: 100%;
  }

  .transport {
    justify-content: center;
  }
}
</style>
