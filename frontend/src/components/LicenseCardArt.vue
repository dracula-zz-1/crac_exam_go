<script setup lang="ts">
// 《业余无线电台操作技术能力验证证书》样式示意。
// 全部用 SVG 绘制，任意尺寸都清晰；编号/日期为占位，姓名与证件号取自本地档案。
const props = withDefaults(
  defineProps<{
    name?: string;
    idCard?: string;
    category?: string;
  }>(),
  { name: "示例用户", idCard: "S11010120000101XXXX", category: "A 类" },
);

function maskId(raw?: string) {
  const v = (raw ?? "").trim();
  if (v.length <= 6) return v || "—";
  return `${v.slice(0, 3)}${"*".repeat(Math.min(6, v.length - 6))}${v.slice(-3)}`;
}
</script>

<template>
  <svg
    class="lic"
    viewBox="0 0 900 580"
    role="img"
    aria-label="业余无线电台操作技术能力验证证书样式示意"
  >
    <defs>
      <linearGradient id="lic-paper" x1="0" y1="0" x2="0.4" y2="1">
        <stop offset="0%" stop-color="#f5faff" />
        <stop offset="45%" stop-color="#eaf3fb" />
        <stop offset="100%" stop-color="#e2eefa" />
      </linearGradient>
      <linearGradient id="lic-photo-bg" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#e6eff8" />
        <stop offset="100%" stop-color="#cfdfef" />
      </linearGradient>
      <!-- 底纹：细密斜线 + 菱形网，模仿证书的防伪花纹 -->
      <pattern
        id="lic-guilloche"
        width="14"
        height="14"
        patternUnits="userSpaceOnUse"
        patternTransform="rotate(35)"
      >
        <path d="M0 0h14" stroke="#c3d9ec" stroke-width="0.5" opacity="0.75" />
        <path d="M0 7h14" stroke="#c3d9ec" stroke-width="0.35" opacity="0.45" />
        <circle cx="7" cy="7" r="2.4" fill="none" stroke="#c3d9ec" stroke-width="0.35" opacity="0.5" />
      </pattern>
      <clipPath id="lic-photo-clip">
        <rect x="648" y="196" width="156" height="196" rx="6" />
      </clipPath>
    </defs>

    <!-- 纸张 + 防伪底纹 -->
    <rect width="900" height="580" rx="10" fill="url(#lic-paper)" />
    <rect x="12" y="12" width="876" height="556" rx="6" fill="url(#lic-guilloche)" opacity="0.55" />

    <!-- 边框：外粗内细的双线 -->
    <rect
      x="16"
      y="16"
      width="868"
      height="548"
      rx="5"
      fill="none"
      stroke="#2e6cb3"
      stroke-width="2.4"
    />
    <rect
      x="26"
      y="26"
      width="848"
      height="528"
      rx="3"
      fill="none"
      stroke="#6f9fcd"
      stroke-width="0.9"
    />
    <g stroke="#2e6cb3" stroke-width="1.8" opacity="0.9">
      <path d="M26 46h34M46 26v34" />
      <path d="M874 46h-34M854 26v34" />
      <path d="M26 534h34M46 554v-34" />
      <path d="M874 534h-34M854 554v-34" />
    </g>

    <!-- 徽标 -->
    <g transform="translate(78 78)" fill="none" stroke="#2e6cb3">
      <circle r="32" stroke-width="1.6" />
      <circle r="25" stroke-width="0.8" opacity="0.9" />
      <circle r="3.4" fill="#2e6cb3" stroke="none" />
      <g stroke-width="1.3">
        <path d="M0 -21 5 -6H20L8 3l5 15L0 9l-13 9 5-15L-20 -6H-5z" opacity="0.85" />
      </g>
      <g stroke-width="0.7" opacity="0.7">
        <path d="M-30 16c6-6 12-6 18 0" />
        <path d="M12 16c6-6 12-6 18 0" />
      </g>
    </g>

    <!-- 标题 -->
    <text x="466" y="90" text-anchor="middle" class="lic-title">
      业余无线电台操作技术能力验证证书
    </text>
    <rect x="322" y="100" width="288" height="22" rx="4" fill="#dbeaf7" opacity="0.85" />
    <text x="466" y="116" text-anchor="middle" class="lic-sub">
      Certificate of Amateur Radio Operation
    </text>

    <!-- 编号 -->
    <text x="476" y="164" class="lic-key">编　号</text>
    <text x="556" y="165" class="lic-no">B1100123456</text>
    <text x="556" y="180" class="lic-en">Certificate No.</text>

    <!-- 字段 -->
    <g>
      <text x="66" y="228" class="lic-key">姓　　名</text>
      <text x="66" y="240" class="lic-en">Name</text>
      <text x="186" y="228" class="lic-val">{{ props.name }}</text>
      <path d="M186 236h404" class="lic-rule" />

      <text x="66" y="286" class="lic-key">证件号码</text>
      <text x="66" y="298" class="lic-en">ID No.</text>
      <text x="186" y="286" class="lic-val mono">{{ maskId(props.idCard) }}</text>
      <path d="M186 294h404" class="lic-rule" />

      <text x="66" y="344" class="lic-key">操作类别</text>
      <text x="66" y="356" class="lic-en">Certificate Class</text>
      <text x="186" y="344" class="lic-val strong">{{ props.category }}</text>
      <path d="M186 352h404" class="lic-rule" />

      <text x="66" y="402" class="lic-key">发证日期</text>
      <text x="66" y="414" class="lic-en">Date of Issue</text>
      <text x="186" y="402" class="lic-val mono">20XX 年 XX 月 XX 日</text>
      <path d="M186 410h404" class="lic-rule" />

      <text x="66" y="460" class="lic-key">核发机构</text>
      <text x="66" y="472" class="lic-en">Authority</text>
      <text x="186" y="460" class="lic-val">XX 省工业和信息化厅</text>
      <path d="M186 468h404" class="lic-rule" />
    </g>

    <!-- 证件照 -->
    <rect x="648" y="196" width="156" height="196" rx="6" fill="url(#lic-photo-bg)" />
    <g clip-path="url(#lic-photo-clip)">
      <circle cx="726" cy="266" r="34" fill="#a9c2d8" />
      <path d="M648 392c0-38 33-62 78-62s78 24 78 62z" fill="#a9c2d8" />
    </g>
    <rect
      x="648"
      y="196"
      width="156"
      height="196"
      rx="6"
      fill="none"
      stroke="#2e6cb3"
      stroke-width="1.6"
    />
    <rect
      x="653"
      y="201"
      width="146"
      height="186"
      rx="4"
      fill="none"
      stroke="#8fb4d6"
      stroke-width="0.7"
    />

    <!-- 页脚 -->
    <path d="M66 500h510" stroke="#a8c4dc" stroke-width="0.8" stroke-dasharray="3 5" />
    <text x="66" y="522" class="lic-foot">
      原操作证号 A000012345　　原发证日期 20XX 年 XX 月 XX 日
    </text>
    <text x="66" y="538" class="lic-en">
      （如无该部分内容，可在相应位置填写“无”）
    </text>
    <text x="854" y="540" text-anchor="end" class="lic-issue">
      中华人民共和国工业和信息化部监制
    </text>
  </svg>
