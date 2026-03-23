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
  gap: 0.5rem;
  max-width: 360px;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  background: var(--md-sys-color-surface-container-highest);
  border: 1px solid var(--app-glass-border);
  border-left-width: 3px;
  border-radius: var(--md-sys-shape-corner-extra-small);
  padding: 0.6em 0.85em;
  font-size: 0.82rem;
  color: var(--md-sys-color-on-surface);
  cursor: pointer;
  backdrop-filter: var(--app-glass-blur);
  box-shadow: var(--md-sys-elevation-3);
  user-select: none;
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.toast-item:hover {
  background: var(--md-sys-color-surface-container-high);
}

.toast-level-info    { border-left-color: var(--md-sys-color-primary); }
.toast-level-success { border-left-color: var(--app-color-success); }
.toast-level-warning { border-left-color: var(--app-color-warning); }
.toast-level-error   { border-left-color: var(--md-sys-color-error); }

.toast-header {
  display: flex;
  align-items: center;
  gap: 0.4em;
}

.toast-category {
  font-size: 0.68rem;
  font-weight: 500;
  text-transform: uppercase;
  color: var(--md-sys-color-on-surface-variant);
  letter-spacing: 0.06em;
  flex-shrink: 0;
}

.toast-title {
  flex: 1;
  font-weight: 600;
  color: var(--md-sys-color-on-surface);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: var(--md-sys-color-on-surface-variant);
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0.1em 0.2em;
  border-radius: var(--md-sys-shape-corner-extra-small);
  transition: background var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard),
              color var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-standard);
}

.toast-close:hover {
  background: rgba(196, 198, 208, 0.12);
  color: var(--md-sys-color-on-surface);
}

.toast-message {
  margin-top: 0.25em;
  color: var(--md-sys-color-on-surface-variant);
  font-size: 0.78rem;
  white-space: pre-wrap;
  word-break: break-word;
}

.toast-enter-active {
  transition: opacity var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-emphasized-decelerate),
              transform var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-spring);
}
.toast-leave-active {
  transition: opacity var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-emphasized-accelerate),
              transform var(--md-sys-motion-duration-short) var(--md-sys-motion-easing-emphasized-accelerate);
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
  transition: transform var(--md-sys-motion-duration-medium) var(--md-sys-motion-easing-standard);
}
</style>
