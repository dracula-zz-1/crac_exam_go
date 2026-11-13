import { createRouter, createWebHashHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { useSessionStore } from "@/stores/session";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "login",
    component: () => import("@/views/LoginView.vue"),
    meta: { public: true },
  },
  {
    path: "/home",
    name: "dashboard",
    component: () => import("@/views/DashboardView.vue"),
  },
  {
    path: "/practice/:mode/:category",
    name: "practice",
    component: () => import("@/views/PracticeView.vue"),
    props: true,
  },
  {
    path: "/exam/:category",
    name: "exam",
    component: () => import("@/views/ExamView.vue"),
    props: true,
  },
  {
    path: "/result/:examId",
    name: "result",
    component: () => import("@/views/ResultView.vue"),
    props: true,
  },
  {
    path: "/errors",
    name: "errors",
    component: () => import("@/views/CollectionView.vue"),
    props: { mode: "error" },
  },
  {
    path: "/favorites",
    name: "favorites",
    component: () => import("@/views/CollectionView.vue"),
    props: { mode: "favorite" },
  },
  {
    path: "/bank",
    name: "bank",
    component: () => import("@/views/BankView.vue"),
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("@/views/SettingsView.vue"),
  },
  { path: "/:pathMatch(.*)*", redirect: "/home" },
];

// 桌面端使用 hash 路由：WebView 刷新时无需服务端做 SPA 回退。
const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
});

router.beforeEach((to) => {
  const session = useSessionStore();
  if (!to.meta.public && !session.user) {
    return { path: "/" };
  }
  if (to.path === "/" && session.user) {
    return { path: "/home" };
  }
  return true;
});

export default router;
