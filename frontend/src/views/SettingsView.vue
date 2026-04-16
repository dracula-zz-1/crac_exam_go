<template>
  <div class="settings-container">
    <!-- 左侧导航菜单 -->
    <div class="nav-menu">
      <div
        v-for="(item, index) in menuItems"
        :key="index"
        :class="['nav-item', { active: currentMenu === index }]"
        @click="switchMenu(index)"
      >
        <el-icon :size="20">
          <component :is="item.icon" />
        </el-icon>
        <span>{{ item.name }}</span>
      </div>
    </div>

    <!-- 右侧内容区域 -->
    <div class="content-area">
      <!-- 导入题库 -->
      <div v-show="currentMenu === 0" class="settings-content no-title">
        <div class="import-section">
          <div class="import-area" @dragover.prevent @dragenter.prevent @drop.prevent="handleWailsDrop">
            <p>拖拽文件到此处，或点击选择文件</p>
            <el-button type="primary" @click="selectFile">选择文件</el-button>
            <div class="format-hint">
              <p>支持格式：JSON、CSV、Excel (.xlsx/.xls) 格式</p>
              <p>PDF 导入请使用包含 5 个 PDF 文件的 ZIP 包</p>
            </div>
          </div>
          
          <!-- 导入结果统计表格 -->
          <div class="import-stats">
            <div v-if="!showImportStats" class="empty-hint">
              暂无导入数据，请导入题库后查看统计信息
            </div>
            <el-table v-else :data="importStatsData" border style="width: 100%">
              <el-table-column prop="category" label="分类" width="100" />
              <el-table-column prop="total" label="总题数" width="100" align="right" />
              <el-table-column prop="single" label="单选题数" width="100" align="right">
                <template #default="{ row }">
                  <span v-if="row.single !== null">{{ row.single }}</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column prop="multiple" label="多选题数" width="100" align="right">
                <template #default="{ row }">
                  <span v-if="row.multiple !== null">{{ row.multiple }}</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column prop="withImage" label="含图题数" width="100" align="right">
                <template #default="{ row }">
                  <span v-if="row.withImage !== null">{{ row.withImage }}</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
            
            <!-- 自定义合计行 -->
            <div v-if="showImportStats" class="summary-row">
              <span class="summary-label">总题数：</span>
              <span class="summary-value">{{ importTotalCount }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 考试统计 -->
      <div v-show="currentMenu === 1" class="settings-content">
        <h1 class="content-title">考试统计</h1>
        
        <div class="statistics-section">
          <!-- 选择区域 -->
          <div class="chart-options">
            <div class="option-row">
              <el-radio-group v-model="selectedCategory" size="default">
                <el-radio-button value="A">A 类</el-radio-button>
                <el-radio-button value="B">B 类</el-radio-button>
                <el-radio-button value="C">C 类</el-radio-button>
              </el-radio-group>
              
              <el-radio-group v-model="selectedTimeRange" size="default">
                <el-radio-button value="7days">近 7 天</el-radio-button>
                <el-radio-button value="30days">近 30 天</el-radio-button>
                <el-radio-button value="180days">近半年</el-radio-button>
              </el-radio-group>
            </div>
          </div>
          
          <!-- 分数趋势图 -->
          <div class="score-trend-container" v-show="showScoreChart">
            <div ref="scoreChartRef" class="score-chart"></div>
          </div>
        </div>
      </div>

      <!-- 数据清理 -->
      <div v-show="currentMenu === 2" class="settings-content">
        <h1 class="content-title">数据清理</h1>
        
        <div class="clear-section">
          <el-alert
            title="警告"
            description="此操作将清空所选数据，请谨慎选择！"
            type="warning"
            :closable="false"
            show-icon
          />
          
          <div class="clear-options" style="margin: 20px 0;">
            <el-checkbox-group v-model="clearOptions" style="display: flex; flex-direction: column; gap: 15px;">
              <el-checkbox label="user_data">用户数据（练习进度、错题本、收藏夹）</el-checkbox>
              <el-checkbox label="practice_progress">练习进度（所有类别的练习进度）</el-checkbox>
              <el-checkbox label="question_bank">题库数据（所有题目）</el-checkbox>
            </el-checkbox-group>
          </div>
          
          <div class="action-area">
            <el-button type="danger" @click="confirmClearData" :disabled="clearOptions.length === 0">
              <el-icon><Delete /></el-icon>
              清空选中数据
            </el-button>
          </div>
        </div>
      </div>

      <!-- 题库管理 -->
      <div v-if="currentMenu === 3" class="settings-content full-height">
        <QuestionBankView />
      </div>

      <!-- 关于 -->
      <div v-show="currentMenu === 4" class="settings-content">
        <h1 class="content-title">关于</h1>
        
        <div class="about-section">
          <div class="app-info">
            <div class="app-logo">
              <el-icon :size="80" color="#409EFF"><Reading /></el-icon>
            </div>
            <h2>业余无线电模拟考试系统</h2>
            <p>版本：1.0.0</p>
            <p>作者：BA4RHH</p>
            <p class="copyright">Copyright © 2024-2025 BA4RHH. All rights reserved.</p>
          </div>
          
          <div class="app-description">
            <h3>系统介绍</h3>
            <p>本系统用于业余无线电操作技术能力模拟考试，支持 A、B、C 三类题目的练习和考试。</p>
            <h3>功能特性</h3>
            <ul>
              <li>逐题练习模式，支持进度保存</li>
              <li>全真模拟考试，限时答题</li>
              <li>错题自动收集，针对性复习</li>
              <li>题目收藏功能，重点题目重点复习</li>
              <li>题库管理，支持多种格式导入</li>
              <li>考试统计分析，学习情况一目了然</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useUserStore } from '@/stores/user'
