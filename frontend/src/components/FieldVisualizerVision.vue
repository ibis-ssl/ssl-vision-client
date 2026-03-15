<script setup lang="ts">
import FieldVisualizer from '@/components/FieldVisualizer.vue'
import SvgVision from '@/components/SvgVision.vue'
import SvgReferee from '@/components/SvgReferee.vue'
import SvgTracked from '@/components/SvgTracked.vue'
import SvgBallPlacement from '@/components/SvgBallPlacement.vue'
import SvgGameEvents from '@/components/SvgGameEvents.vue'
import StatusBar from '@/components/StatusBar.vue'
import type { LayerVisibility } from '@/components/LayerControl.vue'
import { computed, inject, onMounted, onUnmounted, provide, ref, watch, type Ref } from 'vue'
import {
  useTrackedFrame,
  useTrackedSources,
  useVisionDetection,
  useVisionGeometry,
} from '@/composables/vision.ts'
import { useReferee } from '@/composables/referee.ts'
import { useReplay } from '@/composables/replay.ts'
import { useGrSimReplacement } from '@/composables/grsim'
import { useSettings } from '@/composables/settings'
import { Referee_Command } from '@/proto/gc/ssl_gc_referee_message_pb'

const activeSource = ref('vision')
const { field } = useVisionGeometry()
const { detectionFrame, connected: visionConnected } = useVisionDetection(activeSource)
const { referee, connected: refereeConnected } = useReferee()
const { trackedFrame } = useTrackedFrame(activeSource)
const { trackerSources } = useTrackedSources()
const replay = useReplay()
const replayState = replay.state
const { replaceBall } = useGrSimReplacement()
const { config } = useSettings()

// grSim接続状態（将来の実装用に常にtrueと想定）
const grsimConnected = ref(true)

const svgVisionRef = ref<InstanceType<typeof SvgVision>>()

// 選択状態をこのコンポーネントで管理し、配下全体でprovide
const selectedObject = ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>(null)
provide('selectedObject', selectedObject)

function onMoveObject(x: number, y: number) {
  svgVisionRef.value?.onMoveObject(x, y)
}

const sources = computed(() => {
  return {
    vision: 'vision',
    ...trackerSources.value,
  }
})

// レイヤー表示状態の管理
const layerVisibility = ref<LayerVisibility>({
  ball: true,
  referee: true,
  fieldLines: true,
  fouls: true,
})

const lastAutoPlacementCommandCounter = ref<number | null>(null)
const lastAutoPlacementGoalEventKey = ref<string | null>(null)

function getLatestGoalEventKey(): string | null {
  const events = referee.value?.gameEvents
  if (!events || events.length === 0) return null
  for (let i = events.length - 1; i >= 0; i -= 1) {
    const event = events[i]
    if (!event) continue
    if (event.event.case === 'goal') {
      return event.id || `${event.createdTimestamp}`
    }
  }
  return null
}

watch(
  [
    () => referee.value?.commandCounter,
    () => referee.value?.command,
    () => referee.value?.gameEvents,
    () => config.value.autoBallPlacementEnabled,
    () => config.value.autoCenterAfterGoalEnabled,
    () => replayState.value.mode,
  ],
  async ([currentCommandCounter]) => {
    const currentReferee = referee.value
    if (!currentReferee || currentCommandCounter === undefined) {
      return
    }
    if (replayState.value.mode !== 'live') {
      return
    }
    const isSameCommandCounter = lastAutoPlacementCommandCounter.value === currentCommandCounter

    const isBallPlacementCommand =
      currentReferee.command === Referee_Command.BALL_PLACEMENT_YELLOW ||
      currentReferee.command === Referee_Command.BALL_PLACEMENT_BLUE
    const isGoalCommand =
      currentReferee.command === Referee_Command.GOAL_YELLOW ||
      currentReferee.command === Referee_Command.GOAL_BLUE
    const latestGoalEventKey = config.value.autoCenterAfterGoalEnabled ? getLatestGoalEventKey() : null

    try {
      if (
        config.value.autoBallPlacementEnabled &&
        !isSameCommandCounter &&
        isBallPlacementCommand &&
        currentReferee.designatedPosition
      ) {
        const x = currentReferee.designatedPosition.x / 1000
        const y = currentReferee.designatedPosition.y / 1000
        await replaceBall(x, y)
        lastAutoPlacementCommandCounter.value = currentCommandCounter
      } else if (
        config.value.autoCenterAfterGoalEnabled &&
        !isSameCommandCounter &&
        isGoalCommand
      ) {
        await replaceBall(0, 0)
        lastAutoPlacementCommandCounter.value = currentCommandCounter
      } else if (
        config.value.autoCenterAfterGoalEnabled &&
        latestGoalEventKey &&
        lastAutoPlacementGoalEventKey.value !== latestGoalEventKey
      ) {
        await replaceBall(0, 0)
        lastAutoPlacementGoalEventKey.value = latestGoalEventKey
      }
    } catch (error) {
      if (isGoalCommand) {
        console.error('Failed to move ball to center after GOAL:', error)
      } else {
        console.error('Failed to auto place ball:', error)
      }
    }
  },
  { immediate: true }
)

