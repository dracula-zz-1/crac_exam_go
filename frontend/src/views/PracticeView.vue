<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useBankStore } from "@/stores/bank";
import { useToastStore } from "@/stores/toast";
import type { PracticeData, PracticeMode, Question } from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import OptionList from "@/components/OptionList.vue";
import ProgressRing from "@/components/ProgressRing.vue";
import BaseModal from "@/components/BaseModal.vue";

const props = defineProps<{ mode: string; category: string }>();
const router = useRouter();
const route = useRoute();
const session = useSessionStore();
const bank = useBankStore();
const toast = useToastStore();

const MODE_META: Record<PracticeMode, { label: string; icon: string; desc: string }> = {
  practice: { label: "逐题练习", icon: "practice", desc: "先思考，再揭晓答案" },
  random: { label: "随机练习", icon: "sparkles", desc: "错题加权；每轮覆盖全部题目" },
  error: { label: "错题强化", icon: "errors", desc: "把答错的题逐一消化" },
  favorite: { label: "收藏练习", icon: "star", desc: "只看你标记过的重点题" },
};

const mode = computed<PracticeMode>(() =>
  (["practice", "random", "error", "favorite"].includes(props.mode)
    ? props.mode
    : "practice") as PracticeMode,
);
const category = computed(() =>
  ["A", "B", "C"].includes(props.category.toUpperCase())
    ? props.category.toUpperCase()
    : "A",
);

const data = ref<PracticeData | null>(null);
const loading = ref(true);
const error = ref("");
const index = ref(0);
const selected = ref<string[]>([]);
const revealed = ref(false);
const results = ref<Record<number, "correct" | "wrong">>({});
const imageOpen = ref(false);
const navOpen = ref(false);
const finished = ref(false);
const saving = ref(false);

const questions = computed<Question[]>(() => data.value?.questions ?? []);
const question = computed<Question | null>(() => questions.value[index.value] ?? null);
const total = computed(() => questions.value.length);
const isMulti = computed(() => question.value?.type === 2);
const answerLetters = computed(() => sortedLetters(question.value?.T ?? ""));
const isFavorite = computed(() =>
  question.value ? bank.favoriteIds.has(question.value.id) : false,
);
const selectedSorted = computed(() => sortedLetters(selected.value.join("")));
const isCorrect = computed(
  () => !!question.value && selectedSorted.value === answerLetters.value,
);
const hasSelection = computed(() => selected.value.length > 0);

const correctCount = computed(
  () => Object.values(results.value).filter((r) => r === "correct").length,
);
const wrongCount = computed(
  () => Object.values(results.value).filter((r) => r === "wrong").length,
);
const progressPercent = computed(() =>
  total.value ? Math.round(((index.value + 1) / total.value) * 100) : 0,
);
const accuracy = computed(() => {
  const done = correctCount.value + wrongCount.value;
  return done ? Math.round((correctCount.value / done) * 100) : 0;
});

const modeMeta = computed(() => MODE_META[mode.value]);

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

function sortedLetters(raw: string) {
  return [...new Set(raw.toUpperCase().replace(/[^A-D]/g, "").split(""))]
    .sort()
    .join("");
}

async function load(force = false) {
  if (!session.user) return;
  loading.value = true;
  error.value = "";
  finished.value = false;
  results.value = {};
  try {
    const d = await bank.loadPractice(
      session.user.id,
      mode.value,
      category.value,
      force || mode.value === "random",
    );
    data.value = d;
    index.value =
      mode.value === "practice" || mode.value === "random"
        ? Math.min(d.progress, Math.max(0, d.total - 1))
        : 0;
    const startId = Number(route.query.start);
    if (startId) {
      const found = d.questions.findIndex((q) => q.id === startId);
      if (found >= 0) index.value = found;
    }
    resetQuestion();
  } catch (e: any) {
    error.value = e?.message || "加载练习失败";
  } finally {
    loading.value = false;
  }
}

function resetQuestion() {
  selected.value = [];
  revealed.value = false;
}

function select(key: string) {
  if (revealed.value) return;
  if (isMulti.value) {
    selected.value = selected.value.includes(key)
      ? selected.value.filter((k) => k !== key)
      : [...selected.value, key].sort();
  } else {
    selected.value = [key];
  }
}

