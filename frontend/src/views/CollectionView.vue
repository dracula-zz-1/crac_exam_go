<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useBankStore } from "@/stores/bank";
import { useToastStore } from "@/stores/toast";
import type { Category, Question } from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import OptionList from "@/components/OptionList.vue";

const props = defineProps<{ mode: "error" | "favorite" }>();
const router = useRouter();
const route = useRoute();
const session = useSessionStore();
const bank = useBankStore();
const toast = useToastStore();

function queryCategory(): "" | Category {
  const raw = route.query.category;
  if (typeof raw === "string" && ["A", "B", "C"].includes(raw.toUpperCase())) {
    return raw.toUpperCase() as Category;
  }
  return "";
}

const category = ref<"" | Category>(queryCategory());
const questions = ref<Question[]>([]);
const loading = ref(true);
const error = ref("");
const revealed = ref<Set<number>>(new Set());
const visibleCount = ref(15);

const isError = computed(() => props.mode === "error");
const title = computed(() => (isError.value ? "错题本" : "收藏夹"));
const subtitle = computed(() =>
  isError.value
    ? "练习中答错与未作答的题会自动收集到这里，逐题消化它们。"
    : "把重点题目标记下来，考前集中复习。",
);
const visible = computed(() => questions.value.slice(0, visibleCount.value));
const hasMore = computed(() => questions.value.length > visibleCount.value);
const selectedCategory = computed(() => category.value || "A");

