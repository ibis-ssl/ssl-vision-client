<script setup lang="ts">
import type { SSL_GeometryFieldSize } from '@/proto/vision/ssl_vision_geometry_pb.ts'
import { computed } from 'vue'

const props = defineProps<{
  field: SSL_GeometryFieldSize
}>()

const fieldMaxX = computed(() => {
  return props.field.fieldLength / 2000
})
const fieldMaxY = computed(() => {
  return props.field.fieldWidth / 2000
})
const penaltyAreaMinX = computed(() => {
  return fieldMaxX.value - props.field.penaltyAreaDepth / 1000
})
const penaltyAreaMaxY = computed(() => {
  return props.field.penaltyAreaWidth / 2000
})

const lineStyle = computed(() => {
  // lineThicknessは通常10mm程度なので、最小値を0.01（10mm）に設定
  const thickness = Math.max((props.field.lineThickness || 10) / 1000, 0.01)
  return {
    stroke: 'white',
    strokeWidth: thickness,
    strokeOpacity: 1,
    fill: 'none',
  }
})

// Generate SVG path for arc
const generateArcPath = (arc: any) => {
  const cx = arc.center.x / 1000
  const cy = arc.center.y / 1000
  const radius = arc.radius / 1000
  const a1 = arc.a1 || 0
  const a2 = arc.a2 || 0

  // a1, a2 are already in radians
  const startAngle = a1
  const endAngle = a2

  // Calculate the angle difference (handle wrap-around)
  let angleDiff = endAngle - startAngle
  if (angleDiff < 0) {
    angleDiff += 2 * Math.PI
  }

  // 完全な円（またはほぼ完全な円）の場合は、2つの半円に分割
  if (angleDiff >= 2 * Math.PI - 0.01) {
    // 2つの半円を使って完全な円を描画
    const x1 = cx + radius * Math.cos(startAngle)
    const y1 = cy + radius * Math.sin(startAngle)
    const midAngle = startAngle + Math.PI
    const xMid = cx + radius * Math.cos(midAngle)
    const yMid = cy + radius * Math.sin(midAngle)
    const x2 = cx + radius * Math.cos(endAngle)
    const y2 = cy + radius * Math.sin(endAngle)

    return `M ${x1} ${y1} A ${radius} ${radius} 0 0 1 ${xMid} ${yMid} A ${radius} ${radius} 0 0 1 ${x2} ${y2}`
  }

  // 通常の円弧の場合
  const x1 = cx + radius * Math.cos(startAngle)
  const y1 = cy + radius * Math.sin(startAngle)
  const x2 = cx + radius * Math.cos(endAngle)
  const y2 = cy + radius * Math.sin(endAngle)

  // Determine if it's a large arc (>180 degrees)
  const largeArcFlag = angleDiff > Math.PI ? 1 : 0

  return `M ${x1} ${y1} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`
}
</script>

<template>
  <!-- Draw field lines from geometry data if available -->
  <g v-if="field.fieldLines && field.fieldLines.length > 0">
    <line
      v-for="(line, index) in field.fieldLines"
      :key="'line-' + index"
      :x1="line.p1!.x / 1000"
      :y1="line.p1!.y / 1000"
      :x2="line.p2!.x / 1000"
      :y2="line.p2!.y / 1000"
      :style="lineStyle"
    />
  </g>

  <!-- Draw field arcs from geometry data if available -->
  <g v-if="field.fieldArcs && field.fieldArcs.length > 0">
    <path
      v-for="(arc, index) in field.fieldArcs"
      :key="'arc-' + index"
      :d="generateArcPath(arc)"
      :style="lineStyle"
    />
  </g>

  <!-- Fallback: Draw basic field lines if geometry data doesn't include lines -->
  <g v-if="!field.fieldLines || field.fieldLines.length === 0">
    <line :x1="fieldMaxX" :y1="fieldMaxY" :x2="fieldMaxX" :y2="-fieldMaxY" :style="lineStyle" />
    <line :x1="-fieldMaxX" :y1="fieldMaxY" :x2="-fieldMaxX" :y2="-fieldMaxY" :style="lineStyle" />
    <line :x1="fieldMaxX" :y1="fieldMaxY" :x2="-fieldMaxX" :y2="fieldMaxY" :style="lineStyle" />
    <line :x1="fieldMaxX" :y1="-fieldMaxY" :x2="-fieldMaxX" :y2="-fieldMaxY" :style="lineStyle" />

    <line :x1="0" :y1="-fieldMaxY" :x2="0" :y2="fieldMaxY" :style="lineStyle" />

    <circle :cx="0" :cy="0" :r="field.centerCircleRadius / 1000" :style="lineStyle" />

    <line
      :x1="penaltyAreaMinX"
      :y1="penaltyAreaMaxY"
      :x2="penaltyAreaMinX"
      :y2="-penaltyAreaMaxY"
      :style="lineStyle"
    />
    <line
      :x1="penaltyAreaMinX"
      :y1="penaltyAreaMaxY"
      :x2="fieldMaxX"
      :y2="penaltyAreaMaxY"
      :style="lineStyle"
    />
    <line
      :x1="penaltyAreaMinX"
      :y1="-penaltyAreaMaxY"
      :x2="fieldMaxX"
      :y2="-penaltyAreaMaxY"
      :style="lineStyle"
    />

    <line
      :x1="-penaltyAreaMinX"
      :y1="penaltyAreaMaxY"
      :x2="-penaltyAreaMinX"
      :y2="-penaltyAreaMaxY"
      :style="lineStyle"
    />
    <line
      :x1="-penaltyAreaMinX"
      :y1="penaltyAreaMaxY"
      :x2="-fieldMaxX"
      :y2="penaltyAreaMaxY"
      :style="lineStyle"
    />
    <line
      :x1="-penaltyAreaMinX"
      :y1="-penaltyAreaMaxY"
      :x2="-fieldMaxX"
      :y2="-penaltyAreaMaxY"
      :style="lineStyle"
    />
  </g>
</template>
