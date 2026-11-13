<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { onBeforeRouteLeave, useRouter } from "vue-router";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useToastStore } from "@/stores/toast";
import type { Category, ExamConfig, ExamQuestion, ExamStart } from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import OptionList from "@/components/OptionList.vue";
import BaseModal from "@/components/BaseModal.vue";

const props = defineProps<{ category: string }>();
const router = useRouter();
const session = useSessionStore();
const toast = useToastStore();

const CATEGORY = computed<Category>(() =>
  ["A", "B", "C"].includes(props.category.toUpperCase())
    ? (props.category.toUpperCase() as Category)
    : "A",
);

const META: Record<Category, { label: string; desc: string; pass: number; total: number; minutes: number }> = {
  A: { label: "A 类", desc: "基础操作能力验证", pass: 30, total: 40, minutes: 40 },
  B: { label: "B 类", desc: "常用频段与设备操作", pass: 45, total: 60, minutes: 60 },
  C: { label: "C 类", desc: "短波与高级操作能力", pass: 70, total: 90, minutes: 90 },
};

const meta = computed(() => META[CATEGORY.value]);

const config = ref<ExamConfig | null>(null);
const startData = ref<ExamStart | null>(null);
const starting = ref(false);
const startError = ref("");
const answers = ref<Record<number, string>>({});
const index = ref(0);
const endAt = ref(0);
const remaining = ref(0);
const submitted = ref(false);
const submitting = ref(false);
const navOpen = ref(false);
const imageOpen = ref(false);
const submitOpen = ref(false);
const exitOpen = ref(false);
const autoSubmitTriggered = ref(false);
let timer: number | null = null;

const live = computed(() => !!startData.value && !submitted.value);
const questions = computed<ExamQuestion[]>(() => startData.value?.questions ?? []);
const question = computed<ExamQuestion | null>(() => questions.value[index.value] ?? null);
const total = computed(() => questions.value.length);
const answeredCount = computed(
  () => Object.values(answers.value).filter((v) => v.length > 0).length,
);
const unanswered = computed(() => total.value - answeredCount.value);
const isMulti = computed(() => question.value?.type === 2);
const selected = computed(() => {
  const q = question.value;
  if (!q) return [];
  return (answers.value[q.id] ?? "").split("").filter(Boolean);
});

