<template>
  <div class="home-container">
    <!-- 标题 -->
    <div class="header">
      <h1 class="title">业余无线电模拟考试系统</h1>
      
      <!-- 右上角设置按钮 -->
      <div class="settings-btn" @click="showSettings = true">
        <el-icon :size="30"><Setting /></el-icon>
      </div>
    </div>

    <!-- 类别框架容器 -->
    <div class="categories-container">
      <!-- A 类题目 -->
      <div class="category-frame">
        <div class="category-title">A 类题目</div>
        <div class="mode-buttons">
          <div class="mode-btn" @click="startMode('practice', 'A')">
            <el-icon :size="40"><Reading /></el-icon>
            <span>逐题练习</span>
          </div>
          <div class="mode-btn" @click="startMode('exam', 'A')">
            <el-icon :size="40"><Document /></el-icon>
            <span>模拟考试</span>
          </div>
          <div class="mode-btn" @click="startMode('error', 'A')">
            <el-icon :size="40"><DocumentDelete /></el-icon>
            <span>只做错题</span>
          </div>
          <div class="mode-btn" @click="startMode('favorite', 'A')">
            <el-icon :size="40"><Star /></el-icon>
            <span>只做收藏</span>
          </div>
        </div>
      </div>

      <!-- B 类题目 -->
      <div class="category-frame">
        <div class="category-title">B 类题目</div>
        <div class="mode-buttons">
          <div class="mode-btn" @click="startMode('practice', 'B')">
            <el-icon :size="40"><Reading /></el-icon>
            <span>逐题练习</span>
          </div>
          <div class="mode-btn" @click="startMode('exam', 'B')">
            <el-icon :size="40"><Document /></el-icon>
            <span>模拟考试</span>
          </div>
          <div class="mode-btn" @click="startMode('error', 'B')">
            <el-icon :size="40"><DocumentDelete /></el-icon>
            <span>只做错题</span>
          </div>
          <div class="mode-btn" @click="startMode('favorite', 'B')">
            <el-icon :size="40"><Star /></el-icon>
            <span>只做收藏</span>
          </div>
        </div>
      </div>

      <!-- C 类题目 -->
      <div class="category-frame">
        <div class="category-title">C 类题目</div>
        <div class="mode-buttons">
          <div class="mode-btn" @click="startMode('practice', 'C')">
            <el-icon :size="40"><Reading /></el-icon>
            <span>逐题练习</span>
          </div>
          <div class="mode-btn" @click="startMode('exam', 'C')">
            <el-icon :size="40"><Document /></el-icon>
            <span>模拟考试</span>
          </div>
          <div class="mode-btn" @click="startMode('error', 'C')">
            <el-icon :size="40"><DocumentDelete /></el-icon>
            <span>只做错题</span>
          </div>
          <div class="mode-btn" @click="startMode('favorite', 'C')">
            <el-icon :size="40"><Star /></el-icon>
            <span>只做收藏</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户信息区域 -->
    <div class="user-section" v-if="userStore.currentUser">
      <div class="user-info-frame">
        <div class="frame-header">用户信息</div>
        <div class="frame-content">
          <div class="info-column">
            <div class="info-row">
              <span class="info-label">用户名：</span>
              <span class="info-value">{{ userStore.currentUser.username }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">上次登录：</span>
              <span class="info-value">{{ formattedLastLogin }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">最近考试：</span>
              <span class="info-value">
                <template v-if="statistics.lastExamDate">
                  {{ statistics.lastExamCategory }}类 {{ statistics.lastExamScore }}分 ({{ statistics.lastExamDate }})
                </template>
                <template v-else>
                  暂无考试记录
                </template>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">最佳成绩：</span>
              <div class="best-scores-inline">
                <span class="score-item">A 类：{{ statistics.bestScoreA !== null ? statistics.bestScoreA + '分' : '未参加' }}</span>
                <span class="score-item">B 类：{{ statistics.bestScoreB !== null ? statistics.bestScoreB + '分' : '未参加' }}</span>
                <span class="score-item">C 类：{{ statistics.bestScoreC !== null ? statistics.bestScoreC + '分' : '未参加' }}</span>
              </div>
            </div>
          </div>
          
          <div class="logout-frame">
            <div class="logout-btn" @click="handleLogout">
              <el-icon :size="30"><SwitchButton /></el-icon>
              <span>退出登录</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部版权信息 -->
    <div class="footer">
      <p>Copyright © 2024-2025 BA4RHH. All rights reserved.</p>
    </div>

    <!-- 设置对话框 -->
    <el-dialog
      v-model="showSettings"
      title="系统设置"
      width="900"
      :close-on-click-modal="false"
      top="5px"
      :modal-append-to-body="false"
      destroy-on-close
      --el-dialog-padding-primary: 15px;
    >
      <div style="height: 82vh; overflow: hidden;">
        <SettingsView @close="showSettings = false" />
      </div>
    </el-dialog>

    <!-- 练习视图 -->
    <div v-if="currentView === 'practice'" class="fullscreen-view">
      <PracticeView 
        :category="currentCategory" 
        :mode="currentMode"
        @back="currentView = 'home'"
      />
    </div>

    <!-- 考试视图 -->
    <div v-if="currentView === 'exam'" class="fullscreen-view">
      <ExamView 
        :category="currentCategory"
        @back="currentView = 'home'"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { StatisticsService } from '@/api'
import { Reading, Document, DocumentDelete, Star, Setting, SwitchButton } from '@element-plus/icons-vue'
import SettingsView from './SettingsView.vue'
import PracticeView from './PracticeView.vue'
import ExamView from './ExamView.vue'

const router = useRouter()
const userStore = useUserStore()

// 当前视图状态
const currentView = ref<'home' | 'practice' | 'exam'>('home')
const currentMode = ref<'practice' | 'error' | 'favorite'>('practice')
const currentCategory = ref<'A' | 'B' | 'C'>('A')

// 设置对话框
const showSettings = ref(false)

const statistics = ref({
  practiceAccuracy: 0,
  errorCount: 0,
  favoriteCount: 0,
  lastExamDate: '',
  lastExamCategory: '',
  lastExamScore: 0,
  bestScoreA: null as number | null,
  bestScoreB: null as number | null,
  bestScoreC: null as number | null
})

const formattedLastLogin = computed(() => {
  if (!userStore.currentUser?.last_login) return ''
  const date = new Date(userStore.currentUser.last_login)
  return date.toLocaleString('zh-CN')
})

const loadStatistics = async () => {
  if (!userStore.currentUser) return
  
  try {
    const [accuracy, errorCount, favoriteCount, examData] = await Promise.all([
      StatisticsService.GetPracticeAccuracy(userStore.currentUser.user_id),
      StatisticsService.GetErrorCount(userStore.currentUser.user_id),
      StatisticsService.GetFavoriteCount(userStore.currentUser.user_id),
      StatisticsService.GetExamData(userStore.currentUser.user_id, '', '180days')
    ])
    
    // 计算最近考试和最佳成绩
    let lastExamDate = ''
    let lastExamCategory = ''
    let lastExamScore = 0
    let bestScoreA: number | null = null
    let bestScoreB: number | null = null
    let bestScoreC: number | null = null
    
    if (examData && examData.length > 0) {
      // 最近一次考试
      const latestExam = examData[examData.length - 1]
      lastExamDate = new Date(latestExam.exam_date as any).toLocaleString('zh-CN')
      lastExamCategory = latestExam.category
      lastExamScore = latestExam.score
      
      // 计算各类别最佳成绩
      const categoryData: Record<string, typeof examData> = { A: [], B: [], C: [] }
      examData.forEach(exam => {
        if (exam.category in categoryData) {
          categoryData[exam.category].push(exam)
        }
      })
      
      for (const cat of ['A', 'B', 'C']) {
        if (categoryData[cat].length > 0) {
          const maxScore = Math.max(...categoryData[cat].map(e => e.score))
          if (cat === 'A') bestScoreA = maxScore
          else if (cat === 'B') bestScoreB = maxScore
          else if (cat === 'C') bestScoreC = maxScore
        }
      }
    }
    
    statistics.value.practiceAccuracy = accuracy || 0
    statistics.value.errorCount = errorCount || 0
    statistics.value.favoriteCount = favoriteCount || 0
    statistics.value.lastExamDate = lastExamDate
    statistics.value.lastExamCategory = lastExamCategory
    statistics.value.lastExamScore = lastExamScore
    statistics.value.bestScoreA = bestScoreA
    statistics.value.bestScoreB = bestScoreB
    statistics.value.bestScoreC = bestScoreC
  } catch (error) {
    console.error('Load statistics error:', error)
  }
}

onMounted(() => {
  loadStatistics()
})

// 当从缓存中激活组件时，刷新统计数据
onActivated(() => {
  loadStatistics()
})

const startMode = (mode: string, category: string) => {
  userStore.setMode(mode)
  userStore.setCategory(category)
  currentMode.value = mode as 'practice' | 'error' | 'favorite'
  currentCategory.value = category as 'A' | 'B' | 'C'
  
  switch (mode) {
    case 'practice':
      currentView.value = 'practice'
      break
    case 'exam':
      currentView.value = 'exam'
      break
    case 'error':
      currentView.value = 'practice'
      break
    case 'favorite':
      currentView.value = 'practice'
      break
  }
}

// 从其他页面返回首页时调用，刷新统计数据
const refreshHomeData = () => {
  loadStatistics()
}

// 暴露方法给父组件调用
defineExpose({
  refreshHomeData
})

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    userStore.logout()
    // 使用 '/' 而不是 '/login'，因为路由配置中 '/' 对应 LoginView
    router.push('/')
  } catch {
    // 取消退出
  }
}
</script>