async function onModeUpdate(mode: 'live' | 'replay') {
  await replay.setMode(mode)
}

async function onReplayFileSelected(file: File) {
  await replay.uploadFile(file)
  if (replay.state.value.mode !== 'replay') {
    await replay.setMode('replay')
  }
}

function onToggleLayer(layerName: keyof LayerVisibility) {
  layerVisibility.value[layerName] = !layerVisibility.value[layerName]
}

function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  if (target.isContentEditable) return true
  if (target instanceof HTMLInputElement) {
    // Keep shortcuts active while range slider is focused.
    return target.type !== 'range'
  }
  const tag = target.tagName
  return tag === 'TEXTAREA' || tag === 'SELECT'
}

async function onWindowKeyDown(event: KeyboardEvent) {
  if (isTypingTarget(event.target)) return
  if (replayState.value.mode !== 'replay') return
  const seekDeltaNs = 5 * 1e9

  if (event.code === 'Space') {
    event.preventDefault()
    if (event.repeat) return
    if (replayState.value.playing) {
      await replay.pause()
    } else {
      await replay.play()
    }
    return
  }

  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    if (event.shiftKey) {
      const next = Math.max(0, replayState.value.positionNs - seekDeltaNs)
      await replay.seek(next)
      return
    }
    await replay.step(-1)
    return
  }

  if (event.key === 'ArrowRight') {
    event.preventDefault()
    if (event.shiftKey) {
      const max = Math.max(0, replayState.value.durationNs)
      const next = Math.min(max, replayState.value.positionNs + seekDeltaNs)
      await replay.seek(next)
      return
    }
    await replay.step(1)
  }
}

function getDroppedFile(event: DragEvent): File | null {
  const files = event.dataTransfer?.files
  if (!files || files.length === 0) return null
  return files[0] ?? null
}

function hasFilePayload(event: DragEvent): boolean {
  const types = event.dataTransfer?.types
  if (!types) return false
  return Array.from(types).includes('Files')
}

function onWindowDragOver(event: DragEvent) {
  if (!hasFilePayload(event)) return
  event.preventDefault()
  event.stopPropagation()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy'
  }
}

async function onWindowDrop(event: DragEvent) {
  if (!hasFilePayload(event)) return
  event.preventDefault()
  event.stopPropagation()
  const file = getDroppedFile(event)
  if (!file) return
  await onReplayFileSelected(file)
}

onMounted(() => {
  window.addEventListener('keydown', onWindowKeyDown)
  document.addEventListener('dragenter', onWindowDragOver, true)
  document.addEventListener('dragover', onWindowDragOver, true)
  document.addEventListener('drop', onWindowDrop, true)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onWindowKeyDown)
  document.removeEventListener('dragenter', onWindowDragOver, true)
  document.removeEventListener('dragover', onWindowDragOver, true)
  document.removeEventListener('drop', onWindowDrop, true)
})
</script>

<template>
  <div class="visualizer-container">
    <div
      class="field-container"
      :class="{
        'with-referee-info': true,
        'with-replay-controls': replayState.mode === 'replay',
      }"
    >
      <FieldVisualizer :field="field" :show-field-lines="layerVisibility.fieldLines" @move-object="onMoveObject">
        <SvgVision
          ref="svgVisionRef"
          v-if="detectionFrame"
          :detection-frame="detectionFrame"
          :show-ball="layerVisibility.ball"
        />
        <SvgReferee
          v-if="referee && layerVisibility.referee"
          :field="field"
          :referee="referee"
        />
        <SvgBallPlacement
          v-if="referee && layerVisibility.referee && detectionFrame"
          :referee="referee"
          :detection-frame="detectionFrame"
        />
        <SvgGameEvents
          v-if="referee && layerVisibility.fouls"
          :game-events="referee.gameEvents"
        />
        <SvgTracked v-if="trackedFrame" :tracked-frame="trackedFrame" />
      </FieldVisualizer>
    </div>
    <StatusBar
      :referee="referee"
      :active-source="activeSource"
      :sources="sources"
      :vision-connected="visionConnected"
      :referee-connected="refereeConnected"
      :grsim-connected="grsimConnected"
      :replay-state="replayState"
      :replay-loading="replay.loading.value"
      :layer-visibility="layerVisibility"
      @update:active-source="activeSource = $event"
      @update:mode="onModeUpdate"
      @replay-file-selected="onReplayFileSelected"
      @toggle-layer="onToggleLayer"
      @replay-play="replay.play"
      @replay-pause="replay.pause"
      @replay-seek="replay.seek"
      @replay-step="replay.step"
      @replay-rate="replay.setRate"
    />
  </div>
</template>

<style scoped>
.visualizer-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.field-container {
  width: 100%;
  height: 100%;
}

.field-container.with-referee-info {
  height: calc(100% - 4em);
}

.field-container.with-replay-controls {
  height: calc(100% - 8em);
}
</style>
