import { defineStore } from "pinia";
import { ref } from "vue";

export type ToastTone = "success" | "error" | "info" | "warning";

export interface ToastItem {
  id: number;
  tone: ToastTone;
  title: string;
  message?: string;
}

let seq = 0;

export const useToastStore = defineStore("toast", () => {
  const items = ref<ToastItem[]>([]);

  function push(tone: ToastTone, title: string, message?: string, ttl = 3600) {
    const id = ++seq;
    items.value.push({ id, tone, title, message });
    window.setTimeout(() => dismiss(id), ttl);
  }

  function dismiss(id: number) {
    items.value = items.value.filter((t) => t.id !== id);
  }

  const success = (title: string, message?: string) => push("success", title, message);
  const error = (title: string, message?: string) => push("error", title, message, 4600);
  const info = (title: string, message?: string) => push("info", title, message);
  const warning = (title: string, message?: string) => push("warning", title, message, 4200);

  return { items, push, dismiss, success, error, info, warning };
});
