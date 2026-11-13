<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useBankStore } from "@/stores/bank";
import { useToastStore } from "@/stores/toast";
import type {
  BankUpdateStatus,
  Category,
  Dashboard,
  ExamSummary,
  ScorePoint,
} from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import StatTile from "@/components/StatTile.vue";
import ScoreTrend from "@/components/ScoreTrend.vue";
import { nextLevelCategory } from "@/lib/license";

const router = useRouter();
const session = useSessionStore();
// 默认类别 = 用户下一次要考的等级（无证→A，A→B，B→C，C→C）
const defaultCategory = computed(() => nextLevelCategory(session.user?.level));
const bank = useBankStore();
const toast = useToastStore();

const dash = ref<Dashboard | null>(null);
const series = ref<ScorePoint[]>([]);
const loading = ref(true);
const chartLoading = ref(false);
const error = ref("");
const days = ref(30);
const chartCategory = ref<"" | Category>("");
const updateBanner = ref<BankUpdateStatus | null>(null);
let offUpdateStatus: (() => void) | null = null;
let offImportProgress: (() => void) | null = null;

const onlineLoading = ref(false);
const onlineStage = ref("");
const onlinePercent = ref(0);

const CATEGORY_META: Record<Category, { label: string; desc: string; minutes: number }> = {
  A: { label: "A 类", desc: "基础操作能力", minutes: 40 },
  B: { label: "B 类", desc: "常用频段与设备", minutes: 60 },
  C: { label: "C 类", desc: "短波与高级操作", minutes: 90 },
};

const greeting = computed(() => {
  const h = new Date().getHours();
  if (h < 6) return "夜深了";
  if (h < 11) return "早上好";
  if (h < 13) return "中午好";
  if (h < 18) return "下午好";
  return "晚上好";
});

const subline = computed(() => {
  if (!dash.value) return "正在加载学习数据…";
  if (dash.value.totalQuestions === 0) return "题库还是空的，先导入一个 ZIP 题库包吧。";
  if (dash.value.lastExam) {
    return `上次 ${dash.value.lastExam.category} 类考试 ${dash.value.lastExam.score}/${dash.value.lastExam.total} 分，继续保持。`;
  }
  return "今天也来听一段电波，从一题开始。";
});

const categories = computed(() =>
  (["A", "B", "C"] as Category[]).map((cat) => {
    const count = dash.value?.categories.find((c) => c.category === cat);
    const progress = dash.value?.practice.find((p) => p.category === cat);
    const best = dash.value?.bestScores?.[cat] ?? null;
    return {
      cat,
      meta: CATEGORY_META[cat],
      total: count?.total ?? 0,
      single: count?.single ?? 0,
      multiple: count?.multiple ?? 0,
      image: count?.withImage ?? 0,
      percent: progress?.percent ?? 0,
      answered: progress?.answered ?? 0,
      best,
      pass: cat === "A" ? 30 : cat === "B" ? 45 : 70,
    };
  }),
);

const recent = computed<ExamSummary[]>(() => dash.value?.recentExams ?? []);

const chartPassScore = computed(() => {
  if (!chartCategory.value) return undefined;
  return chartCategory.value === "A" ? 30 : chartCategory.value === "B" ? 45 : 70;
});

const chartMaxScore = computed(() => {
  if (!chartCategory.value) return undefined;
  return chartCategory.value === "A" ? 40 : chartCategory.value === "B" ? 60 : 90;
});

const avgProgress = computed(() => {
  const list = dash.value?.practice ?? [];
  if (!list.length) return 0;
  return Math.round(list.reduce((s, p) => s + p.percent, 0) / list.length);
});

