import { defineStore } from "pinia";
import { reactive, ref } from "vue";
import { api } from "@/api";
import type { PracticeData, PracticeMode } from "@/types";

export const useBankStore = defineStore("bank", () => {
  const cache = reactive(new Map<string, PracticeData>());
  const favoriteIds = ref<Set<number>>(new Set());

  function key(mode: PracticeMode, category: string) {
    return `${mode}:${category}`;
  }

  async function loadPractice(
    userId: number,
    mode: PracticeMode,
    category: string,
    force = false,
  ): Promise<PracticeData> {
    const k = key(mode, category);
    if (!force && cache.has(k)) return cache.get(k)!;
    const data = await api.getPracticeData(userId, category, mode);
    data.questions = data.questions ?? [];
    data.favoriteIds = data.favoriteIds ?? [];
    cache.set(k, data);
    favoriteIds.value = new Set(data.favoriteIds);
    return data;
  }

  function getCached(mode: PracticeMode, category: string) {
    return cache.get(key(mode, category));
  }

  function setFavorite(id: number, on: boolean) {
    const next = new Set(favoriteIds.value);
    if (on) next.add(id);
    else next.delete(id);
    favoriteIds.value = next;
    for (const data of cache.values()) {
      const set = new Set(data.favoriteIds);
      if (on) set.add(id);
      else set.delete(id);
      data.favoriteIds = [...set];
    }
  }

  function updateProgress(mode: PracticeMode, category: string, index: number) {
    const data = cache.get(key(mode, category));
    if (data) data.progress = index;
  }

  function invalidate() {
    cache.clear();
    favoriteIds.value = new Set();
  }

  return {
    cache,
    favoriteIds,
    loadPractice,
    getCached,
    setFavorite,
    updateProgress,
    invalidate,
  };
});
