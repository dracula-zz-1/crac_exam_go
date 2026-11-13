import { Events } from "@wailsio/runtime";
import * as UserService from "@bindings/crac_exam_v3/backend/services/userservice";
import * as QuestionService from "@bindings/crac_exam_v3/backend/services/questionservice";
import * as ExamService from "@bindings/crac_exam_v3/backend/services/examservice";
import * as PracticeService from "@bindings/crac_exam_v3/backend/services/practiceservice";
import * as StatisticsService from "@bindings/crac_exam_v3/backend/services/statisticsservice";
import * as SettingsService from "@bindings/crac_exam_v3/backend/services/settingsservice";
import * as ImportService from "@bindings/crac_exam_v3/backend/services/importservice";
import * as UpdateService from "@bindings/crac_exam_v3/backend/services/updateservice";
import * as AppUpdateService from "@bindings/crac_exam_v3/backend/services/appupdateservice";
import { mockApi } from "./mock";
import type {
  AppInfo,
  AppUpdateProgress,
  AppUpdateStatus,
  BankUpdateSettings,
  BankUpdateStatus,
  CategoryCount,
  Dashboard,
  DBStats,
  ExamConfig,
  ExamDetail,
  ExamResult,
  ExamStart,
  ExamSummary,
  ImportProgress,
  ImportResult,
  PracticeData,
  Question,
  QuestionPage,
  ScorePoint,
  User,
} from "@/types";

/**
 * 判断是否运行在真正的 Wails 宿主中。
 * @wailsio/runtime 在普通浏览器里也会注入 window._wails（只提供 UI 预览），
 * 因此不能只看 _wails 是否存在，而要检查底层 WebView 桥。
 */
function hasWailsHost(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as any;
  return Boolean(
    w.chrome?.webview?.postMessage ||
      w.webkit?.messageHandlers?.external?.postMessage ||
      w.wails?.invoke,
  );
}

const isWails = hasWailsHost();

function cleanError(err: unknown): string {
  if (typeof err === "string") return err.replace(/^Error:\s*/, "");
  const message = (err as any)?.message;
  if (typeof message === "string") return message.replace(/^Error:\s*/, "");
  return String(err);
}

async function call<T>(fn: () => Promise<unknown>): Promise<T> {
  try {
    return (await fn()) as T;
  } catch (err) {
    throw new Error(cleanError(err));
  }
}

