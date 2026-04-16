/**
 * Wails runtime 调用封装
 * 提供统一的错误处理和日志记录
 */

import * as UserServiceBindings from '../../wailsjs/go/services/UserService'
import * as QuestionsBankServiceBindings from '../../wailsjs/go/services/QuestionsBankService'
import * as SettingsServiceBindings from '../../wailsjs/go/services/SettingsService'
import * as PracticeServiceBindings from '../../wailsjs/go/services/PracticeService'
import * as ExamServiceBindings from '../../wailsjs/go/services/ExamService'
import * as FavoriteServiceBindings from '../../wailsjs/go/services/FavoriteService'
import * as StatisticsServiceBindings from '../../wailsjs/go/services/StatisticsService'
import type { services, models } from '../../wailsjs/go/models'

// 用户服务 API 调用
export const UserService = {
  Login: async (username: string, idCard: string): Promise<services.UserLoginResponse> => {
    try {
      return await UserServiceBindings.Login(username, idCard)
    } catch (error) {
      console.error('UserService.Login error:', error)
      throw error
    }
  },
  
  Logout: async (): Promise<void> => {
    try {
      // 前端登出只需要清除本地状态
      // 后端没有 Logout 方法
    } catch (error) {
      console.error('UserService.Logout error:', error)
      throw error
    }
  }
}

// 题库服务 API 调用
export const QuestionsBankService = {
  GetPageData: async (
    pageNum: number,
    pageSize: number,
    searchQuery: string,
    filterLA: boolean,
    filterLB: boolean,
    filterLC: boolean
  ): Promise<services.PageDataResult> => {
    try {
      return await QuestionsBankServiceBindings.GetPageData(
        pageNum,
        pageSize,
        searchQuery,
        filterLA,
        filterLB,
        filterLC
      )
    } catch (error) {
      console.error('QuestionsBankService.GetPageData error:', error)
      throw error
    }
  },
  
  GetQuestionDetail: async (questionID: number): Promise<models.Question> => {
    try {
      return await QuestionsBankServiceBindings.GetQuestionByID(questionID)
    } catch (error) {
      console.error('QuestionsBankService.GetQuestionDetail error:', error)
      throw error
    }
  },
  
  UpdateQuestion: async (questionID: number, updatedData: any): Promise<void> => {
    try {
      await QuestionsBankServiceBindings.UpdateQuestion(questionID, updatedData)
    } catch (error) {
      console.error('QuestionsBankService.UpdateQuestion error:', error)
      throw error
    }
  },
  
  DeleteQuestion: async (questionID: number): Promise<void> => {
    try {
      await QuestionsBankServiceBindings.DeleteQuestion(questionID)
    } catch (error) {
      console.error('QuestionsBankService.DeleteQuestion error:', error)
      throw error
    }
  }
}

// 设置服务 API 调用
export const SettingsService = {
  ImportQuestions: async (filePath: string): Promise<services.ImportResult> => {
    try {
      return await SettingsServiceBindings.ImportQuestions(filePath)
    } catch (error) {
      console.error('SettingsService.ImportQuestions error:', error)
      throw error
    }
  },
  
  ClearUserData: async (userID: number): Promise<void> => {
    try {
      await SettingsServiceBindings.ClearUserData(userID)
    } catch (error) {
      console.error('SettingsService.ClearUserData error:', error)
      throw error
    }
  },
  
  ClearQuestionBank: async (): Promise<void> => {
    try {
      await SettingsServiceBindings.ClearQuestionBank()
    } catch (error) {
      console.error('SettingsService.ClearQuestionBank error:', error)
      throw error
    }
  },
  
  GetQuestionsPage: async (
    pageNum: number,
    pageSize: number,
    searchQuery: string,
    filterLA: boolean,
    filterLB: boolean,
    filterLC: boolean
  ): Promise<{ data: models.Question[], total: number }> => {
    try {
      const result = await SettingsServiceBindings.GetQuestionsPage(
        pageNum,
        pageSize,
        searchQuery,
        filterLA,
        filterLB,
        filterLC
      )
      // 类型断言，因为 Wails 返回的是 Record<string, any>
      return result as { data: models.Question[], total: number }
    } catch (error) {
      console.error('SettingsService.GetQuestionsPage error:', error)
      throw error
    }
  },
  
  UpdateQuestion: async (questionID: number, updatedData: any): Promise<void> => {
    try {
      await SettingsServiceBindings.UpdateQuestion(questionID, updatedData)
    } catch (error) {
      console.error('SettingsService.UpdateQuestion error:', error)
      throw error
    }
  },
  
  DeleteQuestion: async (questionID: number): Promise<void> => {
    try {
      await SettingsServiceBindings.DeleteQuestion(questionID)
    } catch (error) {
      console.error('SettingsService.DeleteQuestion error:', error)
      throw error
    }
  },
  
  GetQuestionDetail: async (questionID: number): Promise<models.Question> => {
    try {
      return await SettingsServiceBindings.GetQuestionByID(questionID)
    } catch (error) {
      console.error('SettingsService.GetQuestionDetail error:', error)
      throw error
    }
  },
  
  GetAppInfo: async (): Promise<services.AppInfo> => {
    try {
      return await SettingsServiceBindings.GetAppInfo()
    } catch (error) {
      console.error('SettingsService.GetAppInfo error:', error)
      throw error
    }
  },
  
  GetExamConfig: async (category: string): Promise<any> => {
    try {
      return await SettingsServiceBindings.GetExamConfig(category)
    } catch (error) {
      console.error('SettingsService.GetExamConfig error:', error)
      throw error
    }
  },
  
  GetAllExamConfigs: async (): Promise<Record<string, any>> => {
    try {
      return await SettingsServiceBindings.GetAllExamConfigs()
    } catch (error) {
      console.error('SettingsService.GetAllExamConfigs error:', error)
      throw error
    }
  }
}

