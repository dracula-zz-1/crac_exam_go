import templateUrl from "@/assets/cert-template.png";
import type { User } from "@/types";
import { normalizeLevel } from "@/lib/license";

/** 实拍样证（已去掉照片白边并做圆角），所有坐标都按它的尺寸换算。 */
export const CERT_TEMPLATE_URL = templateUrl;
export const CERT_TEMPLATE_W = 1013;
export const CERT_TEMPLATE_H = 640;

export interface CertificateFields {
  number: string;
  name: string;
  idCard: string;
  category: string; // "A 类" / "B 类" / "C 类"
  year: string;
  month: string;
  day: string;
  org: string; // 核发机构省份（后面“省工业和信息化厅”是证上印好的）
  oldNo: string;
  oldDate: string;
}

interface Region {
  x: number;
  y: number;
  w: number;
  h: number;
  align: "left" | "center";
  /** 字号上限，占模板宽度的比例 */
  size: number;
  serif?: boolean;
  /** 左内边距，占模板宽度比例 */
  pad?: number;
}

/**
 * 空白区域（照片上原本被白块盖住的字段），坐标取自对实拍图的像素测量。
 * 另外“操作类别”在样证上是印好的 A 类，导出时也要盖掉重写。
 */
export const CERT_REGIONS: Record<keyof CertificateFields, Region> = {
  number: { x: 0.4413, y: 0.225, w: 0.2053, h: 0.0516, align: "left", size: 0.0345, serif: true, pad: 0.004 },
  name: { x: 0.2399, y: 0.3266, w: 0.1481, h: 0.0594, align: "left", size: 0.0355, pad: 0.004 },
  idCard: { x: 0.2577, y: 0.4328, w: 0.311, h: 0.0453, align: "left", size: 0.034, serif: true, pad: 0.004 },
  category: { x: 0.2369, y: 0.5281, w: 0.077, h: 0.0672, align: "center", size: 0.036 },
  year: { x: 0.2399, y: 0.6516, w: 0.0711, h: 0.0469, align: "center", size: 0.0295, serif: true },
  month: { x: 0.3623, y: 0.65, w: 0.0346, h: 0.0469, align: "center", size: 0.0295, serif: true },
  day: { x: 0.4482, y: 0.6516, w: 0.0375, h: 0.0422, align: "center", size: 0.0295, serif: true },
  org: { x: 0.2428, y: 0.7656, w: 0.0671, h: 0.0516, align: "center", size: 0.0355 },
  oldNo: { x: 0.2201, y: 0.8719, w: 0.1046, h: 0.0281, align: "left", size: 0.0225, serif: true, pad: 0.003 },
  oldDate: { x: 0.4482, y: 0.8703, w: 0.1688, h: 0.0312, align: "left", size: 0.0225, serif: true, pad: 0.003 },
};

/** 证件照区域（样证上是一块空白）。 */
export const CERT_PHOTO_REGION = { x: 0.7147, y: 0.3469, w: 0.1905, h: 0.4828 };

export const CERT_FONT_SANS = '"Microsoft YaHei UI", "Microsoft YaHei", "SimHei", sans-serif';
export const CERT_FONT_SERIF = '"Times New Roman", "SimSun", "Songti SC", serif';

/** 表单字段中文名，界面上直接用。 */
export const CERT_FIELD_LABELS: Record<keyof CertificateFields, string> = {
  number: "证书编号",
  name: "姓名",
  idCard: "证件号码",
  category: "操作类别",
  year: "发证年",
  month: "月",
  day: "日",
  org: "核发机构（省份）",
  oldNo: "原操作证号",
  oldDate: "原发证日期",
};

/** 依据本地档案生成一份初始内容，用户在此基础上补全。 */
export function defaultCertificateFields(user: User | null, level: string): CertificateFields {
  const cls = normalizeLevel(level || user?.level);
  const cat = cls === "none" ? "A" : cls;
  const now = new Date();
  return {
    number: `${cat}1100123456`,
    name: user?.username ?? "",
    idCard: user?.idCard ?? "",
    category: `${cat} 类`,
    year: String(now.getFullYear()),
    month: String(now.getMonth() + 1),
    day: String(now.getDate()),
    org: "",
    oldNo: "",
    oldDate: "",
  };
}

