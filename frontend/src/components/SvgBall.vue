<script setup lang="ts">
import { computed, inject, ref, type CSSProperties } from 'vue'

const props = defineProps<{
  x: number
  y: number
  height: number
  draggable?: boolean
}>()

const emit = defineEmits<{
  dragEnd: [x: number, y: number]
}>()

const isDragMode = inject<{ value: boolean }>('isDragMode', { value: false })
const isDragging = ref(false)

const radius = computed(() => {
  return 0.021 * (1 + 0.01 * props.height)
})

const style = computed((): CSSProperties => {
  return {
    fill: 'orange',
    fillOpacity: 1,
  }
})
const highlightStyle = computed((): CSSProperties => {
  return {
    stroke: 'orange',
    strokeWidth: 0.01,
    fill: 'none',
    cursor: isDragMode.value && props.draggable ? 'move' : 'default'
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

      // Y座標を反転してemit
      emit('dragEnd', svgPt.x, -svgPt.y)
    }

    event.stopPropagation()
  }
}
</script>

<template>
  <g>
    <circle :cx="x" :cy="-y" :r="radius" :style="style" />

    <!-- ball highlighter -->
    <circle
      :cx="x"
      :cy="-y"
      :r="0.5"
      :style="highlightStyle"
      @mousedown="onMouseDown"
      @mouseup="onMouseUp"
    />
  </g>
</template>