<style scoped>
.home-container {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  justify-content: space-between;
}

.header {
  text-align: center;
  position: relative;
  padding: 30px 0 20px 0;
  flex-shrink: 0;
}

.title {
  font-size: 36px;
  color: #1a73e8;
  margin: 0;
  font-weight: 600;
}

.settings-btn {
  position: absolute;
  top: 25px;
  right: 110px;
  cursor: pointer;
  color: #666;
  transition: color 0.3s;
}

.settings-btn:hover {
  color: #409EFF;
}

.categories-container {
  display: flex;
  justify-content: center;
  gap: 15px;
  padding: 0;
  margin-bottom: 10px;
  flex-shrink: 0;
  width: 100%;
  max-width: 810px;
  margin-left: auto;
  margin-right: auto;
}

.category-frame {
  background: white;
  border-radius: 6px;
  padding: 10px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  width: 250px;
  height: 300px;
}

.category-title {
  text-align: center;
  font-size: 16px;
  color: #333;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.mode-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
  flex: 1;
  justify-content: flex-start;
}

.mode-btn {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 15px;
  cursor: pointer;
  padding: 10px 18px;
  border-radius: 6px;
  transition: all 0.3s;
  background: #f0f9eb;
  color: #67c23a;
  width: 160px;
  margin: 0 auto;
}