async function load() {
  if (!session.user) return;
  loading.value = true;
  error.value = "";
  try {
    dash.value = await api.getDashboard(session.user.id);
    await loadSeries();
  } catch (e: any) {
    error.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadSeries() {
  if (!session.user) return;
  chartLoading.value = true;
  try {
    series.value = await api.getScoreSeries(
      session.user.id,
      chartCategory.value,
      days.value,
    );
  } catch {
    series.value = [];
  } finally {
    chartLoading.value = false;
  }
}

async function loadUpdateStatus() {
  try {
    const st = await api.getBankUpdateStatus();
    updateBanner.value = st.available ? st : null;
  } catch {
    updateBanner.value = null;
  }
}

async function onlineFetch() {
  if (!session.user || onlineLoading.value) return;
  onlineLoading.value = true;
  onlineStage.value = "正在检查 CRAC 官方题库…";
  onlinePercent.value = 0;
  try {
    let st = await api.getBankUpdateStatus();
    if (!st.available) st = await api.checkBankUpdate();
    if (!st.available) {
      toast.warning("未发现可下载的题库", st.message || "请改用离线 ZIP 导入");
      return;
    }
    onlineStage.value = "正在下载并导入题库…";
    const result = await api.applyBankUpdate();
    bank.invalidate();
    toast.success("题库已就绪", result.message);
    await load();
  } catch (e: any) {
    toast.error("在线获取失败", e?.message || "请检查网络，或改用离线 ZIP 导入");
  } finally {
    onlineLoading.value = false;
    onlineStage.value = "";
    onlinePercent.value = 0;
  }
}

watch([days, chartCategory], loadSeries);

onMounted(() => {
  load();
  loadUpdateStatus();
  offUpdateStatus = api.onBankUpdateStatus((st) => {
    updateBanner.value = st.available ? st : null;
    if (st.message.includes("已更新") || st.message.includes("已自动更新")) {
      load();
      loadUpdateStatus();
    }
  });
  offImportProgress = api.onImportProgress((p) => {
    if (!onlineLoading.value) return;
    onlinePercent.value = p.percent;
    onlineStage.value = p.message;
  });
});

onBeforeUnmount(() => {
  offUpdateStatus?.();
  offImportProgress?.();
});

function goPractice(cat: Category) {
  router.push(`/practice/practice/${cat}`);
}
function goRandom(cat: Category) {
  router.push(`/practice/random/${cat}`);
}
function goExam(cat: Category) {
  router.push(`/exam/${cat}`);
}
function goResult(id: number) {
  router.push(`/result/${id}`);
}
function goErrors(cat?: Category) {
  router.push({ path: "/errors", query: cat ? { category: cat } : {} });
}
function goFavorites() {
  router.push("/favorites");
}

function fmtDate(iso: string) {
  const d = new Date(iso);
  return d.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}
function fmtDuration(sec: number) {
  if (!sec) return "—";
  const m = Math.floor(sec / 60);
  return m > 0 ? `${m} 分钟` : `${sec} 秒`;
}
</script>

<template>
  <div class="page dashboard">
    <header class="page-head">
      <div>
        <span class="eyebrow">OVERVIEW · 学习总览</span>
        <h1 class="page-title">{{ greeting }}，{{ session.user?.username }}</h1>
        <p class="page-sub">{{ subline }}</p>
      </div>
      <div class="head-actions">
        <button class="btn btn-ghost" @click="router.push('/settings')">
          <AppIcon name="upload" :size="16" />
          导入题库
        </button>
        <button class="btn btn-accent" @click="goExam(defaultCategory)">
          <AppIcon name="zap" :size="16" />
          开始 {{ defaultCategory }} 类模拟考试
        </button>
      </div>
    </header>

    <div v-if="updateBanner" class="update-banner">
      <span class="update-banner-icon">
        <AppIcon name="refresh" :size="17" />
      </span>
      <div class="update-banner-text">
        <strong>题库有更新</strong>
        <span>{{ updateBanner.remoteVersion || "发现 CRAC 官方新版本" }}</span>
      </div>
      <button class="btn btn-sm btn-accent" @click="router.push('/settings')">
        去更新
      </button>
    </div>

    <div v-if="error" class="error-box">
      <AppIcon name="warning" :size="15" />
      {{ error }}
      <button class="btn btn-sm btn-ghost" @click="load">重试</button>
    </div>

    <div v-if="loading" class="skeleton-grid">
      <div v-for="i in 4" :key="i" class="skeleton" style="height: 104px" />
      <div class="skeleton" style="height: 260px; grid-column: span 2" />
      <div class="skeleton" style="height: 260px; grid-column: span 2" />
    </div>

    <template v-else-if="dash">
      <section class="stat-row rise">
        <StatTile
          label="题库总量"
          :value="dash.totalQuestions"
          suffix="题"
          icon="bank"
          tone="amber"
          :hint="`${dash.withImages} 道带图 · ${dash.categories.length} 个类别`"
        />
        <StatTile
          label="考试通过率"
          :value="dash.passRate"
          suffix="%"
          icon="trophy"
          tone="teal"
          :hint="`${dash.passedExams} / ${dash.totalExams} 次通过`"
        />
        <StatTile
          label="错题待复习"
          :value="dash.errorCount"
          suffix="题"
          icon="errors"
          tone="red"
          hint="答错后自动收集"
        />
        <StatTile
          label="平均练习进度"
          :value="avgProgress"
          suffix="%"
          icon="target"
          tone="blue"
          :hint="`累计学习 ${dash.totalStudyMinutes} 分钟`"
        />
      </section>

      <section v-if="dash.totalQuestions === 0" class="empty-bank surface rise rise-1">
        <div class="empty-bank-art">
          <AppIcon name="folder" :size="34" :stroke-width="1.5" />
        </div>
        <div class="empty-bank-main">
          <h3>还没有题库</h3>
          <p>
            你可以导入本地 CRAC 题库 ZIP，或直接从 CRAC 官网在线获取最新题库。
            在线获取会检查题库页、资料下载栏目与站内搜索。
          </p>
          <div v-if="onlineLoading" class="online-progress">
            <div class="online-progress-top">
              <span>{{ onlineStage || "正在获取题库…" }}</span>
              <strong>{{ onlinePercent }}%</strong>
            </div>
            <div class="progress amber">
              <span :style="{ width: onlinePercent + '%' }" />
            </div>
          </div>
        </div>
        <div class="empty-bank-actions">
          <button class="btn btn-ghost" :disabled="onlineLoading" @click="router.push('/settings')">
            <AppIcon name="upload" :size="16" />
            离线导入 ZIP
          </button>
          <button class="btn btn-accent" :disabled="onlineLoading" @click="onlineFetch">
            <span v-if="onlineLoading" class="spinner dark" />
            <AppIcon v-else name="download" :size="16" />
            {{ onlineLoading ? "正在获取…" : "在线获取官方题库" }}
          </button>
        </div>
      </section>

      <template v-else>
        <section class="surface card-pad category-section rise rise-1">
          <div class="section-head">
            <div>
              <h2 class="section-title">分类训练</h2>
              <p class="section-sub">选择类别，开始逐题练习或全真模拟</p>
            </div>
            <span class="chip chip-outline">A / B / C</span>
          </div>

          <div class="cat-grid">
              <article
                v-for="item in categories"
                :key="item.cat"
                class="cat-card"
                :class="`cat-${item.cat}`"
              >
                <span class="cat-watermark">{{ item.cat }}</span>
                <header class="cat-head">
                  <div>
                    <h3>{{ item.meta.label }}</h3>
                    <p>{{ item.meta.desc }}</p>
                  </div>
                  <span class="chip" :class="`chip-${item.cat === 'A' ? 'amber' : item.cat === 'B' ? 'teal' : 'blue'}`">
                    {{ item.meta.minutes }} 分钟
                  </span>
                </header>

                <div class="cat-meta">
                  <span><strong>{{ item.total }}</strong> 题</span>
                  <span>单选 {{ item.single }}</span>
                  <span>多选 {{ item.multiple }}</span>
                </div>

                <div class="cat-progress">
                  <div class="cat-progress-top">
                    <span>练习进度</span>
                    <strong>{{ item.percent }}%</strong>
                  </div>
                  <div class="progress">
                    <span :style="{ width: item.percent + '%' }" />
                  </div>
                </div>

                <div class="cat-best">
                  <span>最佳 {{ item.best !== null ? item.best + ' 分' : '尚未考试' }}</span>
                  <span class="dim">及格 {{ item.pass }} 分</span>
                </div>

                <div class="cat-actions">
                  <button class="btn btn-sm btn-soft" @click="goPractice(item.cat)">
                    <AppIcon name="practice" :size="14" />
                    练习
                  </button>
                  <button class="btn btn-sm" @click="goExam(item.cat)">
                    <AppIcon name="exam" :size="14" />
                    考试
                  </button>
                  <button
                    class="btn btn-sm btn-ghost"
                    title="随机练习（错题加权）"
                    @click="goRandom(item.cat)"
                  >
                    <AppIcon name="sparkles" :size="14" />
                  </button>
                  <button
                    class="btn btn-sm btn-ghost"
                    title="错题本"
                    @click="goErrors(item.cat)"
                  >
                    <AppIcon name="errors" :size="14" />
                  </button>
                </div>
              </article>
            </div>
        </section>

        <!-- 成绩趋势：整行铺开，图表才看得清 -->
        <section class="surface card-pad trend-section rise rise-1">
          <div class="section-head">
            <div>
              <h2 class="section-title">成绩趋势</h2>
              <p class="section-sub">最近 {{ days }} 天的考试表现</p>
            </div>
            <div class="segmented">
              <button :class="{ active: days === 7 }" @click="days = 7">7 天</button>
              <button :class="{ active: days === 30 }" @click="days = 30">30 天</button>
              <button :class="{ active: days === 180 }" @click="days = 180">半年</button>
            </div>
          </div>

          <div class="chart-cats">
            <button
              class="chart-cat"
              :class="{ active: chartCategory === '' }"
              @click="chartCategory = ''"
            >
              全部
            </button>
            <button
              v-for="cat in (['A', 'B', 'C'] as Category[])"
              :key="cat"
              class="chart-cat"
              :class="{ active: chartCategory === cat }"
              @click="chartCategory = cat"
            >
              {{ cat }} 类
            </button>
          </div>

          <div :class="{ 'chart-dim': chartLoading }">
            <ScoreTrend
              :points="series"
              :pass-score="chartPassScore"
              :max-score="chartMaxScore"
            />
          </div>
        </section>

        <div class="grid-2 dashboard-lower">
          <div class="stack gap-22">
          <section class="surface card-pad rise rise-2">
            <div class="section-head">
              <div>
                <h2 class="section-title">最近考试</h2>
                <p class="section-sub">点击查看试卷复盘</p>
              </div>
              <span class="chip chip-outline">{{ dash.totalExams }} 次记录</span>
            </div>

            <div v-if="recent.length" class="exam-list">
              <button
                v-for="exam in recent"
                :key="exam.id"
                class="exam-row"
                @click="goResult(exam.id)"
              >
                <span class="exam-cat" :class="`cat-${exam.category}`">{{ exam.category }}</span>
                <div class="exam-main">
                  <strong>{{ exam.category }} 类模拟考试</strong>
                  <span>{{ fmtDate(exam.examDate) }} · 用时 {{ fmtDuration(exam.durationSeconds) }}</span>
                </div>
                <div class="exam-score">
                  {{ exam.score }}<span>/{{ exam.total }}</span>
                </div>
                <span class="chip" :class="exam.passed ? 'chip-green' : 'chip-red'">
                  {{ exam.passed ? "通过" : "未通过" }}
                </span>
                <AppIcon name="chevronRight" :size="15" class="exam-arrow" />
              </button>
            </div>

            <div v-else class="empty" style="padding: 34px 0">
              <div class="empty-icon"><AppIcon name="exam" :size="24" /></div>
              <strong class="empty-title">还没有考试记录</strong>
              <span class="empty-text">选一个类别开始模拟考试，成绩会自动记录在这里。</span>
            </div>
          </section>
        </div>

        <aside class="stack gap-22">
          <section class="surface card-pad rise rise-2">
            <div class="section-head">
              <div>
                <h2 class="section-title">学习进度</h2>
                <p class="section-sub">每个类别上次练到哪里</p>
              </div>
            </div>

            <div class="progress-list">
              <div v-for="item in categories" :key="item.cat" class="progress-row">
                <div class="progress-row-head">
                  <span class="progress-cat" :class="`cat-${item.cat}`">{{ item.cat }}</span>
                  <span>{{ item.answered }} / {{ item.total }} 题</span>
                  <strong>{{ item.percent }}%</strong>
                </div>
                <div class="progress" :class="{ amber: item.cat === 'A', blue: item.cat === 'C' }">
                  <span :style="{ width: item.percent + '%' }" />
                </div>
              </div>
            </div>

            <div class="quick-grid">
              <button class="quick" @click="goErrors()">
                <span class="quick-icon red"><AppIcon name="errors" :size="17" /></span>
                <span>
                  <strong>{{ dash.errorCount }}</strong>
                  <em>错题本</em>
                </span>
              </button>
              <button class="quick" @click="goFavorites">
                <span class="quick-icon amber"><AppIcon name="star" :size="17" /></span>
                <span>
                  <strong>{{ dash.favoriteCount }}</strong>
                  <em>收藏夹</em>
                </span>
              </button>
            </div>
          </section>
        </aside>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped>
.head-actions {
  display: flex;
  gap: 10px;
}
.update-banner {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 11px 14px;
  margin-bottom: 16px;
  border-radius: 13px;
  background: linear-gradient(90deg, var(--amber-soft), #fffaf0);
  border: 1px solid rgba(232, 163, 61, 0.28);
}
.update-banner-icon {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--surface);
  color: var(--amber-deep);
  flex: none;
}
.update-banner-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.update-banner-text strong {
  font-size: 12.5px;
  font-weight: 700;
  color: #7a5211;
}
.update-banner-text span {
  font-size: 10.5px;
  color: #8a631f;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.update-banner .btn {
  margin-left: auto;
  flex: none;
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
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 15px;
  margin-bottom: 22px;
}
.category-section {
  margin-bottom: 22px;
}
.dashboard-lower {
  margin-top: 0;
}
.trend-section {
  margin-bottom: 22px;
}
.trend-section .chart-cats {
  margin-bottom: 14px;
}
.empty-bank {
  display: flex;
  align-items: center;
  gap: 22px;
  padding: 28px 30px;
  margin-bottom: 22px;
  flex-wrap: wrap;
}
.empty-bank-main {
  flex: 1;
  min-width: 260px;
}
.empty-bank-actions {
  display: flex;
  gap: 9px;
  flex: none;
  flex-wrap: wrap;
}
.online-progress {
  margin-top: 12px;
  max-width: 420px;
}
.online-progress-top {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--ink-3);
  margin-bottom: 6px;
}
.online-progress-top strong {
  color: var(--amber-deep);
}
.empty-bank .btn {
  margin-left: 0;
}
.empty-bank-art {
  width: 66px;
  height: 66px;
  border-radius: 20px;
  background: var(--amber-soft);
  color: #9a6514;
  display: grid;
  place-items: center;
  flex: none;
}
.empty-bank h3 {
  font-size: 18px;
  font-weight: 750;
}
.empty-bank p {
  margin-top: 6px;
  font-size: 12.5px;
  color: var(--ink-3);
  line-height: 1.7;
  max-width: 620px;
}
.empty-bank .btn {
  margin-left: 0;
  flex: none;
}

.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.cat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 13px;
}
.cat-card {
  position: relative;
  padding: 18px;
  border-radius: 16px;
  border: 1px solid var(--line);
  background: linear-gradient(180deg, var(--cat-soft) 0%, rgba(255, 253, 248, 0) 42%);
  overflow: hidden;
  transition: transform 0.18s var(--ease), box-shadow 0.18s var(--ease);
}
.cat-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-sm);
}
.cat-watermark {
  position: absolute;
  right: 8px;
  top: 2px;
  font-size: 78px;
  line-height: 1;
  font-weight: 850;
  color: var(--cat);
  opacity: 0.09;
  letter-spacing: -0.06em;
  pointer-events: none;
}
.cat-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  position: relative;
}
.cat-head h3 {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.01em;
}
.cat-head p {
  margin-top: 3px;
  font-size: 11px;
  color: var(--ink-3);
}
.cat-meta {
  display: flex;
  gap: 10px;
  margin: 16px 0 14px;
  font-size: 11px;
  color: var(--ink-3);
  flex-wrap: wrap;
}
.cat-meta strong {
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
}
.cat-progress-top {
  display: flex;
  justify-content: space-between;
  font-size: 10.5px;
  color: var(--ink-3);
  margin-bottom: 6px;
}
.cat-progress-top strong {
  color: var(--cat-ink);
}
.cat-progress .progress {
  height: 6px;
}
.cat-progress .progress > span {
  background: var(--cat);
}
.cat-best {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  font-size: 11px;
  color: var(--ink-2);
}
.cat-actions {
  display: flex;
  gap: 6px;
  margin-top: 14px;
}
.cat-actions .btn {
  flex: 1;
  padding: 0 8px;
}
.cat-actions .btn-ghost {
  flex: 0 0 36px;
  padding: 0;
}
.cat-actions .btn-ghost {
  background: rgba(255, 255, 255, 0.6);
}

