<script setup lang="ts">
import AppIcon from "./AppIcon.vue";

withDefaults(
  defineProps<{
    label: string;
    value: string | number;
    suffix?: string;
    icon?: string;
    tone?: "amber" | "teal" | "blue" | "red" | "neutral";
    hint?: string;
  }>(),
  { tone: "neutral" },
);
</script>

<template>
  <div class="stat-tile" :class="`tone-${tone}`">
    <div class="stat-top">
      <span class="stat-label">{{ label }}</span>
      <span v-if="icon" class="stat-icon">
        <AppIcon :name="icon" :size="16" />
      </span>
    </div>
    <div class="stat-value">
      {{ value }}<small v-if="suffix">{{ suffix }}</small>
    </div>
    <div v-if="hint" class="stat-hint">{{ hint }}</div>
  </div>
</template>

<style scoped>
.stat-tile {
  padding: 17px 18px;
  border-radius: 15px;
  background: var(--surface);
  border: 1px solid var(--line);
  box-shadow: var(--shadow-xs);
  position: relative;
  overflow: hidden;
  transition: transform 0.18s var(--ease), box-shadow 0.18s var(--ease);
}
.stat-tile:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}
.stat-tile::after {
  content: "";
  position: absolute;
  right: -22px;
  top: -30px;
  width: 90px;
  height: 90px;
  border-radius: 50%;
  background: var(--tone-soft, var(--surface-2));
  opacity: 0.55;
}
.stat-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  position: relative;
  z-index: 1;
}
.stat-label {
  font-size: 11.5px;
  font-weight: 650;
  color: var(--ink-3);
  letter-spacing: 0.02em;
}
.stat-icon {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  display: grid;
  place-items: center;
  background: var(--tone-soft, var(--surface-2));
  color: var(--tone-ink, var(--ink-2));
}
.stat-value {
  font-size: 27px;
  font-weight: 800;
  letter-spacing: -0.035em;
  line-height: 1;
  color: var(--ink);
  position: relative;
  z-index: 1;
}
.stat-value small {
  font-size: 12px;
  font-weight: 650;
  color: var(--ink-3);
  margin-left: 3px;
  letter-spacing: 0;
}
.stat-hint {
  margin-top: 8px;
  font-size: 10.5px;
  color: var(--muted);
  position: relative;
  z-index: 1;
}
.tone-amber { --tone-soft: var(--amber-soft); --tone-ink: #9a6514; }
.tone-teal { --tone-soft: var(--teal-soft); --tone-ink: #147065; }
.tone-blue { --tone-soft: var(--blue-soft); --tone-ink: #2d5f8f; }
.tone-red { --tone-soft: var(--red-soft); --tone-ink: #a63b2e; }
.tone-neutral { --tone-soft: var(--surface-2); --tone-ink: var(--ink-2); }
</style>