import { SettingsService, StatisticsService, PracticeService } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Reading } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import QuestionBankView from './QuestionBankView.vue'

import { ProcessUnifiedData, ProcessFileContent } from '@wailsjs/go/services/ImportService'

import { OnFileDrop, OnFileDropOff } from '@wailsjs/runtime/runtime'

const userStore = useUserStore()

const currentMenu = ref(0)
const clearOptions = ref<string[]>(['user_data'])

// 导入结果统计
const showImportStats = ref(false)
const importStatsData = ref<any[]>([])

// 考试统计相关
const selectedCategory = ref<string>('A')
const selectedTimeRange = ref<string>('7days')
const showScoreChart = ref(false)
const scoreChartRef = ref<HTMLElement | null>(null)
let scoreChart: echarts.ECharts | null = null

const menuItems = [
  { name: '导入题库', icon: 'Upload' },
  { name: '考试统计', icon: 'TrendCharts' },
  { name: '数据清理', icon: 'Delete' },
  { name: '题库管理', icon: 'Document' },
  { name: '关于', icon: 'InfoFilled' }
]

const switchMenu = async (index: number) => {
  currentMenu.value = index
  
  // 如果切换到考试统计，初始化并加载默认数据
  if (index === 1) {
    selectedCategory.value = 'A'
    selectedTimeRange.value = '7days'
    showScoreChart.value = false
    await nextTick()
    generateScoreTrendChart()
  }
}

// 注意：需要在全局声明 Wails Go 对象
declare global {
  interface Window {
    go?: any
  }
}

