<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSettings } from '@/composables/settings'
import type { LayerVisibility } from '@/components/LayerControl.vue'

const { config, loading, error, successMessage, fetchConfig, updateConfig } = useSettings()

// Props for layer and source control
const props = defineProps<{
  sources: { [key: string]: string }
  mode: 'live' | 'replay'
  replayLoaded: boolean
  replayFileName: string
}>()
const emit = defineEmits<{
  (e: 'update:mode', mode: 'live' | 'replay'): void
  (e: 'replay-file-selected', file: File): void
}>()

const layerVisibility = defineModel<LayerVisibility>('layerVisibility', { required: true })
const activeSource = defineModel<string>('activeSource', { required: true })

const isOpen = ref(false)
const visionPort = ref(10006)
const trackedPort = ref(10010)
const refereePort = ref(10003)
const grSimAddress = ref('127.0.0.1')
const grSimPort = ref(20011)

onMounted(async () => {
  // 設定を取得するが、値は手動で開いた時のみ更新
  await fetchConfig()
})

const togglePanel = () => {
  if (!isOpen.value) {
    // パネルを開く時に設定値を読み込む
    visionPort.value = config.value.visionPort
    trackedPort.value = config.value.trackedPort
    refereePort.value = config.value.refereePort
    grSimAddress.value = config.value.grSimAddress || '127.0.0.1'
    grSimPort.value = config.value.grSimPort || 20011
  }
  isOpen.value = !isOpen.value
}

const handleSave = async () => {
  await updateConfig({
    visionPort: visionPort.value,
    trackedPort: trackedPort.value,
    refereePort: refereePort.value,
    grSimAddress: grSimAddress.value,
    grSimPort: grSimPort.value,
  })
}

const handleReset = () => {
  visionPort.value = config.value.visionPort
  trackedPort.value = config.value.trackedPort
  refereePort.value = config.value.refereePort
  grSimAddress.value = config.value.grSimAddress || '127.0.0.1'
  grSimPort.value = config.value.grSimPort || 20011
}

const toggleLayer = (layerName: keyof LayerVisibility) => {
  layerVisibility.value[layerName] = !layerVisibility.value[layerName]
}

const updateActiveSource = (sourceId: string) => {
  activeSource.value = sourceId
}

const updateMode = (mode: 'live' | 'replay') => {
  emit('update:mode', mode)
}

const onReplayFileSelected = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    emit('replay-file-selected', file)
  }
}
</script>

