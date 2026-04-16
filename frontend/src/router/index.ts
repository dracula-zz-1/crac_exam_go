import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Login',
    component: () => import('@/views/LoginView.vue')
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/HomeView.vue')
  },
  {
    path: '/practice',
    name: 'Practice',
    component: () => import('@/views/PracticeView.vue')
  },
  {
    path: '/exam',
    name: 'Exam',
    component: () => import('@/views/ExamView.vue')
  },
  {
    path: '/error-book',
    name: 'ErrorBook',
    component: () => import('@/views/ErrorBookView.vue')
  },
  {
    path: '/favorite',
    name: 'Favorite',
    component: () => import('@/views/FavoriteView.vue')
  },
  {
    path: '/question-bank',
    name: 'QuestionBank',
    component: () => import('@/views/QuestionBankView.vue')
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue')
  },
  {
    path: '/exam-result',
    name: 'ExamResult',
    component: () => import('@/views/ExamResultView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
