<template>
  <div class="exam-container">
    <!-- 考试前确认界面 -->
    <div class="exam-confirm" v-if="!examStarted">
      <div class="confirm-card">
        <div class="back-button" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          <span>返回</span>
        </div>
        
        <h1 class="main-title">业余无线电台操作技术能力验证</h1>
        <h2 class="sub-title">模拟考试</h2>
        
        <div class="category-letter">{{ userStore.currentCategory }}</div>
        
        <div class="exam-details">
          <div class="detail-line">
            共<span class="highlight">{{ examConfig?.total || 0 }}</span>题
          </div>
          <div class="detail-line">
            其中单选<span class="highlight-red">{{ examConfig?.single || 0 }}</span>题，
            多选<span class="highlight-red">{{ examConfig?.multiple || 0 }}</span>题
          </div>
          <div class="detail-line">
            每题<span class="highlight-red">1</span>分，
            <span class="highlight-red">{{ examConfig?.pass_score || 0 }}</span>分通过
          </div>
          <div class="detail-line">
            考试时间<span class="highlight-red">{{ examConfig?.time_minutes || 0 }}</span>分钟
          </div>
        </div>
        
        <div class="start-button-area">
          <el-button type="primary" @click="startExam">
            开始考试
          </el-button>
        </div>
      </div>
    </div>

    <!-- 考试界面 -->
    <div v-else class="exam-interface">
      <!-- 顶部信息栏 -->
      <div class="header">
        <div class="back-btn" @click="confirmExit">
          <el-icon :size="20"><Back /></el-icon>
          <span>返回</span>
        </div>
        <div class="info-text">
          <span>模拟考试 - {{ userStore.currentCategory }}类</span>
          <span class="total-questions">共 {{ examConfig?.total || 0 }} 题</span>
          <span class="exam-time">考试时间：{{ examConfig?.time_minutes || 0 }}分钟</span>
          <span class="timer">剩余时间：{{ formattedTime }}</span>
        </div>
        <div class="action-buttons">
          <el-button @click="showCardDialog = true">
            <el-icon><Document /></el-icon>
            答题卡
          </el-button>
          <el-button type="danger" @click="confirmSubmit">
            <el-icon><Check /></el-icon>
            交卷
          </el-button>
        </div>
      </div>

    <!-- 题目区域 -->
    <div class="question-area">
      <div class="question-content-wrapper" v-if="currentQuestion">
        <div class="question-text">
          <span class="question-number">第 {{ currentIndex + 1 }}/{{ totalQuestions }} 题</span>
          <span class="question-type">({{ questionTypeText }})</span>
          <span class="question-content">{{ currentQuestion.Q }}</span>
        </div>

        <!-- 图片显示 -->
        <div class="question-image" v-if="currentQuestion.hasImage && currentQuestion.imageBase64">
          <img :src="currentQuestion.imageBase64" alt="题目图片" />
        </div>

        <!-- 选项区域 -->
        <div class="options-area">
          <div
            v-for="option in options"
            :key="option.key"
            :class="['option-item', { 'selected': userAnswers[currentQuestion.id]?.includes(option.key) }]"
            @click="selectAnswer(option.key)"
          >
            <span class="option-key">{{ option.key }}.</span>
            <span class="option-text">{{ option.text }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部导航 - 固定在底部 -->
    <div class="bottom-navigation-fixed">
      <el-button @click="previousQuestion" :disabled="currentIndex === 0">
        <el-icon><ArrowLeft /></el-icon>
        上一题
      </el-button>
      <el-button type="primary" @click="nextQuestion" :disabled="currentIndex === totalQuestions - 1">
        下一题
        <el-icon><ArrowRight /></el-icon>
      </el-button>
    </div>

    <!-- 答题卡对话框 -->
    <el-dialog v-model="showCardDialog" width="900px" :style="{ maxHeight: '90vh', overflow: 'hidden' }" :show-close="false" top="2vh">
      <div class="answer-card" @click.self="showCardDialog = false" style="max-height: 80vh; overflow-y: auto;">
        <div
          v-for="(question, index) in questions"
          :key="question.id"
          :class="['card-item', {
            'answered': userAnswers[question.id],
            'current': index === currentIndex
          }]"
          @click="goToQuestion(index)"
        >
          {{ index + 1 }}
        </div>
      </div>
    </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Back, Document, Check, ArrowLeft } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { SettingsService, ExamService } from '@/api'
import type { Question } from '@/types'

// ExamView 扩展的题目类型（含前端计算字段）
interface ExamQuestion extends Question {
  hasImage: boolean
  imageBase64?: string
}

// 定义 emit
const emit = defineEmits<{
  back: []
}>()

const router = useRouter()

const userStore = useUserStore()

