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
// 乱序后正确答案会换位置，这里跟着重映射到显示字母
const answerLetters = computed(() =>
  sortedLetters((question.value?.T ?? "").replace(/[A-D]/gi, (m) => displayLetter(m))),
);
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
// 只有逐题/随机两种模式共用一个入口，才需要显示切换开关
const switchable = computed(
  () => mode.value === "practice" || mode.value === "random",
);
const answerTone = computed(() => {
  if (!revealed.value) return "idle";
  return isCorrect.value ? "ok" : "bad";
});
const answerNote = computed(() => {
  if (!revealed.value) return "";
  if (isCorrect.value) {
    return mode.value === "error" ? "已掌握，可移出错题本" : "";
  }
  if (!hasSelection.value) return "未作答不会加入错题本";
  if (mode.value === "error") return "这道题已收在错题本";
  return "已自动加入错题本";
});

// 选项乱序：optionOrder[i] 表示第 i 个显示位（A/B/C/D）对应的原始选项字母
const OPTION_LABELS = ["A", "B", "C", "D"];
const optionOrder = ref<string[]>([...OPTION_LABELS]);

function shuffleOptionOrder() {
  const order = [...OPTION_LABELS];
  for (let i = order.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [order[i], order[j]] = [order[j], order[i]];
  }
  optionOrder.value = order;
}

// 原始字母 → 乱序后显示的字母
function displayLetter(original: string) {
  const idx = optionOrder.value.indexOf(original.toUpperCase());
  return idx >= 0 ? OPTION_LABELS[idx] : original.toUpperCase();
}

