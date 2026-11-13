<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { api } from "@/api";
import { useToastStore } from "@/stores/toast";
import type { Category, Question } from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import BaseModal from "@/components/BaseModal.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const router = useRouter();
const toast = useToastStore();

const rows = ref<Question[]>([]);
const total = ref(0);
const totalPages = ref(1);
const page = ref(1);
const pageSize = ref(10);
const search = ref("");
const filters = ref<Record<Category, boolean>>({ A: false, B: false, C: false });
const loading = ref(true);
const error = ref("");

const editOpen = ref(false);
const editing = ref<Question | null>(null);
const saving = ref(false);
const answerChecks = ref<Record<string, boolean>>({ A: false, B: false, C: false, D: false });
const catChecks = ref<Record<Category, boolean>>({ A: false, B: false, C: false });
const fileInput = ref<HTMLInputElement | null>(null);

const deleteTarget = ref<Question | null>(null);
const deleting = ref(false);

let searchTimer: number | null = null;

const activeFilterCount = computed(
  () => Object.values(filters.value).filter(Boolean).length,
);

const pageItems = computed<(number | "…")[]>(() => {
  const n = totalPages.value;
  if (n <= 7) return Array.from({ length: n }, (_, i) => i + 1);
  const items: (number | "…")[] = [1];
  const start = Math.max(2, page.value - 1);
  const end = Math.min(n - 1, page.value + 1);
  if (start > 2) items.push("…");
  for (let i = start; i <= end; i++) items.push(i);
  if (end < n - 1) items.push("…");
  items.push(n);
  return items;
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const result = await api.getPage(
      page.value,
      pageSize.value,
      search.value,
      filters.value.A,
      filters.value.B,
      filters.value.C,
    );
    rows.value = result.data ?? [];
    total.value = result.total ?? 0;
    totalPages.value = result.totalPages ?? 1;
    if (page.value > totalPages.value) {
      page.value = totalPages.value;
    }
  } catch (e: any) {
    error.value = e?.message || "加载题库失败";
  } finally {
    loading.value = false;
  }
}

function toggleFilter(cat: Category) {
  filters.value = { ...filters.value, [cat]: !filters.value[cat] };
  page.value = 1;
  load();
}

function resetFilters() {
  filters.value = { A: false, B: false, C: false };
  search.value = "";
  page.value = 1;
  load();
}

function goPage(n: number | "…") {
  if (typeof n !== "number") return;
  if (n < 1 || n > totalPages.value || n === page.value) return;
  page.value = n;
}

watch(search, () => {
  if (searchTimer) window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(() => {
    page.value = 1;
    load();
  }, 320);
});

watch([page, pageSize], load);
onMounted(load);

function openEdit(q: Question) {
  editing.value = JSON.parse(JSON.stringify(q));
  answerChecks.value = {
    A: q.T.includes("A"),
    B: q.T.includes("B"),
    C: q.T.includes("C"),
    D: q.T.includes("D"),
  };
  catChecks.value = { A: !!q.LA, B: !!q.LB, C: !!q.LC };
  editOpen.value = true;
}

function toggleAnswer(key: string) {
  answerChecks.value = { ...answerChecks.value, [key]: !answerChecks.value[key] };
  syncAnswer();
}

function syncAnswer() {
  if (!editing.value) return;
  const t = ["A", "B", "C", "D"].filter((k) => answerChecks.value[k]).join("");
  editing.value.T = t;
  editing.value.type = t.length > 1 ? 2 : 1;
}

function toggleCat(cat: Category) {
  catChecks.value = { ...catChecks.value, [cat]: !catChecks.value[cat] };
  if (!editing.value) return;
  editing.value.LA = catChecks.value.A ? 1 : 0;
  editing.value.LB = catChecks.value.B ? 1 : 0;
  editing.value.LC = catChecks.value.C ? 1 : 0;
}

function pickImage() {
  fileInput.value?.click();
}

function onImagePicked(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file || !editing.value) return;
  if (file.size > 8 * 1024 * 1024) {
    toast.error("图片过大", "单张图片不能超过 8MB");
    return;
  }
  const reader = new FileReader();
  reader.onload = () => {
    if (editing.value) editing.value.F = String(reader.result ?? "");
  };
  reader.readAsDataURL(file);
  (e.target as HTMLInputElement).value = "";
}

function removeImage() {
  if (editing.value) editing.value.F = "";
}

async function save() {
  if (!editing.value) return;
  const q = editing.value;
  if (!q.J.trim() || !q.Q.trim()) {
    toast.warning("请填写题号与题干");
    return;
  }
  if (!q.T) {
    toast.warning("请至少选择一个正确答案");
    return;
  }
  saving.value = true;
  try {
    await api.updateQuestion(q);
    toast.success("题目已保存");
    editOpen.value = false;
    await load();
  } catch (e: any) {
    toast.error("保存失败", e?.message);
  } finally {
    saving.value = false;
  }
}

