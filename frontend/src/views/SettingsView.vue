<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { api, usingMock } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useBankStore } from "@/stores/bank";
import { useToastStore } from "@/stores/toast";
import type {
  AppInfo,
  AppUpdateProgress,
  AppUpdateStatus,
  BankUpdateSettings,
  BankUpdateStatus,
  DBStats,
  ImportProgress,
  ImportResult,
} from "@/types";
import AppIcon from "@/components/AppIcon.vue";
import BaseModal from "@/components/BaseModal.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const router = useRouter();
const session = useSessionStore();
const bank = useBankStore();
const toast = useToastStore();

const info = ref<AppInfo | null>(null);
const stats = ref<DBStats | null>(null);

// 数据库目录
const dataDirOpen = ref(false);
const dataDirBusy = ref(false);
const pendingDataDir = ref("");

const dragging = ref(false);
const pendingPath = ref("");
const confirmImportOpen = ref(false);
const importing = ref(false);
const progress = ref<ImportProgress | null>(null);
const lastResult = ref<ImportResult | null>(null);

const clearUserOpen = ref(false);
const clearBankOpen = ref(false);
const clearingUser = ref(false);
const clearingBank = ref(false);

const updateStatus = ref<BankUpdateStatus | null>(null);
const updateSettings = ref<BankUpdateSettings>({ autoCheck: true, autoImport: false });
const checking = ref(false);
const applyingUpdate = ref(false);
const updateError = ref("");
const confirmUpdateOpen = ref(false);

const appUpdate = ref<AppUpdateStatus | null>(null);
const appChecking = ref(false);
const appDownloading = ref(false);
const appUpdateError = ref("");
const appProgress = ref<AppUpdateProgress | null>(null);
const appInstallOpen = ref(false);
const appInstalling = ref(false);

// 更新通道：GitHub 官方仓库 / Gitee 镜像仓库
const appUpdateSource = computed<"github" | "gitee">(() =>
  appUpdate.value?.source === "gitee" ? "gitee" : "github",
);
const appUpdateHost = computed(() =>
  appUpdateSource.value === "gitee" ? "gitee.com" : "github.com",
);

let offProgress: (() => void) | null = null;
let offDrop: (() => void) | null = null;
let offUpdateStatus: (() => void) | null = null;
let offAppUpdateStatus: (() => void) | null = null;
let offAppUpdateProgress: (() => void) | null = null;

const stageLabel = computed(() => progress.value?.message ?? "准备导入…");
const progressPercent = computed(() => progress.value?.percent ?? 0);

const resultStats = computed(() => {
  if (!lastResult.value) return [];
  const s = lastResult.value.stats ?? {};
  return [
    { label: "单选题", value: s.single ?? 0, tone: "amber" },
    { label: "多选题", value: s.multiple ?? 0, tone: "teal" },
    { label: "A 类", value: s.categoryA ?? 0, tone: "amber" },
    { label: "B 类", value: s.categoryB ?? 0, tone: "teal" },
    { label: "C 类", value: s.categoryC ?? 0, tone: "blue" },
    { label: "带图题", value: s.withImage ?? 0, tone: "blue" },
  ];
});

async function loadMeta() {
  try {
    const [i, s] = await Promise.all([api.getAppInfo(), api.getDBStats()]);
    info.value = i;
    stats.value = s;
  } catch (e: any) {
    toast.error("读取应用信息失败", e?.message);
  }
}

async function pickDataDir() {
  try {
    const dir = await api.pickDataDir();
    if (!dir) return;
    pendingDataDir.value = dir;
    dataDirOpen.value = true;
  } catch (e: any) {
    toast.error("选择数据目录失败", e?.message);
  }
}

async function applyDataDir(migrate: boolean) {
  if (!pendingDataDir.value) return;
  dataDirBusy.value = true;
  try {
    const message = await api.setDataDir(pendingDataDir.value, migrate);
    dataDirOpen.value = false;
    toast.success("数据目录已更新", message);
    await loadMeta();
  } catch (e: any) {
    toast.error("切换数据目录失败", e?.message);
  } finally {
    dataDirBusy.value = false;
  }
}

async function loadStats() {
  try {
    stats.value = await api.getDBStats();
  } catch {
    /* ignore */
  }
}

async function loadUpdate() {
  try {
    const [st, settings] = await Promise.all([
      api.getBankUpdateStatus(),
      api.getBankUpdateSettings(),
    ]);
    updateStatus.value = st;
    updateSettings.value = settings;
  } catch {
    /* ignore */
  }
}

async function saveUpdateSettings(next: BankUpdateSettings) {
  updateSettings.value = next;
  try {
    await api.setBankAutoUpdate(next.autoCheck, next.autoImport);
  } catch (e: any) {
    toast.error("保存更新设置失败", e?.message);
    await loadUpdate();
  }
}

function toggleAutoCheck() {
  const next = {
    ...updateSettings.value,
    autoCheck: !updateSettings.value.autoCheck,
  };
  if (!next.autoCheck) next.autoImport = false;
  saveUpdateSettings(next);
}

function toggleAutoImport() {
  if (!updateSettings.value.autoCheck) return;
  saveUpdateSettings({
    ...updateSettings.value,
    autoImport: !updateSettings.value.autoImport,
  });
}

