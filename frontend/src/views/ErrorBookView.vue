<template>
  <div class="error-book-container">
    <div class="header">
      <div class="back-btn" @click="goBack">
        <el-icon :size="30"><Back /></el-icon>
        <span>返回</span>
      </div>
      <div class="title">
        {{ userStore.currentCategory }}类错题本
      </div>
      <div style="width: 100px;"></div>
    </div>

    <div class="question-list" v-loading="loading">
      <div
        v-for="(question, index) in questions"
        :key="question.id"
        class="question-item"
        @click="viewQuestion(question)"
      >
        <div class="question-number">{{ index + 1 }}. {{ question.J }}</div>
        <div class="question-content">{{ question.Q }}</div>
        <div class="question-footer">
          <span class="question-type">{{ question.type === 1 ? '单选题' : '多选题' }}</span>
          <span class="correct-answer">正确答案：{{ question.T }}</span>
        </div>
      </div>

      <el-empty v-if="!loading && questions.length === 0" description="暂无错题" />
    </div>

    <!-- 题目详情对话框 -->
    <el-dialog v-model="dialogVisible" title="题目详情" width="800px">
      <div class="question-detail" v-if="currentQuestion">
        <div class="detail-header">
          <span class="question-number">第 {{ currentIndex + 1 }}/{{ totalQuestions }} 题</span>
          <span class="question-type">({{ typeText }})</span>
        </div>
        <div class="detail-content">{{ currentQuestion.Q }}</div>
        
        <div class="detail-image" v-if="currentQuestion.hasImage && currentQuestion.imageBase64">
          <img :src="currentQuestion.imageBase64" alt="题目图片" />
        </div>

        <div class="detail-options">
          <div
            v-for="option in options"
            :key="option.key"
            :class="['option-item', { 
              'selected': userAnswer === option.key,
              'correct': answerRevealed && option.key === currentQuestion.T,
              'wrong': answerRevealed && userAnswer === option.key && option.key !== currentQuestion.T
            }]"
            @click="selectAnswer(option.key)"
          >
            <span class="option-key">{{ option.key }}.</span>
            <span class="option-text">{{ option.text }}</span>
          </div>
        </div>

        <!-- 答案显示区域（带刮刮卡遮罩） -->
        <div class="detail-answer-container">
          <div class="scratch-overlay" v-show="!answerRevealed" @click="revealAnswer">
            <div class="scratch-hint">
              <el-icon :size="40"><View /></el-icon>
              <p>点击查看答案</p>
            </div>
          </div>
          
          <div class="detail-answer" v-show="answerRevealed">
            <strong>正确答案：</strong>{{ currentQuestion.T }}
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="previousQuestion" :disabled="currentIndex === 0 && totalQuestions > 1">上一题</el-button>
        <el-button @click="nextQuestion" :disabled="currentIndex === totalQuestions - 1 && totalQuestions > 1">
          {{ totalQuestions === 1 ? '完成' : '下一题' }}
        </el-button>
        <el-button type="primary" @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { FavoriteService } from '@/api'
import { Back, View } from '@element-plus/icons-vue'

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

const loading = ref(false)
const questions = ref<Question[]>([])
const dialogVisible = ref(false)
const currentQuestion = ref<Question | null>(null)
const currentIndex = ref(0)
const totalQuestions = ref(0)
const userAnswer = ref('')
const answerRevealed = ref(false) // 刮刮卡答案是否已显示
const isWrongAnswer = ref(false) // 标记是否答错

const typeText = computed(() => {
  if (!currentQuestion.value) return ''
  return currentQuestion.value.type === 1 ? '单选题' : '多选题'
})

const options = computed(() => {
  if (!currentQuestion.value) return []
  
  // 返回所有选项
  return [
    { key: 'A', text: currentQuestion.value.A },
    { key: 'B', text: currentQuestion.value.B },
    { key: 'C', text: currentQuestion.value.C },
    { key: 'D', text: currentQuestion.value.D }
  ]
})

