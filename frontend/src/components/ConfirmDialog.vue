<script setup lang="ts">
import BaseModal from "./BaseModal.vue";

withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
    danger?: boolean;
    loading?: boolean;
  }>(),
  { confirmText: "确定", cancelText: "取消", danger: false, loading: false },
);

const emit = defineEmits<{ confirm: []; cancel: [] }>();
</script>

<template>
  <BaseModal :open="open" :title="title" @close="emit('cancel')">
    <div class="confirm-body">
      <div class="confirm-icon" :class="{ danger }">
        <slot name="icon" />
      </div>
      <p v-if="message" class="confirm-message">{{ message }}</p>
      <slot />
    </div>
    <template #footer>
      <button class="btn btn-ghost" @click="emit('cancel')" :disabled="loading">
        {{ cancelText }}
      </button>
      <button
        class="btn"
        :class="danger ? 'btn-danger' : 'btn-accent'"
        @click="emit('confirm')"
        :disabled="loading"
      >
        <span v-if="loading" class="spinner" :class="{ dark: danger }" />
        {{ confirmText }}
      </button>
    </template>
  </BaseModal>
</template>

<style scoped>
.confirm-body {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.confirm-icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: var(--amber-soft);
  color: #9a6514;
  display: grid;
  place-items: center;
  flex: none;
}
.confirm-icon.danger {
  background: var(--red-soft);
  color: #a63b2e;
}
.confirm-message {
  font-size: 13.5px;
  line-height: 1.75;
  color: var(--ink-2);
  padding-top: 9px;
  white-space: pre-line;
}
</style>
