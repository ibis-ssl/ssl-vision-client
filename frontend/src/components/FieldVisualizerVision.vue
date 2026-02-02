<script setup lang="ts">
import FieldVisualizer from '@/components/FieldVisualizer.vue'
import SvgVision from '@/components/SvgVision.vue'
import SvgReferee from '@/components/SvgReferee.vue'
import SvgTracked from '@/components/SvgTracked.vue'
import SvgBallPlacement from '@/components/SvgBallPlacement.vue'
import SvgGameEvents from '@/components/SvgGameEvents.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import RefereeInfo from '@/components/RefereeInfo.vue'
import type { LayerVisibility } from '@/components/LayerControl.vue'
import { computed, inject, provide, ref, type Ref } from 'vue'
import {
  useTrackedFrame,
  useTrackedSources,
  useVisionDetection,
  useVisionGeometry,
} from '@/composables/vision.ts'
import { useReferee } from '@/composables/referee.ts'

const activeSource = ref('vision')
const { field } = useVisionGeometry()
const { detectionFrame } = useVisionDetection(activeSource)
const { referee } = useReferee()
const { trackedFrame } = useTrackedFrame(activeSource)
const { trackerSources } = useTrackedSources()

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
</script>

<template>
  <div class="visualizer-container">
    <SettingsPanel
      :sources="sources"
      v-model:layer-visibility="layerVisibility"
      v-model:active-source="activeSource"
    />
    <div class="field-container" :class="{ 'with-referee-info': referee && layerVisibility.referee }">
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
    <RefereeInfo v-if="referee && layerVisibility.referee" :referee="referee" />
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
</style>