.exam-list {
  display: flex;
  flex-direction: column;
}
.exam-row {
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 12px 6px;
  border-bottom: 1px solid var(--line);
  text-align: left;
  transition: background 0.14s;
  border-radius: 8px;
}
.exam-row:last-child {
  border-bottom: none;
}
.exam-row:hover {
  background: rgba(241, 236, 223, 0.45);
}
.exam-cat {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-weight: 850;
  font-size: 15px;
  background: var(--cat-soft);
  color: var(--cat-ink);
  flex: none;
}
.exam-main {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
  flex: 1;
}
.exam-main strong {
  font-size: 12.8px;
  font-weight: 650;
}
.exam-main span {
  font-size: 10.5px;
  color: var(--muted);
}
.exam-score {
  font-size: 20px;
  font-weight: 850;
  letter-spacing: -0.03em;
}
.exam-score span {
  font-size: 11px;
  color: var(--muted);
  font-weight: 600;
}
.exam-arrow {
  color: var(--muted);
}

.chart-cats {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
}
.chart-cat {
  height: 26px;
  padding: 0 10px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-3);
  background: var(--surface-2);
  transition: all 0.15s;
}
.chart-cat.active {
  background: var(--navy);
  color: #fff;
}
.chart-dim {
  opacity: 0.45;
  transition: opacity 0.2s;
}