const selectFile = async () => {
  try {
    // 使用 HTML5 文件选择
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json,.csv,.xlsx,.xls,.zip'
    
    input.onchange = async (e: Event) => {
      const target = e.target as HTMLInputElement
      const file = target.files?.[0]
      if (!file) return
      
      ElMessage.info(`开始导入文件：${file.name}`)
      
      const fileType = file.name.split('.').pop()?.toLowerCase() || ''
      
      try {
        // 读取文件内容
        const reader = new FileReader()
        reader.onload = async (event) => {
          try {
            let content: string
            
            // 二进制文件使用 base64 编码
            if (fileType === 'xlsx' || fileType === 'xls' || fileType === 'zip') {
              const arrayBuffer = event.target?.result as ArrayBuffer
              // 使用正确的 base64 编码方法
              const bytes = new Uint8Array(arrayBuffer)
              let binary = ''
              for (let i = 0; i < bytes.byteLength; i++) {
                binary += String.fromCharCode(bytes[i])
              }
              content = btoa(binary)
            } else {
              // 文本文件直接使用内容
              content = event.target?.result as string
            }
            
            // 调用后端导入服务（传递文件内容和类型）
            const result = await ProcessFileContent(content, fileType)
            
            if (result && result.success) {
              ElMessage.success(`导入成功！共导入 ${result.imported_count}/${result.total_count} 道题目`)
              
              // 显示统计表格
              showImportStatsTable(result.stats)
            } else {
              ElMessage.error(result?.message || '导入失败')
            }
          } catch (error: any) {
            ElMessage.error('导入失败：' + error.message)
            console.error('Import error:', error)
          }
        }
        
        // 根据文件类型选择读取方式
        if (fileType === 'xlsx' || fileType === 'xls' || fileType === 'zip') {
          reader.readAsArrayBuffer(file)
        } else {
          reader.readAsText(file)
        }
      } catch (error: any) {
        ElMessage.error('读取文件失败：' + error.message)
        console.error('File read error:', error)
      }
    }
    
    input.click()
  } catch (error: any) {
    console.error('导入失败:', error)
    ElMessage.error('导入失败：' + error.message)
  }
}

// 处理 Wails 拖拽事件
const handleWailsDrop = async (event: DragEvent) => {
  event.preventDefault()
  console.log('handleWailsDrop 被调用')
  
  // 尝试获取文件路径
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    console.log('拖拽的文件:', file.name, file)
    
    // 对于本地文件，尝试获取路径
    const filePath = (file as any).path
    if (filePath) {
      console.log('获取到文件路径:', filePath)
      const fileName = file.name
      
      ElMessage.info(`开始导入文件：${fileName}`)
      
      // 直接调用后端导入服务（传递真实文件路径）
      ProcessUnifiedData(filePath).then(result => {
        console.log('导入结果:', result)
        if (result && result.success) {
          ElMessage.success(`导入成功！共导入 ${result.imported_count}/${result.total_count} 道题目`)
          
          console.log('result.stats:', result.stats)
          // 显示统计表格
          showImportStatsTable(result.stats)
        } else {
          ElMessage.error(result?.message || '导入失败')
        }
      }).catch(error => {
        ElMessage.error('导入失败：' + error.message)
        console.error('Import error:', error)
      })
    } else {
      // 没有路径，使用 FileReader 读取内容
      ElMessage.info(`开始导入文件：${file.name}`)
      const fileType = file.name.split('.').pop()?.toLowerCase() || ''
      
      try {
        const reader = new FileReader()
        reader.onload = async (event) => {
          try {
            let content: string
            
            if (fileType === 'xlsx' || fileType === 'xls' || fileType === 'zip') {
              const arrayBuffer = event.target?.result as ArrayBuffer
              const bytes = new Uint8Array(arrayBuffer)
              let binary = ''
              for (let i = 0; i < bytes.byteLength; i++) {
                binary += String.fromCharCode(bytes[i])
              }
              content = btoa(binary)
            } else {
              content = event.target?.result as string
            }
            
            const result = await ProcessFileContent(content, fileType)
            
            console.log('导入结果:', result)
            
            if (result && result.success) {
              ElMessage.success(`导入成功！共导入 ${result.imported_count}/${result.total_count} 道题目`)
              
              console.log('result.stats:', result.stats)
              // 显示统计表格
              showImportStatsTable(result.stats)
            } else {
              ElMessage.error(result?.message || '导入失败')
            }
          } catch (error: any) {
            ElMessage.error('导入失败：' + error.message)
            console.error('Import error:', error)
          }
        }
        
        if (fileType === 'xlsx' || fileType === 'xls' || fileType === 'zip') {
          reader.readAsArrayBuffer(file)
        } else {
          reader.readAsText(file)
        }
      } catch (error: any) {
        ElMessage.error('读取文件失败：' + error.message)
        console.error('File read error:', error)
      }
    }
  }
}