// 练习服务 API 调用
export const PracticeService = {
  GetQuestionsByCategory: async (category: string): Promise<models.Question[]> => {
    try {
      return await PracticeServiceBindings.GetQuestionsByCategory(category)
    } catch (error) {
      console.error('PracticeService.GetQuestionsByCategory error:', error)
      throw error
    }
  },
  
  ShuffleOptions: async (questions: models.Question[]): Promise<models.Question[]> => {
    try {
      return await PracticeServiceBindings.ShuffleOptions(questions)
    } catch (error) {
      console.error('PracticeService.ShuffleOptions error:', error)
      throw error
    }
  },
  
  GetPracticeProgress: async (userID: number, category: string): Promise<number> => {
    try {
      return await PracticeServiceBindings.GetPracticeProgress(userID, category)
    } catch (error) {
      console.error('PracticeService.GetPracticeProgress error:', error)
      throw error
    }
  },
  
  SavePracticeProgress: async (userID: number, category: string, index: number): Promise<void> => {
    try {
      await PracticeServiceBindings.SavePracticeProgress(userID, category, index)
    } catch (error) {
      console.error('PracticeService.SavePracticeProgress error:', error)
      throw error
    }
  },
  
  ResetProgress: async (userID: number, category: string): Promise<void> => {
    try {
      await PracticeServiceBindings.ResetProgress(userID, category)
    } catch (error) {
      console.error('PracticeService.ResetProgress error:', error)
      throw error
    }
  },
  
  SubmitAnswer: async (questionID: number, userAnswer: string): Promise<any> => {
    try {
      // 后端没有 SubmitAnswer 方法，需要前端自己判断
      return {
        question_id: questionID,
        user_answer: userAnswer,
        is_correct: true // TODO: 需要后端实现
      }
    } catch (error) {
      console.error('PracticeService.SubmitAnswer error:', error)
      throw error
    }
  },
  
  GetNextQuestion: async (currentQuestionID: number): Promise<models.Question> => {
    try {
      // 后端没有 GetNextQuestion 方法，前端自己处理
      return currentQuestionID as any // TODO: 需要实现
    } catch (error) {
      console.error('PracticeService.GetNextQuestion error:', error)
      throw error
    }
  }
}

// 考试服务 API 调用
export const ExamService = {
  StartExam: async (category: string): Promise<services.ExamStartResponse> => {
    try {
      // 从 store 获取用户 ID
      const userStore = await import('@/stores/user')
      const userID = userStore.useUserStore().currentUser?.user_id || 0
      // 现在 CreateExam 直接返回 ExamStartResponse 对象
      const result = await ExamServiceBindings.CreateExam(userID, category)
      
      console.log('CreateExam 返回值:', result)
      
      // 检查返回值是否为 null
      if (!result) {
        console.log('题库为空，返回 null')
        return null as any
      }
      
      console.log('题目数量:', result.questions?.length)
      
      return result
    } catch (error) {
      console.error('ExamService.StartExam error:', error)
      throw error
    }
  },
  
  SubmitExam: async (examID: number, answers: Record<number, string>, _startTime: number): Promise<services.ExamResult> => {
    try {
      // 转换答案为后端需要的格式
      const answersRecord: Record<number, { Answer: string; IsCorrect?: boolean }> = {}
      Object.entries(answers).forEach(([key, value]) => {
        answersRecord[Number(key)] = { Answer: value }
      })
      
      // 计算考试时长（秒）
      // const duration = Math.floor((Date.now() - startTime) / 1000)
      
      // 创建当前时间对象传递给后端
      const now = new Date().toISOString()
      
      return await ExamServiceBindings.SubmitExam(examID, answersRecord, now)
    } catch (error) {
      console.error('ExamService.SubmitExam error:', error)
      throw error
    }
  },
  
  GetExamResult: async (examID: number): Promise<services.ExamResult> => {
    try {
      return await ExamServiceBindings.GetExamResult(examID)
    } catch (error) {
      console.error('ExamService.GetExamResult error:', error)
      throw error
    }
  },
  
  InvalidateExam: async (examID: number): Promise<void> => {
    try {
      await ExamServiceBindings.InvalidateExam(examID)
    } catch (error) {
      console.error('ExamService.InvalidateExam error:', error)
      throw error
    }
  },
  
  GetExamQuestions: async (examID: number): Promise<models.ExamQuestionDetail[]> => {
    try {
      return await ExamServiceBindings.GetExamQuestions(examID)
    } catch (error) {
      console.error('ExamService.GetExamQuestions error:', error)
      throw error
    }
  }
}

