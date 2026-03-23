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
  padding: 0.35em 0.85em 0.35em 0.65em;
  border-radius: var(--md-sys-shape-corner-full);
  background: var(--md-sys-color-error-container);
  backdrop-filter: var(--app-glass-blur);
  border: 1px solid var(--app-glass-border);
  box-shadow: var(--md-sys-elevation-2);
  pointer-events: none;
  user-select: none;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--md-sys-color-on-error-container);
  animation: pulse 1.4s ease-in-out infinite;
  flex-shrink: 0;
}

.label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--md-sys-color-on-error-container);
  text-transform: uppercase;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-standard);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