async function checkUpdate() {
  if (checking.value || applyingUpdate.value) return;
  checking.value = true;
  updateError.value = "";
  try {
    updateStatus.value = await api.checkBankUpdate();
    if (updateStatus.value.available) {
      toast.warning("发现新题库", updateStatus.value.remoteVersion || "可立即更新");
    } else {
      toast.success("已是最新题库");
    }
  } catch (e: any) {
    updateError.value = e?.message || "检查更新失败";
    toast.error("检查更新失败", updateError.value);
  } finally {
    checking.value = false;
  }
}

function askApplyUpdate() {
  if (applyingUpdate.value) return;
  confirmUpdateOpen.value = true;
}

async function applyUpdate() {
  confirmUpdateOpen.value = false;
  applyingUpdate.value = true;
  updateError.value = "";
  try {
    const result = await api.applyBankUpdate();
    lastResult.value = result;
    bank.invalidate();
    await loadStats();
    updateStatus.value = await api.getBankUpdateStatus();
    toast.success("题库已更新", result.message);
  } catch (e: any) {
    updateError.value = e?.message || "更新失败";
    toast.error("题库更新失败", updateError.value);
  } finally {
    applyingUpdate.value = false;
  }
}

async function loadAppUpdate() {
  try {
    appUpdate.value = await api.getAppUpdateStatus();
  } catch {
    /* ignore */
  }
}

async function setUpdateSource(source: "github" | "gitee") {
  if (appUpdateSource.value === source || appChecking.value || appDownloading.value) return;
  try {
    await api.setAppUpdateSource(source);
    await loadAppUpdate();
    toast.success(
      source === "gitee" ? "已切换到 Gitee 镜像通道" : "已切换到 GitHub 通道",
      appUpdate.value?.repo || "",
    );
  } catch (e: any) {
    toast.error("切换更新通道失败", e?.message);
  }
}

async function checkAppUpdate() {
  if (appChecking.value || appDownloading.value) return;
  appChecking.value = true;
  appUpdateError.value = "";
  try {
    appUpdate.value = await api.checkAppUpdate();
    if (appUpdate.value.available) {
      toast.warning(
        "发现应用新版本",
        appUpdate.value.latestVersion || appUpdate.value.releaseName || "可立即下载",
      );
    } else {
      toast.success("应用已是最新版本", appUpdate.value.message);
    }
  } catch (e: any) {
    appUpdateError.value = e?.message || "检查应用更新失败";
    toast.error("检查应用更新失败", appUpdateError.value);
    await loadAppUpdate();
  } finally {
    appChecking.value = false;
  }
}

async function downloadAppUpdate() {
  if (appDownloading.value || !appUpdate.value?.asset) return;
  appDownloading.value = true;
  appUpdateError.value = "";
  appProgress.value = { stage: "download", message: "正在下载更新包…", percent: 0 };
  try {
    appUpdate.value = await api.downloadAppUpdate();
    toast.success("更新包下载完成", appUpdate.value.downloadedName || "");
  } catch (e: any) {
    appUpdateError.value = e?.message || "下载更新包失败";
    toast.error("下载更新包失败", appUpdateError.value);
    await loadAppUpdate();
  } finally {
    appDownloading.value = false;
  }
}

function askRunInstaller() {
  if (appInstalling.value) return;
  appInstallOpen.value = true;
}

async function runInstaller() {
  appInstallOpen.value = false;
  appInstalling.value = true;
  try {
    await api.runAppUpdateInstaller();
    toast.success("安装程序已启动", "请按安装向导完成更新");
  } catch (e: any) {
    toast.error("启动安装程序失败", e?.message);
  } finally {
    appInstalling.value = false;
  }
}

async function openAppRelease() {
  try {
    await api.openAppUpdateRelease();
  } catch (e: any) {
    toast.error("打开发布页失败", e?.message);
  }
}

async function openAppUpdateFolder() {
  try {
    await api.openAppUpdateFolder();
  } catch (e: any) {
    toast.error("打开下载目录失败", e?.message);
  }
}

