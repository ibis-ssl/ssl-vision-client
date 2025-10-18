<script setup lang="ts">
export interface LayerVisibility {
  ball: boolean
  referee: boolean
  fieldLines: boolean
  fouls: boolean
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
        id="layer-ball"
        type="checkbox"
        :checked="layers.ball"
        @change="toggleLayer('ball')"
      />
      <span>ボール</span>
    </label>
    <label class="layer-item">
      <input
        id="layer-referee"
        type="checkbox"
        :checked="layers.referee"
        @change="toggleLayer('referee')"
      />
      <span>レフェリー情報</span>
    </label>
    <label class="layer-item">
      <input
        id="layer-field-lines"
        type="checkbox"
        :checked="layers.fieldLines"
        @change="toggleLayer('fieldLines')"
      />
      <span>フィールドライン</span>
    </label>
    <label class="layer-item">
      <input
        id="layer-fouls"
        type="checkbox"
        :checked="layers.fouls"
        @change="toggleLayer('fouls')"
      />
      <span>ファール</span>
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