async function reveal() {
  if (!question.value || revealed.value) return;
  revealed.value = true;
  const correct = isCorrect.value;
  if (hasSelection.value) {
    results.value = { ...results.value, [index.value]: correct ? "correct" : "wrong" };
  }

  if (!correct && hasSelection.value && mode.value !== "error") {
    try {
      await api.addError(session.user!.id, question.value.id, category.value);
      toast.warning("已加入错题本", "稍后可在错题强化中复习");
    } catch (e: any) {
      toast.error("加入错题本失败", e?.message);
    }
  }
  if (correct && mode.value === "error") {
    toast.success("这道题已经掌握", "可以移出错题本了");
  }
}

async function next() {
  if (!revealed.value) {
    await reveal();
    return;
  }
  if (index.value >= total.value - 1) {
    if (mode.value === "random" && session.user) {
      // 标记本轮完成，下次进入随机练习会生成新的一轮
      try {
        await api.saveRandomProgress(session.user.id, category.value, total.value);
      } catch {
        /* ignore */
      }
    }
    finished.value = true;
    return;
  }
  index.value++;
  resetQuestion();
  await saveProgress();
}

async function prev() {
  if (index.value <= 0) return;
  index.value--;
  resetQuestion();
  await saveProgress();
}

async function jumpTo(i: number) {
  if (i < 0 || i >= total.value) return;
  index.value = i;
  resetQuestion();
  navOpen.value = false;
  await saveProgress();
}

async function saveProgress() {
  if (!session.user) return;
  if (mode.value === "random") {
    saving.value = true;
    try {
      await api.saveRandomProgress(session.user.id, category.value, index.value);
    } catch {
      /* 进度保存失败不打断练习 */
    } finally {
      saving.value = false;
    }
    return;
  }
  if (mode.value !== "practice") return;
  saving.value = true;
  try {
    await api.saveProgress(session.user.id, category.value, index.value);
    bank.updateProgress(mode.value, category.value, index.value);
  } catch {
    /* 进度保存失败不打断练习 */
  } finally {
    saving.value = false;
  }
}

async function toggleFavorite() {
  if (!question.value || !session.user) return;
  const q = question.value;
  const wasFavorite = isFavorite.value;
  try {
    if (wasFavorite) {
      await api.removeFavorite(session.user.id, q.id);
      bank.setFavorite(q.id, false);
      toast.info("已取消收藏");
      if (mode.value === "favorite") await removeCurrentFromList();
    } else {
      await api.addFavorite(session.user.id, q.id, category.value);
      bank.setFavorite(q.id, true);
      toast.success("已加入收藏");
    }
  } catch (e: any) {
    toast.error("收藏操作失败", e?.message);
  }
}

async function removeFromErrors() {
  if (!question.value || !session.user) return;
  try {
    await api.removeError(session.user.id, question.value.id);
    toast.success("已移出错题本");
    if (mode.value === "error") await removeCurrentFromList();
  } catch (e: any) {
    toast.error("操作失败", e?.message);
  }
}

async function removeCurrentFromList() {
  if (!data.value) return;
  const id = question.value?.id;
  data.value.questions = data.value.questions.filter((q) => q.id !== id);
  if (index.value >= data.value.questions.length) {
    index.value = Math.max(0, data.value.questions.length - 1);
  }
  resetQuestion();
  if (!data.value.questions.length) {
    finished.value = true;
  }
}

async function restart() {
  if (mode.value === "random") {
    // 重新生成一轮（完成时会清除队列状态）
    await load(true);
    return;
  }
  results.value = {};
  index.value = 0;
  finished.value = false;
  resetQuestion();
  saveProgress();
}

function switchMode(next: "practice" | "random") {
  if (next === mode.value) return;
  router.push(`/practice/${next}/${category.value}`);
}

function back() {
  router.push("/home");
}

function onKey(e: KeyboardEvent) {
  if (imageOpen.value || navOpen.value || finished.value) return;
  const target = e.target as HTMLElement;
  if (["INPUT", "TEXTAREA"].includes(target?.tagName)) return;

  const key = e.key.toUpperCase();
  if (["A", "B", "C", "D", "1", "2", "3", "4"].includes(key)) {
    const letter = "1234".includes(key) ? "ABCD"[Number(key) - 1] : key;
    select(letter);
    e.preventDefault();
  } else if (e.key === "Enter" || e.key === " ") {
    if (!revealed.value) reveal();
    else next();
    e.preventDefault();
  } else if (e.key === "ArrowLeft") {
    prev();
  } else if (e.key === "ArrowRight") {
    next();
  } else if (e.key === "f" || e.key === "F") {
    toggleFavorite();
  }
}