// 使用 Wails 的 OnFileDrop API 实现真正的拖拽导入（获取真实路径）
const enableFileDrop = () => {
  console.log('启用文件拖拽监听器...')
  // 注册文件拖拽监听器
  OnFileDrop((_x: number, _y: number, paths: string[]) => {
    console.log('检测到文件拖拽:', paths)
    if (paths && paths.length > 0) {
      const filePath = paths[0]
      const fileName = filePath.split(/[/\\]/).pop() || ''
      
      ElMessage.info(`开始导入文件：${fileName}`)
      
      // 直接调用后端导入服务（传递真实文件路径）
      ProcessUnifiedData(filePath).then(result => {
        console.log('导入结果:', result)
        if (result && result.success) {
          ElMessage.success(`导入成功！共导入 ${result.imported_count}/${result.total_count} 道题目`)
          
          console.log('result.stats:', result.stats)
          // 显示统计表格
          showImportStatsTable(result.stats)
        } else {
          ElMessage.error(result?.message || '导入失败')
        }
      }).catch(error => {
        ElMessage.error('导入失败：' + error.message)
        console.error('Import error:', error)
      })
    }
  }, false)
}

// 显示导入统计表格
const showImportStatsTable = (stats: any) => {
  console.log('导入统计:', stats)
  
  if (!stats) {
    console.warn('没有统计信息')
    showImportStats.value = false
    return
  }
  
  importStatsData.value = [
    { category: 'A 类', total: stats.a_total || 0, single: stats.a_single || 0, multiple: stats.a_multiple || 0, withImage: stats.a_with_image || 0 },
    { category: 'B 类', total: stats.b_total || 0, single: stats.b_single || 0, multiple: stats.b_multiple || 0, withImage: stats.b_with_image || 0 },
    { category: 'C 类', total: stats.c_total || 0, single: stats.c_single || 0, multiple: stats.c_multiple || 0, withImage: stats.c_with_image || 0 }
  ]
  
  // 保存总题数
  importTotalCount.value = stats.total || 0
  
  console.log('统计表格数据:', importStatsData.value)
  showImportStats.value = true
}

// 导入总题数
const importTotalCount = ref(0)

onMounted(() => {
  // 启用文件拖拽
  enableFileDrop()
})

onUnmounted(() => {
  // 清理拖拽监听器
  OnFileDropOff()
})