const examID = ref<number>(0)
const startTime = ref<number>(Date.now())
const examStarted = ref(false)
const questions = ref<ExamQuestion[]>([])
const currentQuestion = ref<ExamQuestion | null>(null)
const currentIndex = ref(0)
const totalQuestions = ref(0)
const userAnswers = ref<Record<number, string>>({})
const remainingTime = ref(0)
const timerInterval = ref<number | null>(null)
const showCardDialog = ref(false)
const examConfig = ref<any>(null)

// 加载考试配置
const loadExamConfig = async () => {
  try {
    examConfig.value = await SettingsService.GetExamConfig(userStore.currentCategory)
  } catch (error: any) {
    console.error('Load exam config error:', error)
    // 如果加载失败，使用默认配置
    examConfig.value = {
      total: 50,
      time_minutes: 45
    }
  }
}

const formattedTime = computed(() => {
  const minutes = Math.floor(remainingTime.value / 60)
  const seconds = remainingTime.value % 60
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
})

const questionTypeText = computed(() => {
  if (!currentQuestion.value) return ''
  return currentQuestion.value.type === 1 ? '单选题' : '多选题'
})

const options = computed(() => {
  if (!currentQuestion.value) return []
  
  // 单选题和多选题都显示 ABCD 选项
  return [
    { key: 'A', text: currentQuestion.value.A },
    { key: 'B', text: currentQuestion.value.B },
    { key: 'C', text: currentQuestion.value.C },
    { key: 'D', text: currentQuestion.value.D }
  ]
})

const loadExam = async () => {
  try {
    // 调用后端 API 开始考试
    const result: any = await ExamService.StartExam(userStore.currentCategory)
    
    if (!result || !result.questions || result.questions.length === 0) {
      ElMessage.warning({
        message: `"${userStore.currentCategory}类考试题库为空，2 秒后返回首页`,
        duration: 2000
      })
      setTimeout(() => {
        emit('back')
      }, 2000)
      return
    }
    
    // 保存考试 ID 和开始时间
    examID.value = result.exam_id
    startTime.value = Date.now()
    
    // 加载题目
    questions.value = result.questions as any
    totalQuestions.value = questions.value.length
    
    // 使用配置中的考试时间
    if (result.config) {
      remainingTime.value = result.config.time_minutes * 60 // 转换为秒
    }
    
    // 加载第一题
    if (questions.value.length > 0) {
      currentQuestion.value = questions.value[0]
    }
    
    // 标记考试已开始
    examStarted.value = true
    
    // 启动计时器
    startTimer()
    
    ElMessage.success('考试已开始，祝你考试顺利！')
  } catch (error: any) {
    ElMessage.error(error.message || '加载考试失败')
    console.error('Load exam error:', error)
  }
}

const startExam = async () => {
  await loadExam()
}

const startTimer = () => {
  timerInterval.value = window.setInterval(() => {
    if (remainingTime.value > 0) {
      remainingTime.value--
    } else {
      confirmSubmit()
    }
  }, 1000)
}

const selectAnswer = (key: string) => {
  if (!currentQuestion.value) return
  
  const questionId = currentQuestion.value.id
  
  // 多选题支持选择多个答案
  if (currentQuestion.value.type === 2) {
    // 如果是多选题，使用数组存储答案
    const currentAnswer = userAnswers.value[questionId] || ''
    if (currentAnswer.includes(key)) {
      // 如果已选择，取消选择
      userAnswers.value[questionId] = currentAnswer.replace(key, '').split('').sort().join('')
    } else {
      // 如果未选择，添加答案
      userAnswers.value[questionId] = (currentAnswer + key).split('').sort().join('')
    }
  } else {
    // 单选题直接设置答案
    userAnswers.value[questionId] = key
  }
}

const nextQuestion = () => {
  if (currentIndex.value < totalQuestions.value - 1) {
    currentIndex.value++
    loadQuestion(currentIndex.value)
  }
}

const previousQuestion = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
    loadQuestion(currentIndex.value)
  }
}

const loadQuestion = (index: number) => {
  currentQuestion.value = questions.value[index]
}

const goToQuestion = (index: number) => {
  currentIndex.value = index
  loadQuestion(index)
  showCardDialog.value = false
}

const goBack = () => {
  emit('back')
}

