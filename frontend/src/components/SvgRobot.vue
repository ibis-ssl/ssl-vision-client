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
  return {
    stroke: 'black',
    strokeWidth: isSelected.value ? 0.01 : 0.005,
    strokeOpacity: 1,
    fill: props.teamColor == 'YELLOW' ? 'yellow' : 'blue',
    cursor: props.draggable ? 'pointer' : 'default',
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
    <svg-text :x="x" :y="y" :text="robotId" :color="teamColor == 'YELLOW' ? 'black' : 'white'" />
  </g>
</template>
