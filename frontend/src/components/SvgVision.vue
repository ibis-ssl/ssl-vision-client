<script setup lang="ts">
import SvgRobot from '@/components/SvgRobot.vue'
import SvgBall from '@/components/SvgBall.vue'
import type { SSL_DetectionFrame } from '@/proto/vision/ssl_vision_detection_pb.ts'
import { useGrSimReplacement } from '@/composables/grsim'
import { inject, type Ref } from 'vue'

const props = withDefaults(
  defineProps<{
    detectionFrame: SSL_DetectionFrame
    showBall?: boolean
  }>(),
  {
    showBall: true,
  }
)

const { replaceBall, replaceRobot } = useGrSimReplacement()

const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject')!

async function onMoveObject(x: number, y: number) {
  if (!selectedObject.value) return

  try {
    if (selectedObject.value.type === 'ball') {
      await replaceBall(x, y)
    } else if (selectedObject.value.type === 'robot') {
      // ロボットの現在の向きを取得
      const robot =
        selectedObject.value.team === 'YELLOW'
          ? props.detectionFrame.robotsYellow.find((r) => r.robotId === selectedObject.value?.id)
          : props.detectionFrame.robotsBlue.find((r) => r.robotId === selectedObject.value?.id)

      if (robot) {
        // orientationをラジアンから度に変換
        const dirDeg = (robot.orientation * 180) / Math.PI
        await replaceRobot(
          x,
          y,
          dirDeg,
          selectedObject.value.id!,
          selectedObject.value.team === 'YELLOW'
        )
      }
    }
  } catch (error) {
    console.error('Failed to move object:', error)
  }
}

// 親コンポーネントから呼び出せるようにexposeする
defineExpose({
  onMoveObject,
})
</script>

<template>
  <SvgBall
    v-if="props.showBall"
    v-for="(s, i) in detectionFrame.balls"
    :key="'ball-' + i"
    :x="s.x / 1000"
    :y="s.y / 1000"
    :height="0"
    :draggable="true"
  />
  <SvgRobot
    v-for="(s, i) in detectionFrame.robotsYellow"
    :key="'robot-yellow-' + i"
    :x="s.x / 1000"
    :y="s.y / 1000"
    :orientation="s.orientation"
    :id="s.robotId"
    :team-color="'YELLOW'"
    :draggable="true"
  />
  <SvgRobot
    v-for="(s, i) in detectionFrame.robotsBlue"
    :key="'robot-yellow-' + i"
    :x="s.x / 1000"
    :y="s.y / 1000"
    :orientation="s.orientation"
    :id="s.robotId"
    :team-color="'BLUE'"
    :draggable="true"
  />
</template>
