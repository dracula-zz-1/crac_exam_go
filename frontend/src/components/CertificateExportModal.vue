<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useToastStore } from "@/stores/toast";
import {
  CERT_FIELD_LABELS,
  CERT_TEMPLATE_H,
  CERT_TEMPLATE_W,
  canvasToDataURL,
  defaultCertificateFields,
  drawCertificate,
  loadTemplate,
  type CertificateFields,
} from "@/lib/certificate";
import AppIcon from "./AppIcon.vue";
import BaseModal from "./BaseModal.vue";
import ConfirmDialog from "./ConfirmDialog.vue";

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: [] }>();

const session = useSessionStore();
const toast = useToastStore();

const fields = ref<CertificateFields>(defaultCertificateFields(null, ""));
const photo = ref<HTMLImageElement | null>(null);
const photoName = ref("");
const previewRef = ref<HTMLCanvasElement | null>(null);
const saving = ref(false);
const looking = ref(false);
const noticeOpen = ref(false);
let template: HTMLImageElement | null = null;
let primeTimer: number | null = null;

const categories = ["A 类", "B 类", "C 类"];

const fieldEntries = computed(
  () =>
    (
      ["number", "name", "idCard", "year", "month", "day", "org", "oldNo", "oldDate"] as Array<
        keyof CertificateFields
      >
    ).map((key) => ({ key, label: CERT_FIELD_LABELS[key] })),
);

function redraw() {
  const canvas = previewRef.value;
  if (!canvas || !template) return;
  drawCertificate(canvas, { template, fields: fields.value, photo: photo.value, scale: 1.5 });
}

function scheduleRedraw() {
  if (primeTimer !== null) window.clearTimeout(primeTimer);
  primeTimer = window.setTimeout(redraw, 90);
}

watch(fields, scheduleRedraw, { deep: true });

watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    if (!template) {
      try {
        template = await loadTemplate();
      } catch (e: any) {
        toast.error("载入样证失败", e?.message);
        return;
      }
    }
    if (!fields.value.name) {
      fields.value = defaultCertificateFields(session.user, session.user?.level ?? "");
    }
    await nextTick();
    redraw();
  },
);

onBeforeUnmount(() => {
  if (primeTimer !== null) window.clearTimeout(primeTimer);
});

function pickPhoto() {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "image/*";
  input.onchange = () => {
    const file = input.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      const img = new Image();
      img.onload = () => {
        photo.value = img;
        photoName.value = file.name;
        redraw();
      };
      img.src = String(reader.result);
    };
    reader.readAsDataURL(file);
  };
  input.click();
}

function clearPhoto() {
  photo.value = null;
  photoName.value = "";
  redraw();
}

async function openCrac() {
  try {
    await api.openCertificateQueryPage();
    toast.info("已在浏览器打开官方查询页", "可自行核对证书信息");
  } catch (e: any) {
    toast.error("打开浏览器失败", e?.message);
  }
}

/** 用姓名 + 证件号到 CRAC 官方库查询，并把结果填进表单。 */
async function lookup() {
  if (looking.value) return;
  if (!fields.value.name.trim() || !fields.value.idCard.trim()) {
    toast.warning("请先填写姓名和证件号码", "官方查询需要这两项");
    return;
  }
  looking.value = true;
  try {
    const info = await api.lookupCertificate(fields.value.name.trim(), fields.value.idCard.trim());
    if (!info?.found) {
      toast.warning("未查询到记录", info?.message || "请核对姓名与证件号码");
      return;
    }
    if (info.name) fields.value.name = info.name;
    if (info.idCard) fields.value.idCard = info.idCard;
    if (info.certificateNo) fields.value.number = info.certificateNo;
    if (info.category) fields.value.category = `${info.category} 类`;
    if (info.issueDate) {
      const [y, m, d] = info.issueDate.split("-");
      if (y) fields.value.year = y;
      if (m) fields.value.month = String(Number(m));
      if (d) fields.value.day = String(Number(d));
    }
    if (info.province) fields.value.org = info.province;
    toast.success("已从官方库填入证书信息", info.message || "请核对后导出");
  } catch (e: any) {
    toast.error("查询失败", e?.message);
  } finally {
    looking.value = false;
  }
}

function exportImage() {
  if (!previewRef.value) return;
  noticeOpen.value = true;
}

