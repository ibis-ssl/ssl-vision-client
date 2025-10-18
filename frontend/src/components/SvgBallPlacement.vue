<script setup lang="ts">
import { computed } from 'vue'
import type { Referee } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import type { SSL_DetectionFrame } from '@/proto/vision/ssl_vision_detection_pb.ts'

const props = defineProps<{
  referee: Referee
  detectionFrame: SSL_DetectionFrame | null
}>()

// ボールプレイスメントコマンドかチェック
const isBallPlacement = computed(() => {
  return props.referee.command === 16 || props.referee.command === 17 // BALL_PLACEMENT_YELLOW or BLUE
})

// 目的地の座標
const designatedPosition = computed(() => {
  if (!props.referee.designatedPosition) return null
  return {
    x: props.referee.designatedPosition.x / 1000, // mm to m
    y: -props.referee.designatedPosition.y / 1000, // y軸を反転
  }
})

// ボールの現在位置（最も信頼度の高いボールを使用）
const ballPosition = computed(() => {
  if (!props.detectionFrame?.balls || props.detectionFrame.balls.length === 0) {
    return null
  }
  // 最も信頼度の高いボール（通常は最初のボール）
  const ball = props.detectionFrame.balls[0]
  if (!ball || ball.x === undefined || ball.y === undefined) {
    return null
  }
  return {
    x: ball.x / 1000, // mm to m
    y: -ball.y / 1000, // y軸を反転
  }
})

// カプセルを描画するかどうか
const shouldDrawCapsule = computed(() => {
  return isBallPlacement.value && designatedPosition.value && ballPosition.value
})

// チームカラー (16=YELLOW, 17=BLUE)
const teamColor = computed(() => {
  return props.referee.command === 17 ? 'blue' : 'yellow'
})
</script>

<template>
  <g v-if="shouldDrawCapsule">
    <!-- カプセル形状：ボールから目的地までの線分を太く、両端を丸く -->
    <line
      :x1="ballPosition!.x"
      :y1="ballPosition!.y"
      :x2="designatedPosition!.x"
      :y2="designatedPosition!.y"
      :stroke="teamColor === 'blue' ? '#0066cc' : '#ffcc00'"
      stroke-width="1.0"
      stroke-opacity="0.6"
      stroke-linecap="round"
    />
  </g>
</template>

<style scoped></style>
