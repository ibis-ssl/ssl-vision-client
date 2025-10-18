<script setup lang="ts">

import {computed} from "vue";
import type {SSL_GeometryFieldSize} from "@/proto/vision/ssl_vision_geometry_pb.ts";
import type {TeamJson} from "@/proto/gc/ssl_gc_common_pb.ts";

const props = defineProps<{
  field: SSL_GeometryFieldSize
  teamColor: TeamJson
  positiveHalf: boolean
}>()

const fieldMaxX = computed(() => {
  return props.field.fieldLength / 2000
})
const goalMaxX = computed(() => {
  return fieldMaxX.value + props.field.goalDepth / 1000
})
const goalMaxY = computed(() => {
  return props.field.goalWidth / 2000
})
const xMultiplier = computed(() => {
  return props.positiveHalf ? 1 : -1
})

const style = computed(() => {
  // lineThicknessは通常10mm程度なので、最小値を0.01（10mm）に設定
  const thickness = Math.max((props.field.lineThickness || 10) / 1000, 0.01)
  return {
    stroke: props.teamColor == 'YELLOW' ? 'yellow' : 'blue',
    strokeWidth: thickness,
    strokeOpacity: 1,
    fill: 'none',
  }
})

</script>

<template>
  <line
    :x1="xMultiplier * goalMaxX"
    :y1="goalMaxY"
    :x2="xMultiplier * goalMaxX"
    :y2="-goalMaxY"
    :style="style"
  />
  <line
    :x1="xMultiplier * fieldMaxX"
    :y1="goalMaxY"
    :x2="xMultiplier * goalMaxX"
    :y2="goalMaxY"
    :style="style"
  />
  <line
    :x1="xMultiplier * fieldMaxX"
    :y1="-goalMaxY"
    :x2="xMultiplier * goalMaxX"
    :y2="-goalMaxY"
    :style="style"
  />
</template>

<style scoped>

</style>