</template>

<style scoped>
.lic {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 10px;
  box-shadow: 0 10px 26px rgba(22, 27, 33, 0.14);
}
.lic-title {
  font-family: var(--font-sans);
  font-size: 35px;
  font-weight: 800;
  letter-spacing: 0.045em;
  fill: #14487f;
}
.lic-sub {
  font-family: var(--font-sans);
  font-size: 12.5px;
  font-weight: 600;
  letter-spacing: 0.07em;
  fill: #2f6fb0;
}
.lic-key {
  font-family: var(--font-sans);
  font-size: 15px;
  font-weight: 700;
  fill: #24486b;
  letter-spacing: 0.02em;
}
.lic-en {
  font-family: var(--font-sans);
  font-size: 8.5px;
  fill: #7f97ab;
}
.lic-val {
  font-family: var(--font-sans);
  font-size: 16px;
  font-weight: 600;
  fill: #10243a;
}
.lic-val.strong {
  font-size: 18px;
  font-weight: 800;
  fill: #14487f;
}
.lic-val.mono {
  font-family: var(--font-mono);
  font-size: 15.5px;
  letter-spacing: 0.03em;
}
.lic-no {
  font-family: var(--font-mono);
  font-size: 20px;
  font-weight: 700;
  fill: #14487f;
  letter-spacing: 0.05em;
}
.lic-rule {
  stroke: #9dbdd6;
  stroke-width: 0.9;
  stroke-dasharray: 2 4;
}
.lic-foot {
  font-family: var(--font-sans);
  font-size: 9.5px;
  fill: #6d8ba6;
}
.lic-issue {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  fill: #2b5c8f;
}
</style>
