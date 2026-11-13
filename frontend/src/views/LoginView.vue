<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useSessionStore } from "@/stores/session";
import { useToastStore } from "@/stores/toast";
import { api, usingMock } from "@/api";
import AppIcon from "@/components/AppIcon.vue";

const router = useRouter();
const session = useSessionStore();
const toast = useToastStore();

const username = ref("");
const idCard = ref("");
const loading = ref(false);
const error = ref("");
const appVersion = ref("3.0.0");

onMounted(async () => {
  try {
    const info = await api.getAppInfo();
    if (info.version) appVersion.value = info.version;
  } catch {
    // 登录页不因版本信息加载失败而阻塞。
  }
});

async function submit() {
  error.value = "";
  if (!username.value.trim()) {
    error.value = "请输入用户名";
    return;
  }
  if (!idCard.value.trim()) {
    error.value = "请输入证件号";
    return;
  }
  loading.value = true;
  try {
    const user = await api.login(username.value.trim(), idCard.value.trim());
    session.login(user);
    toast.success(`欢迎回来，${user.username}`, "学习数据已加载");
    router.push("/home");
  } catch (e: any) {
    error.value = e?.message || "登录失败，请重试";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="login">
    <section class="visual">
      <div class="visual-top">
        <div class="brand-mark"><AppIcon name="radio" :size="24" :stroke-width="2" /></div>
        <div class="brand-copy">
          <strong>HAM LAB</strong>
          <span>CRAC EXAM SYSTEM</span>
        </div>
      </div>

      <div class="orbit" aria-hidden="true">
        <svg viewBox="0 0 520 520" fill="none">
          <circle cx="260" cy="260" r="70" stroke="rgba(232,163,61,.5)" stroke-width="1.4" />
          <circle cx="260" cy="260" r="130" stroke="rgba(232,163,61,.28)" stroke-width="1.2" />
          <circle cx="260" cy="260" r="192" stroke="rgba(232,163,61,.16)" stroke-width="1" />
          <circle cx="260" cy="260" r="250" stroke="rgba(232,163,61,.08)" stroke-width="1" />
          <circle cx="260" cy="260" r="8" fill="#e8a33d" />
          <circle cx="390" cy="260" r="5" fill="#2a9d8f" />
          <circle cx="150" cy="330" r="4" fill="#e8a33d" opacity=".8" />
          <circle cx="335" cy="140" r="3.5" fill="#f4f0e6" opacity=".7" />
          <path d="M60 330 Q110 290 160 330 T260 330 T360 330 T460 330" stroke="rgba(42,157,143,.4)" stroke-width="1.4" />
          <path d="M100 360 Q140 330 180 360 T260 360 T340 360 T420 360" stroke="rgba(42,157,143,.22)" stroke-width="1.2" />
        </svg>
        <span class="orbit-chip chip-a">A 类</span>
        <span class="orbit-chip chip-b">B 类</span>
        <span class="orbit-chip chip-c">C 类</span>
      </div>

      <div class="visual-copy">
        <h1>听见电波<br /><em>验证实力</em></h1>
        <p>为业余无线电操作能力验证而做的本地模拟考试系统，题库、练习与成绩全部留在你的电脑里。</p>
      </div>

      <div class="visual-foot">
        <span><AppIcon name="shield" :size="14" /> 本地离线</span>
        <span><AppIcon name="zap" :size="14" /> 纯 Go 解析题库</span>
        <span><AppIcon name="target" :size="14" /> A / B / C 全类别</span>
      </div>
    </section>

    <section class="panel">
      <div class="panel-inner">
        <span class="eyebrow">HAM LAB · WELCOME</span>
        <h2>进入学习空间</h2>
        <p class="panel-sub">
          使用本地档案登录。证件号仅用于区分档案，首次登录会自动创建。
        </p>

        <form class="form" @submit.prevent="submit">
          <label class="field">
            <span class="field-label">用户名</span>
            <input
              v-model="username"
              class="input"
              type="text"
              placeholder="例如：BG7XXX"
              autocomplete="username"
              maxlength="32"
            />
          </label>
          <label class="field">
            <span class="field-label">证件号</span>
            <input
              v-model="idCard"
              class="input"
              type="text"
              placeholder="用于区分不同档案"
              autocomplete="off"
              maxlength="32"
            />
          </label>

          <div v-if="error" class="form-error">
            <AppIcon name="warning" :size="14" />
            {{ error }}
          </div>

          <button class="btn btn-accent btn-lg btn-block" type="submit" :disabled="loading">
            <span v-if="loading" class="spinner dark" />
            <template v-else>
              开始学习
              <AppIcon name="arrowRight" :size="17" />
            </template>
          </button>
        </form>

        <div class="divider"><span>本地 · 私有 · 无账号体系</span></div>

        <ul class="notes">
          <li><AppIcon name="check" :size="14" /> 题库与成绩保存在本机数据库</li>
          <li><AppIcon name="check" :size="14" /> 不联网、不上传任何个人信息</li>
          <li><AppIcon name="check" :size="14" /> 支持 ZIP 题库包一键导入</li>
        </ul>
      </div>

      <footer class="panel-foot">
        <span>v{{ appVersion }}</span>
        <span v-if="usingMock" class="mock-tag">浏览器演示模式</span>
        <span>业余无线电模拟考试</span>
      </footer>
    </section>
  </div>
</template>

<style scoped>
.login {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(460px, 0.85fr);
  height: 100vh;
  overflow: hidden;
}

.visual {
  position: relative;
  background: linear-gradient(160deg, #111a22 0%, #0c1218 55%, #101a20 100%);
  color: #f4f0e6;
  padding: 38px 48px 34px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.visual::before {
  content: "";
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 25% 20%, rgba(232, 163, 61, 0.16), transparent 45%),
    radial-gradient(circle at 80% 75%, rgba(42, 157, 143, 0.12), transparent 50%);
}
.visual::after {
  content: "";
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 26px 26px;
  mask-image: radial-gradient(circle at 50% 45%, #000, transparent 75%);
}

.visual-top {
  display: flex;
  align-items: center;
  gap: 13px;
  position: relative;
  z-index: 2;
}
.brand-mark {
  width: 48px;
  height: 48px;
  border-radius: 15px;
  display: grid;
  place-items: center;
  color: #241703;
  background: linear-gradient(135deg, #f4c26b, var(--amber));
  box-shadow: 0 10px 30px rgba(232, 163, 61, 0.32);
}
.brand-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.brand-copy strong {
  font-size: 18px;
  letter-spacing: 0.12em;
}
.brand-copy span {
  font-size: 9.5px;
  letter-spacing: 0.22em;
  color: rgba(244, 240, 230, 0.42);
}

.orbit {
  position: absolute;
  right: -80px;
  top: 50%;
  width: 560px;
  height: 560px;
  transform: translateY(-52%);
  z-index: 1;
  opacity: 0.95;
}
.orbit svg {
  width: 100%;
  height: 100%;
}
.orbit-chip {
  position: absolute;
  padding: 7px 13px;
  border-radius: 99px;
  font-size: 11.5px;
  font-weight: 700;
  letter-spacing: 0.08em;
  background: rgba(16, 24, 32, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(8px);
}
.chip-a { left: 60px; top: 150px; color: #f3c069; }
.chip-b { right: 80px; top: 250px; color: #6fd0c2; }
.chip-c { left: 130px; bottom: 110px; color: #8bb8e4; }

.visual-copy {
  margin-top: auto;
  position: relative;
  z-index: 2;
  max-width: 520px;
}
.visual-copy h1 {
  font-size: clamp(42px, 4.6vw, 66px);
  line-height: 1.04;
  letter-spacing: -0.04em;
  font-weight: 850;
}
.visual-copy h1 em {
  font-style: normal;
  color: var(--amber);
}
.visual-copy p {
  margin-top: 22px;
  font-size: 14px;
  line-height: 1.8;
  color: rgba(244, 240, 230, 0.56);
  max-width: 420px;
}
.visual-foot {
  display: flex;
  gap: 22px;
  margin-top: 38px;
  font-size: 11px;
  color: rgba(244, 240, 230, 0.42);
  position: relative;
  z-index: 2;
  flex-wrap: wrap;
}
.visual-foot span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.panel {
  background: var(--surface);
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 46px 56px 30px;
  position: relative;
}
.panel::before {
  content: "";
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, transparent, var(--line-strong), transparent);
}
.panel-inner {
  width: min(400px, 100%);
  margin: 0 auto;
  animation: rise 0.55s var(--ease) both;
}
.panel-inner h2 {
  margin-top: 14px;
  font-size: 32px;
  font-weight: 850;
  letter-spacing: -0.035em;
}
.panel-sub {
  margin-top: 10px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--ink-3);
}
.form {
  margin-top: 30px;
  display: flex;
  flex-direction: column;
  gap: 17px;
}
.form-error {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  color: #a63b2e;
  background: var(--red-soft);
  padding: 9px 12px;
  border-radius: 9px;
}
.divider {
  margin: 28px 0 18px;
  text-align: center;
  position: relative;
}
.divider::before {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  height: 1px;
  background: var(--line);
}
.divider span {
  position: relative;
  padding: 0 12px;
  background: var(--surface);
  font-size: 10px;
  letter-spacing: 0.14em;
  color: var(--muted);
}
.notes {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.notes li {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--ink-3);
}
.notes :deep(svg) {
  color: var(--teal);
}
.panel-foot {
  margin-top: auto;
  padding-top: 36px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 10.5px;
  color: var(--muted);
}
.mock-tag {
  padding: 3px 8px;
  border-radius: 99px;
  background: var(--amber-soft);
  color: #8a5a10;
  font-weight: 600;
}

@media (max-width: 1080px) {
  .login {
    grid-template-columns: 1fr;
  }
  .visual {
    display: none;
  }
  .panel {
    padding: 32px;
  }
}
</style>
