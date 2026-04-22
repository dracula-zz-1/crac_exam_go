<template>
  <div class="practice-container">
    <!-- 顶部导航 -->
    <div class="header">
      <div class="back-btn" @click="goBack">
        <el-icon :size="30"><Back /></el-icon>
        <span>返回</span>
      </div>
      <div class="info-text">
        {{ infoText }}
      </div>
      <div style="width: 100px;"></div>
    </div>

    <!-- 题目区域 -->
    <div class="question-area" v-if="currentQuestion">
      <div class="question-header">
        <div class="question-text">
          <span class="question-number">第 {{ currentIndex + 1 }}/{{ totalQuestions }} 题</span>
          <span class="question-type">({{ questionTypeText }})</span>
          <span class="question-content">{{ currentQuestion.Q }}</span>
        </div>
        <div class="favorite-btn" @click="toggleFavorite">
          <el-icon :size="30" :color="isFavorite ? '#F7BA2A' : '#C0C4CC'">
            <Star v-if="isFavorite" />
            <StarFilled v-else />
          </el-icon>
          <span>收藏</span>
        </div>
      </div>

      <!-- 图片显示 -->
      <div class="question-image" v-if="currentQuestion.hasImage && currentQuestion.imageBase64">
        <img :src="currentQuestion.imageBase64" alt="题目图片" @click="viewImage" />
      </div>

      <!-- 选项区域 -->
      <div class="options-area">
        <div
          v-for="option in options"
          :key="option.key"
          :class="['option-item', { 
            'selected': userAnswer === option.key || (currentQuestion?.type === 2 && userAnswers.includes(option.key)),
            'correct': showAnswer && currentQuestion.T?.includes(option.key),
            'wrong': showAnswer && (userAnswer === option.key || userAnswers.includes(option.key)) && !currentQuestion.T?.includes(option.key)
          }]"
          @click="selectAnswer(option.key)"
        >
          <span class="option-key">{{ option.key }}.</span>
          <span class="option-text">{{ option.text }}</span>
        </div>
      </div>

      <!-- 答案显示 -->
      <div class="answer-feedback">
        <!-- 刮刮卡遮罩层 - 一直显示直到被点击 -->
        <div class="scratch-overlay" v-show="!answerRevealed" @click="revealAnswer">
          <div class="scratch-hint">
            <el-icon :size="40"><View /></el-icon>
            <p>点击查看答案</p>
          </div>
        </div>
        
        <div :class="['answer-text', userAnswer === currentQuestion.T ? 'correct' : 'wrong']" v-show="answerRevealed">
          <span v-if="userAnswer === currentQuestion.T">✓ 回答正确</span>
          <span v-else>✗ 回答错误</span>
        </div>
        <div class="correct-answer" v-show="answerRevealed">
          正确答案：<strong>{{ currentQuestion.T }}</strong>
        </div>
        <div class="analysis" v-if="currentQuestion.analysis && answerRevealed">
          <strong>解析：</strong>{{ currentQuestion.analysis }}
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="bottom-buttons">
        <el-button @click="previousQuestion" :disabled="currentIndex === 0">
          <el-icon><ArrowLeft /></el-icon>
          上一题
        </el-button>
        <el-button type="primary" @click="nextQuestion">
          <el-icon v-if="currentIndex < totalQuestions - 1"><ArrowRight /></el-icon>
          {{ currentIndex < totalQuestions - 1 ? '下一题' : '完成' }}
        </el-button>
      </div>
    </div>

    <!-- 图片查看对话框 -->
    <el-dialog v-model="imageDialogVisible" title="查看图片" width="auto">
      <div class="image-container">
        <img :src="currentImage" alt="题目图片" style="max-width: 100%;" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Back, ArrowLeft, ArrowRight, Star, StarFilled, View } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { FavoriteService, PracticeService } from '@/api'
import type { Question } from '@/types'

// PracticeView 扩展的题目类型（含前端计算字段）
interface PracticeQuestion extends Question {
  hasImage?: boolean
  imageBase64?: string
  analysis?: string
}

// 定义 props
const props = defineProps<{
  category: 'A' | 'B' | 'C'
  mode: 'practice' | 'error' | 'favorite'
}>()

const userStore = useUserStore()

const currentQuestion = ref<PracticeQuestion | null>(null)
const currentIndex = ref(0)
const totalQuestions = ref(0)
const userAnswer = ref('')
const userAnswers = ref<string[]>([]) // 用于多选题
const showAnswer = ref(false)
const answerRevealed = ref(false) // 答案是否已显示（刮刮卡）
const isWrongAnswer = ref(false) // 标记是否答错
const isFavorite = ref(false)
const imageDialogVisible = ref(false)
const currentImage = ref('')
const isSubmitting = ref(false) // 防止重复提交

const infoText = computed(() => {
  const modeText = props.mode === 'error' ? '错题' : props.mode === 'favorite' ? '收藏' : '练习'
  return `${props.category}类${modeText} - 共${totalQuestions.value}题`
})

