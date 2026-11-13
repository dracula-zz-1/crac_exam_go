<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    value: number;
    size?: number;
    stroke?: number;
    color?: string;
    track?: string;
    label?: string;
    sub?: string;
    showValue?: boolean;
  }>(),
  {
    size: 128,
    stroke: 11,
    color: "var(--teal)",
    track: "var(--surface-3)",
    showValue: true,
  },
);

const radius = computed(() => (props.size - props.stroke) / 2);
const circumference = computed(() => 2 * Math.PI * radius.value);
const clamped = computed(() => Math.max(0, Math.min(100, props.value)));
const offset = computed(
  () => circumference.value * (1 - clamped.value / 100),
);
</script>

<template>
  <div
    class="ring"
    :style="{ width: size + 'px', height: size + 'px' }"
  >
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`">
      <circle
        :cx="size / 2"
        :cy="size / 2"
        :r="radius"
        fill="none"
        :stroke="track"
        :stroke-width="stroke"
      />
      <circle
        class="ring-value"
        :cx="size / 2"
        :cy="size / 2"
        :r="radius"
        fill="none"
        :stroke="color"
        :stroke-width="stroke"
        stroke-linecap="round"
        :stroke-dasharray="circumference"
        :stroke-dashoffset="offset"
        :transform="`rotate(-90 ${size / 2} ${size / 2})`"
      />
    </svg>
    <div class="ring-center">
      <strong v-if="showValue" :style="{ color }">{{ Math.round(value) }}<small>%</small></strong>
      <span v-if="label">{{ label }}</span>
      <em v-if="sub">{{ sub }}</em>
    </div>
  </div>
</template>

<style scoped>
.ring {
  position: relative;
  flex: none;
}
svg {
  display: block;
}
.ring-value {
  transition: stroke-dashoffset 0.7s var(--ease);
}
.ring-center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1px;
  text-align: center;
}
.ring-center strong {
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1;
}
.ring-center strong small {
  font-size: 11px;
  font-weight: 700;
  margin-left: 1px;
}
.ring-center span {
  font-size: 10.5px;
  color: var(--ink-3);
  letter-spacing: 0.04em;
}
.ring-center em {
  font-style: normal;
  font-size: 9.5px;
  color: var(--muted);
}
</style>