watch(
  () => [props.mode, props.category],
  () => load(),
);

onMounted(() => {
  load();
  window.addEventListener("keydown", onKey);
});
onBeforeUnmount(() => window.removeEventListener("keydown", onKey));
</script>

<template>
  <div class="page practice-page">
    <header class="practice-head">
      <button class="back-btn" @click="back">
        <AppIcon name="arrowLeft" :size="17" />
        返回
      </button>
      <div class="head-copy">
        <span class="eyebrow">{{ modeMeta.label }} · {{ category }} 类</span>
        <h1 class="page-title">{{ modeMeta.label }}</h1>
        <p class="page-sub">{{ modeMeta.desc }}</p>
        <div v-if="mode === 'practice' || mode === 'random'" class="mode-switch">
          <div class="segmented">
            <button
              :class="{ active: mode === 'practice' }"
              @click="switchMode('practice')"
            >
              顺序练习
            </button>
            <button
              :class="{ active: mode === 'random' }"
              @click="switchMode('random')"
            >
              随机练习
            </button>
          </div>
        </div>
      </div>
      <div class="head-actions" v-if="question">
        <button
          class="btn btn-sm"
          :class="isFavorite ? 'btn-accent' : 'btn-ghost'"
          @click="toggleFavorite"
        >
          <AppIcon name="star" :size="15" />
          {{ isFavorite ? "已收藏" : "收藏" }}
        </button>
        <button
          v-if="mode === 'error'"
          class="btn btn-sm btn-ghost"
          @click="removeFromErrors"
        >
          <AppIcon name="check" :size="15" />
          移出错题
        </button>
        <button class="btn btn-sm btn-soft" @click="navOpen = true">
          <AppIcon name="grid" :size="15" />
          题目列表
        </button>
      </div>
    </header>

    <div v-if="error" class="error-box">
      <AppIcon name="warning" :size="15" />
      {{ error }}
      <button class="btn btn-sm btn-ghost" @click="load(true)">重试</button>
    </div>

    <div v-if="loading" class="practice-loading">
      <div class="skeleton" style="height: 60px" />
      <div class="skeleton" style="height: 420px" />
    </div>

    <div v-else-if="finished || !total" class="empty surface">
      <div class="empty-icon">
        <AppIcon :name="total ? 'check' : mode === 'error' ? 'errors' : 'folder'" :size="28" />
      </div>
      <template v-if="total">
        <h2 class="empty-title">
          {{ mode === "random" ? "本轮随机练习完成" : "本轮练习完成" }}
        </h2>
        <p class="empty-text">
          共 {{ total }} 题，已作答 {{ correctCount + wrongCount }} 题，
          答对 {{ correctCount }} 题，正确率 {{ accuracy }}%。
        </p>
        <div class="done-actions">
          <button class="btn btn-ghost" @click="restart">再练一遍</button>
          <button class="btn btn-accent" @click="back">返回总览</button>
        </div>
      </template>
      <template v-else>
        <h2 class="empty-title">
          {{ mode === "error" ? "错题本还是空的" : mode === "favorite" ? "还没有收藏题目" : "这个类别还没有题目" }}
        </h2>
        <p class="empty-text">
          {{
            mode === "error"
              ? "练习中答错的题会自动收集到这里，先去练几题吧。"
              : mode === "favorite"
                ? "练习时点右上角收藏，就能在这里集中复习。"
                : "题库为空，请先在设置中导入 ZIP 题库包。"
          }}
        </p>
        <button
          class="btn btn-accent"
          @click="
            router.push(
              mode === 'practice' || mode === 'random'
                ? '/settings'
                : `/practice/practice/${category}`,
            )
          "
        >
          {{ mode === "practice" || mode === "random" ? "去导入题库" : "去练习" }}
        </button>
      </template>
    </div>

    <div v-else-if="question" class="practice-grid">
      <section class="q-card surface rise">
        <div class="q-top">
          <div class="q-meta">
            <span class="chip" :class="isMulti ? 'chip-teal' : 'chip-amber'">
              {{ isMulti ? "多选题" : "单选题" }}
            </span>
            <span class="chip chip-outline">{{ category }} 类</span>
            <span class="q-index mono">{{ index + 1 }} / {{ total }}</span>
          </div>
          <span class="q-code mono">{{ question.J }}</span>
        </div>

        <h2 class="q-stem">{{ question.Q }}</h2>

        <div v-if="question.F" class="q-image" @click="imageOpen = true">
          <img :src="question.F" alt="题目图片" />
          <span class="q-image-hint"><AppIcon name="search" :size="13" /> 点击放大</span>
        </div>

        <OptionList
          :options="options"
          :selected="selected"
          :revealed="revealed"
          :correct="answerLetters"
          :multi="isMulti"
          @select="select"
        />

        <transition name="page">
          <div v-if="revealed" class="feedback" :class="isCorrect ? 'ok' : 'bad'">
            <div class="feedback-icon">
              <AppIcon :name="isCorrect ? 'check' : 'x'" :size="18" :stroke-width="2.4" />
            </div>
            <div class="feedback-body">
              <strong>
                {{ isCorrect ? "回答正确" : hasSelection ? "回答错误" : "未作答" }}
              </strong>
              <p>
                正确答案：<b class="mono">{{ answerLetters }}</b>
                <span v-if="!isCorrect && hasSelection && mode !== 'error'">
                  · 已自动加入错题本
                </span>
                <span v-else-if="!isCorrect && mode === 'error'">· 这道题已收在错题本</span>
                <span v-else-if="!isCorrect && !hasSelection">· 未作答不会加入错题本</span>
              </p>
            </div>
          </div>
        </transition>

        <div class="q-actions">
          <button class="btn btn-ghost" @click="prev" :disabled="index === 0">
            <AppIcon name="chevronLeft" :size="16" />
            上一题
          </button>
          <span class="spacer" />
          <button v-if="!revealed" class="btn btn-accent btn-lg" @click="reveal">
            <AppIcon name="eye" :size="17" />
            查看答案
          </button>
          <button v-else class="btn btn-lg" @click="next">
            {{ index === total - 1 ? "完成本轮" : "下一题" }}
            <AppIcon name="arrowRight" :size="17" />
          </button>
        </div>
      </section>

      <aside class="q-side">
        <section class="surface card-pad side-card">
          <ProgressRing
            :value="progressPercent"
            :size="128"
            :label="mode === 'random' ? '本轮进度' : '本组进度'"
            :sub="`${index + 1} / ${total}`"
            color="var(--amber)"
          />
          <div class="side-stats">
            <div>
              <strong style="color: var(--green)">{{ correctCount }}</strong>
              <span>答对</span>
            </div>
            <div>
              <strong style="color: var(--red)">{{ wrongCount }}</strong>
              <span>答错</span>
            </div>
            <div>
              <strong>{{ accuracy }}%</strong>
              <span>正确率</span>
            </div>
          </div>
        </section>

        <section class="surface card-pad side-card">
          <h3 class="side-title">快捷键</h3>
          <ul class="shortcuts">
            <li><kbd>A</kbd><kbd>B</kbd><kbd>C</kbd><kbd>D</kbd><span>选择答案</span></li>
            <li><kbd>Enter</kbd><span>查看答案 / 下一题</span></li>
            <li><kbd>←</kbd><kbd>→</kbd><span>上一题 / 下一题</span></li>
            <li><kbd>F</kbd><span>收藏当前题</span></li>
          </ul>
        </section>

        <section class="surface card-pad side-card progress-mini">
          <div class="row gap-8">
            <span class="dot" />
            <span>{{ saving ? "正在保存进度…" : "练习进度自动保存" }}</span>
          </div>
          <div class="progress amber">
            <span :style="{ width: progressPercent + '%' }" />
          </div>
        </section>
      </aside>
    </div>

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

    <BaseModal
      :open="navOpen"
      :title="`题目列表 · ${modeMeta.label}`"
      :subtitle="`共 ${total} 题，点击跳转`"
      wide
      @close="navOpen = false"
    >
      <div class="nav-grid">
        <button
          v-for="(q, i) in questions"
          :key="i"
          class="nav-cell"
          :class="[results[i] ?? '', { current: i === index }]"
          @click="jumpTo(i)"
        >
          {{ i + 1 }}
        </button>
      </div>
    </BaseModal>
  </div>