const questionTypeText = computed(() => {
  if (!currentQuestion.value) return ''
  return currentQuestion.value.type === 1 ? '单选题' : '多选题'
})

const options = computed(() => {
  if (!currentQuestion.value) return []
  
  // 单选题和多选题都显示 ABCD 选项（不打乱顺序）
  return [
    { key: 'A', text: currentQuestion.value.A },
    { key: 'B', text: currentQuestion.value.B },
    { key: 'C', text: currentQuestion.value.C },
    { key: 'D', text: currentQuestion.value.D }
  ]
})

const loadQuestion = async () => {
  try {
    let questions: any[] = []
    
    // 根据模式获取不同的题目
    if (props.mode === 'error') {
      // 只做错题模式
      const userID = userStore.currentUser?.user_id || 0
      console.log('加载错题，userID:', userID, 'category:', props.category)
      questions = await FavoriteService.GetErrorQuestionsByCategory(userID, props.category)
      console.log('错题数量:', questions.length)
      if (questions.length > 0) {
        console.log('第一道错题:', questions[0])
      }
    } else if (props.mode === 'favorite') {
      // 只做收藏模式
      const userID = userStore.currentUser?.user_id || 0
      console.log('加载收藏题目，userID:', userID, 'category:', props.category)
      questions = await FavoriteService.GetFavoriteQuestionsByCategory(userID, props.category)
      console.log('收藏题目数量:', questions.length)
    } else {
      // 逐题练习模式
      const result = await PracticeService.GetQuestionsByCategory(props.category)
      questions = result || []
    }
    
    if (questions.length === 0) {
      const modeText = props.mode === 'error' ? '错题' : props.mode === 'favorite' ? '收藏' : '练习'
      ElMessage.warning({
        message: `"${props.category}类${modeText}"题库为空，2 秒后返回首页`,
        duration: 2000
      })
      setTimeout(() => {
        emit('back')
      }, 2000)
      return
    }
    
    // 如果是第一次加载，从保存的进度开始（仅逐题练习模式）；否则使用当前索引
    if (currentQuestion.value === null && props.mode === 'practice') {
      const progress = await PracticeService.GetPracticeProgress(userStore.currentUser!.user_id, props.category)
      currentIndex.value = progress
      // 确保索引不超出范围
      if (currentIndex.value >= questions.length) {
        currentIndex.value = 0
      }
    }
    
    currentQuestion.value = questions[currentIndex.value]
    totalQuestions.value = questions.length
    userAnswer.value = ''
    userAnswers.value = []
    showAnswer.value = false
    answerRevealed.value = false  // 重置刮刮卡状态
    isWrongAnswer.value = false  // 重置答错标记
    isSubmitting.value = false
    
    // 检查是否已收藏（仅收藏模式需要）
    if (props.mode === 'favorite') {
      isFavorite.value = true
    } else {
      const favorites = await FavoriteService.GetFavoriteQuestionsByCategory(userStore.currentUser!.user_id, props.category)
      isFavorite.value = favorites.some((q: any) => q.id === currentQuestion.value!.id)
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载题目失败')
    console.error('Load question error:', error)
  }
}

// 添加排序函数
const sortString = (str: string): string => {
  return str.split('').sort().join('')
}

const selectAnswer = async (key: string) => {
  if (!currentQuestion.value || isSubmitting.value) return
  
  isSubmitting.value = true
  
  if (currentQuestion.value.type === 1) {
    // 单选题
    userAnswer.value = key
  } else {
    // 多选题
    const index = userAnswers.value.indexOf(key)
    if (index > -1) {
      userAnswers.value.splice(index, 1)
    } else {
      userAnswers.value.push(key)
    }
    userAnswer.value = userAnswers.value.join('')
  }
  
  // 只保存答案，不立即显示和判断
  // 答案判断在 nextQuestion/previousQuestion 中进行
  
  isSubmitting.value = false
}

// 刮刮卡：点击查看答案，触发判题逻辑
const revealAnswer = () => {
  if (!currentQuestion.value) return
  
  // 判断答案是否正确（未答题也视为错误）
  const isCorrect = sortString(userAnswer.value) === sortString(currentQuestion.value.T)
  
  // 显示答案
  showAnswer.value = true
  answerRevealed.value = true  // 移除遮罩，显示答案
  
  // 如果答错了（包括未答题），标记并加入错题本
  if (!isCorrect) {
    isWrongAnswer.value = true
    // 未答题也视同答错，需要加入错题本
    FavoriteService.AddToErrorBook(userStore.currentUser!.user_id, currentQuestion.value!.id, props.category)
      .catch(err => {
        console.error('Add to error book error:', err)
        ElMessage.error('加入错题本失败')
      })
  } else {
    isWrongAnswer.value = false
  }
}

const nextQuestion = async () => {
  // 如果已经显示答案
  if (showAnswer.value) {
    // 如果是答错后再次点击，才切换题目
    if (isWrongAnswer.value) {
      if (currentIndex.value < totalQuestions.value - 1) {
        await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value + 1)
        currentIndex.value++
        loadQuestion()
      } else {
        ElMessage.success('练习完成！')
        goBack()
      }
      return
    }
    
    // 答对的情况下直接切换
    if (currentIndex.value < totalQuestions.value - 1) {
      await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value + 1)
      currentIndex.value++
      loadQuestion()
    } else {
      ElMessage.success('练习完成！')
      goBack()
    }
    return
  }
  
  // 答案未显示：先点击遮罩（判题 + 加入错题本）
  revealAnswer()
  
  // 如果答错了，不自动切换，等待用户再次点击
  if (isWrongAnswer.value) {
    return
  }
  
  // 答对的情况下，延迟 1 秒后切换
  setTimeout(async () => {
    if (currentIndex.value < totalQuestions.value - 1) {
      await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value + 1)
      currentIndex.value++
      loadQuestion()
    } else {
      ElMessage.success('练习完成！')
      goBack()
    }
  }, 1000)
}

