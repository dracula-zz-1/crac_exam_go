<template>
  <div class="exam-result-container">
    <div class="result-card">
      <div class="result-header">
        <h1>考试结果</h1>
        <div class="exam-info">
          <span>{{ userStore.currentCategory }}类模拟考试</span>
          <span>{{ examDate }}</span>
        </div>
      </div>

      <div class="score-section">
        <div class="score-circle" :class="scoreClass">
          <div class="score-number">{{ score }}</div>
          <div class="score-label">分</div>
        </div>
        <div class="score-info">
          <div :class="['score-text', passClass]">
            {{ passed ? '✓ 考试通过' : '✗ 未通过' }}
          </div>
          <div class="pass-score">
            通过分数：{{ passScore }} 分
          </div>
        </div>
      </div>

      <div class="statistics-section">
        <div class="stat-item">
          <div class="stat-value">{{ totalQuestions }}</div>
          <div class="stat-label">总题数</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ correctCount }}</div>
          <div class="stat-label">正确数</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ wrongCount }}</div>
          <div class="stat-label">错误数</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ accuracy }}%</div>
          <div class="stat-label">正确率</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ duration }}分钟</div>
          <div class="stat-label">用时</div>
        </div>
      </div>

      <div class="action-buttons">
        <el-button type="primary" @click="reviewPaper">
          <el-icon><Document /></el-icon>
          查看试卷
        </el-button>
        <el-button @click="goHome">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
      </div>
    </div>

    <!-- 查看试卷对话框 - 全部显示，垂直滚动 -->
    <el-dialog 
      v-model="dialogVisible" 
      title="查看试卷" 
      width="900px"
      :close-on-click-modal="false"
      top="15px"
      :style="{ height: 'calc(100vh - 30px)' }"
    >
      <div class="paper-content" style="height: calc(100vh - 180px); overflow-y: auto;">
        <div 
          v-for="(question, index) in questions" 
          :key="question.id" 
          class="paper-question"
        >
          <div class="paper-header">
            <span class="paper-number">第 {{ index + 1 }} 题</span>
            <span class="paper-type">{{ question.type === 1 ? '单选题' : '多选题' }}</span>
            <span :class="['paper-tag', getResultClass(index)]">
              {{ getResultText(index) }}
            </span>
          </div>
          <div class="paper-content-text">{{ question.Q }}</div>
          
          <div class="paper-image" v-if="question.hasImage && question.imageBase64">
            <img :src="question.imageBase64" alt="题目图片" />
          </div>

          <div class="paper-options">
            <div
              v-for="option in getOptions(question)"
              :key="option.key"
              :class="['paper-option', getOptionClass(question, option.key)]"
            >
              <span class="option-key">{{ option.key }}.</span>
              <span class="option-text">{{ option.text }}</span>
            </div>
          </div>

          <div class="paper-answer">
            <div><strong>你的答案：</strong>
              <span :class="{ 'text-wrong': !userAnswers[question.id] }">
                {{ userAnswers[question.id] || '未作答' }}
              </span>
            </div>
            <div><strong>正确答案：</strong>{{ question.T }}</div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button type="primary" @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { HomeFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { SettingsService, ExamService } from '@/api'

interface Question {
  id: number
  J: string
  Q: string
  T: string
  A: string
  B: string
  C: string
  D: string
  type: number
  hasImage: boolean
  imageBase64?: string
}

const router = useRouter()
const userStore = useUserStore()

const score = ref(0)
const totalQuestions = ref(0)
const correctCount = ref(0)
const wrongCount = ref(0)
const duration = ref(0)
const examDate = ref('')
const userAnswers = ref<Record<number, string>>({})
const questions = ref<Question[]>([])

const dialogVisible = ref(false)
const examConfigs = ref<Record<string, any>>({})

// 加载考试配置
const loadExamConfigs = async () => {
  try {
    examConfigs.value = await SettingsService.GetAllExamConfigs()
  } catch (error: any) {
    console.error('Load exam configs error:', error)
    // 如果加载失败，使用默认值
    examConfigs.value = {
      'A': { pass_score: 30 },
      'B': { pass_score: 45 },
      'C': { pass_score: 70 }
    }
  }
}

const passScore = computed(() => {
  // 根据类别获取通过分数（从后端配置）
  const category = userStore.currentCategory
  if (examConfigs.value[category]) {
    return examConfigs.value[category].pass_score
  }
  return 60 // 默认值
})

const passed = computed(() => score.value >= passScore.value)

const accuracy = computed(() => {
  if (totalQuestions.value === 0) return 0
  return Math.round((correctCount.value / totalQuestions.value) * 100)
})

const scoreClass = computed(() => {
  if (score.value >= 90) return 'excellent'
  if (score.value >= 60) return 'good'
  return 'poor'
})

const passClass = computed(() => passed.value ? 'pass' : 'fail')

// 获取选项列表（用于试卷查看）
const getOptions = (question: Question) => {
  return [
    { key: 'A', text: question.A },
    { key: 'B', text: question.B },
    { key: 'C', text: question.C },
    { key: 'D', text: question.D }
  ]
}

// 获取选项样式类（用于试卷查看）
const getOptionClass = (question: Question, optionKey: string) => {
  const userAnswer = userAnswers.value[question.id] || ''
  const isUserSelected = userAnswer.includes(optionKey)
  const isCorrect = question.T.includes(optionKey)
  
  // 用户选择了该选项
  if (isUserSelected) {
    // 且是正确答案，显示绿色
    if (isCorrect) {
      return 'correct'
    }
    // 但是错误答案，显示红色
    return 'wrong'
  }
  
  // 用户没选，但是正确答案（漏选）
  if (isCorrect) {
    return 'correct'
  }
  
  // 用户没选，也不是正确答案
  return ''
}

// 获取结果样式类
const getResultClass = (index: number) => {
  const answer = userAnswers.value[questions.value[index]?.id]
  if (!answer) return 'unanswered'
  if (answer === questions.value[index].T) return 'correct'
  return 'wrong'
}

// 获取结果文本
const getResultText = (index: number) => {
  const answer = userAnswers.value[questions.value[index]?.id]
  if (!answer) return '未答'
  if (answer === questions.value[index].T) return '正确'
  return '错误'
}

const loadExamResult = async () => {
  try {
    // 从路由参数获取考试 ID
    const route = router.currentRoute.value
    const examIdParam = route.query.examId as string
    
    if (!examIdParam) {
      ElMessage.warning('未找到考试记录')
      return
    }
    
    const examId = parseInt(examIdParam)
    
    // 调用后端 API 获取考试结果
    const result: any = await ExamService.GetExamResult(examId)
    
    score.value = result.score || 0
    totalQuestions.value = result.total_count || 0
    correctCount.value = result.correct_count || 0
    wrongCount.value = totalQuestions.value - correctCount.value
    duration.value = Math.floor((result.duration_seconds || 0) / 60)
    examDate.value = new Date().toLocaleString('zh-CN')
    
    // 加载考试题目详情
    try {
      const examQuestions: any[] = await ExamService.GetExamQuestions(examId)
      questions.value = examQuestions.map(q => ({
        id: q.question_id,
        J: '',  // 级别字段（考试题目详情中没有，留空）
        Q: q.question_text,
        T: q.correct_answer,
        A: q.option_a,
        B: q.option_b,
        C: q.option_c,
        D: q.option_d,
        type: q.type,
        hasImage: !!q.image_data && q.image_data.length > 0,
        imageBase64: q.image_data || ''
      }))
      
      // 从考试题目详情中获取用户答案
      userAnswers.value = {}
      examQuestions.forEach(q => {
        if (q.user_answer && q.user_answer.trim() !== '') {
          userAnswers.value[q.question_id] = q.user_answer
        }
      })
    } catch (error) {
      console.error('Load exam questions error:', error)
      questions.value = []
      userAnswers.value = {}
    }
    
    ElMessage.success('考试结果加载成功')
  } catch (error: any) {
    console.error('Load exam result error:', error)
    ElMessage.error('加载考试结果失败：' + (error.message || '未知错误'))
  }
}

const reviewPaper = () => {
  // 打开试卷查看对话框
  dialogVisible.value = true
}

const goHome = () => {
  router.push('/home')
}

onMounted(() => {
  loadExamResult()
  loadExamConfigs() // 加载考试配置
})
</script>

<style scoped>
html, body {
  overflow: hidden !important;
  margin: 0;
  padding: 0;
  height: 100vh;
  width: 100vw;
}

.exam-result-container {
  width: 100%;
  height: 100vh;
  padding: 0;
  background: #f5f7fa;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden !important;
}

.result-card {
  width: 900px;
  background: white;
  border-radius: 10px;
  padding: 15px 25px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  max-height: calc(100vh - 20px);
  overflow-x: hidden !important;
  overflow-y: hidden !important;
  scrollbar-width: none !important;
  -ms-overflow-style: none !important;
}

.result-card::-webkit-scrollbar {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
}

.result-header {
  text-align: center;
  margin-bottom: 12px;
}

.result-header h1 {
  font-size: 24px;
  color: #333;
  margin-bottom: 6px;
}

.exam-info {
  display: flex;
  justify-content: center;
  gap: 20px;
  color: #666;
  font-size: 14px;
}

.score-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 30px;
  margin: 20px 0;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 10px;
}