</template>

<style scoped>
.practice-page {
  max-width: 1200px;
}
.practice-head {
  display: flex;
  align-items: flex-start;
  gap: 22px;
  margin-bottom: 24px;
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
  margin-top: 4px;
  transition: all 0.15s;
}
.back-btn:hover {
  color: var(--ink);
  border-color: var(--line-strong);
}
.head-copy {
  flex: 1;
}
.head-copy .page-title {
  margin-top: 5px;
}
.head-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}
.mode-switch {
  margin-top: 12px;
}
.mode-switch .segmented button {
  height: 28px;
  padding: 0 12px;
}

.practice-loading {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.practice-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 268px;
  gap: 20px;
  align-items: start;
}

.q-card {
  padding: 28px 32px 24px;
}
.q-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}
.q-meta {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-wrap: wrap;
}
.q-index {
  font-size: 11px;
  color: var(--muted);
  margin-left: 4px;
}
.q-code {
  font-size: 11px;
  color: var(--muted);
  letter-spacing: 0.06em;
}
.q-stem {
  font-size: 21px;
  line-height: 1.65;
  font-weight: 700;
  letter-spacing: -0.01em;
  margin-bottom: 20px;
}
.q-image {
  position: relative;
  margin: 0 0 20px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--line);
  cursor: zoom-in;
  background: var(--navy);
}
.q-image img {
  width: 100%;
  max-height: 320px;
  object-fit: contain;
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
.feedback {
  display: flex;
  gap: 12px;
  padding: 15px 17px;
  border-radius: 13px;
  margin-top: 18px;
}
.feedback.ok {
  background: var(--green-soft);
  border: 1px solid rgba(77, 155, 106, 0.25);
}
.feedback.bad {
  background: var(--red-soft);
  border: 1px solid rgba(208, 92, 77, 0.22);
}
.feedback-icon {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #fff;
  flex: none;
}
.feedback.ok .feedback-icon { background: var(--green); }
.feedback.bad .feedback-icon { background: var(--red); }
.feedback-body strong {
  font-size: 13.5px;
  font-weight: 700;
}
.feedback.ok .feedback-body strong { color: #2f7148; }
.feedback.bad .feedback-body strong { color: #a63b2e; }
.feedback-body p {
  margin-top: 4px;
  font-size: 12px;
  color: var(--ink-2);
}
.feedback-body b {
  color: var(--ink);
  font-size: 13px;
}
.q-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--line);
}

.q-side {
  display: flex;
  flex-direction: column;
  gap: 14px;
  position: sticky;
  top: 20px;
}
.side-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px;
}
.side-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  width: 100%;
  gap: 6px;
  margin-top: 8px;
}
.side-stats div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 9px 4px;
  border-radius: 10px;
  background: var(--surface-2);
}
.side-stats strong {
  font-size: 15px;
  font-weight: 800;
}
.side-stats span {
  font-size: 9.5px;
  color: var(--ink-3);
}
.side-title {
  align-self: flex-start;
  font-size: 13px;
  font-weight: 750;
  margin-bottom: 4px;
}
.shortcuts {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}
.shortcuts li {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10.5px;
  color: var(--ink-3);
}
.shortcuts span {
  margin-left: auto;
}
kbd {
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: 5px;
  background: var(--surface-2);
  border: 1px solid var(--line-strong);
  border-bottom-width: 2px;
  font-family: var(--font-mono);
  font-size: 9.5px;
  display: inline-grid;
  place-items: center;
  color: var(--ink-2);
}
.progress-mini {
  align-items: stretch;
  gap: 10px;
}
.progress-mini .row {
  font-size: 11px;
  color: var(--ink-3);
}

.done-actions {
  display: flex;
  gap: 10px;
  margin-top: 8px;
}
.image-full img {
  width: 100%;
  max-height: 70vh;
  object-fit: contain;
  border-radius: 12px;
}
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
  transition: all 0.14s;
}
.nav-cell:hover {
  border-color: var(--line-strong);
  background: var(--surface-2);
}
.nav-cell.correct {
  background: var(--green-soft);
  border-color: rgba(77, 155, 106, 0.35);
  color: #2f7148;
}
.nav-cell.wrong {
  background: var(--red-soft);
  border-color: rgba(208, 92, 77, 0.3);
  color: #a63b2e;
}
.nav-cell.current {
  border-color: var(--amber);
  box-shadow: 0 0 0 3px rgba(232, 163, 61, 0.16);
}

.error-box {
  display: flex;
  align-items: center;
  gap: 9px;
  background: var(--red-soft);
  color: #a63b2e;
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 13px;
  margin-bottom: 18px;
}

@media (max-width: 1180px) {
  .practice-grid {
    grid-template-columns: 1fr;
  }
  .q-side {
    position: static;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
  }
  .progress-mini {
    display: none;
  }
}
</style>
