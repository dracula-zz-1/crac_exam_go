<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useSessionStore } from "@/stores/session";
import { usingMock } from "@/api";
import { nextLevelCategory } from "@/lib/license";
import AppIcon from "./AppIcon.vue";

const route = useRoute();
const router = useRouter();
const session = useSessionStore();

interface NavItem {
  label: string;
  icon: string;
  to: string;
  match: string[];
  hint?: string;
}

// 侧栏入口默认进入“下一次要考”的等级，具体类别可在页面内切换
const defaultCategory = computed(() => nextLevelCategory(session.user?.level));

const learnNav = computed<NavItem[]>(() => [
  { label: "总览", icon: "home", to: "/home", match: ["dashboard"] },
  {
    label: "逐题练习",
    icon: "practice",
    to: `/practice/practice/${defaultCategory.value}`,
    match: ["practice"],
  },
  {
    label: "模拟考试",
    icon: "exam",
    to: `/exam/${defaultCategory.value}`,
    match: ["exam"],
  },
  { label: "错题本", icon: "errors", to: "/errors", match: ["errors"] },
  { label: "收藏夹", icon: "star", to: "/favorites", match: ["favorites"] },
]);

const toolNav: NavItem[] = [
  { label: "题库管理", icon: "bank", to: "/bank", match: ["bank"] },
  { label: "设置", icon: "settings", to: "/settings", match: ["settings"] },
];

const titleMap: Record<string, string> = {
  welcome: "欢迎",
  dashboard: "总览",
  practice: "逐题练习",
  exam: "模拟考试",
  result: "考试结果",
  errors: "错题本",
  favorites: "收藏夹",
  bank: "题库管理",
  settings: "设置",
};

const pageTitle = computed(() => titleMap[String(route.name ?? "")] ?? "工作台");

const today = new Date().toLocaleDateString("zh-CN", {
  month: "long",
  day: "numeric",
  weekday: "long",
});

const maskedId = computed(() => {
  const id = session.user?.idCard ?? "";
  if (id.length <= 6) return id;
  return `${id.slice(0, 3)}••••${id.slice(-3)}`;
});

function isActive(item: NavItem) {
  return item.match.includes(String(route.name ?? ""));
}

function logout() {
  session.logout();
  router.push("/");
}
</script>

<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">
          <AppIcon name="radio" :size="22" :stroke-width="2" />
        </div>
        <div class="brand-text">
          <strong>HAM LAB</strong>
          <span>业余无线电模拟考试</span>
        </div>
      </div>

      <nav class="nav">
        <div class="nav-label">学习</div>
        <router-link
          v-for="item in learnNav"
          :key="item.label"
          :to="item.to"
          class="nav-item"
          :class="{ active: isActive(item) }"
        >
          <AppIcon :name="item.icon" :size="18" />
          <span>{{ item.label }}</span>
          <span class="nav-glow" />
        </router-link>

        <div class="nav-label spaced">工具</div>
        <router-link
          v-for="item in toolNav"
          :key="item.label"
          :to="item.to"
          class="nav-item"
          :class="{ active: isActive(item) }"
        >
          <AppIcon :name="item.icon" :size="18" />
          <span>{{ item.label }}</span>
          <span class="nav-glow" />
        </router-link>
      </nav>

      <div class="sidebar-art">
        <svg viewBox="0 0 220 60" fill="none" aria-hidden="true">
          <path
            d="M0 34 Q18 8 36 34 T72 34 T108 34 T144 34 T180 34 T216 34"
            stroke="rgba(232,163,61,.55)"
            stroke-width="1.6"
          />
          <path
            d="M0 44 Q18 26 36 44 T72 44 T108 44 T144 44 T180 44 T216 44"
            stroke="rgba(42,157,143,.35)"
            stroke-width="1.2"
          />
        </svg>
        <div class="status-line">
          <span class="pulse" />
          离线运行 · 数据保存在本机
        </div>
      </div>

      <div class="user-card" v-if="session.user">
        <div class="avatar">{{ session.user.username.slice(0, 1) }}</div>
        <div class="user-meta">
          <strong>{{ session.user.username }}</strong>
          <span class="mono">{{ maskedId }}</span>
        </div>
        <button class="logout" title="退出登录" @click="logout">
          <AppIcon name="logout" :size="17" />
        </button>
      </div>
    </aside>

    <main class="main">
      <header class="topbar">
        <div class="crumb">
          <span class="crumb-root">HAM LAB</span>
          <span class="crumb-sep">/</span>
          <span class="crumb-current">{{ pageTitle }}</span>
        </div>
        <div class="topbar-right">
          <span v-if="usingMock" class="chip chip-amber">
            <span class="dot" style="background: #c9821f" />
            浏览器演示数据
          </span>
          <span class="date">{{ today }}</span>
        </div>
      </header>
      <div class="content">
        <slot />
      </div>
    </main>
  </div>