.score-circle {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 6px solid;
}

.score-circle.excellent {
  border-color: #67c23a;
  background: #f0f9eb;
}

.score-circle.good {
  border-color: #409EFF;
  background: #ecf5ff;
}

.score-circle.poor {
  border-color: #f56c6c;
  background: #fef0f0;
}

.score-number {
  font-size: 36px;
  font-weight: bold;
  color: #333;
}

.score-label {
  font-size: 14px;
  color: #666;
}

.score-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.score-text {
  font-size: 20px;
  font-weight: bold;
}

.score-text.pass {
  color: #67c23a;
}

.score-text.fail {
  color: #f56c6c;
}

.pass-score {
  font-size: 14px;
  color: #666;
}

.statistics-section {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
  margin: 15px 0;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 10px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.answer-review-section {
  margin: 15px 0;
}

.answer-review-section h2 {
  font-size: 18px;
  color: #333;
  margin-bottom: 10px;
}

.review-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 6px;
}

.review-item {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #e6e6e6;
  border-radius: 5px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s;
  font-size: 14px;
}

.review-item.correct {
  border-color: #67c23a;
  background: #f0f9eb;
  color: #67c23a;
}

.review-item.wrong {
  border-color: #f56c6c;
  background: #fef0f0;
  color: #f56c6c;
}

