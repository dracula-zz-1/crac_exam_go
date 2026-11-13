<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "@/api";
import type { ExamDetail, ExamResult } from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import OptionList from "@/components/OptionList.vue";
import ProgressRing from "@/components/ProgressRing.vue";

const props = defineProps<{ examId: string }>();
const router = useRouter();

const result = ref<ExamResult | null>(null);
const details = ref<ExamDetail[]>([]);
const loading = ref(true);
const error = ref("");
const filter = ref<"all" | "wrong" | "unanswered">("all");

const wrongDetails = computed(() => details.value.filter((d) => !d.isCorrect && d.userAnswer));
const unansweredDetails = computed(() => details.value.filter((d) => !d.userAnswer));
const filtered = computed(() => {
  if (filter.value === "wrong") return wrongDetails.value;
  if (filter.value === "unanswered") return unansweredDetails.value;
  return details.value;
});
const scorePercent = computed(() =>
  result.value && result.value.totalCount
    ? Math.round((result.value.score / result.value.totalCount) * 100)
    : 0,
);
const accent = computed(() => (result.value?.passed ? "var(--teal)" : "var(--red)"));

async function load() {
  loading.value = true;
  error.value = "";
  const id = Number(props.examId);
  if (!id) {
    error.value = "无效的考试记录";
    loading.value = false;
    return;
  }
  try {
    const [r, d] = await Promise.all([api.getResult(id), api.getDetails(id)]);
    result.value = r;
    details.value = d;
  } catch (e: any) {
    error.value = e?.message || "加载考试结果失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function fmtDate(iso: string) {
  return new Date(iso).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
function fmtDuration(sec: number) {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return m ? `${m} 分 ${s} 秒` : `${s} 秒`;
}
function optionsOf(d: ExamDetail) {
  return [
    { key: "A", text: d.optionA },
    { key: "B", text: d.optionB },
    { key: "C", text: d.optionC },
    { key: "D", text: d.optionD },
  ];
}
function selectedOf(d: ExamDetail) {
  return (d.userAnswer || "").split("").filter(Boolean);
}
</script>

<template>
  <div class="page result-page">
    <div v-if="loading" class="result-loading">
      <div class="skeleton" style="height: 260px" />
      <div class="skeleton" style="height: 360px" />
    </div>

    <div v-else-if="error" class="empty surface">
      <div class="empty-icon"><AppIcon name="warning" :size="26" /></div>
      <h2 class="empty-title">{{ error }}</h2>
      <p class="empty-text">考试记录可能已被删除，或尚未交卷。</p>
      <button class="btn btn-ghost" @click="router.push('/home')">返回总览</button>
    </div>

    <template v-else-if="result">
      <header class="result-head">
        <div>
          <span class="eyebrow">RESULT · {{ result.category }} 类</span>
          <h1 class="page-title">{{ result.passed ? "恭喜，通过了" : "这次差一点" }}</h1>
          <p class="page-sub">
            {{ fmtDate(result.examDate) }} · 用时 {{ fmtDuration(result.durationSeconds) }}
          </p>
        </div>
        <div class="head-actions">
          <button class="btn btn-ghost" @click="router.push('/home')">
            <AppIcon name="home" :size="16" />
            返回总览
          </button>
          <button class="btn btn-accent" @click="router.push(`/exam/${result.category}`)">
            <AppIcon name="refresh" :size="16" />
            再考一次
          </button>
        </div>
      </header>

      <section class="score-hero surface" :class="{ failed: !result.passed }">
        <div class="score-ring">
          <ProgressRing
            :value="scorePercent"
            :size="176"
            :stroke="13"
            :color="accent"
            label="得分率"
            :sub="`${result.score} / ${result.totalCount}`"
          />
        </div>
        <div class="score-main">
          <div class="score-badge" :class="result.passed ? 'pass' : 'fail'">
            <AppIcon :name="result.passed ? 'trophy' : 'target'" :size="15" />
            {{ result.passed ? "通过" : "未通过" }}
          </div>
          <div class="score-line">
            <strong :style="{ color: accent }">{{ result.score }}</strong>
            <span>/ {{ result.totalCount }} 分</span>
          </div>
          <p>
            及格线 <b>{{ result.passScore }}</b> 分，
            {{ result.passed ? `高出及格线 ${result.score - result.passScore} 分` : `距离及格还差 ${result.passScore - result.score} 分` }}。
          </p>
          <div class="score-actions">
            <button class="btn btn-sm btn-ghost" @click="router.push({ path: '/errors', query: { category: result.category } })">
              <AppIcon name="errors" :size="14" />
              复习错题
            </button>
            <button class="btn btn-sm btn-ghost" @click="router.push(`/practice/practice/${result.category}`)">
              <AppIcon name="practice" :size="14" />
              继续练习
            </button>
          </div>
        </div>
        <div class="score-stats">
          <div>
            <strong style="color: var(--green)">{{ result.correctCount }}</strong>
            <span>答对</span>
          </div>
          <div>
            <strong style="color: var(--red)">{{ wrongDetails.length }}</strong>
            <span>答错</span>
          </div>
          <div>
            <strong style="color: var(--amber-deep)">{{ unansweredDetails.length }}</strong>
            <span>未答</span>
          </div>
          <div>
            <strong>{{ fmtDuration(result.durationSeconds) }}</strong>
            <span>用时</span>
          </div>
        </div>
      </section>

      <section class="review surface">
        <div class="review-head">
          <div>
            <h2 class="section-title">试卷复盘</h2>
            <p class="section-sub">逐题查看正确答案与你的作答</p>
          </div>
          <div class="segmented">
            <button :class="{ active: filter === 'all' }" @click="filter = 'all'">
              全部 {{ details.length }}
            </button>
            <button :class="{ active: filter === 'wrong' }" @click="filter = 'wrong'">
              答错 {{ wrongDetails.length }}
            </button>
            <button :class="{ active: filter === 'unanswered' }" @click="filter = 'unanswered'">
              未答 {{ unansweredDetails.length }}
            </button>
          </div>
        </div>

        <div v-if="!filtered.length" class="empty" style="padding: 40px 0">
          <div class="empty-icon"><AppIcon name="check" :size="24" /></div>
          <strong class="empty-title">这里没有题目</strong>
          <span class="empty-text">切换上方筛选看看其他题目。</span>
        </div>

        <div v-else class="review-list">
          <article v-for="d in filtered" :key="d.id" class="review-item">
            <div class="review-meta">
              <span class="review-seq mono">{{ d.seq }}</span>
              <span class="chip" :class="d.type === 2 ? 'chip-teal' : 'chip-amber'">
                {{ d.type === 2 ? "多选" : "单选" }}
              </span>
              <span
                class="chip"
                :class="d.isCorrect ? 'chip-green' : d.userAnswer ? 'chip-red' : 'chip-outline'"
              >
                {{ d.isCorrect ? "正确" : d.userAnswer ? "错误" : "未答" }}
              </span>
              <span class="review-answers mono">
                你的答案 {{ d.userAnswer || "—" }} · 正确答案 {{ d.correctAnswer }}
              </span>
            </div>
            <h3 class="review-stem">{{ d.questionText }}</h3>
            <div v-if="d.imageData" class="review-image">
              <img :src="d.imageData" alt="题目图片" />
            </div>
            <OptionList
              :options="optionsOf(d)"
              :selected="selectedOf(d)"
              revealed
              :correct="d.correctAnswer"
              disabled
            />
          </article>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.result-page {
  max-width: 1120px;
}
.head-actions {
  display: flex;
  gap: 10px;
}
.result-loading {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.score-hero {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 34px;
  align-items: center;
  padding: 30px 36px;
  margin-bottom: 22px;
  position: relative;
  overflow: hidden;
}
.score-hero::after {
  content: "";
  position: absolute;
  right: -70px;
  top: -90px;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(42, 157, 143, 0.14), transparent 65%);
}
.score-hero.failed::after {
  background: radial-gradient(circle, rgba(208, 92, 77, 0.12), transparent 65%);
}
.score-ring {
  position: relative;
  z-index: 1;
}
.score-main {
  position: relative;
  z-index: 1;
}
.score-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 27px;
  padding: 0 11px;
  border-radius: 99px;
  font-size: 11.5px;
  font-weight: 700;
}
.score-badge.pass {
  background: var(--teal-soft);
  color: #147065;
}
.score-badge.fail {
  background: var(--red-soft);
  color: #a63b2e;
}
.score-line {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 10px;
}
.score-line strong {
  font-size: 58px;
  font-weight: 880;
  letter-spacing: -0.05em;
  line-height: 0.95;
}
.score-line span {
  font-size: 15px;
  color: var(--ink-3);
  font-weight: 600;
}
.score-main p {
  margin-top: 10px;
  font-size: 12.5px;
  color: var(--ink-3);
}
.score-main p b {
  color: var(--ink);
}
.score-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}
.score-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(90px, 1fr));
  gap: 8px;
  position: relative;
  z-index: 1;
}
.score-stats div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--surface-2);
}
.score-stats strong {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.score-stats span {
  font-size: 10px;
  color: var(--ink-3);
}

.review {
  padding: 24px 26px 28px;
}
.review-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 20px;
}
.review-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.review-item {
  border: 1px solid var(--line);
  border-radius: 15px;
  padding: 18px 20px;
  background: var(--surface);
}
.review-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.review-seq {
  width: 26px;
  height: 26px;
  border-radius: 8px;
  background: var(--navy);
  color: #f4f0e6;
  display: grid;
  place-items: center;
  font-size: 10.5px;
  font-weight: 700;
}
.review-answers {
  margin-left: auto;
  font-size: 10.5px;
  color: var(--ink-3);
}
.review-stem {
  font-size: 16px;
  line-height: 1.65;
  font-weight: 650;
  margin-bottom: 14px;
}
.review-image {
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 14px;
  border: 1px solid var(--line);
  background: var(--navy);
}
.review-image img {
  width: 100%;
  max-height: 260px;
  object-fit: contain;
}

@media (max-width: 1180px) {
  .score-hero {
    grid-template-columns: auto 1fr;
  }
  .score-stats {
    grid-column: span 2;
    grid-template-columns: repeat(4, 1fr);
  }
}
</style>