function askDelete(q: Question) {
  deleteTarget.value = q;
}

async function confirmDelete() {
  const q = deleteTarget.value;
  if (!q) return;
  deleting.value = true;
  try {
    await api.deleteQuestion(q.id);
    toast.success("题目已删除");
    deleteTarget.value = null;
    if (rows.value.length === 1 && page.value > 1) page.value--;
    else await load();
  } catch (e: any) {
    toast.error("删除失败", e?.message);
  } finally {
    deleting.value = false;
  }
}

function categoriesOf(q: Question) {
  const out: Category[] = [];
  if (q.LA) out.push("A");
  if (q.LB) out.push("B");
  if (q.LC) out.push("C");
  return out;
}

function questionCode(q: Question) {
  return [q.J, q.P, q.I].filter(Boolean).join(" · ");
}
</script>

<template>
  <div class="page bank-page">
    <header class="page-head">
      <div>
        <span class="eyebrow">QUESTION BANK · 题库管理</span>
        <h1 class="page-title">题库管理</h1>
        <p class="page-sub">搜索、筛选、编辑与删除题目。删除题目会同步清理收藏与错题引用。</p>
      </div>
      <div class="head-actions">
        <button class="btn btn-ghost" @click="load">
          <AppIcon name="refresh" :size="16" />
          刷新
        </button>
        <button class="btn btn-accent" @click="router.push('/settings')">
          <AppIcon name="upload" :size="16" />
          导入题库
        </button>
      </div>
    </header>

    <section class="toolbar surface">
      <div class="search-box">
        <AppIcon name="search" :size="16" />
        <input
          v-model="search"
          class="search-input"
          type="text"
          placeholder="搜索题干、选项、题号…"
        />
        <button v-if="search" class="search-clear" @click="search = ''">
          <AppIcon name="x" :size="14" />
        </button>
      </div>

      <div class="filter-pills">
        <button
          v-for="cat in (['A', 'B', 'C'] as Category[])"
          :key="cat"
          class="filter-pill"
          :class="[`cat-${cat}`, { active: filters[cat] }]"
          @click="toggleFilter(cat)"
        >
          {{ cat }} 类
        </button>
        <button v-if="activeFilterCount || search" class="reset-btn" @click="resetFilters">
          重置
        </button>
      </div>

      <div class="toolbar-right">
        <span class="total-text">共 <b>{{ total }}</b> 题</span>
        <select v-model.number="pageSize" class="select page-size">
          <option :value="10">10 条 / 页</option>
          <option :value="20">20 条 / 页</option>
          <option :value="50">50 条 / 页</option>
        </select>
      </div>
    </section>

    <div v-if="error" class="error-box">
      <AppIcon name="warning" :size="15" />
      {{ error }}
      <button class="btn btn-sm btn-ghost" @click="load">重试</button>
    </div>

    <div v-if="loading" class="table-loading">
      <div v-for="i in 5" :key="i" class="skeleton" style="height: 54px" />
    </div>

    <div v-else-if="!rows.length" class="empty surface">
      <div class="empty-icon"><AppIcon name="bank" :size="28" /></div>
      <h2 class="empty-title">{{ total ? "没有匹配的题目" : "题库还是空的" }}</h2>
      <p class="empty-text">
        {{ total ? "换个关键词或重置筛选条件试试。" : "导入 CRAC 官方题库 ZIP 包后，题目会出现在这里。" }}
      </p>
      <button v-if="!total" class="btn btn-accent" @click="router.push('/settings')">去导入题库</button>
      <button v-else class="btn btn-ghost" @click="resetFilters">重置筛选</button>
    </div>

    <template v-else>
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th style="width: 150px">题号</th>
              <th>题干</th>
              <th style="width: 82px">题型</th>
              <th style="width: 130px">分类</th>
              <th style="width: 70px">图片</th>
              <th style="width: 130px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="q in rows" :key="q.id">
              <td class="mono code-cell">{{ questionCode(q) }}</td>
              <td>
                <div class="stem-cell">{{ q.Q }}</div>
                <div class="answer-hint mono">答案 {{ q.T }}</div>
              </td>
              <td>
                <span class="chip" :class="q.type === 2 ? 'chip-teal' : 'chip-amber'">
                  {{ q.type === 2 ? "多选" : "单选" }}
                </span>
              </td>
              <td>
                <div class="cat-tags">
                  <span
                    v-for="cat in categoriesOf(q)"
                    :key="cat"
                    class="cat-tag"
                    :class="`cat-${cat}`"
                  >
                    {{ cat }}
                  </span>
                  <span v-if="!categoriesOf(q).length" class="dim">未分类</span>
                </div>
              </td>
              <td>
                <span v-if="q.F" class="img-flag"><AppIcon name="image" :size="15" /></span>
                <span v-else class="dim">—</span>
              </td>
              <td>
                <div class="row-actions">
                  <button class="icon-btn" title="编辑" @click="openEdit(q)">
                    <AppIcon name="edit" :size="16" />
                  </button>
                  <button class="icon-btn danger" title="删除" @click="askDelete(q)">
                    <AppIcon name="trash" :size="16" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <button class="page-btn" :disabled="page === 1" @click="goPage(page - 1)">
          <AppIcon name="chevronLeft" :size="15" />
        </button>
        <template v-for="(item, i) in pageItems" :key="i">
          <span v-if="item === '…'" class="page-ellipsis">…</span>
          <button
            v-else
            class="page-btn"
            :class="{ active: item === page }"
            @click="goPage(item)"
          >
            {{ item }}
          </button>
        </template>
        <button class="page-btn" :disabled="page === totalPages" @click="goPage(page + 1)">
          <AppIcon name="chevronRight" :size="15" />
        </button>
        <span class="page-info">第 {{ page }} / {{ totalPages }} 页</span>
      </div>
    </template>

    <!-- 编辑题目 -->
    <BaseModal
      :open="editOpen"
      title="编辑题目"
      subtitle="答案按所选选项自动生成，题型随答案数量变化"
      wide
      :close-on-mask="false"
      @close="editOpen = false"
    >
      <div v-if="editing" class="edit-form">
        <div class="grid-3">
          <label class="field">
            <span class="field-label">题号 J</span>
            <input v-model="editing.J" class="input" />
          </label>
          <label class="field">
            <span class="field-label">编号 P</span>
            <input v-model="editing.P" class="input" />
          </label>
          <label class="field">
            <span class="field-label">编号 I</span>
            <input v-model="editing.I" class="input" />
          </label>
        </div>

        <label class="field">
          <span class="field-label">题干</span>
          <textarea v-model="editing.Q" class="textarea" rows="3" />
        </label>

        <div class="grid-2-edit">
          <label class="field">
            <span class="field-label">选项 A</span>
            <textarea v-model="editing.A" class="textarea" rows="2" />
          </label>
          <label class="field">
            <span class="field-label">选项 B</span>
            <textarea v-model="editing.B" class="textarea" rows="2" />
          </label>
          <label class="field">
            <span class="field-label">选项 C</span>
            <textarea v-model="editing.C" class="textarea" rows="2" />
          </label>
          <label class="field">
            <span class="field-label">选项 D</span>
            <textarea v-model="editing.D" class="textarea" rows="2" />
          </label>
        </div>

        <div class="edit-row">
          <div class="field">
            <span class="field-label">正确答案</span>
            <div class="answer-picks">
              <button
                v-for="key in ['A', 'B', 'C', 'D']"
                :key="key"
                class="answer-pick"
                :class="{ active: answerChecks[key] }"
                @click="toggleAnswer(key)"
              >
                {{ key }}
              </button>
            </div>
          </div>
          <div class="field">
            <span class="field-label">所属类别</span>
            <div class="answer-picks">
              <button
                v-for="cat in (['A', 'B', 'C'] as Category[])"
                :key="cat"
                class="answer-pick cat-pick"
                :class="[`cat-${cat}`, { active: catChecks[cat] }]"
                @click="toggleCat(cat)"
              >
                {{ cat }} 类
              </button>
            </div>
          </div>
          <div class="field">
            <span class="field-label">题型</span>
            <span class="chip" :class="editing.type === 2 ? 'chip-teal' : 'chip-amber'">
              {{ editing.type === 2 ? "多选题" : "单选题" }}
            </span>
          </div>
        </div>

        <div class="field">
          <span class="field-label">题目图片</span>
          <div class="image-editor">
            <div v-if="editing.F" class="image-preview">
              <img :src="editing.F" alt="题目图片" />
            </div>
            <div class="image-actions">
              <button class="btn btn-sm btn-ghost" @click="pickImage">
                <AppIcon name="upload" :size="14" />
                {{ editing.F ? "更换图片" : "上传图片" }}
              </button>
              <button v-if="editing.F" class="btn btn-sm btn-danger" @click="removeImage">
                <AppIcon name="trash" :size="14" />
                删除图片
              </button>
              <span class="field-hint">支持 PNG / JPG / SVG，单张不超过 8MB</span>
            </div>
          </div>
          <input ref="fileInput" type="file" accept="image/*" class="hide" @change="onImagePicked" />
        </div>
      </div>

      <template #footer>
        <button class="btn btn-ghost" @click="editOpen = false">取消</button>
        <button class="btn btn-accent" :disabled="saving" @click="save">
          <span v-if="saving" class="spinner dark" />
          保存修改
        </button>
      </template>
    </BaseModal>

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除题目"
      :message="deleteTarget ? `确定删除「${deleteTarget.J}」吗？\n该题会同时从收藏夹与错题本中移除，考试历史快照不受影响。` : ''"
      confirm-text="删除"
      danger
      :loading="deleting"
      @cancel="deleteTarget = null"
      @confirm="confirmDelete"
    >
      <template #icon><AppIcon name="trash" :size="20" /></template>
    </ConfirmDialog>
  </div>