const options = computed(() => {
  const q = question.value;
  if (!q) return [];
  const text: Record<string, string> = { A: q.A, B: q.B, C: q.C, D: q.D };
  return optionOrder.value.map((original, i) => ({
    key: OPTION_LABELS[i],
    text: text[original] ?? "",
  }));
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
  shuffleOptionOrder();
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
      <div class="head-mode">
        <div v-if="switchable" class="segmented mode-tabs">
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
        <template v-else>
          <h1 class="mode-title">{{ modeMeta.label }}</h1>
          <span class="mode-note">{{ category }} 类 · {{ modeMeta.desc }}</span>
        </template>
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
        <div class="q-body">
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

          <OptionList
            :options="options"
            :selected="selected"
            :revealed="revealed"
            :correct="answerLetters"
            :multi="isMulti"
            @select="select"
          />
        </div>

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
        <!-- 图片区：位置固定，有配图才显示卡片（点击放大），无配图留空但占位不变 -->
        <section
          v-if="question.F"
          class="surface media-card"
          @click="imageOpen = true"
        >
          <img :src="question.F" alt="题目图片" />
        </section>
        <div v-else class="media-placeholder" />

        <section class="surface card-pad side-card progress-card">
          <ProgressRing
            :value="progressPercent"
            :size="108"
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

        <section class="surface side-card answer-card" :class="answerTone">
          <h3 class="side-title">答案</h3>
          <template v-if="revealed">
            <div class="answer-state">
              <AppIcon
                :name="isCorrect ? 'check' : hasSelection ? 'x' : 'info'"
                :size="15"
                :stroke-width="2.6"
              />
              <strong>
                {{ isCorrect ? "回答正确" : hasSelection ? "回答错误" : "未作答" }}
              </strong>
            </div>
            <div class="answer-key">
              <span>正确答案</span>
              <b class="mono">{{ answerLetters }}</b>
            </div>
            <p class="answer-note">{{ answerNote }}</p>
          </template>
          <p v-else class="answer-idle">
            选出答案后，点「查看答案」揭晓正确答案
          </p>
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
/* 整页锁定为一屏高度，练习过程中不出现纵向滚动条 */
.practice-page {
  max-width: 1200px;
  height: calc(100vh - 64px);
  padding-top: 18px;
  padding-bottom: 20px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
/* 切页动画不再上下位移，避免瞬间把一屏页面挤出滚动条 */
.practice-page.page-enter-from,
.practice-page.page-leave-to {
  transform: none;
}
.practice-head {
  flex: none;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
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
  transition: all 0.15s;
}
.back-btn:hover {
  color: var(--ink);
  border-color: var(--line-strong);
}
.head-mode {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.mode-tabs button {
  height: 32px;
  padding: 0 16px;
  font-size: 13px;
}
.mode-title {
  font-size: 19px;
  font-weight: 800;
  letter-spacing: -0.01em;
  white-space: nowrap;
}
.mode-note {
  font-size: 12px;
  color: var(--ink-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.head-actions {
  flex: none;
  display: flex;
  gap: 8px;
}

.practice-loading {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.practice-grid {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 268px;
  gap: 18px;
  align-items: stretch;
}

.q-card {
  padding: 20px 26px 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}
/* 题干与选项区域：内容过长时只在卡片内部滚动，不撑高页面 */
.q-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 6px;
}
.q-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
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
  font-size: 18px;
  line-height: 1.55;
  font-weight: 700;
  letter-spacing: -0.01em;
  margin-bottom: 14px;
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
.q-actions .btn {
  height: 44px;
}
.q-actions .btn-lg {
  height: 46px;
}

.q-side {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  height: 100%;
}

/* 图片区：固定占住右栏顶部这块位置，无配图时用占位块代替，下方卡片位置不变 */
.media-card {
  flex: 1;
  min-height: 0;
  padding: 12px;
  display: grid;
  place-items: center;
  overflow: hidden;
  background: var(--navy);
  cursor: zoom-in;
}
.media-card img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.media-placeholder {
  flex: 1;
  min-height: 0;
}

.side-card {
  flex: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 18px;
}
.progress-card {
  padding-top: 16px;
  padding-bottom: 16px;
}
.side-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  width: 100%;
  gap: 6px;
  margin-top: 6px;
}
.side-stats div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 8px 4px;
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
  margin-bottom: 2px;
}

/* 答案区：固定高度，揭晓答案前后尺寸不变 */
.answer-card {
  height: 150px;
  padding: 16px 18px;
  justify-content: flex-start;
  align-items: stretch;
  gap: 7px;
  overflow: hidden;
  border-left: 3px solid var(--line-strong);
}
.answer-card.ok { border-left-color: var(--green); }
.answer-card.bad { border-left-color: var(--red); }
.answer-state {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13.5px;
  font-weight: 700;
}
.answer-card.ok .answer-state { color: #2f7148; }
.answer-card.bad .answer-state { color: #a63b2e; }
.answer-card.idle .answer-state { color: var(--ink-2); }
.answer-key {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 11px;
  border-radius: 10px;
  background: var(--surface-2);
}
.answer-key span {
  font-size: 11px;
  color: var(--ink-3);
}
.answer-key b {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.06em;
  color: var(--ink);
}
.answer-note {
  font-size: 11px;
  line-height: 1.5;
  color: var(--ink-3);
}
.answer-idle {
  font-size: 11.5px;
  line-height: 1.6;
  color: var(--muted);
}

.progress-mini {
  align-items: stretch;
  gap: 10px;
  padding-top: 14px;
  padding-bottom: 14px;
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

/* 窗口极窄时（低于应用最小宽度）才退化为可滚动的单列布局 */
@media (max-width: 1040px) {
  .practice-page {
    height: auto;
    min-height: calc(100vh - 64px);
    overflow: visible;
    padding-bottom: 40px;
  }
  .practice-grid {
    grid-template-columns: 1fr;
  }
  .q-card {
    overflow: visible;
  }
  .q-body {
    overflow: visible;
  }
  .q-side {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    height: auto;
  }
  .media-card,
  .media-placeholder {
    min-height: 180px;
    grid-column: span 2;
  }
}
</style>