async function load() {
  if (!session.user) return;
  loading.value = true;
  error.value = "";
  revealed.value = new Set();
  visibleCount.value = 15;
  try {
    questions.value = isError.value
      ? await api.getErrorQuestions(session.user.id, category.value)
      : await api.getFavoriteQuestions(session.user.id, category.value);
  } catch (e: any) {
    error.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

function setCategory(cat: "" | Category) {
  if (category.value === cat) return;
  category.value = cat;
  load();
}

function toggle(id: number) {
  const next = new Set(revealed.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  revealed.value = next;
}

function categoryOf(q: Question): Category {
  if (q.LA) return "A";
  if (q.LB) return "B";
  return "C";
}

function optionsOf(q: Question) {
  return [
    { key: "A", text: q.A },
    { key: "B", text: q.B },
    { key: "C", text: q.C },
    { key: "D", text: q.D },
  ];
}

function practice(q: Question) {
  router.push(
    `/practice/${props.mode}/${categoryOf(q)}?start=${q.id}`,
  );
}

function practiceGroup() {
  router.push(`/practice/${props.mode}/${selectedCategory.value}`);
}

async function remove(q: Question) {
  if (!session.user) return;
  try {
    if (isError.value) await api.removeError(session.user.id, q.id);
    else {
      await api.removeFavorite(session.user.id, q.id);
      bank.setFavorite(q.id, false);
    }
    questions.value = questions.value.filter((x) => x.id !== q.id);
    const next = new Set(revealed.value);
    next.delete(q.id);
    revealed.value = next;
    toast.success(isError.value ? "已移出错题本" : "已取消收藏");
  } catch (e: any) {
    toast.error("操作失败", e?.message);
  }
}

watch(
  () => route.query.category,
  () => {
    const next = queryCategory();
    if (next && next !== category.value) {
      category.value = next;
      load();
    }
  },
);

onMounted(load);
</script>

<template>
  <div class="page collection-page">
    <header class="page-head">
      <div>
        <span class="eyebrow">{{ isError ? "REVIEW · 错题强化" : "FAVORITES · 重点收藏" }}</span>
        <h1 class="page-title">{{ title }}</h1>
        <p class="page-sub">{{ subtitle }}</p>
      </div>
      <div class="head-actions">
        <button class="btn btn-ghost" @click="router.push('/home')">
          <AppIcon name="home" :size="16" />
          返回总览
        </button>
        <button class="btn btn-accent" :disabled="!questions.length" @click="practiceGroup">
          <AppIcon name="practice" :size="16" />
          开始{{ isError ? "错题" : "收藏" }}练习
        </button>
      </div>
    </header>

    <div class="collection-bar surface">
      <div class="segmented">
        <button :class="{ active: category === '' }" @click="setCategory('')">全部</button>
        <button :class="{ active: category === 'A' }" @click="setCategory('A')">A 类</button>
        <button :class="{ active: category === 'B' }" @click="setCategory('B')">B 类</button>
        <button :class="{ active: category === 'C' }" @click="setCategory('C')">C 类</button>
      </div>
      <span class="count chip chip-outline">
        共 {{ questions.length }} 题
      </span>
    </div>

    <div v-if="error" class="error-box">
      <AppIcon name="warning" :size="15" />
      {{ error }}
      <button class="btn btn-sm btn-ghost" @click="load">重试</button>
    </div>

    <div v-if="loading" class="list-loading">
      <div v-for="i in 4" :key="i" class="skeleton" style="height: 122px" />
    </div>

    <div v-else-if="!questions.length" class="empty surface">
      <div class="empty-icon">
        <AppIcon :name="isError ? 'errors' : 'star'" :size="28" />
      </div>
      <h2 class="empty-title">{{ isError ? "错题本还是空的" : "还没有收藏题目" }}</h2>
      <p class="empty-text">
        {{
          isError
            ? "练习或考试中答错的题目会自动出现在这里。先从一题开始吧。"
            : "练习时点击右上角收藏，就能把重点题目放到这里。"
        }}
      </p>
      <button class="btn btn-accent" @click="router.push(`/practice/practice/${selectedCategory}`)">
        <AppIcon name="practice" :size="16" />
        去练习
      </button>
    </div>

    <div v-else class="collection-list">
      <article v-for="q in visible" :key="q.id" class="q-row surface">
        <div class="q-row-top">
          <span class="seq mono">{{ q.J }}</span>
          <span class="chip" :class="q.type === 2 ? 'chip-teal' : 'chip-amber'">
            {{ q.type === 2 ? "多选" : "单选" }}
          </span>
          <span
            v-for="cat in (['A', 'B', 'C'] as Category[])"
            :key="cat"
            v-show="
              (cat === 'A' && q.LA) ||
              (cat === 'B' && q.LB) ||
              (cat === 'C' && q.LC)
            "
            class="chip chip-outline cat-mini"
            :class="`cat-${cat}`"
          >
            {{ cat }}
          </span>
          <span class="q-answer mono">答案 {{ q.T }}</span>
          <span v-if="isError && q.wrongCount" class="chip chip-red">
            错 {{ q.wrongCount }} 次
          </span>
          <div class="q-row-actions">
            <button class="btn btn-sm btn-soft" @click="toggle(q.id)">
              <AppIcon :name="revealed.has(q.id) ? 'eyeOff' : 'eye'" :size="14" />
              {{ revealed.has(q.id) ? "收起" : "答案" }}
            </button>
            <button class="btn btn-sm btn-ghost" @click="practice(q)">
              <AppIcon name="practice" :size="14" />
              练习
            </button>
            <button class="btn btn-sm btn-ghost" @click="remove(q)">
              <AppIcon :name="isError ? 'check' : 'trash'" :size="14" />
              {{ isError ? "移出" : "取消" }}
            </button>
          </div>
        </div>

        <h3 class="q-stem" @click="toggle(q.id)">{{ q.Q }}</h3>

        <div v-if="q.F" class="q-thumb" @click="toggle(q.id)">
          <img :src="q.F" alt="题目图片" />
        </div>

        <div v-if="revealed.has(q.id)" class="q-reveal">
          <OptionList
            :options="optionsOf(q)"
            :selected="[]"
            revealed
            :correct="q.T"
            disabled
          />
        </div>
      </article>

      <button v-if="hasMore" class="load-more btn btn-ghost" @click="visibleCount += 15">
        加载更多（剩余 {{ questions.length - visibleCount }} 题）
      </button>
    </div>
  </div>
</template>

<style scoped>
.collection-page {
  max-width: 1120px;
}
.head-actions {
  display: flex;
  gap: 10px;
}
.collection-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  margin-bottom: 18px;
}
.count {
  font-size: 11px;
}
.list-loading,
.collection-list {
  display: flex;
  flex-direction: column;
  gap: 13px;
}
.q-row {
  padding: 17px 20px;
  transition: box-shadow 0.16s, transform 0.16s;
}
.q-row:hover {
  box-shadow: var(--shadow-sm);
}
.q-row-top {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.seq {
  font-size: 10.5px;
  color: var(--muted);
  letter-spacing: 0.06em;
  margin-right: 2px;
}
.cat-mini {
  padding: 0 7px;
  height: 21px;
  font-size: 9.5px;
  color: var(--cat-ink);
  border-color: var(--cat);
  background: var(--cat-soft);
}
.q-answer {
  font-size: 10.5px;
  color: var(--ink-3);
  margin-left: 2px;
}
.q-row-actions {
  margin-left: auto;
  display: flex;
  gap: 6px;
}
.q-stem {
  font-size: 15px;
  line-height: 1.65;
  font-weight: 650;
  cursor: pointer;
  transition: color 0.14s;
}
.q-stem:hover {
  color: var(--amber-deep);
}
.q-thumb {
  margin-top: 12px;
  max-width: 260px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--line);
  cursor: zoom-in;
  background: transparent;
}
.q-thumb img {
  display: block;
  width: 100%;
  max-height: 140px;
  object-fit: contain;
}
.q-reveal {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed var(--line);
}
.load-more {
  align-self: center;
  margin-top: 6px;
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
  margin-bottom: 16px;
}
</style>