</template>

<style scoped>
.bank-page {
  max-width: 1240px;
}
.head-actions {
  display: flex;
  gap: 10px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 12px;
  border-radius: 10px;
  background: var(--surface-2);
  border: 1px solid transparent;
  transition: all 0.16s;
  min-width: 260px;
  flex: 1;
  color: var(--ink-3);
}
.search-box:focus-within {
  background: var(--surface);
  border-color: var(--amber);
  box-shadow: 0 0 0 3px rgba(232, 163, 61, 0.12);
}
.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 12.5px;
  color: var(--ink);
}
.search-clear {
  color: var(--muted);
  padding: 2px;
  border-radius: 5px;
}
.search-clear:hover {
  background: var(--surface-3);
  color: var(--ink);
}
.filter-pills {
  display: flex;
  align-items: center;
  gap: 6px;
}
.filter-pill {
  height: 34px;
  padding: 0 12px;
  border-radius: 9px;
  font-size: 11.5px;
  font-weight: 650;
  color: var(--ink-3);
  background: var(--surface);
  border: 1px solid var(--line);
  transition: all 0.15s;
}
.filter-pill:hover {
  border-color: var(--line-strong);
}
.filter-pill.active {
  background: var(--cat-soft);
  border-color: var(--cat);
  color: var(--cat-ink);
}
.reset-btn {
  font-size: 11px;
  color: var(--amber-deep);
  padding: 0 4px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}