function fmtDateTime(iso?: string) {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function shortHash(hash?: string) {
  if (!hash) return "";
  return hash.slice(0, 12) + "…";
}

function fmtSize(bytes?: number) {
  if (!bytes) return "—";
  return (bytes / 1024 / 1024).toFixed(1) + " MB";
}

async function chooseZip() {
  if (importing.value) return;
  try {
    const path = await api.pickZipFile();
    if (path) {
      pendingPath.value = path;
      confirmImportOpen.value = true;
    }
  } catch (e: any) {
    toast.error("选择文件失败", e?.message);
  }
}

function handleDropped(paths: string[]) {
  if (importing.value) return;
  const zip = paths.find((p) => p.toLowerCase().endsWith(".zip"));
  if (!zip) {
    toast.warning("只支持 ZIP", "请拖入包含题库 PDF 的 ZIP 压缩包");
    return;
  }
  pendingPath.value = zip;
  confirmImportOpen.value = true;
}

function onHtmlDrop(e: DragEvent) {
  dragging.value = false;
  if (!usingMock) return; // 桌面端由 Wails 原生拖放事件处理
  const file = e.dataTransfer?.files?.[0];
  if (!file) return;
  if (!file.name.toLowerCase().endsWith(".zip")) {
    toast.warning("只支持 ZIP", "请拖入 ZIP 题库包");
    return;
  }
  pendingPath.value = file.name;
  confirmImportOpen.value = true;
}

async function runImport() {
  if (!pendingPath.value) return;
  confirmImportOpen.value = false;
  importing.value = true;
  lastResult.value = null;
  progress.value = { stage: "start", message: "开始导入…", percent: 2 };
  try {
    const result = await api.importZip(pendingPath.value);
    lastResult.value = result;
    if (result.success) {
      bank.invalidate();
      toast.success("题库导入完成", result.message);
      await loadStats();
    } else {
      toast.error("导入失败", result.message);
    }
  } catch (e: any) {
    toast.error("导入失败", e?.message || "请检查 ZIP 内容");
  } finally {
    importing.value = false;
    pendingPath.value = "";
  }
}

async function clearUserData() {
  if (!session.user) return;
  clearingUser.value = true;
  try {
    await api.clearUserData(session.user.id);
    bank.invalidate();
    clearUserOpen.value = false;
    toast.success("学习数据已清空", "题库与考试历史快照保留");
    await loadStats();
  } catch (e: any) {
    toast.error("清空失败", e?.message);
  } finally {
    clearingUser.value = false;
  }
}

async function clearBank() {
  clearingBank.value = true;
  try {
    await api.clearQuestionBank();
    bank.invalidate();
    clearBankOpen.value = false;
    toast.success("题库已清空", "已交卷的考试复盘仍可查看");
    await loadStats();
  } catch (e: any) {
    toast.error("清空失败", e?.message);
  } finally {
    clearingBank.value = false;
  }
}

function fileName(path: string) {
  return path.split(/[/\\]/).pop() || path;
}

onMounted(() => {
  loadMeta();
  loadUpdate();
  loadAppUpdate();
  offProgress = api.onImportProgress((p) => {
    progress.value = p;
    if (p.stage === "done") {
      bank.invalidate();
      loadStats();
    }
  });
  offDrop = api.onFilesDropped(handleDropped);
  offUpdateStatus = api.onBankUpdateStatus((st) => {
    const before = updateStatus.value?.message;
    updateStatus.value = st;
    updateSettings.value = {
      autoCheck: st.autoCheck,
      autoImport: st.autoImport,
    };
    if (st.message.includes("已更新") && st.message !== before) {
      bank.invalidate();
      loadStats();
    }
  });
  offAppUpdateStatus = api.onAppUpdateStatus((st) => {
    appUpdate.value = st;
  });
  offAppUpdateProgress = api.onAppUpdateProgress((p) => {
    appProgress.value = p;
  });
});

onBeforeUnmount(() => {
  offProgress?.();
  offDrop?.();
  offUpdateStatus?.();
  offAppUpdateStatus?.();
  offAppUpdateProgress?.();
});
</script>

<template>
  <div class="page settings-page" data-file-drop-target @drop.prevent="onHtmlDrop">
    <header class="page-head">
      <div>
        <span class="eyebrow">SETTINGS · 设置</span>
        <h1 class="page-title">设置</h1>
        <p class="page-sub">导入题库、管理本地数据，查看应用信息。</p>
      </div>
      <button class="btn btn-ghost" @click="router.push('/home')">
        <AppIcon name="home" :size="16" />
        返回总览
      </button>
    </header>

    <div class="settings-grid">
      <div class="stack gap-20">
        <!-- 题库更新 -->
        <section class="surface card-pad update-card">
          <div class="section-head">
            <div>
              <h2 class="section-title">题库更新</h2>
              <p class="section-sub">自动从 CRAC 官方资料页发现最新题库</p>
            </div>
            <span class="chip" :class="updateSettings.autoCheck ? 'chip-teal' : 'chip-outline'">
              {{ updateSettings.autoCheck ? "自动检查已开启" : "自动检查已关闭" }}
            </span>
          </div>

          <div class="update-current">
            <div class="update-icon" :class="{ available: updateStatus?.available }">
              <AppIcon name="refresh" :size="20" />
            </div>
            <div class="update-main">
              <strong>{{ updateStatus?.currentVersion || "尚未导入题库" }}</strong>
              <p>
                来源：{{ updateStatus?.currentSource || "本地文件" }}
                <span v-if="updateStatus?.currentImportedAt">
                  · 导入于 {{ fmtDateTime(updateStatus.currentImportedAt) }}
                </span>
              </p>
              <p v-if="updateStatus?.currentHash" class="mono update-hash">
                SHA-256 {{ shortHash(updateStatus.currentHash) }}
              </p>
            </div>
            <div class="update-actions">
              <button
                class="btn btn-sm btn-ghost"
                :disabled="checking || applyingUpdate"
                @click="checkUpdate"
              >
                <span v-if="checking" class="spinner dark" />
                <AppIcon v-else name="search" :size="14" />
                {{ checking ? "检查中…" : "检查更新" }}
              </button>
              <button
                v-if="updateStatus?.available"
                class="btn btn-sm btn-accent"
                :disabled="applyingUpdate"
                @click="askApplyUpdate"
              >
                <AppIcon name="download" :size="15" />
                立即更新
              </button>
            </div>
          </div>

          <div v-if="updateStatus?.available" class="update-available">
            <AppIcon name="sparkles" :size="16" />
            <div>
              <strong>发现新版本：{{ updateStatus.remoteVersion }}</strong>
              <p>
                {{ fmtSize(updateStatus.remoteSize) }}
                <span v-if="updateStatus.remoteModified"> · {{ updateStatus.remoteModified }}</span>
              </p>
            </div>
          </div>
          <div v-else-if="updateStatus?.checked" class="update-ok">
            <AppIcon name="check" :size="15" />
            已是最新题库
          </div>

          <div v-if="updateError || updateStatus?.error" class="update-error">
            <AppIcon name="warning" :size="14" />
            {{ updateError || updateStatus?.error }}
          </div>

          <div v-if="applyingUpdate" class="import-progress">
            <div class="import-progress-top">
              <span>{{ progress?.message || "正在下载并更新题库…" }}</span>
              <strong>{{ progress?.percent || 0 }}%</strong>
            </div>
            <div class="progress amber">
              <span :style="{ width: (progress?.percent || 0) + '%' }" />
            </div>
          </div>

          <div class="update-toggles">
            <button class="toggle-row" @click="toggleAutoCheck">
              <span class="toggle" :class="{ on: updateSettings.autoCheck }"><i /></span>
              <span>
                <strong>启动时自动检查更新</strong>
                <em>后台访问 CRAC 官网，不阻塞使用</em>
              </span>
            </button>
            <button
              class="toggle-row"
              :class="{ disabled: !updateSettings.autoCheck }"
              @click="toggleAutoImport"
            >
              <span class="toggle" :class="{ on: updateSettings.autoImport }"><i /></span>
              <span>
                <strong>发现新题库后自动导入</strong>
                <em>会替换题库；错题/收藏/进度按 J/P/I 保留</em>
              </span>
            </button>
          </div>

          <p class="update-note">
            官方题库可能发布在新页面：程序会同时检查题库页、资料下载栏目与站内搜索，并取最新日期且带 ZIP 附件的版本。
          </p>
        </section>

        <!-- 应用更新 -->
        <section class="surface card-pad app-update-card">
          <div class="section-head">
            <div>
              <h2 class="section-title">应用更新</h2>
              <p class="section-sub">选择更新通道，从 GitHub 或 Gitee Release 检查并下载安装包</p>
            </div>
            <div class="segmented update-source-switch">
              <button
                :class="{ active: appUpdateSource === 'github' }"
                :disabled="appChecking || appDownloading"
                @click="setUpdateSource('github')"
              >
                GitHub
              </button>
              <button
                :class="{ active: appUpdateSource === 'gitee' }"
                :disabled="appChecking || appDownloading"
                @click="setUpdateSource('gitee')"
              >
                Gitee 镜像
              </button>
            </div>
          </div>

          <div class="update-current">
            <div class="update-icon" :class="{ available: appUpdate?.available }">
              <AppIcon name="refresh" :size="20" />
            </div>
            <div class="update-main">
              <strong>当前版本 v{{ appUpdate?.currentVersion || info?.version || "3.0.0" }}</strong>
              <p v-if="appUpdate?.latestVersion">
                最新版本 {{ appUpdate.latestVersion }}
                <span v-if="appUpdate.releaseName"> · {{ appUpdate.releaseName }}</span>
              </p>
              <p v-if="!appUpdate?.latestVersion" class="update-source-line">
                更新通道：{{ appUpdateSource === "gitee" ? "Gitee 镜像" : "GitHub" }}
              </p>
              <p class="update-source-line">
                {{ appUpdateHost }}/{{ appUpdate?.repo || "dracula-zz-1/crac_exam_go" }}
              </p>
              <p v-if="appUpdate?.publishedAt">发布于 {{ fmtDateTime(appUpdate.publishedAt) }}</p>
              <p v-if="appUpdate?.downloadedName" class="mono update-hash">
                已下载：{{ appUpdate.downloadedName }}
              </p>
            </div>
            <div class="update-actions app-update-actions">
              <button
                class="btn btn-sm btn-ghost"
                :disabled="appChecking || appDownloading"
                @click="checkAppUpdate"
              >
                <span v-if="appChecking" class="spinner dark" />
                <AppIcon v-else name="search" :size="14" />
                {{ appChecking ? "检查中…" : "检查更新" }}
              </button>
              <button
                v-if="appUpdate?.available && appUpdate?.asset"
                class="btn btn-sm btn-accent"
                :disabled="appDownloading || appInstalling"
                @click="downloadAppUpdate"
              >
                <span v-if="appDownloading" class="spinner" />
                <AppIcon v-else name="download" :size="15" />
                {{ appDownloading ? "下载中…" : "下载更新" }}
              </button>
              <button
                v-if="appUpdate?.downloadedPath"
                class="btn btn-sm btn-accent"
                :disabled="appInstalling"
                @click="askRunInstaller"
              >
                <AppIcon name="zap" :size="15" />
                运行安装
              </button>
            </div>
          </div>

          <div v-if="appUpdate?.available" class="update-available">
            <AppIcon name="sparkles" :size="16" />
            <div>
              <strong>发现应用新版本：{{ appUpdate.latestVersion }}</strong>
              <p v-if="appUpdate.asset">
                {{ appUpdate.asset.name }} · {{ fmtSize(appUpdate.asset.size) }}
              </p>
              <p v-else>Release 未提供 .exe / .msi 安装包，请打开发布页手动下载。</p>
            </div>
          </div>
          <div v-else-if="appUpdate?.checked" class="update-ok">
            <AppIcon name="check" :size="15" />
            {{ appUpdate.message || "已是最新版本" }}
          </div>

          <div v-if="appUpdateError || appUpdate?.error" class="update-error">
            <AppIcon name="warning" :size="14" />
            {{ appUpdateError || appUpdate?.error }}
          </div>

          <div v-if="appDownloading" class="import-progress">
            <div class="import-progress-top">
              <span>{{ appProgress?.message || "正在下载更新包…" }}</span>
              <strong>{{ appProgress?.percent ?? appUpdate?.downloadPercent ?? 0 }}%</strong>
            </div>
            <div class="progress amber">
              <span
                :style="{
                  width: (appProgress?.percent ?? appUpdate?.downloadPercent ?? 0) + '%',
                }"
              />
            </div>
          </div>

          <div v-if="appUpdate?.notes" class="app-update-notes">{{ appUpdate.notes }}</div>

          <div class="app-update-links">
            <button class="btn btn-sm btn-ghost" @click="openAppRelease">
              <AppIcon name="arrowUpRight" :size="14" />
              打开发布页
            </button>
            <button
              v-if="appUpdate?.downloadedPath"
              class="btn btn-sm btn-ghost"
              @click="openAppUpdateFolder"
            >
              <AppIcon name="folder" :size="14" />
              打开下载目录
            </button>
          </div>

          <p class="update-note">
            更新包从当前通道（{{ appUpdateSource === "gitee" ? "Gitee 镜像" : "GitHub" }}）
            的 Release 获取，仅支持 .exe / .msi。国内网络可选 Gitee 镜像通道加快下载；
            运行安装程序前请确认下载来源可信，更新不会删除本地题库与学习数据。
          </p>
        </section>

        <!-- 导入题库 -->
        <section class="surface card-pad">
          <div class="section-head">
            <div>
              <h2 class="section-title">导入题库</h2>
              <p class="section-sub">仅支持 ZIP；ZIP 内读取 5 份 PDF，使用纯 Go 解析</p>
            </div>
            <span class="chip chip-amber">ZIP only</span>
          </div>

          <div class="import-note">
            <AppIcon name="info" :size="16" />
            <div>
              <strong>ZIP 需包含以下文件</strong>
              <p>
                总题库.pdf · 总题库附图标记.pdf · A类题库.pdf · B类题库.pdf · C类题库.pdf
                <br />
                文件名包含上述关键字即可，允许放在子目录中。
              </p>
            </div>
          </div>

          <!-- data-file-drop-target 是 Wails 桌面端文件拖放的必需标记：
               拖放落在没有该属性的元素上时，原生层会直接忽略这次拖放 -->
          <div
            class="dropzone"
            data-file-drop-target
            :class="{ dragging, disabled: importing }"
            @dragover.prevent="dragging = true"
            @dragenter.prevent="dragging = true"
            @dragleave.prevent="dragging = false"
            @drop.prevent="onHtmlDrop"
            @click="chooseZip"
          >
            <div class="dropzone-icon">
              <AppIcon name="upload" :size="28" :stroke-width="1.7" />
            </div>
            <strong>{{ importing ? "正在导入题库…" : "把 ZIP 拖到这里，或点击选择文件" }}</strong>
            <span>导入会替换题库，并按 J/P/I 保留错题、收藏与练习进度</span>
          </div>

          <div v-if="importing" class="import-progress">
            <div class="import-progress-top">
              <span>{{ stageLabel }}</span>
              <strong>{{ progressPercent }}%</strong>
            </div>
            <div class="progress amber">
              <span :style="{ width: progressPercent + '%' }" />
            </div>
          </div>

          <div v-if="lastResult" class="import-result" :class="{ failed: !lastResult.success }">
            <div class="result-head">
              <div class="result-icon">
                <AppIcon :name="lastResult.success ? 'check' : 'warning'" :size="18" :stroke-width="2.3" />
              </div>
              <div>
                <strong>{{ lastResult.message }}</strong>
                <p v-if="lastResult.success">
                  耗时 {{ (lastResult.elapsedMs / 1000).toFixed(1) }} 秒
                  <span v-if="lastResult.duplicateJ"> · 题号 J 重复值 {{ lastResult.duplicateJ }} 个</span>
                  <span v-if="lastResult.duplicateJpi"> · 跳过 J/P/I 重复题目 {{ lastResult.duplicateJpi }} 道</span>
                  <span
                    v-if="
                      lastResult.preservedFavorites ||
                      lastResult.preservedErrors ||
                      lastResult.preservedProgress
                    "
                  >
                    · 保留 收藏 {{ lastResult.preservedFavorites }} / 错题
                    {{ lastResult.preservedErrors }} / 进度 {{ lastResult.preservedProgress }}
                  </span>
                </p>
              </div>
            </div>

            <div v-if="lastResult.success" class="result-stats">
              <div v-for="item in resultStats" :key="item.label" class="result-stat">
                <strong>{{ item.value }}</strong>
                <span>{{ item.label }}</span>
              </div>
            </div>

            <ul v-if="lastResult.warnings?.length" class="warnings">
              <li v-for="(w, i) in lastResult.warnings" :key="i">
                <AppIcon name="warning" :size="13" />
                {{ w }}
              </li>
            </ul>
          </div>
        </section>

        <!-- 数据管理 -->
        <section class="surface card-pad">
          <div class="section-head">
            <div>
              <h2 class="section-title">数据管理</h2>
              <p class="section-sub">所有数据保存在本机 SQLite 数据库</p>
            </div>
            <button class="btn btn-sm btn-ghost" @click="loadStats">
              <AppIcon name="refresh" :size="14" />
              刷新
            </button>
          </div>

          <div class="db-stats">
            <div>
              <strong>{{ stats?.questions ?? "—" }}</strong>
              <span>题目</span>
            </div>
            <div>
              <strong>{{ stats?.examRecords ?? "—" }}</strong>
              <span>考试记录</span>
            </div>
            <div>
              <strong>{{ stats?.errorRecords ?? "—" }}</strong>
              <span>错题记录</span>
            </div>
            <div>
              <strong>{{ stats?.favorites ?? "—" }}</strong>
              <span>收藏</span>
            </div>
            <div>
              <strong>{{ stats?.users ?? "—" }}</strong>
              <span>档案</span>
            </div>
          </div>

          <div class="danger-list">
            <div class="danger-row">
              <div>
                <strong>清空我的学习数据</strong>
                <p>删除当前档案的考试记录、错题本、收藏与练习进度，题库保留。</p>
              </div>
              <button class="btn btn-sm btn-danger" @click="clearUserOpen = true">
                <AppIcon name="trash" :size="14" />
                清空
              </button>
            </div>
            <div class="danger-row">
              <div>
                <strong>清空题库</strong>
                <p>删除全部题目及依赖的错题、收藏与练习进度；已交卷考试的快照保留。</p>
              </div>
              <button class="btn btn-sm btn-danger" @click="clearBankOpen = true">
                <AppIcon name="trash" :size="14" />
                清空
              </button>
            </div>
          </div>
        </section>
      </div>

      <aside class="stack gap-20">
        <!-- 关于 -->
        <section class="surface card-pad about-card">
          <div class="about-mark"><AppIcon name="radio" :size="24" :stroke-width="2" /></div>
          <h2>{{ info?.name ?? "业余无线电模拟考试" }}</h2>
          <p class="about-version">
            v{{ info?.version ?? "3.0.0" }}
            <span v-if="usingMock" class="chip chip-amber">演示模式</span>
          </p>
          <p class="about-desc">{{ info?.description }}</p>

          <div class="about-meta">
            <div>
              <span>作者</span>
              <strong>{{ info?.author ?? "BA4RHH" }}</strong>
            </div>
            <div>
              <span>Go</span>
              <strong class="mono">{{ info?.goVersion ?? "—" }}</strong>
            </div>
          </div>

          <div class="data-dir">
            <span>数据目录</span>
            <div class="data-dir-box">
              <code class="data-dir-path" :title="info?.dataDir ?? ''">
                {{ info?.dataDir ?? "—" }}
              </code>
              <button
                class="data-dir-more"
                :disabled="dataDirBusy"
                title="更改数据目录"
                @click="pickDataDir"
              >
                …
              </button>
            </div>
          </div>

          <p class="copyright">{{ info?.copyright }}</p>
        </section>

        <section class="surface card-pad tips-card">
          <h3 class="section-title">小提示</h3>
          <ul>
            <li><AppIcon name="check" :size="14" /> 导入前建议先保留原始题库 ZIP</li>
            <li><AppIcon name="check" :size="14" /> PDF 解析在本地完成，不联网</li>
            <li><AppIcon name="check" :size="14" /> 导入是原子替换，失败不会破坏旧题库</li>
            <li><AppIcon name="check" :size="14" /> 题库中重复题号会在结果里提示</li>
          </ul>
        </section>
      </aside>
    </div>

    <ConfirmDialog
      :open="confirmImportOpen"
      title="确认导入题库"
      :message="pendingPath ? `即将导入：${fileName(pendingPath)}\n\n导入会替换现有题库；错题、收藏与练习进度会按 J/P/I 尽量保留（找不到对应题目的条目会丢弃）。已交卷的考试复盘会保留。` : ''"
      confirm-text="开始导入"
      danger
      :loading="false"
      @cancel="confirmImportOpen = false"
      @confirm="runImport"
    >
      <template #icon><AppIcon name="upload" :size="20" /></template>
    </ConfirmDialog>

    <ConfirmDialog
      :open="confirmUpdateOpen"
      title="更新题库"
      :message="`将下载并替换为 CRAC 官方最新题库${updateStatus?.remoteVersion ? `：${updateStatus.remoteVersion}` : ''}。\n\n错题、收藏与练习进度会按 J/P/I 尽量保留；已交卷的考试复盘快照保留。`"
      confirm-text="下载并更新"
      :loading="applyingUpdate"
      @cancel="confirmUpdateOpen = false"
      @confirm="applyUpdate"
    >
      <template #icon><AppIcon name="download" :size="20" /></template>
    </ConfirmDialog>

    <ConfirmDialog
      :open="appInstallOpen"
      title="运行应用更新"
      :message="`即将启动已下载的安装程序${appUpdate?.downloadedName ? `：${appUpdate.downloadedName}` : ''}。\n\n安装过程中应用可能会关闭；本地题库、考试记录与学习数据不会被删除。`"
      confirm-text="运行安装程序"
      :loading="appInstalling"
      @cancel="appInstallOpen = false"
      @confirm="runInstaller"
    >
      <template #icon><AppIcon name="zap" :size="20" /></template>
    </ConfirmDialog>

    <ConfirmDialog
      :open="clearUserOpen"
      title="清空学习数据"
      message="将删除当前档案的全部考试记录、错题、收藏与练习进度。此操作不可恢复，题库不受影响。"
      confirm-text="确认清空"
      danger
      :loading="clearingUser"
      @cancel="clearUserOpen = false"
      @confirm="clearUserData"
    >
      <template #icon><AppIcon name="trash" :size="20" /></template>
    </ConfirmDialog>

    <ConfirmDialog
      :open="clearBankOpen"
      title="清空题库"
      message="将删除全部题目，以及所有错题、收藏与练习进度。已交卷考试的历史快照会保留，但题库需要重新导入。此操作不可恢复。"
      confirm-text="确认清空"
      danger
      :loading="clearingBank"
      @cancel="clearBankOpen = false"
      @confirm="clearBank"
    >
      <template #icon><AppIcon name="trash" :size="20" /></template>
    </ConfirmDialog>

    <!-- 更换数据目录 -->
    <BaseModal
      :open="dataDirOpen"
      title="切换数据目录"
      subtitle="切换后需要重启应用才会生效"
      @close="dataDirOpen = false"
    >
      <div class="data-dir-info">
        <div class="data-dir-icon"><AppIcon name="folder" :size="22" /></div>
        <div>
          <strong>新的数据目录</strong>
          <p class="mono">{{ pendingDataDir }}</p>
          <p>
            选择「复制现有数据」会把当前的题库、考试记录、错题与收藏一起复制到新目录；
            选择「仅切换」则新目录从空数据开始。
          </p>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-ghost" :disabled="dataDirBusy" @click="dataDirOpen = false">
          取消
        </button>
        <button class="btn btn-soft" :disabled="dataDirBusy" @click="applyDataDir(false)">
          仅切换（空数据）
        </button>
        <button class="btn btn-accent" :disabled="dataDirBusy" @click="applyDataDir(true)">
          <span v-if="dataDirBusy" class="spinner dark" />
          <AppIcon v-else name="download" :size="15" />
          复制现有数据并切换
        </button>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.settings-page {
  max-width: 1180px;
}
.settings-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(320px, 0.8fr);
  gap: 20px;
  align-items: start;
}
.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
}

