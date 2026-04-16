import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface User {
  user_id: number
  username: string
  password: string
  last_login: string
}

export const useUserStore = defineStore('user', () => {
  const currentUser = ref<User | null>(null)
  const currentMode = ref<string>('')
  const currentCategory = ref<string>('')

  function setUser(user: User | null) {
    currentUser.value = user
  }

  function setMode(mode: string) {
    currentMode.value = mode
  }

  function setCategory(category: string) {
    currentCategory.value = category
  }

  function logout() {
    currentUser.value = null
    currentMode.value = ''
    currentCategory.value = ''
  }

  return {
    currentUser,
    currentMode,
    currentCategory,
    setUser,
    setMode,
    setCategory,
    logout
  }
})
