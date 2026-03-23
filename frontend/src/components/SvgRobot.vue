<script setup lang="ts">
import { computed, inject, type Ref } from 'vue'
import SvgText from '@/components/SvgText.vue'

const props = defineProps<{
  x: number
  y: number
  orientation: number
  id: number
  teamColor: 'YELLOW' | 'BLUE'
  draggable?: boolean
  visibility?: number
  isKicker?: boolean
}>()

const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject')!

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

const isSelected = computed(() => {
  return (
    selectedObject.value?.type === 'robot' &&
    selectedObject.value?.id === props.id &&
    selectedObject.value?.team === props.teamColor
  )
})

const style = computed(() => {
  const opacity = Math.max(props.visibility ?? 1, 0.15)
  return {
    stroke: 'black',
    strokeWidth: 0.005,
    strokeOpacity: opacity,
    fill: props.teamColor == 'YELLOW' ? 'yellow' : 'blue',
    fillOpacity: opacity,
    cursor: props.draggable ? 'pointer' : 'default',
  }
})

const highlightStyle = computed(() => {
  return {
    stroke: '#00ff88',
    strokeWidth: 0.015,
    fill: 'none',
  }
})

function onClick(event: MouseEvent) {
  if (props.draggable) {
    selectedObject.value = {
      type: 'robot',
      id: props.id,
      team: props.teamColor,
    }
    event.stopPropagation()
  }
}
</script>

<template>
  <g>
    <path :d="botShape" :style="style" @click="onClick" />
    <!-- 選択時のハイライトサークル -->
    <circle
      v-if="isSelected"
      :cx="x"
      :cy="-y"
      :r="radius * 1.3"
      :style="highlightStyle"
    />
    <!-- キッカーハイライト -->
    <circle
      v-if="isKicker"
      :cx="x"
      :cy="-y"
      :r="radius * 1.5"
      stroke="red"
      :stroke-width="0.008"
      stroke-dasharray="0.02 0.01"
      fill="none"
    />
    <svg-text :x="x" :y="y" :text="robotId" :color="teamColor == 'YELLOW' ? 'black' : 'white'" />
  </g>
</template>
