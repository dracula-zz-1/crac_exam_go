export interface User {
  id: number;
  username: string;
  idCard: string;
  lastLogin: string;
}

export interface Question {
  id: number;
  J: string;
  P: string;
  I: string;
  Q: string;
  T: string;
  A: string;
  B: string;
  C: string;
  D: string;
  F: string;
  LA: number;
  LB: number;
  LC: number;
  type: number;
  userId: number;
  wrongCount?: number;
}

export interface ExamQuestion {
  id: number;
  J: string;
  Q: string;
  A: string;
  B: string;
  C: string;
  D: string;
  F: string;
  type: number;
}

export interface ExamConfig {
  total: number;
  single: number;
  multiple: number;
  timeMinutes: number;
  passScore: number;
}

export interface ExamStart {
  examId: number;
  category: string;
  questions: ExamQuestion[];
  total: number;
  single: number;
  multiple: number;
  timeLimitSeconds: number;
  passScore: number;
  startedAt: string;
}

export interface ExamResult {
  examId: number;
  category: string;
  examDate: string;
  durationSeconds: number;
  score: number;
  correctCount: number;
  totalCount: number;
  passScore: number;
  passed: boolean;
}

export interface ExamDetail {
  id: number;
  examId: number;
  questionId: number;
  seq: number;
  questionText: string;
  optionA: string;
  optionB: string;
  optionC: string;
  optionD: string;
  correctAnswer: string;
  userAnswer: string;
  isCorrect: boolean;
  type: number;
  imageData: string;
}

export interface ExamSummary {
  id: number;
  category: string;
  examDate: string;
  score: number;
  total: number;
  correctCount: number;
  passScore: number;
  passed: boolean;
  durationSeconds: number;
}

export interface PracticeData {
  mode: string;
  category: string;
  questions: Question[];
  favoriteIds: number[];
  progress: number;
  total: number;
}

export interface QuestionPage {
  data: Question[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface CategoryCount {
  category: string;
  total: number;
  single: number;
  multiple: number;
  withImage: number;
}

export interface CategoryProgress {
  category: string;
  answered: number;
  total: number;
  percent: number;
  lastAccess: number;
}

export interface Dashboard {
  user: User;
  totalQuestions: number;
  withImages: number;
  categories: CategoryCount[];
  totalExams: number;
  passedExams: number;
  passRate: number;
  bestScores: Record<string, number>;
  recentExams: ExamSummary[];
  errorCount: number;
  favoriteCount: number;
  practice: CategoryProgress[];
  lastExam: ExamSummary | null;
  averageScore: number;
  totalStudyMinutes: number;
}

export interface ScorePoint {
  date: string;
  score: number;
  total: number;
  passed: boolean;
  category: string;
}

export interface ImportResult {
  success: boolean;
  message: string;
  imported: number;
  total: number;
  stats: Record<string, number>;
  warnings: string[];
  duplicateJ: number;
  duplicateJpi: number;
  elapsedMs: number;
  preservedFavorites: number;
  preservedErrors: number;
  preservedProgress: number;
}

export interface BankUpdateSettings {
  autoCheck: boolean;
  autoImport: boolean;
}

export interface BankUpdateStatus {
  checked: boolean;
  checking: boolean;
  available: boolean;
  currentVersion: string;
  currentSource: string;
  currentUrl: string;
  currentHash: string;
  currentImportedAt: string;
  remoteVersion: string;
  remoteUrl: string;
  remoteSize: number;
  remoteModified: string;
  autoCheck: boolean;
  autoImport: boolean;
  message: string;
  error: string;
}

export interface ImportProgress {
  stage: string;
  message: string;
  percent: number;
}

export interface AppReleaseAsset {
  name: string;
  downloadUrl: string;
  size: number;
  contentType: string;
  digest: string;
}

export interface AppUpdateStatus {
  checked: boolean;
  checking: boolean;
  downloading: boolean;
  available: boolean;
  currentVersion: string;
  latestVersion: string;
  releaseName: string;
  releaseUrl: string;
  publishedAt: string;
  notes: string;
  repo: string;
  asset: AppReleaseAsset | null;
  assets: AppReleaseAsset[];
  downloadPercent: number;
  downloadedPath: string;
  downloadedName: string;
  message: string;
  error: string;
}

export interface AppUpdateProgress {
  stage: string;
  message: string;
  percent: number;
}

export interface DBStats {
  users: number;
  questions: number;
  examRecords: number;
  errorRecords: number;
  favorites: number;
}

export interface AppInfo {
  name: string;
  version: string;
  author: string;
  description: string;
  copyright: string;
  goVersion: string;
  dataDir: string;
}

export type Category = "A" | "B" | "C";
export type PracticeMode = "practice" | "random" | "error" | "favorite";
