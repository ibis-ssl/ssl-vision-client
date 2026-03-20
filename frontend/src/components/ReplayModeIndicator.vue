<script setup lang="ts">
defineProps<{ visible: boolean }>()
</script>

<template>
  <Transition name="fade">
    <div v-if="visible" class="replay-indicator" role="status" aria-label="Replay mode active">
      <span class="pulse-dot" aria-hidden="true" />
      <span class="label">REPLAY</span>
    </div>
  </Transition>
</template>

<style scoped>
.replay-indicator {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 50;
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  padding: 0.3em 0.75em 0.3em 0.55em;
  border-radius: 999px;
  background: rgba(200, 30, 30, 0.85);
  backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 100, 100, 0.5);
  pointer-events: none;
  user-select: none;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #fff;
  animation: pulse 1.4s ease-in-out infinite;
  flex-shrink: 0;
}

.label {
  font-family: monospace;
  font-size: 0.78em;
  font-weight: 800;
  letter-spacing: 0.12em;
  color: #fff;
  text-transform: uppercase;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.25; }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
