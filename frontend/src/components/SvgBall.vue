<script setup lang="ts">
import { computed, inject, type CSSProperties, type Ref } from 'vue'

const props = defineProps<{
  x: number
  y: number
  height: number
  draggable?: boolean
}>()

const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject')!

const radius = computed(() => {
  return 0.021 * (1 + 0.01 * props.height)
})

const isSelected = computed(() => {
  return selectedObject.value?.type === 'ball'
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
    strokeWidth: isSelected.value ? 0.02 : 0.01,
    fill: 'none',
    cursor: props.draggable ? 'pointer' : 'default'
  }
})

function onClick(event: MouseEvent) {
  if (props.draggable) {
    selectedObject.value = {
      type: 'ball',
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
      @click="onClick"
    />
  </g>
</template>