.progress-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.progress-row-head {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 11px;
  color: var(--ink-3);
  margin-bottom: 6px;
}
.progress-row-head strong {
  margin-left: auto;
  color: var(--ink);
}
.progress-cat {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  display: grid;
  place-items: center;
  font-size: 11px;
  font-weight: 800;
  background: var(--cat-soft);
  color: var(--cat-ink);
}

.quick-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid var(--line);
}
.quick {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  border-radius: 12px;
  background: var(--surface-2);
  text-align: left;
  transition: transform 0.15s, background 0.15s;
}
.quick:hover {
  transform: translateY(-1px);
  background: var(--surface-3);
}
.quick-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  flex: none;
}
.quick-icon.red { background: var(--red-soft); color: #a63b2e; }
.quick-icon.amber { background: var(--amber-soft); color: #9a6514; }
.quick span:last-child {
  display: flex;
  flex-direction: column;
  gap: 1px;
}
.quick strong {
  font-size: 15px;
  font-weight: 800;
  line-height: 1;
}
.quick em {
  font-style: normal;
  font-size: 10px;
  color: var(--ink-3);
}

@media (max-width: 1240px) {
  .stat-row { grid-template-columns: repeat(2, 1fr); }
  .skeleton-grid { grid-template-columns: repeat(2, 1fr); }
  .cat-grid { grid-template-columns: 1fr; }
}
</style>