const realApi = {
  isMock: false,

  login: (username: string, idCard: string) =>
    call<User>(() => UserService.Login(username, idCard)),

  getDashboard: (userId: number) =>
    call<Dashboard>(() => StatisticsService.GetDashboard(userId)),

  getPracticeData: (userId: number, category: string, mode: string) =>
    call<PracticeData>(() => PracticeService.GetPracticeData(userId, category, mode)),

  saveProgress: (userId: number, category: string, index: number) =>
    call<void>(() => PracticeService.SaveProgress(userId, category, index)),

  saveRandomProgress: (userId: number, category: string, index: number) =>
    call<void>(() => PracticeService.SaveRandomProgress(userId, category, index)),

  addFavorite: (userId: number, questionId: number, category: string) =>
    call<boolean>(() => PracticeService.AddFavorite(userId, questionId, category)),

  removeFavorite: (userId: number, questionId: number) =>
    call<void>(() => PracticeService.RemoveFavorite(userId, questionId)),

  addError: (userId: number, questionId: number, category: string) =>
    call<boolean>(() => PracticeService.AddError(userId, questionId, category)),

  removeError: (userId: number, questionId: number) =>
    call<void>(() => PracticeService.RemoveError(userId, questionId)),

  getErrorQuestions: (userId: number, category: string) =>
    call<Question[]>(() => PracticeService.GetErrorQuestions(userId, category)),

  getFavoriteQuestions: (userId: number, category: string) =>
    call<Question[]>(() => PracticeService.GetFavoriteQuestions(userId, category)),

  getPage: (
    page: number,
    pageSize: number,
    search: string,
    fa: boolean,
    fb: boolean,
    fc: boolean,
  ) =>
    call<QuestionPage>(() =>
      QuestionService.GetPage(page, pageSize, search, fa, fb, fc),
    ),

  getQuestion: (id: number) =>
    call<Question | null>(() => QuestionService.GetQuestion(id)),

  updateQuestion: (q: Question) =>
    call<void>(() => QuestionService.UpdateQuestion(q as any)),

  deleteQuestion: (id: number) =>
    call<void>(() => QuestionService.DeleteQuestion(id)),

  getCategoryCounts: () =>
    call<CategoryCount[]>(() => QuestionService.GetCategoryCounts()),

  getConfig: (category: string) =>
    call<ExamConfig>(() => ExamService.GetConfig(category)),

  createExam: (userId: number, category: string) =>
    call<ExamStart>(() => ExamService.CreateExam(userId, category)),

  submitExam: (examId: number, answers: Record<string, string>) =>
    call<ExamResult>(() => ExamService.SubmitExam(examId, answers)),

  getResult: (examId: number) =>
    call<ExamResult>(() => ExamService.GetResult(examId)),

  getDetails: (examId: number) =>
    call<ExamDetail[]>(() => ExamService.GetDetails(examId)),

  invalidateExam: (examId: number) =>
    call<void>(() => ExamService.InvalidateExam(examId)),

  getUserExamRecords: (userId: number, limit: number) =>
    call<ExamSummary[]>(() => ExamService.GetUserExamRecords(userId, limit)),

  getScoreSeries: (userId: number, category: string, days: number) =>
    call<ScorePoint[]>(() =>
      StatisticsService.GetScoreSeries(userId, category, days),
    ),

  pickZipFile: () =>
    call<string>(() => ImportService.PickZipFile()),

  importZip: (path: string) =>
    call<ImportResult>(() => ImportService.ImportZip(path)),

  getBankUpdateStatus: () =>
    call<BankUpdateStatus>(() => UpdateService.GetStatus()),

  getBankUpdateSettings: () =>
    call<BankUpdateSettings>(() => UpdateService.GetSettings()),

  setBankAutoUpdate: (autoCheck: boolean, autoImport: boolean) =>
    call<void>(() => UpdateService.SetAutoUpdate(autoCheck, autoImport)),

  checkBankUpdate: () =>
    call<BankUpdateStatus>(() => UpdateService.Check()),

  applyBankUpdate: () =>
    call<ImportResult>(() => UpdateService.Apply()),

  onBankUpdateStatus(cb: (s: BankUpdateStatus) => void): () => void {
    return Events.On("bank:update-status", (event: any) => cb(event.data));
  },

  getAppUpdateStatus: () =>
    call<AppUpdateStatus>(() => AppUpdateService.GetStatus()),

  setAppUpdateRepo: (repo: string) =>
    call<void>(() => AppUpdateService.SetRepo(repo)),

  checkAppUpdate: () =>
    call<AppUpdateStatus>(() => AppUpdateService.Check()),

  downloadAppUpdate: () =>
    call<AppUpdateStatus>(() => AppUpdateService.Download()),

  runAppUpdateInstaller: () =>
    call<void>(() => AppUpdateService.RunInstaller()),

  openAppUpdateRelease: () =>
    call<void>(() => AppUpdateService.OpenReleasePage()),

  openAppUpdateFolder: () =>
    call<void>(() => AppUpdateService.OpenDownloadFolder()),

  onAppUpdateStatus(cb: (s: AppUpdateStatus) => void): () => void {
    return Events.On("app:update-status", (event: any) => cb(event.data));
  },

  onAppUpdateProgress(cb: (p: AppUpdateProgress) => void): () => void {
    return Events.On("app:update-progress", (event: any) => cb(event.data));
  },

  clearUserData: (userId: number) =>
    call<void>(() => SettingsService.ClearUserData(userId)),

  clearQuestionBank: () =>
    call<void>(() => SettingsService.ClearQuestionBank()),

  getDBStats: () =>
    call<DBStats>(() => SettingsService.GetDBStats()),

  getAppInfo: () =>
    call<AppInfo>(() => SettingsService.GetAppInfo()),

  onImportProgress(cb: (p: ImportProgress) => void): () => void {
    return Events.On("import:progress", (event: any) => cb(event.data));
  },

  onFilesDropped(cb: (paths: string[]) => void): () => void {
    return Events.On("files-dropped", (event: any) => cb(event.data));
  },
};

export const api = isWails ? realApi : mockApi;
export const usingMock = !isWails;
