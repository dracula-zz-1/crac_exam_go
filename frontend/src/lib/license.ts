import type { Category } from "@/types";

/**
 * 业余无线电操作等级，由低到高：A 类 → B 类 → C 类。
 * "none" 表示还没有操作证（准备考 A 类）。
 */
export type LicenseLevel = "none" | "A" | "B" | "C";

export const LICENSE_LEVELS: Array<{
  value: LicenseLevel;
  label: string;
  desc: string;
}> = [
  { value: "none", label: "还没有操作证", desc: "准备考取 A 类操作证" },
  { value: "A", label: "A 类", desc: "已持有 A 类操作证" },
  { value: "B", label: "B 类", desc: "已持有 B 类操作证" },
  { value: "C", label: "C 类", desc: "已持有 C 类操作证" },
];

/** 归一化后端返回的等级字段。 */
export function normalizeLevel(level?: string | null): LicenseLevel {
  const v = (level ?? "").trim().toUpperCase();
  if (v === "A" || v === "B" || v === "C") return v;
  return "none";
}

/** 当前等级对应的“即将应考”等级：无证→A，A→B，B→C，C→C。 */
export function nextLevelCategory(level?: string | null): Category {
  switch (normalizeLevel(level)) {
    case "A":
      return "B";
    case "B":
    case "C":
      return "C";
    default:
      return "A";
  }
}

/** 等级的中文说明，用于界面展示。 */
export function levelLabel(level?: string | null): string {
  const v = normalizeLevel(level);
  return v === "none" ? "尚未持证" : `${v} 类操作证`;
}
