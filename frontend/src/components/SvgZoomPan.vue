<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

const svg = ref<SVGElement>()
const zoom = ref(1.0)
const translation = ref({ x: 0, y: 0 })
const activeTranslation = ref({ x: 0, y: 0 })
const mouseDownPoint = ref<{
  x: number
  y: number
} | null>(null)
const lastMousePosition = ref<{
  clientX: number
  clientY: number
  target: EventTarget | null
} | null>(null)

// 親から提供されるselectedObjectをinject
const selectedObject = inject<Ref<{
  type: 'robot' | 'ball'
  id?: number
  team?: 'YELLOW' | 'BLUE'
} | null>>('selectedObject')!

// 移動イベントの定義
const emit = defineEmits<{
  moveObject: [x: number, y: number]
}>()

function onScroll(event: WheelEvent) {
  const x = (event.offsetX - translation.value.x) / zoom.value
  const y = (event.offsetY - translation.value.y) / zoom.value

  const newZoom = zoom.value - event.deltaY / 150
  if (newZoom < 1) {
    zoom.value = 1
    translation.value = { x: 0, y: 0 }
  } else {
    const dz = newZoom - zoom.value
    translation.value.x -= x * dz
    translation.value.y -= y * dz
    zoom.value = newZoom
  }

  event.preventDefault()
}

function onMouseMove(event: MouseEvent) {
  // マウス位置を記録（エンターキーでの移動に使用）
  lastMousePosition.value = {
    clientX: event.clientX,
    clientY: event.clientY,
    target: event.target,
  }

  if (mouseDownPoint.value !== null) {
    activeTranslation.value = {
      x: event.clientX - mouseDownPoint.value.x,
      y: event.clientY - mouseDownPoint.value.y,
    }
  }
}

function onMouseDown(event: MouseEvent) {
  mouseDownPoint.value = { x: event.clientX, y: event.clientY }
}

function onMouseUp() {
  if (mouseDownPoint.value !== null) {
    translation.value = {
      x: translation.value.x + activeTranslation.value.x,
      y: translation.value.y + activeTranslation.value.y,
    }
    activeTranslation.value = { x: 0, y: 0 }
    mouseDownPoint.value = null
  }
}

function moveSelectedObject(clientX: number, clientY: number, target?: EventTarget | null) {
  if (selectedObject.value !== null) {
    // SVG座標を取得（イベントターゲットからownerSVGElementを取得）
    let svgElement: SVGSVGElement | null = null

    if (target && target instanceof SVGElement) {
      svgElement = target.ownerSVGElement
    }

    // フォールバック：svg.value を使用
    if (!svgElement) {
      svgElement = svg.value as SVGSVGElement | null
    }

    if (svgElement) {
      const pt = svgElement.createSVGPoint()
      pt.x = clientX
      pt.y = clientY
      const ctm = svgElement.getScreenCTM()
      if (ctm) {
        const svgPt = pt.matrixTransform(ctm.inverse())
        // Y座標を反転してemit
        emit('moveObject', svgPt.x, -svgPt.y)
      }
    }
  }
}

function onKeyDown(event: KeyboardEvent) {
  if (event.key === ' ') {
    zoom.value = 1
    translation.value = { x: 0, y: 0 }
  }
  if (event.key === 'Escape') {
    selectedObject.value = null
  }
  if (event.key === 'Enter' && lastMousePosition.value) {
    moveSelectedObject(
      lastMousePosition.value.clientX,
      lastMousePosition.value.clientY,
      lastMousePosition.value.target
    )
    event.preventDefault()
  }
}

function onDoubleClick(event: MouseEvent) {
  moveSelectedObject(event.clientX, event.clientY, event.target)
  event.preventDefault()
}

const transform = computed(() => {
  const translationX = translation.value.x + activeTranslation.value.x
  const translationY = translation.value.y + activeTranslation.value.y
  return `translate(${translationX}, ${translationY}) scale(${zoom.value})`
})

onMounted(() => {
  document.addEventListener('keydown', onKeyDown)
  svg.value?.addEventListener('wheel', onScroll)
  svg.value?.addEventListener('mousemove', onMouseMove)
  svg.value?.addEventListener('mousedown', onMouseDown)
  svg.value?.addEventListener('mouseup', onMouseUp)
  svg.value?.addEventListener('dblclick', onDoubleClick)
})
onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeyDown)
  svg.value?.removeEventListener('wheel', onScroll)
  svg.value?.removeEventListener('mousemove', onMouseMove)
  svg.value?.removeEventListener('mousedown', onMouseDown)
  svg.value?.removeEventListener('mouseup', onMouseUp)
  svg.value?.removeEventListener('dblclick', onDoubleClick)
})
</script>

<template>
  <svg width="100%" height="100%" ref="svg">
    <g :transform="transform">
      <slot />
    </g>
  </svg>
</template>