// 生成考试成绩趋势图
const generateScoreTrendChart = async () => {
  try {
    const userID = userStore.currentUser?.user_id
    if (!userID) {
      ElMessage.warning('请先登录')
      return
    }
    
    // 获取考试数据
    const examData = await StatisticsService.GetExamData(userID, selectedCategory.value, selectedTimeRange.value)
    
    if (!examData || examData.length === 0) {
      ElMessage.warning('没有找到考试数据')
      showScoreChart.value = false
      return
    }
    
    showScoreChart.value = true
    await nextTick()
    
    // 初始化图表
    if (!scoreChartRef.value) return
    
    if (!scoreChart) {
      scoreChart = echarts.init(scoreChartRef.value)
    }
    
    // 处理数据 - 生成完整日期范围（基于当前时间和所选时间段）
    const allDates: string[] = []
    const endDate = new Date()  // 当前时间
    let daysToSubtract = 0
    
    // 根据选择的时间段计算开始日期
    if (selectedTimeRange.value === '7days') {
      daysToSubtract = 6  // 近 7 天：从今天往前推 6 天（包括今天共 7 天）
    } else if (selectedTimeRange.value === '30days') {
      daysToSubtract = 29  // 近 30 天：从今天往前推 29 天
    } else if (selectedTimeRange.value === '180days') {
      daysToSubtract = 179  // 近半年：从今天往前推 179 天
    }
    
    const startDate = new Date()
    startDate.setDate(startDate.getDate() - daysToSubtract)
    
    // 生成从开始到结束的所有日期
    const currentDate = new Date(startDate)
    while (currentDate <= endDate) {
      const dateStr = currentDate.toLocaleDateString('zh-CN', {
        month: '2-digit',
        day: '2-digit'
      })
      allDates.push(dateStr)
      currentDate.setDate(currentDate.getDate() + 1)
    }
    
    console.log('日期范围:', {
      startDate: startDate.toLocaleDateString('zh-CN'),
      endDate: endDate.toLocaleDateString('zh-CN'),
      totalDays: allDates.length,
      allDates: allDates
    })
    
    // 处理数据 - 只处理选中类别的数据
    const selectedData: { date: string; score: number | null }[] = []
    
    // 先按日期整理已有数据（只处理选中的类别）
    const examDataByDate: Record<string, number | null> = {}
    examData.forEach((exam: any) => {
      const date = new Date(exam.exam_date as any).toLocaleDateString('zh-CN', {
        month: '2-digit',
        day: '2-digit'
      })
      if (exam.category === selectedCategory.value) {
        examDataByDate[date] = exam.score
      }
    })
    
    // 为每个日期填充数据（有考试填分数，没有填 null）
    allDates.forEach(date => {
      if (examDataByDate[date] !== undefined) {
        selectedData.push({ date, score: examDataByDate[date] })
      } else {
        selectedData.push({ date, score: null })
      }
    })
    
    // 构建 series - 只显示选中的类别
    const series = []
    const categoryColors: Record<string, string> = { A: '#409EFF', B: '#67C23A', C: '#E6A23C' }
    const categoryNames: Record<string, string> = { A: 'A 类', B: 'B 类', C: 'C 类' }
    
    if (selectedData.length > 0) {
      series.push({
        name: categoryNames[selectedCategory.value],
        type: 'line',
        data: selectedData.map(item => ({
          value: item.score,
          date: item.date
        })),
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        // 确保所有数据点都显示标记，包括只有一个数据点的情况
        symbolRotate: 0,
        showAllSymbol: true,
        itemStyle: {
          color: categoryColors[selectedCategory.value]
        },
        label: {
          show: (params: any) => params.value !== null,  // 只在有数据时显示标签
          position: 'top',
          formatter: '{c}'
        },
        connectNulls: false  // 不连接空数据点
      })
    }
    
    // 设置图表配置
    const option = {
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          const date = params[0].data.date
          const param = params[0]
          if (param.value !== null) {
            return `<div style="font-weight: bold;">${date}</div>
                    <div style="margin: 5px 0;">${param.marker} ${param.seriesName}: ${param.value}分</div>`
          } else {
            return `<div style="font-weight: bold;">${date}</div>
                    <div style="margin: 5px 0; color: #999;">无考试记录</div>`
          }
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '10%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: allDates,  // 使用完整的日期数组
        axisLabel: {
          rotate: 45,
          interval: getLabelInterval(allDates.length),  // 根据天数计算间隔
          formatter: (value: string) => value
        }
      },
      yAxis: {
        type: 'value',
        name: '成绩 (分)',
        min: 0,
        max: getCategoryMaxScore(selectedCategory.value),
        interval: 10
      },
      series: series
    }
    
    scoreChart.setOption(option)
    
    console.log('图表数据:', {
      allDates: allDates,
      totalDates: allDates.length,
      seriesData: series[0]?.data || []
    })
    
    // 响应式调整
    window.addEventListener('resize', () => {
      scoreChart?.resize()
    })
  
  ElMessage.success('统计图表生成成功')
} catch (error: any) {
  console.error('生成统计图表失败:', error)
  ElMessage.error('生成统计图表失败：' + (error.message || '未知错误'))
}
}

// 监听选择和时间的变化，自动刷新图表
watch([selectedCategory, selectedTimeRange], () => {
  if (currentMenu.value === 1) {
    generateScoreTrendChart()
  }
}, { immediate: false })

// 获取类别最高分
const getCategoryMaxScore = (category: string): number => {
  switch (category) {
    case 'A': return 40
    case 'B': return 60
    case 'C': return 90
    default: return 100
  }
}

