<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSettings } from '@/composables/settings'

const { config, loading, error, successMessage, fetchConfig, updateConfig } = useSettings()

const isOpen = ref(false)
const visionPort = ref(10006)
const trackedPort = ref(10010)
const refereePort = ref(10003)

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
  }
  isOpen.value = !isOpen.value
}

const handleSave = async () => {
  await updateConfig({
    visionPort: visionPort.value,
    trackedPort: trackedPort.value,
    refereePort: refereePort.value,
  })
}

const handleReset = () => {
  visionPort.value = config.value.visionPort
  trackedPort.value = config.value.trackedPort
  refereePort.value = config.value.refereePort
}
</script>

<template>
  <div id="settings-panel" :class="{ open: isOpen }">
    <button class="toggle-button" @click="togglePanel" :title="isOpen ? '設定を閉じる' : '設定を開く'">
      {{ isOpen ? '×' : '⚙' }}
    </button>

    <div v-if="isOpen" class="panel-content">
      <h3>ポート設定</h3>

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

      <div class="button-group">
        <button @click="handleReset" :disabled="loading" class="btn-secondary">リセット</button>
        <button @click="handleSave" :disabled="loading" class="btn-primary">
          {{ loading ? '保存中...' : '保存して適用' }}
        </button>
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

input[type='number'] {
  flex: 1;
  padding: 0.5em;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  color: white;
  font-size: 0.9em;
  font-family: monospace;
}

input[type='number']:focus {
  outline: none;
  border-color: rgba(100, 150, 255, 0.8);
  background-color: rgba(255, 255, 255, 0.15);
}

input[type='number']:disabled {
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
