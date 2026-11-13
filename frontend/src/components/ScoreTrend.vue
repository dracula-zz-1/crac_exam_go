<script setup lang="ts">
import { computed } from "vue";
import type { ScorePoint } from "@/types";
import AppIcon from "./AppIcon.vue";

const props = withDefaults(
  defineProps<{
    points: ScorePoint[];
    maxScore?: number;
    passScore?: number;
    color?: string;
  }>(),
  { color: "var(--teal)" },
);

const W = 900;
const H = 280;
const pad = { l: 46, r: 24, t: 22, b: 38 };

const maxVal = computed(() => {
  const totals = props.points.map((p) => p.total || 0);
  const max = Math.max(props.maxScore ?? 0, props.passScore ?? 0, ...totals, 10);
  return Math.ceil(max / 10) * 10;
});

const plotW = W - pad.l - pad.r;
const plotH = H - pad.t - pad.b;

function x(i: number) {
  if (props.points.length <= 1) return pad.l + plotW / 2;
  return pad.l + (i / (props.points.length - 1)) * plotW;
}
function y(v: number) {
  const ratio = Math.max(0, Math.min(1, v / maxVal.value));
  return pad.t + (1 - ratio) * plotH;
}

const linePath = computed(() =>
  props.points
    .map((p, i) => `${i === 0 ? "M" : "L"} ${x(i).toFixed(1)} ${y(p.score).toFixed(1)}`)
    .join(" "),
);

const areaPath = computed(() => {
  if (!props.points.length) return "";
  const bottom = pad.t + plotH;
  return `${linePath.value} L ${x(props.points.length - 1).toFixed(1)} ${bottom} L ${x(0).toFixed(1)} ${bottom} Z`;
});

const gridLines = computed(() => {
  const rows = [0, 0.25, 0.5, 0.75, 1];
  return rows.map((r) => ({
    y: pad.t + r * plotH,
    label: Math.round(maxVal.value * (1 - r)),
  }));
});

const passY = computed(() =>
  props.passScore ? y(props.passScore) : null,
);

function fmt(iso: string) {
  const d = new Date(iso);
  return `${String(d.getMonth() + 1).padStart(2, "0")}/${String(d.getDate()).padStart(2, "0")}`;
}

const xLabels = computed(() => {
  const n = props.points.length;
  if (!n) return [];
  if (n === 1) return [{ x: x(0), text: fmt(props.points[0].date) }];
  const count = Math.min(5, n);
  const idx = [
    ...new Set(
      Array.from({ length: count }, (_, k) =>
        Math.round((k * (n - 1)) / (count - 1)),
      ),
    ),
  ];
  return idx.map((i) => ({ x: x(i), text: fmt(props.points[i].date) }));
});
</script>

<template>
  <div class="trend">
    <div v-if="!points.length" class="trend-empty">
      <div class="trend-empty-icon"><AppIcon name="sparkles" :size="22" /></div>
      <strong>还没有成绩曲线</strong>
      <span>完成一次模拟考试后，这里会显示你的进步轨迹</span>
    </div>
    <svg v-else :viewBox="`0 0 ${W} ${H}`" class="trend-svg" preserveAspectRatio="none">
      <defs>
        <linearGradient id="trend-fill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="color" stop-opacity="0.22" />
          <stop offset="100%" :stop-color="color" stop-opacity="0.01" />
        </linearGradient>
      </defs>

      <g>
        <line
          v-for="g in gridLines"
          :key="g.y"
          :x1="pad.l"
          :x2="W - pad.r"
          :y1="g.y"
          :y2="g.y"
          stroke="var(--line)"
          stroke-dasharray="3 5"
          stroke-width="1"
        />
        <text
          v-for="g in gridLines"
          :key="`t-${g.y}`"
          :x="pad.l - 10"
          :y="g.y + 3"
          text-anchor="end"
          class="axis-text"
        >
          {{ g.label }}
        </text>
      </g>

      <line
        v-if="passY !== null"
        :x1="pad.l"
        :x2="W - pad.r"
        :y1="passY"
        :y2="passY"
        stroke="var(--amber)"
        stroke-width="1.2"
        stroke-dasharray="5 5"
        opacity="0.7"
      />
      <text
        v-if="passY !== null"
        :x="W - pad.r"
        :y="passY - 6"
        text-anchor="end"
        class="pass-text"
      >
        及格线 {{ passScore }}
      </text>

      <path :d="areaPath" fill="url(#trend-fill)" />
      <path
        :d="linePath"
        fill="none"
        :stroke="color"
        stroke-width="2.4"
        stroke-linecap="round"
        stroke-linejoin="round"
      />

      <g v-for="(p, i) in points" :key="i">
        <circle
          :cx="x(i)"
          :cy="y(p.score)"
          r="5.5"
          :fill="p.passed ? color : 'var(--red)'"
          stroke="var(--surface)"
          stroke-width="2.5"
        >
          <title>{{ fmt(p.date) }} · {{ p.category }} 类 · {{ p.score }}/{{ p.total }}{{ p.passed ? " · 通过" : "" }}</title>
        </circle>
      </g>

      <g>
        <text
          v-for="l in xLabels"
          :key="l.text + l.x"
          :x="l.x"
          :y="H - 8"
          text-anchor="middle"
          class="axis-text"
        >
          {{ l.text }}
        </text>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.trend {
  width: 100%;
  min-height: 260px;
}
.trend-svg {
  width: 100%;
  height: auto;
  display: block;
}
.axis-text {
  font-size: 13px;
  fill: var(--muted);
  font-family: var(--font-sans);
}
.pass-text {
  font-size: 12.5px;
  fill: #a8751f;
  font-family: var(--font-sans);
  font-weight: 600;
}
.trend-empty {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-align: center;
}
.trend-empty-icon {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: var(--surface-2);
  color: var(--ink-3);
  display: grid;
  place-items: center;
  margin-bottom: 2px;
}
.trend-empty strong {
  font-size: 13.5px;
  font-weight: 700;
}
.trend-empty span {
  font-size: 11.5px;
  color: var(--muted);
  max-width: 280px;
  line-height: 1.6;
}
</style>