// 根据总天数计算标签显示间隔，目标是显示大约 7 个标签
const getLabelInterval = (totalDays: number): number => {
  if (totalDays <= 7) {
    return 0  // 7 天以内显示所有标签
  } else {
    // 计算间隔，使得显示的标签数约为 7 个
    // 例如：30 天 -> 间隔 4 -> 显示 8 个标签
    //      180 天 -> 间隔 25 -> 显示 8 个标签
    return Math.ceil(totalDays / 7) - 1
  }
}

const confirmClearData = async () => {
  try {
    if (clearOptions.value.length === 0) {
      ElMessage.warning('请至少选择一项要清理的数据')
      return
    }
    
    const clearUserData = clearOptions.value.includes('user_data')
    const clearPracticeProgress = clearOptions.value.includes('practice_progress')
    const clearQuestionBank = clearOptions.value.includes('question_bank')
    
    let confirmMessage = '确定要清空所选数据吗？此操作不可恢复！'
    if (clearUserData && clearPracticeProgress && clearQuestionBank) {
      confirmMessage = '确定要清空用户数据、练习进度和题库数据吗？此操作不可恢复！'
    } else if (clearUserData && clearPracticeProgress) {
      confirmMessage = '确定要清空用户数据和练习进度吗？此操作不可恢复！'
    } else if (clearUserData && clearQuestionBank) {
      confirmMessage = '确定要清空用户数据和题库数据吗？此操作不可恢复！'
    } else if (clearPracticeProgress && clearQuestionBank) {
      confirmMessage = '确定要清空练习进度和题库数据吗？此操作不可恢复！'
    } else if (clearUserData) {
      confirmMessage = '确定要清空用户数据（练习进度、错题本、收藏夹）吗？此操作不可恢复！'
    } else if (clearPracticeProgress) {
      confirmMessage = '确定要清空练习进度（所有类别的练习进度）吗？此操作不可恢复！'
    } else if (clearQuestionBank) {
      confirmMessage = '确定要清空题库数据（所有题目）吗？此操作不可恢复！'
    }
    
    await ElMessageBox.confirm(confirmMessage, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const userID = userStore.currentUser?.user_id
    if (!userID) {
      ElMessage.warning('请先登录')
      return
    }
    
    // 按顺序执行清理操作
    if (clearQuestionBank) {
      await SettingsService.ClearQuestionBank()
    }
    if (clearPracticeProgress) {
      await PracticeService.ResetProgress(userID, 'all')
    }
    if (clearUserData) {
      await SettingsService.ClearUserData(userID)
    }
    
    ElMessage.success('数据已清空')
    clearOptions.value = []
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '清空失败')
      console.error('Clear data error:', error)
    }
  }
}
</script>

<style scoped>
.settings-container {
  display: flex;
  height: 100%;
  background: #f5f7fa;
}

.nav-menu {
  width: 200px;
  background: white;
  padding: 20px 0;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px 25px;
  cursor: pointer;
  transition: all 0.3s;
  color: #666;
  font-size: 15px;
}

.nav-item:hover {
  background: #f5f7fa;
  color: #409EFF;
}

.nav-item.active {
  background: #ecf5ff;
  color: #409EFF;
  border-right: 3px solid #409EFF;
}

.content-area {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
}