</template>

<style scoped>
.shell {
  display: grid;
  grid-template-columns: 252px minmax(0, 1fr);
  height: 100vh;
  position: relative;
  z-index: 1;
}

.sidebar {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 26px 18px 20px;
  background: linear-gradient(180deg, #111a22 0%, #0d141b 100%);
  color: #f4f0e6;
  overflow: hidden;
}
.sidebar::before {
  content: "";
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 20% 0%, rgba(232, 163, 61, 0.12), transparent 45%),
    radial-gradient(circle at 90% 80%, rgba(42, 157, 143, 0.1), transparent 50%);
  pointer-events: none;
}
.sidebar::after {
  content: "";
  position: absolute;
  top: 0;
  right: 0;
  width: 1px;
  height: 100%;
  background: linear-gradient(180deg, transparent, rgba(232, 163, 61, 0.25), transparent);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 8px 26px;
  position: relative;
}
.brand-mark {
  width: 42px;
  height: 42px;
  border-radius: 13px;
  display: grid;
  place-items: center;
  color: #241703;
  background: linear-gradient(135deg, #f4c26b, var(--amber));
  box-shadow: 0 8px 20px rgba(232, 163, 61, 0.28);
}
.brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.brand-text strong {
  font-size: 17px;
  letter-spacing: 0.08em;
}
.brand-text span {
  font-size: 10.5px;
  color: rgba(244, 240, 230, 0.5);
  letter-spacing: 0.02em;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  position: relative;
}
.nav-label {
  padding: 8px 12px 6px;
  font-size: 10px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(244, 240, 230, 0.32);
  font-weight: 700;
}
.nav-label.spaced {
  margin-top: 18px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 44px;
  padding: 0 13px;
  border-radius: 12px;
  font-size: 13.5px;
  font-weight: 550;
  color: rgba(244, 240, 230, 0.66);
  position: relative;
  transition: color 0.16s, background 0.16s, transform 0.16s;
}
.nav-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.05);
}
.nav-item.active {
  color: #fff;
  background: rgba(232, 163, 61, 0.13);
}
.nav-item.active::before {
  content: "";
  position: absolute;
  left: -18px;
  top: 10px;
  bottom: 10px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--amber);
}
.nav-item.active :deep(svg) {
  color: var(--amber);
}
.nav-glow {
  margin-left: auto;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: transparent;
}
.nav-item.active .nav-glow {
  background: var(--amber);
  box-shadow: 0 0 12px rgba(232, 163, 61, 0.8);
}

.sidebar-art {
  margin-top: auto;
  padding: 18px 10px 12px;
  position: relative;
}
.sidebar-art svg {
  width: 100%;
  height: 48px;
  opacity: 0.8;
}
.status-line {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 10.5px;
  color: rgba(244, 240, 230, 0.38);
  padding: 6px 4px 0;
}
.pulse {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4ecdc4;
  box-shadow: 0 0 0 0 rgba(78, 205, 196, 0.5);
  animation: pulse 2.4s infinite;
}
@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(78, 205, 196, 0.45); }
  70% { box-shadow: 0 0 0 7px rgba(78, 205, 196, 0); }
  100% { box-shadow: 0 0 0 0 rgba(78, 205, 196, 0); }
}

.user-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.07);
  position: relative;
}
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 11px;
  display: grid;
  place-items: center;
  font-weight: 800;
  font-size: 15px;
  color: #241703;
  background: linear-gradient(135deg, #f4c26b, var(--amber));
  flex: none;
}
.user-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.user-meta strong {
  font-size: 12.5px;
  font-weight: 650;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.user-meta span {
  font-size: 10px;
  color: rgba(244, 240, 230, 0.4);
}
.logout {
  margin-left: auto;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  color: rgba(244, 240, 230, 0.5);
  transition: all 0.16s;
}
.logout:hover {
  background: rgba(208, 92, 77, 0.2);
  color: #ff9c8d;
}

.main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
.topbar {
  height: 64px;
  flex: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 38px;
  border-bottom: 1px solid rgba(227, 221, 207, 0.7);
  background: rgba(246, 243, 236, 0.72);
  backdrop-filter: blur(14px);
  position: relative;
  z-index: 5;
}
.crumb {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 12px;
}
.crumb-root {
  letter-spacing: 0.14em;
  font-weight: 700;
  color: var(--ink-3);
}
.crumb-sep {
  color: var(--muted);
}
.crumb-current {
  font-weight: 700;
  color: var(--ink);
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 14px;
}
.date {
  font-size: 12px;
  color: var(--ink-3);
}
.content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
</style>