async function doExport() {
  noticeOpen.value = false;
  const canvas = previewRef.value;
  if (!canvas) return;
  saving.value = true;
  try {
    // 导出用 2 倍尺寸，重新画一遍保证清晰度
    const out = document.createElement("canvas");
    drawCertificate(out, { template: template!, fields: fields.value, photo: photo.value, scale: 2 });
    const dataURL = canvasToDataURL(out);
    const path = await api.saveCertificateImage(
      `操作证_${fields.value.name || "未命名"}.png`,
      dataURL,
    );
    if (!path) return; // 用户取消
    toast.success("已导出电子操作证", path);
  } catch (e: any) {
    toast.error("导出失败", e?.message);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <BaseModal
    :open="open"
    title="生成电子操作证"
    subtitle="按你本地档案与填写内容套用到样证底图上，导出为 PNG 图片"
    wide
    :close-on-mask="false"
    @close="emit('close')"
  >
    <div class="cert-body">
      <div class="cert-preview">
        <canvas ref="previewRef" class="cert-canvas" />
        <p class="cert-preview-hint">
          预览为按比例缩小显示，导出的是 {{ CERT_TEMPLATE_W * 2 }} × {{ CERT_TEMPLATE_H * 2 }} 的高清 PNG。
        </p>
      </div>

      <div class="cert-form">
        <div class="cert-form-row">
          <label v-for="f in fieldEntries.slice(0, 1)" :key="f.key" class="field cert-wide">
            <span class="field-label">{{ f.label }}</span>
            <input v-model="fields[f.key]" class="input" placeholder="如 B1100123456" />
          </label>
        </div>

        <div class="cert-form-grid">
          <label class="field">
            <span class="field-label">姓名</span>
            <input v-model="fields.name" class="input" />
          </label>
          <label class="field">
            <span class="field-label">证件号码</span>
            <input v-model="fields.idCard" class="input" placeholder="18 位身份证号" />
          </label>
          <label class="field">
            <span class="field-label">核发机构省份</span>
            <input v-model="fields.org" class="input" placeholder="如 江苏" />
          </label>
          <label class="field">
            <span class="field-label">原操作证号</span>
            <input v-model="fields.oldNo" class="input" placeholder="没有可留空" />
          </label>
          <label class="field">
            <span class="field-label">原发证日期</span>
            <input v-model="fields.oldDate" class="input" placeholder="如 20XX 年 XX 月 XX 日" />
          </label>
          <div class="field">
            <span class="field-label">操作类别</span>
            <div class="segmented cert-cats">
              <button
                v-for="c in categories"
                :key="c"
                :class="{ active: fields.category === c }"
                @click="fields.category = c"
              >
                {{ c }}
              </button>
            </div>
          </div>
          <div class="field cert-date">
            <span class="field-label">发证日期</span>
            <div class="cert-date-inputs">
              <input v-model="fields.year" class="input" inputmode="numeric" placeholder="年" />
              <input v-model="fields.month" class="input" inputmode="numeric" placeholder="月" />
              <input v-model="fields.day" class="input" inputmode="numeric" placeholder="日" />
            </div>
          </div>
          <div class="field">
            <span class="field-label">证件照</span>
            <div class="cert-photo-pick">
              <button class="btn btn-sm btn-soft" @click="pickPhoto">
                <AppIcon name="upload" :size="14" />
                {{ photo ? "更换" : "选择照片" }}
              </button>
              <button v-if="photo" class="btn btn-sm btn-ghost" @click="clearPhoto">移除</button>
              <span v-if="photoName" class="dim cert-photo-name">{{ photoName }}</span>
            </div>
          </div>
        </div>

        <div class="cert-tip">
          <AppIcon name="info" :size="14" />
          <span>
            姓名与证件号已按本地档案预填，点「用姓名 + 证件号查询」可直接从 CRAC 官方库取回证书编号、
            发证日期与核发机构；也可以点「打开官方查询页」自行核对。
          </span>
        </div>
      </div>
    </div>

    <template #footer>
      <button class="btn btn-ghost" @click="openCrac">
        <AppIcon name="arrowUpRight" :size="14" />
        打开官方查询页
      </button>
      <button class="btn btn-soft" :disabled="looking" @click="lookup">
        <span v-if="looking" class="spinner dark" />
        <AppIcon v-else name="search" :size="14" />
        {{ looking ? "查询中…" : "用姓名 + 证件号查询" }}
      </button>
      <span class="spacer" />
      <button class="btn btn-ghost" :disabled="saving" @click="emit('close')">关闭</button>
      <button class="btn btn-accent" :disabled="saving" @click="exportImage">
        <span v-if="saving" class="spinner dark" />
        <AppIcon v-else name="download" :size="15" />
        导出 PNG
      </button>
    </template>

    <ConfirmDialog
      :open="noticeOpen"
      title="导出前请确认"
      message="本图依据你的本地档案与填写内容生成，仅为样式样例，不具备法律效力，一切以官方签发的证书为准。请勿用于伪造、冒用等不当用途。"
      confirm-text="仍要导出"
      cancel-text="再检查一下"
      :loading="saving"
      @cancel="noticeOpen = false"
      @confirm="doExport"
    >
      <template #icon><AppIcon name="warning" :size="20" /></template>
    </ConfirmDialog>
  </BaseModal>
</template>

<style scoped>
.cert-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.cert-preview {
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.cert-canvas {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 10px;
  box-shadow: 0 8px 22px rgba(22, 27, 33, 0.16);
}
.cert-preview-hint {
  font-size: 11px;
  color: var(--muted);
  text-align: center;
}
.cert-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.cert-form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}
.cert-wide {
  grid-column: span 3;
}
.cert-cats {
  width: 100%;
}
.cert-cats button {
  flex: 1;
  height: 34px;
}
.cert-date-inputs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
}
.cert-photo-pick {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.cert-photo-name {
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}
.cert-tip {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  padding: 10px 12px;
  border-radius: 11px;
  background: var(--amber-soft);
  color: #8a5a10;
  font-size: 11.5px;
  line-height: 1.6;
}
.cert-tip :deep(svg) {
  margin-top: 1px;
}

@media (max-width: 720px) {
  .cert-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .cert-wide {
    grid-column: span 2;
  }
}
</style>