.settings-content {
  background: white;
  border-radius: 10px;
  padding: 15px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.settings-content.full-height {
  padding: 0;
  background: transparent;
  box-shadow: none;
}

.content-title {
  font-size: 24px;
  color: #333;
  margin-bottom: 30px;
  padding-bottom: 15px;
  border-bottom: 2px solid #f0f0f0;
}

.settings-content.no-title {
  padding: 15px;
}

.settings-content.no-title .content-title {
  display: none;
}

.import-section {
  max-width: 800px;
}

.import-area {
  margin-top: 0;
  padding: 15px 20px;
  border: 2px dashed #dcdfe6;
  border-radius: 10px;
  text-align: center;
  transition: all 0.3s;
  height: 200px;
}

.import-area:hover {
  border-color: #409EFF;
  background: #f5f7fa;
}

.import-area p {
  margin: 20px 0;
  color: #666;
}

.format-hint {
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 5px;
  font-size: 13px;
  color: #666;
}

.format-hint p {
  margin: 5px 0;
}

.import-stats {
  margin-top: 5px;
  padding: 20px;
  background: #fafafa;
  border-radius: 10px;
  min-height: 200px;
}

.import-stats h3 {
  font-size: 18px;
  color: #333;
  margin-bottom: 15px;
  text-align: center;
}

.import-stats .empty-hint {
  text-align: center;
  color: #999;
  padding: 40px 0;
  font-size: 14px;
}

.import-stats .summary-row {
  margin-top: 15px;
  padding: 15px 20px;
  background: #f5f7fa;
  border-radius: 5px;
  text-align: right;
  font-size: 16px;
}

.import-stats .summary-label {
  font-weight: bold;
  color: #333;
  margin-right: 10px;
}

.import-stats .summary-value {
  font-weight: bold;
  color: #409EFF;
  font-size: 18px;
}

.statistics-section {
  margin-top: 30px;
}

.category-chart-container {
  margin-top: 40px;
  padding: 30px;
  background: #fafafa;
  border-radius: 10px;
}

.chart-title {
  font-size: 18px;
  color: #333;
  margin-bottom: 20px;
  text-align: center;
}

.category-chart {
  width: 100%;
  height: 400px;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 30px;
  border-radius: 10px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.stat-card:nth-child(2) {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  box-shadow: 0 4px 12px rgba(240, 147, 251, 0.3);
}

.stat-card:nth-child(3) {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  box-shadow: 0 4px 12px rgba(79, 172, 254, 0.3);
}

/* 小型统计卡片 - 减少高度 */
.stat-card-small {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 15px 20px;
  border-radius: 8px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.stat-card-small:nth-child(2) {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  box-shadow: 0 2px 8px rgba(240, 147, 251, 0.3);
}

.stat-card-small:nth-child(3) {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  box-shadow: 0 2px 8px rgba(79, 172, 254, 0.3);
}

.stat-value {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 10px;
}

.stat-label {
  font-size: 16px;
  opacity: 0.9;
}

/* 小型统计数值和标签 */
.stat-value-small {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 5px;
}

.stat-label-small {
  font-size: 13px;
  opacity: 0.9;
}

.progress-section {
  max-width: 600px;
}

.progress-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.category-select {
  margin: 20px 0;
}

/* 考试统计图表样式 */
.chart-options {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 10px;
  margin-bottom: 20px;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.score-trend-container {
  background: white;
  padding: 20px;
  border-radius: 10px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.score-chart {
  width: 100%;
  height: 265px;
  margin-top: 20px;
}

.category-select .label {
  margin-right: 15px;
  font-weight: 500;
}

.action-area {
  margin-top: 20px;
  text-align: right;
}

.clear-section {
  max-width: 600px;
}

.clear-section .action-area {
  margin-top: 30px;
}

.question-bank-placeholder {
  text-align: center;
  padding: 60px 20px;
}

.about-section {
  max-width: 800px;
}

.app-info {
  text-align: center;
  padding: 40px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 10px;
  margin-bottom: 30px;
}

.app-logo {
  margin-bottom: 20px;
}

.app-info h2 {
  font-size: 24px;
  margin-bottom: 10px;
}

.app-info p {
  font-size: 16px;
  opacity: 0.9;
  margin: 5px 0;
}

.copyright {
  margin-top: 15px;
  font-size: 14px;
  opacity: 0.8;
}

.app-description {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 10px;
}

.app-description h3 {
  font-size: 18px;
  color: #333;
  margin: 20px 0 10px;
}

.app-description h3:first-child {
  margin-top: 0;
}

.app-description ul {
  list-style: disc;
  margin-left: 20px;
  line-height: 2;
  color: #666;
}
</style>
