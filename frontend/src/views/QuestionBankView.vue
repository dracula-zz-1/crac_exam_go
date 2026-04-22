<template>
  <div class="question-bank-container">
    <div class="toolbar">
      <div class="search-group">
        <el-input
          v-model="searchQuery"
          placeholder="搜索题目编号、内容、答案等"
          clearable
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
      
      <div class="filter-group">
        <el-checkbox v-model="filterLA" @change="handleFilterChange">A 类</el-checkbox>
        <el-checkbox v-model="filterLB" @change="handleFilterChange">B 类</el-checkbox>
        <el-checkbox v-model="filterLC" @change="handleFilterChange">C 类</el-checkbox>
      </div>
    </div>

    <el-table
      :data="tableData"
      style="width: 100%"
      @row-dblclick="handleRowDblClick"
      @cell-mouse-enter="handleCellMouseEnter"
      @cell-mouse-leave="handleCellMouseLeave"
      v-loading="loading"
    >
      <el-table-column label="编号" width="120">
        <template #default="{ row }">
          <span>{{ formatQuestionNumber(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="Q" label="题干" min-width="250" show-overflow-tooltip>
        <template #default="{ row }">
          <el-tooltip :disabled="!shouldShowTooltip(row.Q)" placement="top">
            <template #content>
              <div class="tooltip-content">{{ row.Q }}</div>
            </template>
            <span :class="{ 'truncated': isTextTruncated(row.Q) }">{{ row.Q }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="类别" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.LA === 1" type="success" size="small" style="margin-right: 2px;">A</el-tag>
          <el-tag v-if="row.LB === 1" type="primary" size="small" style="margin-right: 2px;">B</el-tag>
          <el-tag v-if="row.LC === 1" type="warning" size="small">C</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button link type="danger" @click="handleDelete(row)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pageNum"
        :page-size="5"
        layout="total, prev, pager, next"
        :total="total"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑题目"
      width="900px"
      :close-on-click-modal="false"
      top="5vh"
      :style="{ height: '85vh' }"
    >
      <el-form :model="editingQuestion" label-width="90px" style="height: 100%;">
        <!-- JPI 和题目类型、ABC 类显示为一行 -->
        <el-form-item label="编号">
          <span style="font-size: 14px;">{{ formatQuestionNumber(editingQuestion) }}</span>
          <span style="margin-left: 20px; font-size: 14px;">
            类型：{{ editingQuestion.type === 2 ? '多选题' : '单选题' }}
          </span>
          <span style="margin-left: 20px; font-size: 14px;">
            <el-checkbox v-model="editingQuestion.LA" :true-label="1" :false-label="0" style="margin-left: 10px;">A 类</el-checkbox>
            <el-checkbox v-model="editingQuestion.LB" :true-label="1" :false-label="0" style="margin-left: 10px;">B 类</el-checkbox>
            <el-checkbox v-model="editingQuestion.LC" :true-label="1" :false-label="0" style="margin-left: 10px;">C 类</el-checkbox>
          </span>
        </el-form-item>
        
        <el-form-item label="题目内容">
          <el-input
            v-model="editingQuestion.Q"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
        <el-form-item label="选项 A">
          <el-input v-model="editingQuestion.A">
            <template #append>
              <el-checkbox v-model="isCheckedA" @change="updateAnswer('A')" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="选项 B">
          <el-input v-model="editingQuestion.B">
            <template #append>
              <el-checkbox v-model="isCheckedB" @change="updateAnswer('B')" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="选项 C">
          <el-input v-model="editingQuestion.C">
            <template #append>
              <el-checkbox v-model="isCheckedC" @change="updateAnswer('C')" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="选项 D">
          <el-input v-model="editingQuestion.D">
            <template #append>
              <el-checkbox v-model="isCheckedD" @change="updateAnswer('D')" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="图片管理">
          <div v-if="editingQuestion.F && editingQuestion.F.trim() !== ''" style="position: relative; width: 150px; height: 150px; display: flex; align-items: center; justify-content: center; background: #f5f5f5; border-radius: 4px; flex-shrink: 0;">
            <img 
              :src="editingQuestion.F" 
              alt="题目图片" 
              style="max-width: 100%; max-height: 100%; cursor: pointer; object-fit: contain;"
              @click="handleViewImage(editingQuestion)"
            />
            <el-button 
              type="danger" 
              size="small" 
              @click="handleDeleteImage"
              style="position: absolute; top: 5px; right: 5px; width: 24px; height: 24px; padding: 0; border-radius: 50%; display: flex; align-items: center; justify-content: center;"
            >
              <el-icon :size="14"><Close /></el-icon>
            </el-button>
          </div>
          <div v-else>
            <el-button type="primary" size="small" @click="handleUploadImage">上传图片</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 图片查看对话框 -->
    <el-dialog
      v-model="imageDialogVisible"
      width="60vw"
      :close-on-click-modal="true"
      :show-close="false"
      :style="{ background: 'transparent', boxShadow: 'none' }"
      class="image-viewer-dialog"
    >
      <div class="image-container" style="position: relative; width: 100%; height: 60vh; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.8); border-radius: 8px; overflow: hidden;">
        <img :src="currentImage" alt="题目图片" style="max-width: 100%; max-height: 100%; display: block; object-fit: contain;" />
        <el-button 
          type="primary" 
          size="small" 
          @click="handleReplaceImageFromDialog"
          style="position: absolute; top: 10px; right: 10px; width: 40px; height: 40px; padding: 0; border-radius: 50%; display: flex; align-items: center; justify-content: center; z-index: 10; background: rgba(64, 158, 255, 0.8);"
        >
          <el-icon :size="24"><Upload /></el-icon>
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { SettingsService } from '@/api'
import type { Question } from '@/types'

const loading = ref(false)
const tableData = ref<Question[]>([])
const total = ref(0)
const pageNum = ref(1)
const searchQuery = ref('')
const filterLA = ref(false)
const filterLB = ref(false)
const filterLC = ref(false)

const editDialogVisible = ref(false)
const editingQuestion = ref<Question>({} as Question)

// 选项 checkbox 状态
const isCheckedA = ref(false)
const isCheckedB = ref(false)
const isCheckedC = ref(false)
const isCheckedD = ref(false)

const imageDialogVisible = ref(false)
const currentImage = ref('')

const loadTableData = async () => {
  loading.value = true
  try {
    const result = await SettingsService.GetQuestionsPage(
      pageNum.value,
      5,
      searchQuery.value,
      filterLA.value,
      filterLB.value,
      filterLC.value
    )
    if (result && result.data && Array.isArray(result.data)) {
      tableData.value = result.data.map((q: any) => ({
        ...q,
        LA: q.LA === 1 ? 1 : 0,
        LB: q.LB === 1 ? 1 : 0,
        LC: q.LC === 1 ? 1 : 0,
        type: q.type ? Number(q.type) : 0,
      }))
      total.value = result.total
    } else {
      tableData.value = []
      total.value = 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载数据失败')
    console.error('Load data error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pageNum.value = 1
  loadTableData()
}

const handleFilterChange = () => {
  pageNum.value = 1
  loadTableData()
}

const handleCurrentChange = () => {
  loadTableData()
}

const handleRowDblClick = (row: Question) => {
  handleEdit(row)
}

const handleCellMouseEnter = () => {
  // TODO: 显示 tooltip
}

const handleCellMouseLeave = () => {
  // TODO: 隐藏 tooltip
}

const handleViewImage = async (row: Question) => {
  try {
    // 如果已经有 F 数据（base64），直接显示
    if (row.F) {
      currentImage.value = row.F
      imageDialogVisible.value = true
      return
    }
    
    const detail = await SettingsService.GetQuestionDetail(row.id)
    if (detail && (detail as any).image_data) {
      currentImage.value = (detail as any).image_data
      imageDialogVisible.value = true
    } else {
      ElMessage.warning('该题目没有图片')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取图片失败')
    console.error('View image error:', error)
  }
}

const handleUploadImage = async () => {
  try {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    
    input.onchange = async (e: Event) => {
      const target = e.target as HTMLInputElement
      const file = target.files?.[0]
      if (!file) return
      
      const reader = new FileReader()
      reader.onload = async (event) => {
        const base64 = event.target?.result as string
        editingQuestion.value.F = base64
        ElMessage.success('图片上传成功')
      }
      reader.readAsDataURL(file)
    }
    
    input.click()
  } catch (error: any) {
    ElMessage.error(error.message || '上传图片失败')
    console.error('Upload image error:', error)
  }
}

const handleDeleteImage = async () => {
  try {
    await ElMessageBox.confirm('确定要删除这道题目的图片吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    editingQuestion.value.F = ''
    ElMessage.success('图片已删除')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除图片失败')
      console.error('Delete image error:', error)
    }
  }
}

const handleReplaceImageFromDialog = async () => {
  try {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    
    input.onchange = async (e: Event) => {
      const target = e.target as HTMLInputElement
      const file = target.files?.[0]
      if (!file) return
      
      const reader = new FileReader()
      reader.onload = async (event) => {
        const base64 = event.target?.result as string
        // 更新当前显示的图片
        currentImage.value = base64
        // 同时更新编辑表单中的图片
        editingQuestion.value.F = base64
        ElMessage.success('图片上传成功')
        // 关闭图片查看框
        imageDialogVisible.value = false
      }
      reader.readAsDataURL(file)
    }
    
    input.click()
  } catch (error: any) {
    ElMessage.error(error.message || '上传图片失败')
    console.error('Upload image error:', error)
  }
}

const handleEdit = async (row: Question) => {
  editingQuestion.value = { ...row }
  
  // 根据答案 T 字段设置 checkbox 状态
  const answer = (row.T || '').toUpperCase()
  isCheckedA.value = answer.includes('A')
  isCheckedB.value = answer.includes('B')
  isCheckedC.value = answer.includes('C')
  isCheckedD.value = answer.includes('D')
  
  // 根据 LA/LB/LC 设置 ABC 类 checkbox 状态
  editingQuestion.value.LA = row.LA === 1 ? 1 : 0
  editingQuestion.value.LB = row.LB === 1 ? 1 : 0
  editingQuestion.value.LC = row.LC === 1 ? 1 : 0
  
  // 确保图片 URL 有正确的格式
  if (editingQuestion.value.F && !editingQuestion.value.F.startsWith('data:image')) {
    // 如果是纯 base64 数据，添加前缀
    editingQuestion.value.F = `data:image/png;base64,${editingQuestion.value.F}`
  }
  
  editDialogVisible.value = true
}

const updateAnswer = (_option: string) => {
  const checkedMap: Record<string, boolean> = {
    'A': isCheckedA.value,
    'B': isCheckedB.value,
    'C': isCheckedC.value,
    'D': isCheckedD.value
  }
  
  // 构建答案字符串
  const answers = Object.entries(checkedMap)
    .filter(([_, checked]) => checked)
    .map(([option]) => option)
    .join('')
  
  editingQuestion.value.T = answers
  
  // 根据选中数量确定单选还是多选
  const selectedCount = answers.length
  editingQuestion.value.type = selectedCount > 1 ? 2 : 1
}

const handleSaveEdit = async () => {
  try {
    // 根据 checkbox 状态更新答案
    updateAnswer('')
    
    // 创建纯对象，避免响应式代理问题
    const saveData: any = {
      id: editingQuestion.value.id,
      J: editingQuestion.value.J,
      P: editingQuestion.value.P,
      I: editingQuestion.value.I,
      Q: editingQuestion.value.Q,
      T: editingQuestion.value.T,
      A: editingQuestion.value.A,
      B: editingQuestion.value.B,
      C: editingQuestion.value.C,
      D: editingQuestion.value.D,
      F: editingQuestion.value.F,
      type: editingQuestion.value.type,
      // LA/LB/LC 已经是数字类型，直接使用
      LA: editingQuestion.value.LA,
      LB: editingQuestion.value.LB,
      LC: editingQuestion.value.LC
    }
    
    console.log('保存数据:', saveData)
    
    const questionID = editingQuestion.value.id
    await SettingsService.UpdateQuestion(questionID, saveData)
    
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    loadTableData()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
    console.error('Save edit error:', error)
  }
}

const handleDelete = async (row: Question) => {
  try {
    await ElMessageBox.confirm('确定要删除这道题目吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const questionID = parseInt(row.J)
    await SettingsService.DeleteQuestion(questionID)
    
    ElMessage.success('删除成功')
    loadTableData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
      console.error('Delete error:', error)
    }
  }
}

const shouldShowTooltip = (text: string): boolean => {
  // 判断是否需要显示 tooltip
  if (!text) return false
  return text.length > 100
}

const isTextTruncated = (text: string): boolean => {
  // 判断文本是否被截断 - 根据实际显示宽度判断
  if (!text) return false
  return text.length > 100
}

const formatQuestionNumber = (row: Question): string => {
  // 拼接 J、P、I 三个字段，用逗号分隔
  return [row.J, row.P, row.I].filter(Boolean).join(',')
}

onMounted(() => {
  loadTableData()
})
</script>

<style scoped>
.question-bank-container {
  padding: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 减少编辑对话框 form-item 的间距 */
:deep(.el-dialog__body) {
  padding: 10px 20px;
}

:deep(.el-form-item) {
  margin-bottom: 10px;
}

:deep(.el-form-item__label) {
  line-height: 32px;
}

/* 图片管理项特殊处理 */
:deep(.el-form-item[label="图片管理"]) {
  margin-bottom: 15px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  gap: 20px;
  flex-shrink: 0;
}

.search-group {
  display: flex;
  gap: 10px;
  flex: 1;
  max-width: 500px;
}

.filter-group {
  display: flex;
  gap: 15px;
}

.action-group {
  display: flex;
  gap: 10px;
}

.pagination-container {
  margin-top: 10px;
  display: flex;
  justify-content: center;
  flex-shrink: 0;
}

.tooltip-content {
  max-width: 500px;
  white-space: pre-wrap;
  word-break: break-word;
}

.truncated {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-viewer-dialog {
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-viewer-dialog .el-dialog {
  margin: 0 !important;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.image-container {
  text-align: center;
}

/* 让表格占据剩余空间 */
.el-table {
  flex: 1;
  overflow: auto;
}

/* 移除表格滚动条 */
.el-table__body-wrapper::-webkit-scrollbar {
  display: none;
}

/* 调整表格行高 */
:deep(.el-table__row) {
  height: 40px !important;
}

/* 调整表格头部高度 */
:deep(.el-table__header-wrapper) {
  height: 40px;
}
</style>