const confirmExit = async () => {
  try {
    await ElMessageBox.confirm('确定要退出考试吗？当前答题记录将不会保存', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    // ✅ 修复 BUG-002: 退出时作废考试记录
    if (examID.value) {
      await ExamService.InvalidateExam(examID.value)
    }
    goBack()
  } catch (error) {
    // 用户取消退出
  }
}

const confirmSubmit = async () => {
  const answeredCount = Object.keys(userAnswers.value).length
  const unansweredCount = totalQuestions.value - answeredCount
  
  if (unansweredCount > 0) {
    try {
      await ElMessageBox.confirm(
        `还有 ${unansweredCount} 道题未作答，确定要交卷吗？`,
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }
  }
  
  submitExam()
}

const submitExam = async () => {
  try {
    if (examID.value === 0) {
      ElMessage.error('考试未开始，无法提交')
      return
    }
    
    // 调用后端 API 提交考试
    await ExamService.SubmitExam(examID.value, userAnswers.value, startTime.value)
    
    ElMessage.success('考试已提交')
    if (timerInterval.value) {
      clearInterval(timerInterval.value)
    }
    
    // 跳转到考试结果页面，并传递考试 ID
    router.push({ path: '/exam-result', query: { examId: examID.value.toString() } })
  } catch (error: any) {
    ElMessage.error(error.message || '提交考试失败')
    console.error('Submit exam error:', error)
  }
}

onMounted(() => {
  loadExamConfig() // 只加载考试配置，不立即开始考试
})

onUnmounted(() => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value)
  }
})
</script>

<style scoped>
.exam-container {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 考试前确认界面 */
.exam-confirm {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background: #f5f7fa;
}

.confirm-card {
  background: white;
  padding: 30px 50px 50px;
  border-radius: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  text-align: center;
  min-width: 600px;
  position: relative;
}

.back-button {
  position: absolute;
  top: 20px;
  left: 20px;
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  color: #666;
  padding: 10px;
  border-radius: 5px;
  transition: all 0.3s;
}

.back-button:hover {
  background: #e6e6e6;
  color: #409EFF;
}

.main-title {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 10px;
  font-weight: bold;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.sub-title {
  font-size: 28px;
  color: #409EFF;
  margin-bottom: 20px;
  font-weight: bold;
}

.category-letter {
  font-size: 80px;
  font-weight: bold;
  color: #E6A23C;
  margin: 20px 0;
  text-shadow: 3px 3px 6px rgba(0, 0, 0, 0.15);
}

.exam-details {
  margin: 30px 0;
  font-size: 18px;
  line-height: 2;
  color: #606266;
}

.detail-line {
  margin: 8px 0;
}

.highlight {
  color: #E6A23C;
  font-weight: bold;
  font-size: 20px;
}

.highlight-red {
  color: #f56c6c;
  font-weight: bold;
  font-size: 20px;
}

.start-button-area {
  margin-top: 40px;
}

.start-button-area .el-button {
  width: 200px;
  height: 50px;
  font-size: 20px;
  font-weight: bold;
  background: linear-gradient(135deg, #66b1ff 0%, #409eff 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

.start-button-area .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.4);
}

/* 考试界面 */
.exam-interface {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.header {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: white;
  padding: 15px 30px;
  border-radius: 10px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  padding: 8px 15px;
  border-radius: 5px;
  color: #606266;
  font-size: 14px;
  transition: all 0.3s;
}

.back-btn:hover {
  background: #f5f7fa;
  color: #409EFF;
}

.info-text {
  display: flex;
  gap: 30px;
  font-size: 18px;
  font-weight: 600;
  align-items: center;
}

.total-questions {
  color: #409EFF;
}

.exam-time {
  color: #67c23a;
}

.timer {
  color: #f56c6c;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.question-area {
  flex: 1;
  overflow-y: auto;
  background: white;
  border-radius: 10px;
  padding: 30px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  margin-bottom: 0;
}

.question-content-wrapper {
  min-height: 100%;
}

.question-text {
  font-size: 18px;
  line-height: 1.8;
  color: #333;
  margin-bottom: 20px;
}

.question-number {
  font-weight: 600;
  margin-right: 10px;
}

.question-type {
  color: #666;
  margin-right: 10px;
}

.question-image {
  text-align: center;
  margin: 20px 0;
}

.question-image img {
  max-width: 100%;
  max-height: 400px;
  border-radius: 5px;
}

.options-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin: 20px 0;
}

.option-item {
  display: flex;
  align-items: flex-start;
  padding: 15px 20px;
  border: 2px solid #e6e6e6;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.option-item:hover {
  border-color: #409EFF;
  background: #f0f9ff;
}

.option-item.selected {
  border-color: #409EFF;
  background: #ecf5ff;
}

.option-key {
  font-weight: 600;
  margin-right: 10px;
  color: #409EFF;
}

.option-text {
  flex: 1;
  font-size: 16px;
  line-height: 1.6;
  color: #333;
}

.bottom-navigation-fixed {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 20px;
  padding: 20px 40px;
  background: transparent;
  z-index: 100;
}

.answer-card {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 12px;
  padding: 15px;
  scrollbar-width: none !important;
  -ms-overflow-style: none !important;
}

.answer-card::-webkit-scrollbar {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
}

.card-item {
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #e6e6e6;
  border-radius: 5px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s;
}

.card-item:hover {
  border-color: #409EFF;
  background: #f0f9ff;
}

.card-item.answered {
  border-color: #67c23a;
  background: #f0f9eb;
  color: #67c23a;
}

.card-item.current {
  border-color: #409EFF;
  background: #ecf5ff;
}
</style>
