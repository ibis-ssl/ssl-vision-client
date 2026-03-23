<script setup lang="ts">
import { useToast } from '@/composables/toast'

const { toasts, removeToast } = useToast()
</script>

<template>
  <TransitionGroup name="toast" tag="div" class="toast-container">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="toast-item"
      :class="[`toast-level-${toast.level}`, `toast-cat-${toast.category}`]"
      @click="removeToast(toast.id)"
    >
      <div class="toast-header">
        <span class="toast-category">{{ toast.category }}</span>
        <span class="toast-title">{{ toast.title }}</span>
        <button class="toast-close" @click.stop="removeToast(toast.id)">&times;</button>
      </div>
      <div v-if="toast.message" class="toast-message">{{ toast.message }}</div>
    </div>
  </TransitionGroup>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  max-width: 360px;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  background: rgba(8, 14, 22, 0.95);
  border: 1px solid #2b3d55;
  border-radius: 6px;
  padding: 0.5em 0.75em;
  font-family: monospace;
  font-size: 0.82rem;
  color: #e8f0fc;
  cursor: pointer;
  backdrop-filter: blur(8px);
  border-left-width: 3px;
  user-select: none;
}

.toast-item:hover {
  background: rgba(20, 32, 46, 0.98);
}

.toast-level-info    { border-left-color: #79c0ff; }
.toast-level-success { border-left-color: #56d364; }
.toast-level-warning { border-left-color: #e3b341; }
.toast-level-error   { border-left-color: #f85149; }

.toast-header {
  display: flex;
  align-items: center;
  gap: 0.4em;
}

.toast-category {
  font-size: 0.7rem;
  text-transform: uppercase;
  color: #6e8caa;
  letter-spacing: 0.05em;
  flex-shrink: 0;
}

.toast-title {
  flex: 1;
  font-weight: bold;
  color: #c9d7e8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: #6e8caa;
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0;
}

.toast-close:hover {
  color: #e8f0fc;
}

.toast-message {
  margin-top: 0.2em;
  color: #9bb3cf;
  font-size: 0.78rem;
  white-space: pre-wrap;
  word-break: break-word;
}

/* TransitionGroup アニメーション */
.toast-enter-active {
  transition: opacity 0.25s ease-out, transform 0.25s ease-out;
}
.toast-leave-active {
  transition: opacity 0.2s ease-in, transform 0.2s ease-in;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(110%);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(110%);
}
.toast-move {
  transition: transform 0.25s ease;
}
</style>