// 错题本和收藏服务 API 调用
export const FavoriteService = {
  GetErrorQuestionsByCategory: async (userID: number, category: string): Promise<models.Question[]> => {
    try {
      return await PracticeServiceBindings.GetErrorQuestions(userID, category)
    } catch (error) {
      console.error('FavoriteService.GetErrorQuestionsByCategory error:', error)
      throw error
    }
  },
  
  GetFavoriteQuestionsByCategory: async (userID: number, category: string): Promise<models.Question[]> => {
    try {
      return await FavoriteServiceBindings.GetFavoriteQuestions(userID, category)
    } catch (error) {
      console.error('FavoriteService.GetFavoriteQuestionsByCategory error:', error)
      throw error
    }
  },
  
  AddToFavorite: async (userID: number, questionID: number, category: string): Promise<void> => {
    try {
      await FavoriteServiceBindings.AddFavoriteQuestion(userID, questionID, category)
    } catch (error) {
      console.error('FavoriteService.AddToFavorite error:', error)
      throw error
    }
  },
  
  RemoveFromFavorite: async (userID: number, questionID: number): Promise<void> => {
    try {
      await FavoriteServiceBindings.RemoveFavoriteQuestion(userID, questionID)
    } catch (error) {
      console.error('FavoriteService.RemoveFromFavorite error:', error)
      throw error
    }
  },
  
  AddToErrorBook: async (userID: number, questionID: number, category: string): Promise<void> => {
    try {
      // 后端函数签名：AddErrorQuestion(questionID, category, userID)
      await PracticeServiceBindings.AddErrorQuestion(questionID, category, userID)
    } catch (error) {
      console.error('FavoriteService.AddToErrorBook error:', error)
      throw error
    }
  }
}

// 统计服务 API 调用
export const StatisticsService = {
  GetExamData: async (userID: number, category: string, timeRange: string): Promise<models.ExamStatisticsData[]> => {
    try {
      return await StatisticsServiceBindings.GetExamData(userID, category, timeRange)
    } catch (error) {
      console.error('StatisticsService.GetExamData error:', error)
      throw error
    }
  },
  
  CalculateExamStatistics: async (examData: any[]): Promise<services.ExamStatisticsResult> => {
    try {
      return await StatisticsServiceBindings.CalculateExamStatistics(examData)
    } catch (error) {
      console.error('StatisticsService.CalculateExamStatistics error:', error)
      throw error
    }
  },
  
  GetPracticeAccuracy: async (userID: number): Promise<number> => {
    try {
      // 后端没有直接的方法，使用 GetUserStatistics
      const stats = await StatisticsServiceBindings.GetUserStatistics(userID)
      return stats.avg_practice_rate || 0
    } catch (error) {
      console.error('StatisticsService.GetPracticeAccuracy error:', error)
      throw error
    }
  },
  
  GetErrorCount: async (_userID: number, _category?: string): Promise<number> => {
    try {
      // 后端没有直接的方法，需要从 GetUserStatistics 获取
      const stats = await StatisticsServiceBindings.GetUserStatistics(_userID)
      return stats.total_errors || 0
    } catch (error) {
      console.error('StatisticsService.GetErrorCount error:', error)
      throw error
    }
  },
  
  GetFavoriteCount: async (userID: number, category?: string): Promise<number> => {
    try {
      // 后端没有直接的方法，使用 FavoriteServiceBindings.GetFavoriteCount
      return await FavoriteServiceBindings.GetFavoriteCount(userID, category || '')
    } catch (error) {
      console.error('StatisticsService.GetFavoriteCount error:', error)
      throw error
    }
  }
}