function pickFont(ctx: CanvasRenderingContext2D, size: number, serif?: boolean) {
  ctx.font = `700 ${size}px ${serif ? CERT_FONT_SERIF : CERT_FONT_SANS}`;
}

/** 在给定宽度内自适应字号（中文名字、长身份证号、三字省份都能放下）。 */
function fitSize(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  maxSize: number,
  serif?: boolean,
) {
  let size = maxSize;
  pickFont(ctx, size, serif);
  while (size > 8 && ctx.measureText(text).width > maxWidth) {
    size -= 1;
    pickFont(ctx, size, serif);
  }
  return size;
}

export interface DrawOptions {
  template: CanvasImageSource;
  fields: CertificateFields;
  photo?: CanvasImageSource | null;
  /** 画布放大倍数，导出用 2，预览用 1.5 */
  scale?: number;
}

/** 把证书画到 canvas 上：样证底图 + 照片 + 各字段文字。 */
export function drawCertificate(canvas: HTMLCanvasElement, opts: DrawOptions) {
  const scale = opts.scale ?? 2;
  const W = Math.round(CERT_TEMPLATE_W * scale);
  const H = Math.round(CERT_TEMPLATE_H * scale);
  canvas.width = W;
  canvas.height = H;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.clearRect(0, 0, W, H);
  // 底色用白色，圆角处不会出现透明
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, W, H);
  ctx.drawImage(opts.template, 0, 0, W, H);

  // 证件照：等比铺满照片框
  const pr = CERT_PHOTO_REGION;
  const px = pr.x * W;
  const py = pr.y * H;
  const pw = pr.w * W;
  const ph = pr.h * H;
  if (opts.photo) {
    const img = opts.photo as HTMLImageElement;
    const iw = img.naturalWidth || img.width;
    const ih = img.naturalHeight || img.height;
    if (iw > 0 && ih > 0) {
      const ratio = Math.max(pw / iw, ph / ih);
      const dw = iw * ratio;
      const dh = ih * ratio;
      ctx.save();
      ctx.beginPath();
      ctx.rect(px, py, pw, ph);
      ctx.clip();
      ctx.drawImage(img, px + (pw - dw) / 2, py + (ph - dh) / 2, dw, dh);
      ctx.restore();
    }
  }

  // 字段
  ctx.textBaseline = "middle";
  ctx.fillStyle = "#0d0d0d";
  (Object.keys(CERT_REGIONS) as Array<keyof CertificateFields>).forEach((key) => {
    const r = CERT_REGIONS[key];
    const text = (opts.fields[key] ?? "").toString().trim();
    if (!text) return;
    const boxX = r.x * W;
    const boxY = r.y * H;
    const boxW = r.w * W;
    const boxH = r.h * H;
    const pad = (r.pad ?? 0) * W;
    // 操作类别与姓名等印在证上的字段没有白块，需要先盖一块底
    if (key === "category") {
      ctx.save();
      ctx.fillStyle = "#ffffff";
      roundRect(ctx, boxX - 4 * scale, boxY - 4 * scale, boxW + 8 * scale, boxH + 8 * scale, 3 * scale);
      ctx.fill();
      ctx.restore();
    }
    const size = fitSize(ctx, text, boxW - pad * 2, r.size * W, r.serif);
    ctx.fillStyle = "#0d0d0d";
    if (r.align === "center") {
      ctx.textAlign = "center";
      ctx.fillText(text, boxX + boxW / 2, boxY + boxH / 2 + size * 0.04);
    } else {
      ctx.textAlign = "left";
      ctx.fillText(text, boxX + pad, boxY + boxH / 2 + size * 0.04);
    }
  });
}

function roundRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number,
) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}

/** 预加载样证底图。 */
export function loadTemplate(): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("样证底图加载失败"));
    img.src = CERT_TEMPLATE_URL;
  });
}

/** 把画布内容导出成 PNG dataURL。 */
export function canvasToDataURL(canvas: HTMLCanvasElement) {
  return canvas.toDataURL("image/png");
}