.update-current {
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 14px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: var(--surface-2);
}
.update-icon {
  width: 42px;
  height: 42px;
  border-radius: 13px;
  display: grid;
  place-items: center;
  background: var(--surface);
  color: var(--ink-3);
  border: 1px solid var(--line);
  flex: none;
}
.update-icon.available {
  background: var(--amber-soft);
  color: #9a6514;
  border-color: transparent;
}
.update-main {
  flex: 1;
  min-width: 0;
}
.update-main strong {
  font-size: 13px;
  font-weight: 700;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.update-main p {
  margin-top: 3px;
  font-size: 10.5px;
  color: var(--ink-3);
}
.update-hash {
  font-size: 9.5px !important;
  color: var(--muted) !important;
  overflow-wrap: anywhere;
}
.update-actions {
  display: flex;
  gap: 6px;
  flex: none;
}
.update-available {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  margin-top: 12px;
  padding: 11px 13px;
  border-radius: 11px;
  background: var(--amber-soft);
  color: #7a5211;
}
.update-available strong {
  font-size: 12px;
  font-weight: 700;
}
.update-available p {
  margin-top: 3px;
  font-size: 10.5px;
  color: #8a631f;
}
.update-ok {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 12px;
  font-size: 11.5px;
  color: var(--green);
}
.update-error {
  display: flex;
  align-items: flex-start;
  gap: 7px;
  margin-top: 12px;
  padding: 9px 11px;
  border-radius: 10px;
  background: var(--red-soft);
  color: #a63b2e;
  font-size: 10.5px;
  line-height: 1.55;
}
.update-toggles {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 14px;
}
.toggle-row {
  display: flex;
  align-items: flex-start;
  gap: 11px;
  padding: 10px 12px;
  border-radius: 11px;
  border: 1px solid var(--line);
  background: var(--surface);
  text-align: left;
  transition: background 0.15s;
}
.toggle-row:hover {
  background: var(--surface-2);
}
.toggle-row.disabled {
  opacity: 0.45;
  pointer-events: none;
}
.toggle {
  width: 34px;
  height: 20px;
  border-radius: 99px;
  background: var(--surface-3);
  position: relative;
  flex: none;
  margin-top: 1px;
  transition: background 0.16s;
}
.toggle i {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 3px rgba(22, 27, 33, 0.2);
  transition: transform 0.16s var(--ease);
}
.toggle.on {
  background: var(--teal);
}
.toggle.on i {
  transform: translateX(14px);
}
.toggle-row strong {
  display: block;
  font-size: 11.5px;
  font-weight: 650;
}
.toggle-row em {
  display: block;
  margin-top: 2px;
  font-style: normal;
  font-size: 10px;
  color: var(--ink-3);
}
.update-note {
  margin-top: 12px;
  font-size: 10px;
  color: var(--muted);
  line-height: 1.6;
}

.app-update-actions {
  flex-wrap: wrap;
  justify-content: flex-end;
  max-width: 260px;
}

.app-update-notes {
  margin-top: 12px;
  max-height: 150px;
  overflow: auto;
  white-space: pre-wrap;
  padding: 11px 13px;
  border-radius: 11px;
  border: 1px solid var(--line);
  background: var(--surface-2);
  font-size: 10.5px;
  line-height: 1.65;
  color: var(--ink-3);
}

.app-update-links {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 14px;
}

.update-source-switch {
  flex: none;
}
.update-source-switch button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.update-source-line {
  font-family: var(--font-mono);
  letter-spacing: 0.01em;
}

.import-note {
  display: flex;
  gap: 11px;
  padding: 13px 15px;
  border-radius: 12px;
  background: var(--amber-soft);
  color: #7a5211;
  margin-bottom: 16px;
}
.import-note strong {
  font-size: 12px;
  font-weight: 700;
}
.import-note p {
  margin-top: 4px;
  font-size: 11.5px;
  line-height: 1.7;
  color: #8a631f;
}

.dropzone {
  border: 1.5px dashed var(--line-strong);
  border-radius: 16px;
  padding: 34px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-align: center;
  cursor: pointer;
  background: linear-gradient(180deg, var(--surface-2), var(--surface));
  transition: all 0.18s var(--ease);
}
.dropzone:hover,
.dropzone.dragging {
  border-color: var(--amber);
  background: #fffaf0;
  transform: translateY(-1px);
}
.dropzone.disabled {
  pointer-events: none;
  opacity: 0.6;
}
.dropzone-icon {
  width: 58px;
  height: 58px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  background: var(--surface);
  color: var(--amber-deep);
  box-shadow: var(--shadow-xs);
  border: 1px solid var(--line);
  margin-bottom: 4px;
}
.dropzone strong {
  font-size: 13.5px;
  font-weight: 700;
}
.dropzone span {
  font-size: 11.5px;
  color: var(--ink-3);
  line-height: 1.6;
}

.import-progress {
  margin-top: 18px;
}
.import-progress-top {
  display: flex;
  justify-content: space-between;
  font-size: 11.5px;
  color: var(--ink-3);
  margin-bottom: 7px;
}
.import-progress-top strong {
  color: var(--amber-deep);
}
.import-progress .progress {
  height: 9px;
}

.import-result {
  margin-top: 18px;
  border: 1px solid rgba(77, 155, 106, 0.28);
  background: var(--green-soft);
  border-radius: 14px;
  padding: 15px 16px;
}
.import-result.failed {
  border-color: rgba(208, 92, 77, 0.28);
  background: var(--red-soft);
}
.result-head {
  display: flex;
  gap: 11px;
  align-items: flex-start;
}
.result-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--surface);
  color: var(--green);
  flex: none;
}
.import-result.failed .result-icon {
  color: var(--red);
}
.result-head strong {
  font-size: 13px;
  font-weight: 700;
}
.result-head p {
  margin-top: 3px;
  font-size: 11px;
  color: var(--ink-3);
}
.result-stats {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 7px;
  margin-top: 13px;
}
.result-stat {
  background: rgba(255, 253, 248, 0.75);
  border-radius: 9px;
  padding: 8px 4px;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.result-stat strong {
  font-size: 15px;
  font-weight: 800;
}
.result-stat span {
  font-size: 9.5px;
  color: var(--ink-3);
}
.warnings {
  list-style: none;
  margin: 12px 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.warnings li {
  display: flex;
  gap: 6px;
  align-items: flex-start;
  font-size: 10.5px;
  color: #8a631f;
  line-height: 1.55;
}

.db-stats {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 20px;
}
.db-stats div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 12px 6px;
  border-radius: 11px;
  background: var(--surface-2);
}
.db-stats strong {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.db-stats span {
  font-size: 9.5px;
  color: var(--ink-3);
}

.danger-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.danger-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 13px 15px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: var(--surface);
}
.danger-row strong {
  font-size: 12.5px;
  font-weight: 700;
}
.danger-row p {
  margin-top: 3px;
  font-size: 10.5px;
  color: var(--ink-3);
  line-height: 1.55;
}
.danger-row .btn {
  margin-left: auto;
  flex: none;
}

.about-card {
  text-align: center;
}
.about-mark {
  width: 54px;
  height: 54px;
  border-radius: 17px;
  margin: 0 auto 14px;
  display: grid;
  place-items: center;
  color: #241703;
  background: linear-gradient(135deg, #f4c26b, var(--amber));
  box-shadow: 0 10px 26px rgba(232, 163, 61, 0.28);
}
.about-card h2 {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.about-version {
  margin-top: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  font-size: 11.5px;
  color: var(--ink-3);
}
.about-desc {
  margin-top: 10px;
  font-size: 11.5px;
  line-height: 1.7;
  color: var(--ink-3);
}
.about-meta {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 16px;
}
.about-meta div {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 9px;
  border-radius: 9px;
  background: var(--surface-2);
}
.about-meta span {
  font-size: 9.5px;
  color: var(--ink-3);
}
.about-meta strong {
  font-size: 11.5px;
  font-weight: 650;
}
.data-dir {
  margin-top: 12px;
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.data-dir span {
  font-size: 9.5px;
  color: var(--ink-3);
}
.data-dir-box {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 7px 7px 10px;
  border-radius: 8px;
  background: var(--navy);
}
.data-dir-path {
  flex: 1;
  min-width: 0;
  font-family: var(--font-mono);
  font-size: 9.5px;
  color: #d8d2c4;
  line-height: 1.6;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.data-dir-more {
  flex: none;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  display: grid;
  place-items: center;
  color: #d8d2c4;
  background: rgba(255, 255, 255, 0.09);
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  transition: background 0.15s, color 0.15s;
}
.data-dir-more:hover {
  background: rgba(232, 163, 61, 0.28);
  color: #f4c26b;
}
.data-dir-more:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.data-dir-info {
  display: flex;
  gap: 14px;
}
.data-dir-info strong {
  font-size: 13.5px;
  font-weight: 700;
}
.data-dir-info p {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.65;
  color: var(--ink-3);
  overflow-wrap: anywhere;
}
.data-dir-icon {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: var(--amber-soft);
  color: #9a6514;
  display: grid;
  place-items: center;
  flex: none;
}
.tech-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  justify-content: center;
  margin-top: 16px;
}
.copyright {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--line);
  font-size: 9.5px;
  color: var(--muted);
}

.tips-card ul {
  list-style: none;
  margin: 12px 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.tips-card li {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  font-size: 11px;
  color: var(--ink-3);
  line-height: 1.55;
}
.tips-card li :deep(svg) {
  color: var(--teal);
  margin-top: 1px;
}

@media (max-width: 1180px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
  .result-stats {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