.mode-btn:hover {
  background: #67c23a;
  color: white;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.3);
}

.mode-btn span {
  font-size: 16px;
  font-weight: 500;
}

/* 用户信息区域容器 */
.user-section {
  padding: 0 0 10px 0;
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  width: auto;
  min-width: 100%;
}

/* LabelFrame 样式框架 */
.user-info-frame {
  background: white;
  border: 2px solid #dcdfe6;
  border-radius: 6px;
  overflow: hidden;
  width: 815px;
  margin: 0 auto;
  box-sizing: border-box;
}

/* LabelFrame 标题栏 */
.frame-header {
  background: linear-gradient(to right, #f0f9eb, #e8f5e9);
  padding: 6px 12px;
  border-bottom: 1px solid #dcdfe6;
  font-size: 13px;
  font-weight: 600;
  color: #67c23a;
}

/* LabelFrame 内容区 */
.frame-content {
  padding: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 信息列 */
.info-column {
  flex: 1;
}

/* 用户信息行 */
.info-row {
  display: flex;
  margin-bottom: 8px;
  font-size: 13px;
  align-items: center;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  color: #666;
  width: 70px;
  flex-shrink: 0;
}

.info-value {
  color: #333;
  font-weight: 600;
}

.best-scores-inline {
  display: flex;
  gap: 30px;
  margin-left: 5px;
}

.score-item {
  color: #333;
  font-weight: 600;
  font-size: 13px;
  min-width: 100px;
}

/* 退出登录按钮区域 */
.logout-frame {
  flex-shrink: 0;
  margin-left: 20px;
}

.logout-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 5px;
  transition: all 0.3s;
  background: #fef0f0;
  color: #f56c6c;
  font-size: 12px;
}

.logout-btn:hover {
  background: #f56c6c;
  color: white;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 108, 108, 0.3);
}

/* 底部版权信息 */
.footer {
  text-align: center;
  padding: 10px 0 20px 0;
  flex-shrink: 0;
}

.footer p {
  margin: 0;
  font-size: 13px;
  color: #999;
}

/* 全屏视图 */
.fullscreen-view {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: #f5f7fa;
  z-index: 1000;
}
</style>
