<script setup lang="ts">
import AppIcon from "./AppIcon.vue";

const props = withDefaults(
  defineProps<{
    options: Array<{ key: string; text: string }>;
    selected: string[];
    revealed?: boolean;
    correct?: string;
    disabled?: boolean;
    multi?: boolean;
  }>(),
  { revealed: false, disabled: false, multi: false },
);

const emit = defineEmits<{ select: [key: string] }>();

function classes(key: string) {
  const isSelected = props.selected.includes(key);
  const isCorrect = props.revealed && (props.correct ?? "").includes(key);
  const isWrong = props.revealed && isSelected && !(props.correct ?? "").includes(key);
  return {
    selected: isSelected && !props.revealed,
    correct: isCorrect,
    wrong: isWrong,
    disabled: props.disabled,
  };
}

function mark(key: string) {
  const isSelected = props.selected.includes(key);
  if (!props.revealed) return isSelected;
  if ((props.correct ?? "").includes(key)) return true;
  return false;
}
</script>

<template>
  <div class="options">
    <button
      v-for="opt in options"
      :key="opt.key"
      class="option"
      :class="classes(opt.key)"
      :disabled="disabled"
      @click="emit('select', opt.key)"
    >
      <span class="option-key">{{ opt.key }}</span>
      <span class="option-text">{{ opt.text }}</span>
      <span class="option-mark" v-if="mark(opt.key)">
        <AppIcon :name="revealed ? 'check' : 'check'" :size="13" :stroke-width="2.6" />
      </span>
    </button>
  </div>
</template>

<style scoped>
.options {
  display: flex;
  flex-direction: column;
  gap: 11px;
}
.option-mark {
  margin-left: auto;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: currentColor;
  color: #fff;
  flex: none;
  margin-top: 1px;
}
.option.selected .option-mark {
  background: var(--amber);
  color: #241703;
}
.option.correct .option-mark {
  background: var(--green);
  color: #fff;
}
.option.wrong .option-mark {
  background: var(--red);
  color: #fff;
}
</style>
