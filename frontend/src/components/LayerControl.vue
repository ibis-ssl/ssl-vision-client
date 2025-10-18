<script setup lang="ts">
export interface LayerVisibility {
  ball: boolean
  referee: boolean
}

const layers = defineModel<LayerVisibility>({ required: true })

const toggleLayer = (layerName: keyof LayerVisibility) => {
  layers.value[layerName] = !layers.value[layerName]
}
</script>

<template>
  <div id="layer-control">
    <span class="control-label">レイヤー:</span>
    <label class="layer-item">
      <input
        type="checkbox"
        :checked="layers.ball"
        @change="toggleLayer('ball')"
      />
      <span>ボール</span>
    </label>
    <label class="layer-item">
      <input
        type="checkbox"
        :checked="layers.referee"
        @change="toggleLayer('referee')"
      />
      <span>レフェリー情報</span>
    </label>
  </div>
</template>

<style scoped>
#layer-control {
  display: flex;
  align-items: center;
  gap: 1em;
  padding: 0.5em 1em;
  background-color: rgba(0, 0, 0, 0.7);
  color: white;
  width: 100%;
  position: absolute;
  top: 0;
  left: 0;
  z-index: 100;
}

.control-label {
  font-weight: bold;
  margin-right: 0.5em;
}

.layer-item {
  display: flex;
  align-items: center;
  gap: 0.3em;
  cursor: pointer;
  user-select: none;
}

.layer-item input[type="checkbox"] {
  cursor: pointer;
}

.layer-item span {
  font-size: 0.9em;
}
</style>
