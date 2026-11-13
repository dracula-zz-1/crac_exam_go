<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "@/api";
import { useSessionStore } from "@/stores/session";
import { useToastStore } from "@/stores/toast";
import { LICENSE_LEVELS, nextLevelCategory } from "@/lib/license";
import AppIcon from "@/components/AppIcon.vue";
import LicenseCardArt from "@/components/LicenseCardArt.vue";

const router = useRouter();
const session = useSessionStore();
const toast = useToastStore();

const picked = ref<"none" | "A" | "B" | "C">("none");
const saving = ref(false);

const nextCat = computed(() => nextLevelCategory(picked.value));

async function submit() {
  if (!session.user || saving.value) return;
  saving.value = true;
  try {
    const user = await api.setUserLevel(session.user.id, picked.value);
    session.login(user);
    toast.success("已记录你的操作等级", `练习与考试默认按 ${nextCat.value} 类进行`);
    router.push("/home");
  } catch (e: any) {
    toast.error("保存失败", e?.message);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="page welcome-page">
    <header class="page-head">
      <div>
        <span class="eyebrow">WELCOME · 首次使用</span>
        <h1 class="page-title">你好，{{ session.user?.username }}</h1>
        <p class="page-sub">
          先记录你目前的业余无线电操作等级：之后从「逐题练习 / 模拟考试」进入时，
          会默认按你下一次要考的等级出题。
        </p>
      </div>
    </header>

    <div class="welcome-grid">
      <section class="surface card-pad level-card">
        <h2 class="section-title">你现在的等级</h2>
        <p class="section-sub">等级由低到高：A 类 → B 类 → C 类</p>

        <div class="level-list">
          <button
            v-for="lv in LICENSE_LEVELS"
            :key="lv.value"
            class="level-item"
            :class="{ active: picked === lv.value }"
            @click="picked = lv.value"
          >
            <span class="level-badge">{{ lv.value === "none" ? "—" : lv.value }}</span>
            <span class="level-text">
              <strong>{{ lv.label }}</strong>
              <em>{{ lv.desc }}</em>
            </span>
            <AppIcon
              v-if="picked === lv.value"
              name="check"
              :size="16"
              :stroke-width="2.6"
              class="level-check"
            />
          </button>
        </div>

        <div class="level-next">
          <AppIcon name="target" :size="15" />
          默认按 <b>{{ nextCat }} 类</b> 开始练习与模拟考试
        </div>

        <button class="btn btn-accent btn-lg btn-block" :disabled="saving" @click="submit">
          <span v-if="saving" class="spinner dark" />
          <template v-else>
            开始使用
            <AppIcon name="arrowRight" :size="17" />
          </template>
        </button>
      </section>

      <aside class="welcome-art">
        <LicenseCardArt />
        <p class="welcome-note">
          上图是《业余无线电台操作技术能力验证证书》的示意图（数据为示例），
          实际证书样式以核发机构签发为准。
        </p>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.welcome-page {
  max-width: 1160px;
}
.welcome-grid {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(0, 1.25fr);
  gap: 22px;
  align-items: start;
}
.level-card {
  display: flex;
  flex-direction: column;
}
.level-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 18px 0 16px;
}
.level-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border: 1.5px solid var(--line);
  border-radius: 14px;
  background: var(--surface);
  text-align: left;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
}
.level-item:hover {
  border-color: var(--line-strong);
  background: #fffefb;
}
.level-item.active {
  border-color: var(--amber);
  background: #fffaf0;
  box-shadow: 0 0 0 4px rgba(232, 163, 61, 0.12);
}
.level-badge {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 15px;
  font-weight: 800;
  background: var(--surface-2);
  color: var(--ink-2);
  flex: none;
}
.level-item.active .level-badge {
  background: var(--amber);
  color: #241703;
}
.level-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}
.level-text strong {
  font-size: 14px;
  font-weight: 700;
}
.level-text em {
  font-style: normal;
  font-size: 11.5px;
  color: var(--ink-3);
}
.level-check {
  margin-left: auto;
  color: var(--amber-deep);
  flex: none;
}
.level-next {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 11px 14px;
  border-radius: 12px;
  background: var(--teal-soft);
  color: #147065;
  font-size: 12.5px;
  margin-bottom: 18px;
}
.level-next b {
  font-weight: 800;
}
.welcome-art {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.welcome-note {
  font-size: 11.5px;
  line-height: 1.7;
  color: var(--muted);
}

@media (max-width: 1040px) {
  .welcome-grid {
    grid-template-columns: 1fr;
  }
  .welcome-art {
    order: -1;
  }
}
</style>
