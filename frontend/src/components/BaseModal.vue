<script setup lang="ts">
import { onBeforeUnmount, watch } from "vue";
import AppIcon from "./AppIcon.vue";

const props = withDefaults(
  defineProps<{
    open: boolean;
    title?: string;
    subtitle?: string;
    wide?: boolean;
    closeOnMask?: boolean;
  }>(),
  { wide: false, closeOnMask: true },
);

const emit = defineEmits<{ close: [] }>();

function onKey(e: KeyboardEvent) {
  if (e.key === "Escape") emit("close");
}

watch(
  () => props.open,
  (open) => {
    document.body.style.overflow = open ? "hidden" : "";
    if (open) window.addEventListener("keydown", onKey);
    else window.removeEventListener("keydown", onKey);
  },
);

onBeforeUnmount(() => {
  document.body.style.overflow = "";
  window.removeEventListener("keydown", onKey);
});

function maskClick() {
  if (props.closeOnMask) emit("close");
}
</script>

<template>
  <Teleport to="body">
    <transition name="modal">
      <div v-if="open" class="modal-mask" @click.self="maskClick">
        <div class="modal" :class="{ wide }" role="dialog" aria-modal="true">
          <div class="modal-head">
            <div>
              <h3 class="modal-title">{{ title }}</h3>
              <p v-if="subtitle" class="modal-sub">{{ subtitle }}</p>
            </div>
            <button class="icon-btn" @click="emit('close')">
              <AppIcon name="x" :size="17" />
            </button>
          </div>
          <div class="modal-body">
            <slot />
          </div>
          <div v-if="$slots.footer" class="modal-foot">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-active .modal,
.modal-leave-active .modal {
  transition: transform 0.22s var(--ease), opacity 0.2s;
}
.modal-enter-from .modal,
.modal-leave-to .modal {
  transform: translateY(10px) scale(0.98);
  opacity: 0;
}
</style>
