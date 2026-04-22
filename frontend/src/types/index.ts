export interface Question {
  id: number
  J: string
  P: string
  I: string
  Q: string
  T: string
  A: string
  B: string
  C: string
  D: string
  F: string
  LA: number
  LB: number
  LC: number
  type: number
  user_id?: number
  type_text?: string
  has_image?: string
}

export interface ExamQuestionDetail {
  id: number
  exam_id: number
  question_id: number
  question_text: string
  option_a: string
  option_b: string
  option_c: string
  option_d: string
  correct_answer: string
  user_answer: string
  is_correct: boolean
  type: number
  image_data: string
}

export interface ExamRecord {
  id: number
  category: string
  exam_date: string
  duration_seconds: number
  user_id: number
  score: number
  total_questions: number
  correct_count: number
}

export interface User {
  id: number
  username: string
  id_card: string
  last_login: string
}

export interface ExamStartResponse {
  exam_id: number
  questions: Question[]
  config: ExamConfig
}

export interface ExamConfig {
  total: number
  single: number
  multiple: number
  time_minutes: number
  pass_score: number
}

export interface ExamResult {
  exam_id: number
  category: string
  exam_date: string
  duration_seconds: number
  score: number
  correct_count: number
  total_count: number
  pass_exam: boolean
  pass_score: number
}

export interface UserLoginResponse {
  success: boolean
  message?: string
  user_info?: User
}

export interface UserStatisticsResult {
  total_exams: number
  total_practices: number
  total_errors: number
  total_favorites: number
  avg_exam_score: number
  avg_practice_rate: number
  exam_pass_rate: number
  last_exam_date?: string
  last_practice_date?: string
}

export interface PageDataResult {
  data: Question[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ImportResult {
  success: boolean
  message: string
  imported_count: number
  total_count: number
  stats: Record<string, number>
}

export interface ExamStatisticsResult {
  max_score: number
  latest_score: number
  latest_duration: number
  avg_pass_rate: number
  total_exams: number
}

export interface AppInfo {
  name: string
  version: string
}

export interface ImportStatsData {
  [category: string]: number
}
