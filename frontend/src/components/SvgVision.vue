<script setup lang="ts">
import SvgRobot from '@/components/SvgRobot.vue'
import SvgBall from '@/components/SvgBall.vue'
import type { SSL_DetectionFrame } from '@/proto/vision/ssl_vision_detection_pb.ts'
import { useGrSimReplacement } from '@/composables/grsim'

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

async function onBallDragEnd(x: number, y: number) {
  try {
    await replaceBall(x, y)
  } catch (error) {
    console.error('Failed to replace ball:', error)
  }
}

async function onRobotDragEnd(
  x: number,
  y: number,
  orientation: number,
  id: number,
  yellowTeam: boolean
) {
  try {
    // orientationをラジアンから度に変換
    const dirDeg = (orientation * 180) / Math.PI
    await replaceRobot(x, y, dirDeg, id, yellowTeam)
  } catch (error) {
    console.error('Failed to replace robot:', error)
  }
}
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
    @drag-end="onBallDragEnd"
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
    @drag-end="(x, y, orientation) => onRobotDragEnd(x, y, orientation, s.robotId, true)"
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
    @drag-end="(x, y, orientation) => onRobotDragEnd(x, y, orientation, s.robotId, false)"
  />
</template>