<template>
  <div id="settings-panel" :class="{ open: isOpen }">
    <button class="toggle-button" @click="togglePanel" :title="isOpen ? '設定を閉じる' : '設定を開く'">
      {{ isOpen ? '×' : '⚙' }}
    </button>

    <div v-if="isOpen" class="panel-content">
      <h3>設定</h3>

      <!-- レイヤー設定 -->
      <div class="section">
        <div class="section-title">レイヤー</div>
        <div class="layer-controls">
          <label class="layer-item">
            <input
              type="checkbox"
              :checked="layerVisibility.ball"
              @change="toggleLayer('ball')"
            />
            <span>Ball</span>
          </label>
          <label class="layer-item">
            <input
              type="checkbox"
              :checked="layerVisibility.referee"
              @change="toggleLayer('referee')"
            />
            <span>Referee</span>
          </label>
          <label class="layer-item">
            <input
              type="checkbox"
              :checked="layerVisibility.fieldLines"
              @change="toggleLayer('fieldLines')"
            />
            <span>Field Lines</span>
          </label>
        </div>
      </div>

      <!-- ソース選択 -->
      <div class="section">
        <div class="section-title">Source</div>
        <div class="source-controls">
          <label
            v-for="[sourceId, sourceName] of Object.entries(props.sources)"
            :key="sourceId"
            class="source-item"
          >
            <input
              type="radio"
              :value="sourceId"
              name="source-selector"
              :checked="sourceId === activeSource"
              @click="updateActiveSource(sourceId)"
            />
            <span>{{ sourceName }}</span>
          </label>
        </div>
      </div>

      <div class="section">
        <div class="section-title">Data Mode</div>
        <div class="source-controls">
          <label class="source-item">
            <input
              type="radio"
              value="live"
              name="data-mode"
              :checked="props.mode === 'live'"
              @change="updateMode('live')"
            />
            <span>Live UDP</span>
          </label>
          <label class="source-item">
            <input
              type="radio"
              value="replay"
              name="data-mode"
              :checked="props.mode === 'replay'"
              @change="updateMode('replay')"
            />
            <span>Replay Log</span>
          </label>
        </div>
      </div>

      <div class="section">
        <div class="section-title">Replay File</div>
        <div class="setting-group">
          <input
            id="replay-file"
            name="replay-file"
            type="file"
            accept=".log,.gz,.log.gz"
            :disabled="loading"
            @change="onReplayFileSelected"
          />
          <div class="label-text" v-if="props.replayFileName">
            読込中ファイル: {{ props.replayFileName }}
          </div>
          <div class="label-text" v-else-if="props.replayLoaded">
            ログを読み込み済み
          </div>
          <div class="label-text" v-else>
            .log または .log.gz を選択
          </div>
        </div>
      </div>

      <!-- ポート設定 -->
      <div class="section">
        <div class="section-title">Port</div>

        <div v-if="error" class="message error">{{ error }}</div>
        <div v-if="successMessage" class="message success">{{ successMessage }}</div>

        <div class="setting-group">
        <label for="vision-port">
          <span class="label-text">Vision</span>
          <div class="input-group">
            <span class="ip-display">{{ config.visionIP }}:</span>
            <input
              id="vision-port"
              name="vision-port"
              type="number"
              v-model.number="visionPort"
              min="1"
              max="65535"
              :disabled="loading"
            />
          </div>
        </label>
      </div>

      <div class="setting-group">
        <label for="tracked-port">
          <span class="label-text">Tracked</span>
          <div class="input-group">
            <span class="ip-display">{{ config.trackedIP }}:</span>
            <input
              id="tracked-port"
              name="tracked-port"
              type="number"
              v-model.number="trackedPort"
              min="1"
              max="65535"
              :disabled="loading"
            />
          </div>
        </label>
      </div>

      <div class="setting-group">
        <label for="referee-port">
          <span class="label-text">Referee</span>
          <div class="input-group">
            <span class="ip-display">{{ config.refereeIP }}:</span>
            <input
              id="referee-port"
              name="referee-port"
              type="number"
              v-model.number="refereePort"
              min="1"
              max="65535"
              :disabled="loading"
            />
          </div>
        </label>
      </div>
      </div>

      <!-- grSim設定 -->
      <div class="section">
        <div class="section-title">grSim</div>

        <div class="setting-group">
          <label for="grsim-address">
            <span class="label-text">Address</span>
            <input
              id="grsim-address"
              name="grsim-address"
              type="text"
              v-model="grSimAddress"
              :disabled="loading"
              placeholder="127.0.0.1"
            />
          </label>
        </div>

        <div class="setting-group">
          <label for="grsim-port">
            <span class="label-text">Port</span>
            <input
              id="grsim-port"
              name="grsim-port"
              type="number"
              v-model.number="grSimPort"
              min="1"
              max="65535"
              :disabled="loading"
            />
          </label>
        </div>

        <div class="button-group">
          <button @click="handleReset" :disabled="loading" class="btn-secondary">リセット</button>
          <button @click="handleSave" :disabled="loading" class="btn-primary">
            {{ loading ? '保存中...' : '保存して適用' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
#settings-panel {
  position: absolute;
  top: 3em;
  right: 1em;
  z-index: 200;
  background-color: rgba(0, 0, 0, 0.9);
  color: white;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  min-width: 280px;
}

.toggle-button {
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  font-size: 1.5em;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.toggle-button:hover {
  background-color: rgba(50, 50, 50, 0.9);
  transform: scale(1.1);
}

#settings-panel.open .toggle-button {
  position: absolute;
  top: 0.5em;
  right: 0.5em;
}

.panel-content {
  padding: 2.5em 1.5em 1.5em;
}

h3 {
  margin: 0 0 1em;
  font-size: 1.2em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
  padding-bottom: 0.5em;
}

.section {
  margin-bottom: 1.5em;
}

.section-title {
  font-size: 0.95em;
  font-weight: bold;
  margin-bottom: 0.5em;
  color: rgba(255, 255, 255, 0.9);
}

.layer-controls,
.source-controls {
  display: flex;
  flex-direction: column;
  gap: 0.5em;
}

.layer-item,
.source-item {
  display: flex;
  align-items: center;
  gap: 0.5em;
  cursor: pointer;
  user-select: none;
  padding: 0.3em;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.layer-item:hover,
.source-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.layer-item input[type='checkbox'],
.source-item input[type='radio'] {
  cursor: pointer;
}

.layer-item span,
.source-item span {
  font-size: 0.9em;
}

.setting-group {
  margin-bottom: 1em;
}

.label-text {
  display: block;
  font-size: 0.9em;
  margin-bottom: 0.3em;
  color: rgba(255, 255, 255, 0.8);
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.3em;
}

.ip-display {
  font-size: 0.9em;
  color: rgba(255, 255, 255, 0.6);
  font-family: monospace;
}

input[type='number'],
input[type='text'] {
  flex: 1;
  padding: 0.5em;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  color: white;
  font-size: 0.9em;
  font-family: monospace;
}

input[type='number']:focus,
input[type='text']:focus {
  outline: none;
  border-color: rgba(100, 150, 255, 0.8);
  background-color: rgba(255, 255, 255, 0.15);
}

input[type='number']:disabled,
input[type='text']:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.button-group {
  display: flex;
  gap: 0.5em;
  margin-top: 1.5em;
}

button {
  flex: 1;
  padding: 0.6em;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9em;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #4caf50;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #45a049;
}

.btn-secondary {
  background-color: rgba(255, 255, 255, 0.2);
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: rgba(255, 255, 255, 0.3);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.message {
  padding: 0.5em;
  margin-bottom: 1em;
  border-radius: 4px;
  font-size: 0.9em;
}

.message.error {
  background-color: rgba(244, 67, 54, 0.2);
  border: 1px solid rgba(244, 67, 54, 0.5);
  color: #ffcdd2;
}

.message.success {
  background-color: rgba(76, 175, 80, 0.2);
  border: 1px solid rgba(76, 175, 80, 0.5);
  color: #c8e6c9;
}
</style>
