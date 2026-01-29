<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import SvgText from '@/components/SvgText.vue'

const props = defineProps<{
  x: number
  y: number
  orientation: number
  id: number
  teamColor: 'YELLOW' | 'BLUE'
  draggable?: boolean
}>()

const emit = defineEmits<{
  dragEnd: [x: number, y: number, orientation: number]
}>()

const isDragMode = inject<{ value: boolean }>('isDragMode', { value: false })
const isDragging = ref(false)

const center2Dribbler = 0.075
const radius = 0.09

const robotId = computed(() => {
  return String(props.id)
})

const botShape = computed(() => {
  const orient2CornerAngle = Math.acos(center2Dribbler / radius)
  const botRightX = props.x + Math.cos(-props.orientation + orient2CornerAngle) * radius
  const botRightY = -props.y + Math.sin(-props.orientation + orient2CornerAngle) * radius
  const botLeftX = props.x + Math.cos(-props.orientation - orient2CornerAngle) * radius
  const botLeftY = -props.y + Math.sin(-props.orientation - orient2CornerAngle) * radius

  return (
    `M ${botRightX} ${botRightY}` +
    ` A ${radius} ${radius} 0 1 1 ${botLeftX} ${botLeftY}` +
    ` L ${botRightX} ${botRightY}`
  )
})

const style = computed(() => {
  return {
    stroke: 'black',
    strokeWidth: 0.005,
    strokeOpacity: 1,
    fill: props.teamColor == 'YELLOW' ? 'yellow' : 'blue',
    cursor: isDragMode.value && props.draggable ? 'move' : 'default',
  }
})

function onMouseDown(event: MouseEvent) {
  if (isDragMode.value && props.draggable) {
    isDragging.value = true
    event.stopPropagation()
  }
}

function onMouseUp(event: MouseEvent) {
  if (isDragging.value) {
    isDragging.value = false

    // SVG座標を取得
    const svg = (event.target as SVGElement).ownerSVGElement
    if (svg) {
      const pt = svg.createSVGPoint()
      pt.x = event.clientX
      pt.y = event.clientY
      const svgPt = pt.matrixTransform(svg.getScreenCTM()?.inverse())

      // Y座標を反転してemit（orientationは現在の値を維持）
      emit('dragEnd', svgPt.x, -svgPt.y, props.orientation)
    }

    event.stopPropagation()
  }
}
</script>

<template>
  <g>
    <path :d="botShape" :style="style" @mousedown="onMouseDown" @mouseup="onMouseUp" />
    <svg-text :x="x" :y="y" :text="robotId" :color="teamColor == 'YELLOW' ? 'black' : 'white'" />
  </g>
</template>
