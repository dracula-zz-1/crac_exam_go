import { defineStore } from "pinia";
import { ref } from "vue";
import type { User } from "@/types";

const STORAGE_KEY = "crac.session.v3";

function restore(): User | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export const useSessionStore = defineStore("session", () => {
  const user = ref<User | null>(restore());

  function login(u: User) {
    user.value = u;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(u));
  }

  function logout() {
    user.value = null;
    localStorage.removeItem(STORAGE_KEY);
  }

  return { user, login, logout };
});