.total-text {
  font-size: 11.5px;
  color: var(--ink-3);
}
.total-text b {
  color: var(--ink);
}
.page-size {
  width: 116px;
  height: 34px;
  padding: 0 30px 0 10px;
  font-size: 11.5px;
  border-radius: 9px;
}

.table-loading {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.code-cell {
  font-size: 11px;
  color: var(--ink-3);
  white-space: nowrap;
}
.stem-cell {
  font-size: 12.8px;
  line-height: 1.6;
  max-width: 560px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.answer-hint {
  margin-top: 3px;
  font-size: 10px;
  color: var(--muted);
}
.cat-tags {
  display: flex;
  gap: 4px;
  align-items: center;
}
.cat-tag {
  width: 23px;
  height: 23px;
  border-radius: 7px;
  display: grid;
  place-items: center;
  font-size: 10.5px;
  font-weight: 800;
  background: var(--cat-soft);
  color: var(--cat-ink);
}
.img-flag {
  color: var(--teal);
}
.row-actions {
  display: flex;
  gap: 2px;
}
.icon-btn.danger:hover {
  background: var(--red-soft);
  color: #a63b2e;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  margin-top: 18px;
}
.page-btn {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  border-radius: 8px;
  font-size: 11.5px;
  font-weight: 600;
  color: var(--ink-3);
  border: 1px solid transparent;
  transition: all 0.14s;
}
.page-btn:hover:not(:disabled) {
  background: var(--surface-2);
  color: var(--ink);
}
.page-btn.active {
  background: var(--navy);
  color: #fff;
}
.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.page-ellipsis {
  color: var(--muted);
  font-size: 11px;
  padding: 0 2px;
}
.page-info {
  margin-left: 10px;
  font-size: 11px;
  color: var(--muted);
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.grid-2-edit {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.edit-row {
  display: flex;
  gap: 28px;
  align-items: flex-start;
  flex-wrap: wrap;
}
.answer-picks {
  display: flex;
  gap: 7px;
}
.answer-pick {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  border: 1px solid var(--line-strong);
  background: var(--surface);
  font-size: 12.5px;
  font-weight: 750;
  color: var(--ink-3);
  transition: all 0.15s;
}
.answer-pick:hover {
  border-color: var(--amber);
}
.answer-pick.active {
  background: var(--amber);
  border-color: var(--amber);
  color: #241703;
}
.cat-pick {
  width: auto;
  padding: 0 13px;
}
.cat-pick.active {
  background: var(--cat-soft);
  border-color: var(--cat);
  color: var(--cat-ink);
}
.image-editor {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.image-preview {
  width: 150px;
  height: 110px;
  border-radius: 10px;
  border: 1px solid var(--line);
  overflow: hidden;
  background: var(--navy);
}
.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.image-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
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
