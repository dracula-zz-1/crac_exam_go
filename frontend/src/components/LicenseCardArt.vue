<script setup lang="ts">
// 业余无线电台操作技术能力验证证书的示意图。
// 全部用 SVG 绘制，任意尺寸都清晰；数据为示例，不指向真实个人。
withDefaults(defineProps<{ sample?: boolean }>(), { sample: true });
</script>

<template>
  <svg
    class="lic"
    viewBox="0 0 720 456"
    role="img"
    aria-label="业余无线电台操作技术能力验证证书示例"
  >
    <defs>
      <linearGradient id="lic-paper" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#fffdf7" />
        <stop offset="55%" stop-color="#fdf7e9" />
        <stop offset="100%" stop-color="#f7efdd" />
      </linearGradient>
      <linearGradient id="lic-frame" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#c8a04a" />
        <stop offset="50%" stop-color="#e2c076" />
        <stop offset="100%" stop-color="#b98f38" />
      </linearGradient>
      <pattern id="lic-guilloche" width="44" height="44" patternUnits="userSpaceOnUse">
        <g fill="none" stroke="#d9c9a3" stroke-width="0.7" opacity="0.55">
          <circle cx="22" cy="22" r="17" />
          <circle cx="22" cy="22" r="11.5" />
          <circle cx="22" cy="22" r="6" />
          <path d="M0 22h44M22 0v44" stroke-dasharray="2 6" />
        </g>
      </pattern>
      <clipPath id="lic-photo">
        <rect x="556" y="168" width="110" height="146" rx="4" />
      </clipPath>
    </defs>

    <!-- 纸张与花纹 -->
    <rect width="720" height="456" rx="14" fill="url(#lic-paper)" />
    <rect x="6" y="6" width="708" height="444" rx="10" fill="url(#lic-guilloche)" opacity="0.7" />
    <rect
      x="14"
      y="14"
      width="692"
      height="428"
      rx="8"
      fill="none"
      stroke="url(#lic-frame)"
      stroke-width="3"
    />
    <rect
      x="22"
      y="22"
      width="676"
      height="412"
      rx="6"
      fill="none"
      stroke="#c8a04a"
      stroke-width="0.9"
      opacity="0.75"
    />

    <!-- 抬头 -->
    <g transform="translate(52 44)">
      <circle r="20" cx="20" cy="20" fill="none" stroke="#b98f38" stroke-width="1.6" />
      <circle r="13" cx="20" cy="20" fill="none" stroke="#b98f38" stroke-width="1" />
      <path
        d="M20 8v24M8 20h24M11 11l18 18M29 11 11 29"
        stroke="#b98f38"
        stroke-width="0.8"
        opacity="0.75"
      />
    </g>

    <text x="360" y="62" text-anchor="middle" class="lic-title">
      业余无线电台操作技术能力验证证书
    </text>
    <text x="360" y="82" text-anchor="middle" class="lic-sub">
      Certificate of Amateur Radio Operation
    </text>
    <path d="M212 94h296" stroke="#c8a04a" stroke-width="1.2" opacity="0.85" />

    <!-- 编号 -->
    <text x="470" y="122" class="lic-key">编　号</text>
    <text x="536" y="122" class="lic-no">B1100123456</text>

    <!-- 字段 -->
    <g class="lic-fields">
      <text x="56" y="160" class="lic-key">姓　　名</text>
      <text x="150" y="160" class="lic-val">示例用户</text>
      <text x="56" y="166" class="lic-en">Name</text>

      <text x="56" y="204" class="lic-key">证件号码</text>
      <text x="150" y="204" class="lic-val mono">S1101012000010XXXX</text>
      <text x="56" y="210" class="lic-en">ID No.</text>

      <text x="56" y="248" class="lic-key">操作类别</text>
      <text x="150" y="248" class="lic-val strong">A 类</text>
      <text x="56" y="254" class="lic-en">Certificate Class</text>

      <text x="56" y="292" class="lic-key">发证日期</text>
      <text x="150" y="292" class="lic-val mono">20XX 年 XX 月 XX 日</text>
      <text x="56" y="298" class="lic-en">Date of Issue</text>

      <text x="56" y="336" class="lic-key">核发机构</text>
      <text x="150" y="336" class="lic-val">XX 省工业和信息化厅</text>
      <text x="56" y="342" class="lic-en">Authority</text>
    </g>

    <!-- 证件照 -->
    <rect x="556" y="168" width="110" height="146" rx="4" fill="#e8dfc9" />
    <g clip-path="url(#lic-photo)">
      <circle cx="611" cy="222" r="26" fill="#cbbfa2" />
      <path d="M556 320c0-30 25-48 55-48s55 18 55 48z" fill="#cbbfa2" />
    </g>
    <rect
      x="556"
      y="168"
      width="110"
      height="146"
      rx="4"
      fill="none"
      stroke="#c8a04a"
      stroke-width="1.2"
    />
    <text x="611" y="330" text-anchor="middle" class="lic-en">照片</text>

    <!-- 页脚 -->
    <path d="M56 372h456" stroke="#d9c9a3" stroke-width="0.9" stroke-dasharray="4 5" />
    <text x="56" y="394" class="lic-en">
      原操作证号 A000012345　原发证日期 20XX 年 XX 月 XX 日
    </text>
    <text x="56" y="414" class="lic-en">（如无该部分内容，可在相应位置填写“无”）</text>
    <text x="612" y="414" text-anchor="end" class="lic-en">中华人民共和国工业和信息化部监制</text>

    <!-- 水印 -->
    <text
      v-if="sample"
      x="360"
      y="262"
      text-anchor="middle"
      class="lic-watermark"
      transform="rotate(-16 360 262)"
    >
      示例
    </text>
  </svg>
</template>

<style scoped>
.lic {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 12px;
  box-shadow: 0 10px 26px rgba(22, 27, 33, 0.12);
}
.lic-title {
  font-family: var(--font-sans);
  font-size: 27px;
  font-weight: 800;
  letter-spacing: 0.06em;
  fill: #1d4f74;
}
.lic-sub {
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  fill: #6b7c8c;
}
.lic-key {
  font-family: var(--font-sans);
  font-size: 14px;
  font-weight: 700;
  fill: #24343f;
}
.lic-val {
  font-family: var(--font-sans);
  font-size: 15px;
  fill: #1d2b34;
}
.lic-val.strong {
  font-size: 17px;
  font-weight: 800;
  fill: #1d4f74;
}
.lic-val.mono {
  font-family: var(--font-mono);
  font-size: 14px;
  letter-spacing: 0.02em;
}
.lic-en {
  font-family: var(--font-sans);
  font-size: 9px;
  fill: #8a97a1;
}
.lic-no {
  font-family: var(--font-mono);
  font-size: 15px;
  font-weight: 700;
  fill: #a3502f;
  letter-spacing: 0.06em;
}
.lic-watermark {
  font-family: var(--font-sans);
  font-size: 96px;
  font-weight: 900;
  fill: rgba(232, 163, 61, 0.16);
  letter-spacing: 0.12em;
  pointer-events: none;
}
</style>