const loadQuestions = async () => {
  loading.value = true
  try {
    // 调用后端 API 获取错题
    const userID = userStore.currentUser?.user_id || 0
    const category = userStore.currentCategory
    console.log('加载错题，userID:', userID, 'category:', category)
    
    const result = await FavoriteService.GetErrorQuestionsByCategory(userID, category)
    console.log('错题原始数据:', result)
    console.log('错题数量:', result.length)
    
    questions.value = result as any
    totalQuestions.value = questions.value.length
    
    console.log('最终题目数量:', totalQuestions.value)
  } catch (error: any) {
    console.error('Load error questions error:', error)
    ElMessage.error('加载错题失败：' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const viewQuestion = (question: Question) => {
  currentQuestion.value = question
  currentIndex.value = questions.value.findIndex(q => q.id === question.id)
  userAnswer.value = ''
  answerRevealed.value = false  // 重置刮刮卡状态
  isWrongAnswer.value = false  // 重置答错标记
  dialogVisible.value = true
}

// 添加排序函数
const sortString = (str: string): string => {
  return str.split('').sort().join('')
}

// 选择答案
const selectAnswer = (key: string) => {
  if (!currentQuestion.value || answerRevealed.value) return  // 已显示答案后不能修改
  
  userAnswer.value = key
}

// 刮刮卡：点击查看答案，触发判题逻辑
const revealAnswer = () => {
  if (!currentQuestion.value) return
  
  // 判断答案是否正确（未答题也视为错误）
  const isCorrect = sortString(userAnswer.value) === sortString(currentQuestion.value.T)
  
  // 显示答案
  answerRevealed.value = true
  
  // 如果答错了（包括未答题），标记
  if (!isCorrect) {
    isWrongAnswer.value = true
    // 错题本中的题目答错，不需要再加入错题本
  } else {
    isWrongAnswer.value = false
  }
}

const previousQuestion = () => {
  // 如果已经显示答案且答错，需要再次点击才能切换
  if (answerRevealed.value && isWrongAnswer.value) {
    if (currentIndex.value > 0) {
      currentIndex.value--
      currentQuestion.value = questions.value[currentIndex.value]
      userAnswer.value = ''
      answerRevealed.value = false
      isWrongAnswer.value = false
    }
    return
  }
  
  // 答案未显示或答对：先判题
  if (!answerRevealed.value) {
    revealAnswer()
    // 如果答错，不切换
    if (isWrongAnswer.value) {
      return
    }
  }
  
  // 答对的情况下直接切换
  if (currentIndex.value > 0) {
    currentIndex.value--
    currentQuestion.value = questions.value[currentIndex.value]
    userAnswer.value = ''
    answerRevealed.value = false
    isWrongAnswer.value = false
  }
}

const nextQuestion = () => {
  // 如果已经显示答案且答错，需要再次点击才能切换
  if (answerRevealed.value && isWrongAnswer.value) {
    if (currentIndex.value < totalQuestions.value - 1) {
      currentIndex.value++
      currentQuestion.value = questions.value[currentIndex.value]
      userAnswer.value = ''
      answerRevealed.value = false
      isWrongAnswer.value = false
    } else if (totalQuestions.value === 1) {
      dialogVisible.value = false
    }
    return
  }
  
  // 答案未显示或答对：先判题
  if (!answerRevealed.value) {
    revealAnswer()
    // 如果答错，不切换
    if (isWrongAnswer.value) {
      return
    }
  }
  
  // 答对的情况下直接切换
  if (currentIndex.value < totalQuestions.value - 1) {
    currentIndex.value++
    currentQuestion.value = questions.value[currentIndex.value]
    userAnswer.value = ''
    answerRevealed.value = false
    isWrongAnswer.value = false
  } else if (totalQuestions.value === 1) {
    dialogVisible.value = false
  }
}

const goBack = () => {
  router.push('/home')
}

onMounted(() => {
  loadQuestions()
})
</script>

<style scoped>
.error-book-container {
  width: 100%;
  height: 100vh;
  padding: 20px;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  color: #666;
  padding: 10px;
  border-radius: 5px;
  transition: all 0.3s;
}

.back-btn:hover {
  background: #e6e6e6;
  color: #409EFF;
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.question-list {
  flex: 1;
  background: white;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  overflow-y: auto;
}

.question-item {
  padding: 20px;
  border-bottom: 1px solid #e6e6e6;
  cursor: pointer;
  transition: all 0.3s;
}

.question-item:hover {
  background: #f5f7fa;
}

.question-number {
  font-weight: 600;
  margin-bottom: 10px;
  color: #409EFF;
}

.question-content {
  font-size: 16px;
  line-height: 1.6;
  color: #333;
  margin-bottom: 10px;
}

.question-footer {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
}

.question-type {
  color: #909399;
}

.correct-answer {
  color: #67c23a;
  font-weight: 500;
}

.question-detail {
  padding: 15px;
}

.detail-header {
  margin-bottom: 10px;
  font-size: 16px;
}

.detail-number {
  font-weight: 600;
  margin-right: 10px;
}

.detail-type {
  color: #666;
}

.detail-content {
  font-size: 18px;
  line-height: 1.8;
  color: #333;
  margin-bottom: 10px;
}

.detail-image {
  text-align: center;
  margin: 0 0 10px 0;
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
  margin: 0 0 5px 0;
}

.option-item {
  padding: 12px 15px;
  border: 2px solid #e6e6e6;
  border-radius: 5px;
  display: flex;
  gap: 10px;
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

/* 答案显示区域 */
.detail-answer-container {
  position: relative;
  margin-top: 20px;
  margin-bottom: 30px;
  min-height: 60px;
}

.detail-answer {
  padding: 15px;
  background: #f0f9eb;
  border-radius: 5px;
  color: #67c23a;
  font-size: 16px;
}

/* 刮刮卡遮罩层 */
.scratch-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 5px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
}

.scratch-overlay:hover {
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  transform: scale(1.02);
}

.scratch-hint {
  text-align: center;
  color: white;
}

.scratch-hint p {
  margin: 10px 0 0 0;
  font-size: 16px;
  font-weight: 600;
}

/* 对话框按钮区域样式 */
:deep(.el-dialog__footer) {
  padding: 10px 20px 15px 20px;
  margin-bottom: 5px;
}
</style>