const timeLimit = computed(() => startData.value?.timeLimitSeconds ?? meta.value.minutes * 60);
const timePercent = computed(() =>
  timeLimit.value ? Math.max(0, Math.min(100, (remaining.value / timeLimit.value) * 100)) : 0,
);
const timeText = computed(() => {
  const s = Math.max(0, remaining.value);
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${String(m).padStart(2, "0")}:${String(sec).padStart(2, "0")}`;
});
const timeDanger = computed(() => remaining.value <= 300);

const options = computed(() => {
  const q = question.value;
  if (!q) return [];
  return [
    { key: "A", text: q.A },
    { key: "B", text: q.B },
    { key: "C", text: q.C },
    { key: "D", text: q.D },
  ];
});

function normalize(raw: string) {
  return [...new Set(raw.toUpperCase().replace(/[^A-D]/g, "").split(""))].sort().join("");
}

async function loadConfig() {
  try {
    config.value = await api.getConfig(CATEGORY.value);
  } catch {
    config.value = {
      total: meta.value.total,
      single: CATEGORY.value === "A" ? 32 : CATEGORY.value === "B" ? 45 : 70,
      multiple: CATEGORY.value === "A" ? 8 : CATEGORY.value === "B" ? 15 : 20,
      timeMinutes: meta.value.minutes,
      passScore: meta.value.pass,
    };
  }
}

async function start() {
  if (!session.user) return;
  starting.value = true;
  startError.value = "";
  try {
    const data = await api.createExam(session.user.id, CATEGORY.value);
    startData.value = data;
    answers.value = {};
    index.value = 0;
    submitted.value = false;
    autoSubmitTriggered.value = false;
    endAt.value = Date.now() + data.timeLimitSeconds * 1000;
    remaining.value = data.timeLimitSeconds;
    startTimer();
    toast.success("考试开始", `${data.total} 题 · ${Math.round(data.timeLimitSeconds / 60)} 分钟`);
  } catch (e: any) {
    startError.value = e?.message || "开始考试失败";
  } finally {
    starting.value = false;
  }
}

function startTimer() {
  stopTimer();
  tick();
  timer = window.setInterval(tick, 500);
}
function stopTimer() {
  if (timer !== null) {
    clearInterval(timer);
    timer = null;
  }
}
function tick() {
  if (!startData.value || submitted.value) return;
  remaining.value = Math.max(0, Math.ceil((endAt.value - Date.now()) / 1000));
  if (remaining.value <= 0 && !autoSubmitTriggered.value) {
    autoSubmitTriggered.value = true;
    doSubmit(true);
  }
}

function select(key: string) {
  const q = question.value;
  if (!q || submitting.value) return;
  if (q.type === 2) {
    const current = (answers.value[q.id] ?? "").split("").filter(Boolean);
    const next = current.includes(key)
      ? current.filter((k) => k !== key)
      : [...current, key];
    answers.value = { ...answers.value, [q.id]: normalize(next.join("")) };
  } else {
    answers.value = { ...answers.value, [q.id]: key };
  }
}

function goTo(i: number) {
  if (i < 0 || i >= total.value) return;
  index.value = i;
  navOpen.value = false;
}
function next() {
  if (index.value < total.value - 1) index.value++;
}
function prev() {
  if (index.value > 0) index.value--;
}

function askSubmit() {
  if (!live.value) return;
  submitOpen.value = true;
}

async function doSubmit(auto = false) {
  if (!startData.value || submitting.value || submitted.value) return;
  submitting.value = true;
  submitOpen.value = false;
  stopTimer();
  try {
    const answerMap: Record<string, string> = {};
    for (const q of questions.value) {
      const value = answers.value[q.id] ?? "";
      if (value) answerMap[String(q.id)] = value;
    }
    const result = await api.submitExam(startData.value.examId, answerMap);
    submitted.value = true;
    toast.success(auto ? "时间到，已自动交卷" : "交卷成功", `得分 ${result.score} / ${result.totalCount}`);
    router.replace(`/result/${result.examId}`);
  } catch (e: any) {
    submitting.value = false;
    toast.error("交卷失败", e?.message || "请稍后重试");
    if (remaining.value > 0) startTimer();
  }
}

async function exitExam() {
  if (startData.value) {
    try {
      await api.invalidateExam(startData.value.examId);
    } catch {
      /* ignore */
    }
  }
  stopTimer();
  router.push("/home");
}

function confirmExit() {
  if (!live.value) {
    router.push("/home");
    return;
  }
  exitOpen.value = true;
}

function onKey(e: KeyboardEvent) {
  if (!live.value || navOpen.value || imageOpen.value || submitOpen.value || exitOpen.value)
    return;
  const target = e.target as HTMLElement;
  if (["INPUT", "TEXTAREA"].includes(target?.tagName)) return;
  const key = e.key.toUpperCase();
  if (["A", "B", "C", "D", "1", "2", "3", "4"].includes(key)) {
    select("1234".includes(key) ? "ABCD"[Number(key) - 1] : key);
    e.preventDefault();
  } else if (e.key === "ArrowRight") {
    next();
  } else if (e.key === "ArrowLeft") {
    prev();
  }
}

onBeforeRouteLeave(async (_to, _from, done) => {
  if (!live.value) return done();
  const ok = window.confirm("考试正在进行，离开将作废本次考试，确定离开吗？");
  if (!ok) return done(false);
  try {
    if (startData.value) await api.invalidateExam(startData.value.examId);
  } catch {
    /* ignore */
  }
  stopTimer();
  done();
});

onMounted(() => {
  loadConfig();
  window.addEventListener("keydown", onKey);
});
onBeforeUnmount(() => {
  stopTimer();
  window.removeEventListener("keydown", onKey);
});
</script>

<template>
  <div class="page exam-page" :class="{ 'exam-live': !!startData }">
    <!-- 开考确认 -->
    <template v-if="!startData">
      <button class="back-btn" @click="router.push('/home')">
        <AppIcon name="arrowLeft" :size="17" />
        返回总览
      </button>

      <section class="intro cat" :class="`cat-${CATEGORY}`">
        <div class="intro-main">
          <span class="eyebrow">MOCK EXAM · {{ CATEGORY }} 类</span>
          <h1>{{ CATEGORY }} 类模拟考试</h1>
          <p>{{ meta.desc }}。按真实考试规则随机组卷，开始后计时不可暂停。</p>

          <div class="intro-config">
            <div class="cfg">
              <strong>{{ config?.total ?? meta.total }}</strong>
              <span>总题数</span>
            </div>
            <div class="cfg">
              <strong>{{ config?.single ?? "—" }}</strong>
              <span>单选题</span>
            </div>
            <div class="cfg">
              <strong>{{ config?.multiple ?? "—" }}</strong>
              <span>多选题</span>
            </div>
            <div class="cfg">
              <strong>{{ config?.timeMinutes ?? meta.minutes }}<small>分钟</small></strong>
              <span>考试时间</span>
            </div>
            <div class="cfg accent">
              <strong>{{ config?.passScore ?? meta.pass }}<small>分</small></strong>
              <span>及格线</span>
            </div>
          </div>

          <ul class="rules">
            <li><AppIcon name="clock" :size="15" /> 开始后由服务端计时，时间到自动交卷</li>
            <li><AppIcon name="shield" :size="15" /> 题目与选项随机打乱，答案由服务端判定</li>
            <li><AppIcon name="errors" :size="15" /> 答错与未作答的题目会自动进入错题本</li>
          </ul>

          <div v-if="startError" class="intro-error">
            <AppIcon name="warning" :size="15" />
            <div>
              <strong>{{ startError }}</strong>
              <button class="btn btn-sm btn-ghost" @click="router.push('/settings')">去导入题库</button>
            </div>
          </div>

          <div class="intro-actions">
            <button class="btn btn-accent btn-lg" :disabled="starting" @click="start">
              <span v-if="starting" class="spinner dark" />
              <template v-else>
                开始考试
                <AppIcon name="arrowRight" :size="17" />
              </template>
            </button>
            <span class="intro-hint">共 {{ config?.total ?? meta.total }} 题 · 每题 1 分</span>
          </div>
        </div>

        <div class="intro-art" aria-hidden="true">
          <span class="intro-letter">{{ CATEGORY }}</span>
          <svg viewBox="0 0 360 360" fill="none">
            <circle cx="180" cy="180" r="58" stroke="var(--cat)" stroke-opacity=".5" stroke-width="1.4" />
            <circle cx="180" cy="180" r="102" stroke="var(--cat)" stroke-opacity=".28" stroke-width="1.2" />
            <circle cx="180" cy="180" r="146" stroke="var(--cat)" stroke-opacity=".14" stroke-width="1" />
            <circle cx="180" cy="180" r="7" fill="var(--cat)" />
            <path d="M40 230 Q90 190 140 230 T240 230 T340 230" stroke="var(--cat)" stroke-opacity=".4" stroke-width="1.4" />
          </svg>
        </div>
      </section>
    </template>

    <!-- 考试中 -->
    <template v-else>
      <header class="exam-bar surface">
        <button class="back-icon" @click="confirmExit">
          <AppIcon name="arrowLeft" :size="18" />
        </button>
        <div class="exam-title">
          <span class="chip" :class="`chip-${CATEGORY === 'A' ? 'amber' : CATEGORY === 'B' ? 'teal' : 'blue'}`">
            {{ CATEGORY }} 类
          </span>
          <strong>模拟考试</strong>
        </div>
        <div class="exam-progress">
          <span>已答 <b>{{ answeredCount }}</b> / {{ total }}</span>
          <div class="progress" style="width: 180px">
            <span :style="{ width: (answeredCount / Math.max(1, total)) * 100 + '%' }" />
          </div>
        </div>
        <div class="timer-pill" :class="{ danger: timeDanger }">
          <AppIcon name="clock" :size="16" />
          <span class="mono">{{ timeText }}</span>
        </div>
        <div class="exam-actions">
          <button class="btn btn-sm btn-ghost" @click="navOpen = true">
            <AppIcon name="grid" :size="14" />
            答题卡
          </button>
          <button class="btn btn-sm btn-accent" :disabled="submitting" @click="askSubmit">
            <AppIcon name="check" :size="14" />
            交卷
          </button>
        </div>
      </header>

      <div class="time-track">
        <span :style="{ width: timePercent + '%' }" :class="{ danger: timeDanger }" />
      </div>

      <div class="exam-grid" v-if="question">
        <section class="q-card surface">
          <div class="q-body">
            <div class="q-top">
              <div class="q-meta">
                <span class="chip" :class="isMulti ? 'chip-teal' : 'chip-amber'">
                  {{ isMulti ? "多选题" : "单选题" }}
                </span>
                <span class="q-index mono">第 {{ index + 1 }} / {{ total }} 题</span>
              </div>
              <span class="q-code mono">{{ question.J }}</span>
            </div>

            <h2 class="q-stem">{{ question.Q }}</h2>

            <div v-if="question.F" class="q-image" @click="imageOpen = true">
              <img :src="question.F" alt="题目图片" />
              <span class="q-image-hint">
                <AppIcon name="search" :size="13" /> 点击放大
              </span>
            </div>

            <OptionList
              :options="options"
              :selected="selected"
              :multi="isMulti"
              @select="select"
            />
          </div>

          <div class="q-actions">
            <button class="btn btn-ghost" :disabled="index === 0" @click="prev">
              <AppIcon name="chevronLeft" :size="16" />
              上一题
            </button>
            <span class="spacer" />
            <button class="btn btn-soft" @click="navOpen = true">答题卡</button>
            <button class="btn" :disabled="index === total - 1" @click="next">
              下一题
              <AppIcon name="chevronRight" :size="16" />
            </button>
          </div>
        </section>

        <aside class="exam-side">
          <section class="surface card-pad side-card">
            <div class="side-timer" :class="{ danger: timeDanger }">
              <span class="mono">{{ timeText }}</span>
              <em>剩余时间</em>
            </div>
            <div class="side-stats">
              <div>
                <strong>{{ answeredCount }}</strong>
                <span>已答</span>
              </div>
              <div>
                <strong style="color: var(--red)">{{ unanswered }}</strong>
                <span>未答</span>
              </div>
              <div>
                <strong>{{ Math.round((answeredCount / Math.max(1, total)) * 100) }}%</strong>
                <span>完成</span>
              </div>
            </div>
          </section>

          <section class="surface card-pad side-card nav-card">
            <h3 class="side-title">答题卡</h3>
            <div class="mini-grid">
              <button
                v-for="(q, i) in questions"
                :key="q.id"
                class="mini-cell"
                :class="{
                  answered: (answers[q.id] ?? '').length > 0,
                  current: i === index,
                }"
                @click="goTo(i)"
              >
                {{ i + 1 }}
              </button>
            </div>
            <div class="legend">
              <span><i class="lg current" /> 当前</span>
              <span><i class="lg answered" /> 已答</span>
              <span><i class="lg" /> 未答</span>
            </div>
          </section>

          <button class="btn btn-accent btn-block" @click="askSubmit">
            交卷并查看成绩
          </button>
        </aside>
      </div>
    </template>

    <!-- 题目图片 -->
    <BaseModal
      :open="imageOpen"
      title="查看题目图片"
      wide
      @close="imageOpen = false"
    >
      <div class="image-full" v-if="question?.F">
        <img :src="question.F" alt="题目图片" />
      </div>
    </BaseModal>

    <!-- 答题卡 -->
    <BaseModal
      :open="navOpen"
      title="答题卡"
      :subtitle="`已答 ${answeredCount} / ${total} 题`"
      wide
      @close="navOpen = false"
    >
      <div class="nav-grid">
        <button
          v-for="(q, i) in questions"
          :key="q.id"
          class="nav-cell"
          :class="{
            answered: (answers[q.id] ?? '').length > 0,
            current: i === index,
          }"
          @click="goTo(i)"
        >
          {{ i + 1 }}
        </button>
      </div>
      <template #footer>
        <button class="btn btn-ghost" @click="navOpen = false">继续答题</button>
        <button class="btn btn-accent" @click="navOpen = false; askSubmit()">交卷</button>
      </template>
    </BaseModal>

    <!-- 交卷确认 -->
    <BaseModal
      :open="submitOpen"
      title="确认交卷"
      :subtitle="unanswered > 0 ? `还有 ${unanswered} 道题未作答` : '所有题目均已作答'"
      @close="submitOpen = false"
    >
      <div class="submit-info">
        <div class="submit-icon"><AppIcon name="exam" :size="22" /></div>
        <div>
          <strong>交卷后不能再修改答案</strong>
          <p>系统会立即判分，并显示成绩与试卷复盘。</p>
          <div v-if="unanswered > 0" class="submit-warn">
            未作答的 {{ unanswered }} 道题将按错误计入，并加入错题本。
          </div>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-ghost" @click="submitOpen = false">再检查一下</button>
        <button class="btn btn-accent" :disabled="submitting" @click="doSubmit(false)">
          <span v-if="submitting" class="spinner dark" />
          确认交卷
        </button>
      </template>
    </BaseModal>

    <!-- 退出确认 -->
    <BaseModal
      :open="exitOpen"
      title="退出考试"
      subtitle="当前答题记录将不会保存"
      @close="exitOpen = false"
    >
      <div class="submit-info">
        <div class="submit-icon danger"><AppIcon name="warning" :size="22" /></div>
        <div>
          <strong>确定要退出吗？</strong>
          <p>本次考试会被作废，不会计入成绩统计。</p>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-ghost" @click="exitOpen = false">继续考试</button>
        <button class="btn btn-danger" @click="exitExam">退出并作废</button>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.exam-page {
  max-width: 1240px;
}
/* 切页动画不再上下位移，避免瞬间把一屏页面挤出滚动条 */
.exam-page.exam-live.page-enter-from,
.exam-page.exam-live.page-leave-to {
  transform: none;
}
/* 考试进行中：整页锁定一屏，答题过程中不出现纵向滚动条 */
.exam-page.exam-live {
  height: calc(100vh - 64px);
  padding-top: 16px;
  padding-bottom: 18px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.exam-page.exam-live .exam-bar,
.exam-page.exam-live .time-track {
  flex: none;
}
.exam-page.exam-live .exam-grid {
  flex: 1;
  min-height: 0;
  align-items: stretch;
}
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  border-radius: 10px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink-3);
  background: var(--surface);
  border: 1px solid var(--line);
  margin-bottom: 20px;
}
.back-btn:hover {
  color: var(--ink);
}

.intro {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 20px;
  padding: 38px 42px;
  border-radius: 26px;
  border: 1px solid var(--line);
  background: linear-gradient(135deg, var(--cat-soft) 0%, var(--surface) 55%);
  position: relative;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}
.intro-main {
  position: relative;
  z-index: 2;
}
.intro h1 {
  margin-top: 12px;
  font-size: 44px;
  font-weight: 850;
  letter-spacing: -0.04em;
}
.intro > .intro-main > p {
  margin-top: 12px;
  font-size: 14px;
  line-height: 1.8;
  color: var(--ink-3);
  max-width: 560px;
}
.intro-config {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
  margin: 28px 0;
  max-width: 640px;
}
.cfg {
  padding: 13px 14px;
  border-radius: 13px;
  background: rgba(255, 253, 248, 0.78);
  border: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.cfg strong {
  font-size: 21px;
  font-weight: 850;
  letter-spacing: -0.03em;
}
.cfg strong small {
  font-size: 10px;
  font-weight: 600;
  color: var(--ink-3);
  margin-left: 1px;
}
.cfg span {
  font-size: 10px;
  color: var(--ink-3);
}
.cfg.accent {
  background: var(--cat-soft);
  border-color: transparent;
}
.cfg.accent strong {
  color: var(--cat-ink);
}
.rules {
  list-style: none;
  margin: 0 0 26px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.rules li {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--ink-2);
}
.rules :deep(svg) {
  color: var(--cat);
}
.intro-error {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  background: var(--red-soft);
  color: #a63b2e;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 12.5px;
  margin-bottom: 18px;
  max-width: 640px;
}
.intro-error div {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.intro-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}
.intro-hint {
  font-size: 11.5px;
  color: var(--muted);
}
.intro-art {
  position: relative;
  display: grid;
  place-items: center;
}
.intro-art svg {
  width: 330px;
  height: 330px;
}
.intro-letter {
  position: absolute;
  font-size: 190px;
  font-weight: 900;
  line-height: 1;
  color: var(--cat);
  opacity: 0.1;
  letter-spacing: -0.08em;
}

.exam-bar {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 12px 18px;
  border-radius: 16px;
  margin-bottom: 8px;
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(14px);
}
.back-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: var(--ink-3);
}
.back-icon:hover {
  background: var(--surface-2);
  color: var(--ink);
}
.exam-title {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 13.5px;
}
.exam-title strong {
  font-weight: 750;
}
.exam-progress {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  font-size: 11.5px;
  color: var(--ink-3);
}
.exam-progress b {
  color: var(--ink);
}
.exam-progress .progress {
  height: 6px;
}
.timer-pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 38px;
  padding: 0 15px;
  border-radius: 11px;
  background: var(--navy);
  color: #f4f0e6;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.04em;
  transition: background 0.2s;
}
.timer-pill.danger {
  background: var(--red);
  animation: blink 1.2s infinite;
}
@keyframes blink {
  50% { opacity: 0.72; }
}
.exam-actions {
  display: flex;
  gap: 8px;
}
.time-track {
  height: 3px;
  border-radius: 99px;
  background: var(--surface-3);
  overflow: hidden;
  margin-bottom: 18px;
}
.time-track span {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, var(--teal), #58bcae);
  transition: width 0.5s linear;
}
.time-track span.danger {
  background: linear-gradient(90deg, #d05c4d, #f08a7d);
}

.exam-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 20px;
  align-items: start;
}
.q-card {
  padding: 20px 26px 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}
/* 题干 / 图片 / 选项：超长时只在卡片内部滚动 */
.q-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 6px;
}
/* 考试答题期间压缩选项行高，保证带图的题目也能一屏放下 */
.q-body :deep(.option) {
  padding: 12px 15px;
}
.q-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.q-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}
.q-index {
  font-size: 11px;
  color: var(--muted);
}
.q-code {
  font-size: 11px;
  color: var(--muted);
  letter-spacing: 0.06em;
}
.q-stem {
  font-size: 18px;
  line-height: 1.55;
  font-weight: 700;
  letter-spacing: -0.01em;
  margin-bottom: 14px;
}
.q-image {
  position: relative;
  margin-bottom: 14px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--line);
  background: var(--navy);
  cursor: zoom-in;
}
.q-image img {
  width: 100%;
  max-height: 160px;
  object-fit: contain;
  margin: 0 auto;
}
.q-image-hint {
  position: absolute;
  right: 10px;
  bottom: 10px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 9px;
  border-radius: 99px;
  font-size: 10px;
  color: #f4f0e6;
  background: rgba(16, 24, 32, 0.65);
  backdrop-filter: blur(6px);
}
.q-actions {
  flex: none;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--line);
}

.exam-side {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  height: 100%;
}
.nav-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.nav-card .mini-grid {
  overflow-y: auto;
  min-height: 0;
  padding-right: 4px;
}
.side-card {
  padding: 20px;
}
.side-timer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 14px;
  border-radius: 14px;
  background: var(--surface-2);
}
.side-timer span {
  font-size: 30px;
  font-weight: 800;
  letter-spacing: 0.02em;
}
.side-timer em {
  font-style: normal;
  font-size: 10px;
  color: var(--ink-3);
}
.side-timer.danger {
  background: var(--red-soft);
  color: #a63b2e;
}
.side-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin-top: 12px;
}
.side-stats div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 8px 4px;
  border-radius: 9px;
  background: var(--surface-2);
}
.side-stats strong {
  font-size: 14px;
  font-weight: 800;
}
.side-stats span {
  font-size: 9.5px;
  color: var(--ink-3);
}
.side-title {
  font-size: 13px;
  font-weight: 750;
  margin-bottom: 10px;
}
.mini-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 5px;
}
.mini-cell {
  aspect-ratio: 1;
  border-radius: 6px;
  border: 1px solid var(--line);
  background: var(--surface);
  font-size: 9.5px;
  font-weight: 650;
  color: var(--ink-3);
  transition: all 0.12s;
}
.mini-cell.answered {
  background: var(--teal-soft);
  border-color: rgba(42, 157, 143, 0.35);
  color: #147065;
}
.mini-cell.current {
  border-color: var(--amber);
  box-shadow: 0 0 0 2px rgba(232, 163, 61, 0.2);
}
.legend {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  font-size: 9.5px;
  color: var(--ink-3);
}
.legend span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.lg {
  width: 9px;
  height: 9px;
  border-radius: 3px;
  background: var(--surface);
  border: 1px solid var(--line-strong);
}
.lg.answered { background: var(--teal-soft); border-color: rgba(42, 157, 143, 0.35); }
.lg.current { border-color: var(--amber); background: #fff; }

.nav-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(42px, 1fr));
  gap: 8px;
}
.nav-cell {
  height: 40px;
  border-radius: 9px;
  border: 1px solid var(--line);
  background: var(--surface);
  font-size: 11.5px;
  font-weight: 650;
  color: var(--ink-2);
}
.nav-cell.answered {
  background: var(--teal-soft);
  border-color: rgba(42, 157, 143, 0.35);
  color: #147065;
}
.nav-cell.current {
  border-color: var(--amber);
  box-shadow: 0 0 0 3px rgba(232, 163, 61, 0.16);
}

.submit-info {
  display: flex;
  gap: 15px;
}
.submit-icon {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: var(--amber-soft);
  color: #9a6514;
  display: grid;
  place-items: center;
  flex: none;
}
.submit-icon.danger {
  background: var(--red-soft);
  color: #a63b2e;
}
.submit-info strong {
  font-size: 13.5px;
  font-weight: 700;
}
.submit-info p {
  margin-top: 5px;
  font-size: 12px;
  line-height: 1.65;
  color: var(--ink-3);
}
.submit-warn {
  margin-top: 10px;
  padding: 8px 11px;
  border-radius: 9px;
  background: var(--red-soft);
  font-size: 11.5px;
  color: #a63b2e;
}

.image-full img {
  width: 100%;
  max-height: 70vh;
  object-fit: contain;
  border-radius: 12px;
}

/* 窗口极窄时（低于应用最小宽度）才退化为可滚动的单列布局 */
@media (max-width: 1040px) {
  .intro-grid,
  .intro {
    grid-template-columns: 1fr;
  }
  .intro-art {
    display: none;
  }
  .exam-page.exam-live {
    height: auto;
    min-height: calc(100vh - 64px);
    overflow: visible;
    padding-bottom: 40px;
  }
  .exam-grid {
    grid-template-columns: 1fr;
  }
  .q-card {
    overflow: visible;
  }
  .q-body {
    overflow: visible;
  }
  .exam-side {
    display: grid;
    grid-template-columns: 1fr 1fr;
    height: auto;
  }
  .nav-card {
    min-height: 260px;
  }
  .exam-side > .btn {
    grid-column: span 2;
  }
}
</style>