const previousQuestion = async () => {
  // 如果已经显示答案
  if (showAnswer.value) {
    // 如果是答错后再次点击，才切换题目
    if (isWrongAnswer.value) {
      if (currentIndex.value > 0) {
        await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value - 1)
        currentIndex.value--
        loadQuestion()
      }
      return
    }
    
    // 答对的情况下直接切换
    if (currentIndex.value > 0) {
      await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value - 1)
      currentIndex.value--
      loadQuestion()
    }
    return
  }
  
  // 答案未显示：先点击遮罩（判题 + 加入错题本）
  revealAnswer()
  
  // 如果答错了，不自动切换，等待用户再次点击
  if (isWrongAnswer.value) {
    return
  }
  
  // 答对的情况下，延迟 1 秒后切换
  setTimeout(async () => {
    if (currentIndex.value > 0) {
      await PracticeService.SavePracticeProgress(userStore.currentUser!.user_id, props.category, currentIndex.value - 1)
      currentIndex.value--
      loadQuestion()
    }
  }, 1000)
}

const toggleFavorite = async () => {
  try {
    if (isFavorite.value) {
      await FavoriteService.RemoveFromFavorite(userStore.currentUser!.user_id, currentQuestion.value!.id)
      ElMessage.success('已取消收藏')
    } else {
      await FavoriteService.AddToFavorite(userStore.currentUser!.user_id, currentQuestion.value!.id, props.category)
      ElMessage.success('已加入收藏')
    }
    isFavorite.value = !isFavorite.value
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
    console.error('Toggle favorite error:', error)
  }
}

const viewImage = () => {
  if (currentQuestion.value?.imageBase64) {
    currentImage.value = currentQuestion.value.imageBase64
    imageDialogVisible.value = true
  }
}

// 定义 emit
const emit = defineEmits<{
  back: []
}>()

const goBack = () => {
  emit('back')
}

onMounted(() => {
  loadQuestion()
})
</script>

<style scoped>
.practice-container {
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

.info-text {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.question-area {
  flex: 1;
  background: white;
  border-radius: 10px;
  padding: 30px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
}

.question-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 15px;
}

.question-text {
  flex: 1;
  font-size: 18px;
  line-height: 1.8;
  color: #333;
}

.question-number {
  font-weight: 600;
  margin-right: 10px;
}

.question-type {
  color: #666;
  margin-right: 10px;
}

.question-content {
  color: #333;
}

.favorite-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  padding: 8px;
  border-radius: 5px;
  transition: all 0.3s;
  gap: 3px;
  margin-left: 15px;
}

.favorite-btn:hover {
  background: #f5f5f5;
}

.favorite-btn span {
  font-size: 12px;
  color: #666;
}

.question-image {
  text-align: center;
  margin: 0 0 10px 0;
}

.question-image img {
  max-width: 100%;
  max-height: 400px;
  cursor: pointer;
  border-radius: 5px;
}

.options-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0 0 5px 0;
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

.option-item.correct {
  border-color: #67c23a !important;
  background: #f0f9eb !important;
}

.option-item.wrong {
  border-color: #f56c6c !important;
  background: #fef0f0 !important;
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

.answer-feedback {
  margin-top: auto;
  margin-bottom: 30px;
  padding: 15px 20px;
  border-radius: 8px;
  background: #f5f7fa;
  position: relative;
  min-height: 80px;
}

/* 刮刮卡遮罩层 */
.scratch-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
  z-index: 10;
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

.answer-text {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.answer-text.correct {
  color: #67c23a;
}

.answer-text.wrong {
  color: #f56c6c;
}

.correct-answer {
  font-size: 14px;
  margin-bottom: 8px;
  color: #333;
}

.analysis {
  font-size: 13px;
  line-height: 1.5;
  color: #666;
}

.bottom-buttons {
  display: flex;
  justify-content: center;
  gap: 20px;
  padding-top: 15px;
  border-top: 1px solid #e6e6e6;
  margin-top: 15px;
  margin-bottom: 5px;
}

.bottom-buttons .el-button {
  min-width: 120px;
}

.image-container {
  text-align: center;
  padding: 20px;
}
</style>
