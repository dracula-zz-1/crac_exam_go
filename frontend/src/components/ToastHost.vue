<script setup lang="ts">
import { useToastStore } from "@/stores/toast";
import AppIcon from "./AppIcon.vue";

const toast = useToastStore();

const iconFor = (tone: string) =>
  tone === "success" ? "check" : tone === "error" ? "x" : tone === "warning" ? "warning" : "info";
</script>

<template>
  <div class="toast-host">
    <transition-group name="toast">
      <div
        v-for="item in toast.items"
        :key="item.id"
        class="toast-item"
        :class="`tone-${item.tone}`"
      >
        <div class="toast-icon">
          <AppIcon :name="iconFor(item.tone)" :size="16" :stroke-width="2.2" />
        </div>
        <div class="toast-body">
          <strong>{{ item.title }}</strong>
          <p v-if="item.message">{{ item.message }}</p>
        </div>
        <button class="toast-close" @click="toast.dismiss(item.id)">
          <AppIcon name="x" :size="13" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-host {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 200;
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: min(360px, calc(100vw - 48px));
  pointer-events: none;
}
.toast-item {
  pointer-events: auto;
  display: flex;
  align-items: flex-start;
  gap: 11px;
  padding: 13px 14px;
  border-radius: 14px;
  background: rgba(16, 24, 32, 0.96);
  color: #f4f0e6;
  box-shadow: var(--shadow-lg);
  border: 1px solid rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
}
.toast-icon {
  width: 28px;
  height: 28px;
  border-radius: 9px;
  display: grid;
  place-items: center;
  flex: none;
  margin-top: 1px;
}
.tone-success .toast-icon { background: rgba(77, 155, 106, 0.22); color: #7fd39d; }
.tone-error .toast-icon { background: rgba(208, 92, 77, 0.22); color: #ff9c8d; }
.tone-warning .toast-icon { background: rgba(232, 163, 61, 0.22); color: #f3c069; }
.tone-info .toast-icon { background: rgba(75, 127, 181, 0.22); color: #8bb8e4; }
.toast-body {
  flex: 1;
  min-width: 0;
}
.toast-body strong {
  display: block;
  font-size: 12.8px;
  font-weight: 650;
  line-height: 1.4;
}
.toast-body p {
  margin-top: 3px;
  font-size: 11.5px;
  line-height: 1.55;
  color: rgba(244, 240, 230, 0.62);
  word-break: break-word;
}
.toast-close {
  color: rgba(244, 240, 230, 0.4);
  padding: 3px;
  border-radius: 6px;
  flex: none;
}
.toast-close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}
.toast-enter-active,
.toast-leave-active {
  transition: all 0.24s var(--ease);
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
