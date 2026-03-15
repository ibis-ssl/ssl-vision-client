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

watch(
  [
    () => referee.value?.commandCounter,
    () => referee.value?.command,
    () => config.value.autoBallPlacementEnabled,
    () => replayState.value.mode,
  ],
  async ([currentCommandCounter]) => {
    const currentReferee = referee.value
    if (!currentReferee || currentCommandCounter === undefined) {
      return
    }
    if (!config.value.autoBallPlacementEnabled || replayState.value.mode !== 'live') {
      return
    }
    if (lastAutoPlacementCommandCounter.value === currentCommandCounter) {
      return
    }

    const isBallPlacementCommand =
      currentReferee.command === Referee_Command.BALL_PLACEMENT_YELLOW ||
      currentReferee.command === Referee_Command.BALL_PLACEMENT_BLUE
    const isGoalCommand =
      currentReferee.command === Referee_Command.GOAL_YELLOW ||
      currentReferee.command === Referee_Command.GOAL_BLUE

    try {
      if (isBallPlacementCommand && currentReferee.designatedPosition) {
        const x = currentReferee.designatedPosition.x / 1000
        const y = currentReferee.designatedPosition.y / 1000
        await replaceBall(x, y)
        lastAutoPlacementCommandCounter.value = currentCommandCounter
      } else if (isGoalCommand) {
        await replaceBall(0, 0)
        lastAutoPlacementCommandCounter.value = currentCommandCounter
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
  document.addEventListener('dragenter', onWindowDragOver, true)
  document.addEventListener('dragover', onWindowDragOver, true)
  document.addEventListener('drop', onWindowDrop, true)
  window.addEventListener('dragover', onWindowDragOver, true)
  window.addEventListener('drop', onWindowDrop, true)
})

onUnmounted(() => {
  document.removeEventListener('dragenter', onWindowDragOver, true)
  document.removeEventListener('dragover', onWindowDragOver, true)
  document.removeEventListener('drop', onWindowDrop, true)
  window.removeEventListener('dragover', onWindowDragOver, true)
  window.removeEventListener('drop', onWindowDrop, true)
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