.review-item.unanswered {
  border-color: #909399;
  background: #f5f7fa;
  color: #909399;
}

.review-item:hover {
  transform: scale(1.1);
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 15px;
  margin-top: 20px;
}

.action-buttons .el-button {
  min-width: 120px;
  height: 45px;
  font-size: 16px;
}

.question-detail {
  padding: 20px;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
  font-size: 16px;
}

.question-number {
  font-weight: 600;
  color: #409EFF;
}

.question-type {
  color: #666;
}

.result-tag {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
}

.result-tag.correct {
  background: #f0f9eb;
  color: #67c23a;
}

.result-tag.wrong {
  background: #fef0f0;
  color: #f56c6c;
}

.result-tag.unanswered {
  background: #f5f7fa;
  color: #909399;
}

.detail-content {
  font-size: 18px;
  line-height: 1.8;
  color: #333;
  margin-bottom: 20px;
}

.detail-image {
  text-align: center;
  margin: 20px 0;
}

.detail-image img {
  max-width: 100%;
  max-height: 400px;
  border-radius: 5px;
}

.detail-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 20px 0;
}

.option-item {
  padding: 12px 15px;
  border: 2px solid #e6e6e6;
  border-radius: 5px;
  display: flex;
  gap: 10px;
}

.option-item.selected {
  border-color: #409EFF;
  background: #ecf5ff;
}

.option-item.correct {
  border-color: #67c23a;
  background: #f0f9eb;
}

.option-key {
  font-weight: 600;
  color: #409EFF;
}

.option-text {
  flex: 1;
  color: #333;
}

.detail-answer {
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 5px;
  font-size: 16px;
}

.detail-answer div {
  margin: 8px 0;
}

.text-wrong {
  color: #f56c6c;
  font-weight: 500;
}

/* 试卷查看样式 */
.paper-content {
  padding: 10px 20px;
}

.paper-question {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  background: #fafafa;
}

.paper-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
  font-size: 16px;
}

.paper-number {
  font-weight: 600;
  color: #333;
}

.paper-type {
  color: #666;
}

.paper-tag {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
}

.paper-tag.correct {
  background: #f0f9eb;
  color: #67c23a;
}

.paper-tag.wrong {
  background: #fef0f0;
  color: #f56c6c;
}

.paper-tag.unanswered {
  background: #f4f4f5;
  color: #909399;
}

.paper-content-text {
  font-size: 18px;
  line-height: 1.8;
  color: #333;
  margin-bottom: 20px;
}

.paper-image {
  text-align: center;
  margin: 15px 0;
}

.paper-image img {
  max-width: 100%;
  max-height: 400px;
  border-radius: 5px;
}

.paper-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 15px 0;
}

.paper-option {
  padding: 12px 15px;
  border: 2px solid #e6e6e6;
  border-radius: 5px;
  display: flex;
  gap: 10px;
}

.paper-option.selected {
  border-color: #409EFF;
  background: #ecf5ff;
}

.paper-option.correct {
  border-color: #67c23a;
  background: #f0f9eb;
}

.paper-option.wrong {
  border-color: #f56c6c;
  background: #fef0f0;
}

.paper-option-key {
  font-weight: 600;
  color: #409EFF;
}

.paper-option-text {
  flex: 1;
  color: #333;
}

.paper-answer {
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 5px;
  font-size: 16px;
}

.paper-answer div {
  margin: 8px 0;
}
</style>
