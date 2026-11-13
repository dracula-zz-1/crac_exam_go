/**
 * 浏览器预览用的内存实现。
 * 仅在非 Wails 环境（npm run dev 直接打开浏览器）启用，不影响桌面端。
 */
import type {
  AppInfo,
  AppUpdateProgress,
  AppUpdateStatus,
  BankUpdateSettings,
  BankUpdateStatus,
  CategoryCount,
  Dashboard,
  DBStats,
  ExamDetail,
  ExamQuestion,
  ExamResult,
  ExamStart,
  ExamSummary,
  ImportProgress,
  ImportResult,
  PracticeData,
  PracticeMode,
  Question,
  QuestionPage,
  ScorePoint,
  User,
} from "@/types";

const CATEGORIES = ["A", "B", "C"] as const;
const CONFIG = {
  A: { total: 40, single: 32, multiple: 8, timeMinutes: 40, passScore: 30 },
  B: { total: 60, single: 45, multiple: 15, timeMinutes: 60, passScore: 45 },
  C: { total: 90, single: 70, multiple: 20, timeMinutes: 90, passScore: 70 },
} as const;

const IMG = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(`
<svg xmlns="http://www.w3.org/2000/svg" width="640" height="220" viewBox="0 0 640 220">
  <rect width="640" height="220" rx="18" fill="#101820"/>
  <g stroke="#e8a33d" stroke-width="3" fill="none" stroke-linecap="round">
    <path d="M40 120 Q80 60 120 120 T200 120 T280 120 T360 120 T440 120 T520 120 T600 120"/>
    <path d="M40 150 Q80 110 120 150 T200 150 T280 150 T360 150 T440 150 T520 150 T600 150" opacity=".35"/>
  </g>
  <text x="40" y="52" fill="#f6f3ec" font-family="sans-serif" font-size="18" font-weight="700">信号波形示意图</text>
</svg>`)}`;

interface Template {
  q: string;
  opts: [string, string, string, string];
  t: string;
  type: number;
  category?: string;
}

const TEMPLATES: Template[] = [
  {
    q: "我国负责无线电管理工作的主管部门是？",
    opts: ["国务院", "工业和信息化部", "国家广播电视总局", "公安部"],
    t: "B",
    type: 1,
  },
  {
    q: "业余无线电台在发射时，应当使用下列哪种呼号？",
    opts: ["自编呼号", "电台执照指配的呼号", "任意英文名", "手机号码"],
    t: "B",
    type: 1,
  },
  {
    q: "VHF 频段通常是指下列哪一段频率范围？",
    opts: ["3-30 kHz", "30-300 MHz", "300-3000 MHz", "3-30 GHz"],
    t: "B",
    type: 1,
  },
  {
    q: "下列哪些行为属于业余无线电操作中禁止的行为？",
    opts: ["参与应急通信", "进行技术实验", "播送商业广告", "参加竞赛活动"],
    t: "C",
    type: 1,
  },
  {
    q: "天线的驻波比（SWR）越接近下列哪个数值，匹配状态越好？",
    opts: ["0", "1", "2", "5"],
    t: "B",
    type: 1,
  },
  {
    q: "关于频率容限，下列说法正确的是？",
    opts: ["发射频率可以任意偏移", "应保持在执照允许的容限内", "只要对方能听清即可", "与天线长度无关"],
    t: "B",
    type: 1,
  },
  {
    q: "下列哪些属于无线电干扰的常见来源？",
    opts: ["雷电", "劣质开关电源", "同频段其他发射机", "以上都是"],
    t: "D",
    type: 1,
  },
  {
    q: "进行业余无线电通信时，以下哪些做法是正确的？",
    opts: ["使用明码通信", "隐瞒真实身份", "长时间占用应急频率", "播放音乐"],
    t: "A",
    type: 1,
  },
  {
    q: "以下哪些措施有助于降低电磁辐射对周围的影响？",
    opts: ["提高发射功率", "合理架设天线并控制功率", "缩短天线馈线", "使用定向天线"],
    t: "BCD",
    type: 2,
  },
  {
    q: "关于 QSL 卡片，下列说法正确的有？",
    opts: ["用于确认通联", "可以记录通联时间与频率", "可替代电台执照", "可用于申请奖状"],
    t: "ABD",
    type: 2,
  },
];

function buildQuestions(category: string): Question[] {
  const list: Question[] = [];
  for (let i = 0; i < 50; i++) {
    const tpl = TEMPLATES[i % TEMPLATES.length];
    const type = i < 40 ? 1 : 2;
    const answer = type === 1 ? tpl.t.slice(0, 1) : tpl.type === 2 ? tpl.t : "AC";
    list.push({
      id: 1000 + CATEGORIES.indexOf(category as never) * 100 + i,
      J: `LY${String(1000 + i).padStart(4, "0")}`,
      P: String(i + 1),
      I: `LK${String(100 + i).padStart(4, "0")}`,
      Q: tpl.q,
      T: answer,
      A: tpl.opts[0],
      B: tpl.opts[1],
      C: tpl.opts[2],
      D: tpl.opts[3],
      F: i === 3 ? IMG : "",
      LA: category === "A" ? 1 : 0,
      LB: category === "B" ? 1 : 0,
      LC: category === "C" ? 1 : 0,
      type,
      userId: 0,
    });
  }
  return list;
}

const state = {
  user: null as User | null,
  questions: {
    A: buildQuestions("A"),
    B: buildQuestions("B"),
    C: buildQuestions("C"),
  } as Record<string, Question[]>,
  favorites: new Map<number, string>(),
  errors: new Map<number, string>(),
  progress: {} as Record<string, number>,
  results: [] as ExamResult[],
  details: new Map<number, ExamDetail[]>(),
  activeExam: null as null | {
    start: ExamStart;
    correct: Map<number, string>;
    submitted: ExamResult | null;
  },
  nextExamId: 9000,
};

// 初始化一些演示数据
(function seed() {
  state.progress.A = 9;
  state.progress.B = 3;
  const a = state.questions.A;
  const b = state.questions.B;
  [1, 4, 7, 12, 18, 22].forEach((i) => state.errors.set(a[i].id, "A"));
  [2, 5, 9].forEach((i) => state.errors.set(b[i].id, "B"));
  [10, 16, 25, 33].forEach((i) => state.favorites.set(a[i].id, "A"));
  [6, 15].forEach((i) => state.favorites.set(b[i].id, "B"));

  const now = Date.now();
  const samples: Array<[string, number, number]> = [
    ["A", 22, 12],
    ["A", 27, 8],
    ["A", 34, 5],
    ["B", 39, 4],
    ["A", 36, 2],
    ["B", 48, 1],
  ];
  for (const [category, score, daysAgo] of samples) {
    const id = ++state.nextExamId;
    const questions = state.questions[category];
    const total = category === "A" ? 40 : 60;
    const pass = CONFIG[category as "A"].passScore;
    const result: ExamResult = {
      examId: id,
      category,
      examDate: new Date(now - daysAgo * 86400000).toISOString(),
      durationSeconds: 900 + score * 20,
      score,
      correctCount: score,
      totalCount: total,
      passScore: pass,
      passed: score >= pass,
    };
    state.results.push(result);
    state.details.set(id, buildDetails(id, category, questions, total, score));
  }
})();

function buildDetails(
  examId: number,
  category: string,
  questions: Question[],
  total: number,
  score: number,
): ExamDetail[] {
  const out: ExamDetail[] = [];
  for (let i = 0; i < total; i++) {
    const q = questions[i % questions.length];
    const isCorrect = i < score;
    const userAnswer = isCorrect ? q.T : q.T === "A" ? "B" : "A";
    out.push({
      id: examId * 1000 + i,
      examId,
      questionId: q.id,
      seq: i + 1,
      questionText: q.Q,
      optionA: q.A,
      optionB: q.B,
      optionC: q.C,
      optionD: q.D,
      correctAnswer: q.T,
      userAnswer,
      isCorrect,
      type: q.type,
      imageData: q.F,
    });
  }
  return out;
}

const updateState: BankUpdateSettings = { autoCheck: true, autoImport: false };

function mockUpdateStatus(): BankUpdateStatus {
  return {
    checked: false,
    checking: false,
    available: false,
    currentVersion: "浏览器演示题库 · 2026-11",
    currentSource: "内置演示数据",
    currentUrl: "https://www.crac.org.cn/?p=1287",
    currentHash: "",
    currentImportedAt: new Date().toISOString(),
    remoteVersion: "",
    remoteUrl: "",
    remoteSize: 0,
    remoteModified: "",
    autoCheck: updateState.autoCheck,
    autoImport: updateState.autoImport,
    message: "",
    error: "",
  };
}

let mockDataDir = "C:\\Users\\Demo\\AppData\\Roaming\\CracExamV3";

const appUpdateState: AppUpdateStatus = {
  checked: false,
  checking: false,
  downloading: false,
  available: false,
  currentVersion: "3.0.0",
  latestVersion: "",
  releaseName: "",
  releaseUrl: "https://github.com/dracula-zz-1/crac_exam_go/releases",
  publishedAt: "",
  notes: "",
  source: "github",
  repo: "dracula-zz-1/crac_exam_go",
  asset: null,
  assets: [],
  downloadPercent: 0,
  downloadedPath: "",
  downloadedName: "",
  message: "浏览器演示模式：不会真正下载安装包",
  error: "",
};

function mockAppUpdateStatus(): AppUpdateStatus {
  return {
    ...appUpdateState,
    assets: [...appUpdateState.assets],
  };
}

const wait = (ms = 160) => new Promise((r) => setTimeout(r, ms));

function findQuestion(id: number): Question | undefined {
  for (const cat of CATEGORIES) {
    const found = state.questions[cat].find((q) => q.id === id);
    if (found) return found;
  }
  return undefined;
}

function categoryQuestionCount(category: string) {
  const list = state.questions[category] ?? [];
  return {
    category,
    total: list.length,
    single: list.filter((q) => q.type === 1).length,
    multiple: list.filter((q) => q.type === 2).length,
    withImage: list.filter((q) => q.F).length,
  } as CategoryCount;
}

function summaryOf(r: ExamResult): ExamSummary {
  return {
    id: r.examId,
    category: r.category,
    examDate: r.examDate,
    score: r.score,
    total: r.totalCount,
    correctCount: r.correctCount,
    passScore: r.passScore,
    passed: r.passed,
    durationSeconds: r.durationSeconds,
  };
}

function dashboard(userId: number): Dashboard {
  const user = state.user ?? {
    id: userId,
    username: "演示用户",
    idCard: "DEMO",
    lastLogin: new Date().toISOString(),
  };
  const categories = CATEGORIES.map(categoryQuestionCount);
  const results = [...state.results].sort(
    (a, b) => +new Date(b.examDate) - +new Date(a.examDate),
  );
  const passed = results.filter((r) => r.passed).length;
  const best: Record<string, number> = {};
  for (const r of results) best[r.category] = Math.max(best[r.category] ?? 0, r.score);
  const practice = CATEGORIES.map((cat) => {
    const total = state.questions[cat].length;
    const answered = state.progress[cat] ?? 0;
    return {
      category: cat,
      answered,
      total,
      percent: total ? Math.round((answered / total) * 100) : 0,
      lastAccess: 0,
    };
  });
  const totalDuration = results.reduce((s, r) => s + r.durationSeconds, 0);
  return {
    user,
    totalQuestions: CATEGORIES.reduce((s, c) => s + state.questions[c].length, 0),
    withImages: CATEGORIES.reduce(
      (s, c) => s + state.questions[c].filter((q) => q.F).length,
      0,
    ),
    categories,
    totalExams: results.length,
    passedExams: passed,
    passRate: results.length ? Math.round((passed / results.length) * 1000) / 10 : 0,
    bestScores: best,
    recentExams: results.slice(0, 6).map(summaryOf),
    errorCount: state.errors.size,
    favoriteCount: state.favorites.size,
    practice,
    lastExam: results[0] ? summaryOf(results[0]) : null,
    averageScore: results.length
      ? Math.round((results.reduce((s, r) => s + r.score, 0) / results.length) * 10) / 10
      : 0,
    totalStudyMinutes: Math.round(totalDuration / 60),
  };
}

let progressListener: ((p: ImportProgress) => void) | null = null;

export const mockApi = {
  isMock: true,

  async login(username: string, idCard: string): Promise<User> {
    await wait();
    state.user = {
      id: 1,
      username: username.trim(),
      idCard: idCard.trim(),
      lastLogin: new Date().toISOString(),
    };
    return state.user;
  },

  async getDashboard(userId: number): Promise<Dashboard> {
    await wait();
    return dashboard(userId);
  },

  async getPracticeData(
    _userId: number,
    category: string,
    mode: string,
  ): Promise<PracticeData> {
    await wait();
    let questions: Question[] = [];
    if (mode === "practice") {
      questions = state.questions[category] ?? [];
    } else if (mode === "random") {
      questions = [...(state.questions[category] ?? [])].sort(() => Math.random() - 0.5);
    } else if (mode === "error") {
      questions = (state.questions[category] ?? []).filter((q) => state.errors.has(q.id));
    } else {
      questions = (state.questions[category] ?? []).filter((q) => state.favorites.has(q.id));
    }
    return {
      mode,
      category,
      questions: [...questions],
      favoriteIds: [...state.favorites.keys()],
      progress: mode === "practice" ? state.progress[category] ?? 0 : 0,
      total: questions.length,
    };
  },

  async saveProgress(_userId: number, category: string, index: number): Promise<void> {
    state.progress[category] = index;
    await wait(30);
  },

  async saveRandomProgress(_userId: number, _category: string, _index: number): Promise<void> {
    await wait(20);
  },

  async addFavorite(_userId: number, questionId: number, category: string): Promise<boolean> {
    await wait(60);
    if (state.favorites.has(questionId)) return false;
    state.favorites.set(questionId, category);
    return true;
  },

  async removeFavorite(_userId: number, questionId: number): Promise<void> {
    await wait(60);
    state.favorites.delete(questionId);
  },

  async addError(_userId: number, questionId: number, category: string): Promise<boolean> {
    await wait(60);
    if (state.errors.has(questionId)) return false;
    state.errors.set(questionId, category);
    return true;
  },

  async removeError(_userId: number, questionId: number): Promise<void> {
    await wait(60);
    state.errors.delete(questionId);
  },

  async getErrorQuestions(_userId: number, category: string): Promise<Question[]> {
    await wait();
    const all = category
      ? state.questions[category] ?? []
      : CATEGORIES.flatMap((c) => state.questions[c]);
    return all.filter((q) => state.errors.has(q.id));
  },

  async getFavoriteQuestions(_userId: number, category: string): Promise<Question[]> {
    await wait();
    const all = category
      ? state.questions[category] ?? []
      : CATEGORIES.flatMap((c) => state.questions[c]);
    return all.filter((q) => state.favorites.has(q.id));
  },

  async getPage(
    page: number,
    pageSize: number,
    search: string,
    fa: boolean,
    fb: boolean,
    fc: boolean,
  ): Promise<QuestionPage> {
    await wait();
    const all = CATEGORIES.flatMap((c) => state.questions[c]);
    const filtered = all.filter((q) => {
      const catOk =
        (!fa && !fb && !fc) ||
        (fa && q.LA) ||
        (fb && q.LB) ||
        (fc && q.LC);
      const s = search.trim();
      const searchOk =
        !s || [q.Q, q.J, q.P, q.I, q.A, q.B, q.C, q.D].some((v) => v.includes(s));
      return catOk && searchOk;
    });
    const total = filtered.length;
    const start = (page - 1) * pageSize;
    return {
      data: filtered.slice(start, start + pageSize),
      total,
      page,
      pageSize,
      totalPages: Math.max(1, Math.ceil(total / pageSize)),
    };
  },

  async getQuestion(id: number): Promise<Question | null> {
    await wait(80);
    return findQuestion(id) ?? null;
  },

  async updateQuestion(q: Question): Promise<void> {
    await wait();
    for (const cat of CATEGORIES) {
      const idx = state.questions[cat].findIndex((x) => x.id === q.id);
      if (idx >= 0) {
        state.questions[cat][idx] = { ...q };
        return;
      }
    }
    throw new Error("题目不存在");
  },

  async deleteQuestion(id: number): Promise<void> {
    await wait();
    for (const cat of CATEGORIES) {
      state.questions[cat] = state.questions[cat].filter((q) => q.id !== id);
    }
    state.favorites.delete(id);
    state.errors.delete(id);
  },

  async getCategoryCounts(): Promise<CategoryCount[]> {
    await wait(80);
    return CATEGORIES.map(categoryQuestionCount);
  },

  async getConfig(category: string) {
    await wait(40);
    return { ...CONFIG[category as "A"] };
  },

  async createExam(_userId: number, category: string): Promise<ExamStart> {
    await wait(320);
    const cfg = CONFIG[category as "A"];
    const list = state.questions[category] ?? [];
    const singles = list.filter((q) => q.type === 1).slice(0, cfg.single);
    const multiples = list.filter((q) => q.type === 2).slice(0, cfg.multiple);
    if (singles.length < cfg.single || multiples.length < cfg.multiple) {
      throw new Error(`${category} 类题库题目不足，请先导入题库`);
    }
    const picked = [...singles, ...multiples].sort(() => Math.random() - 0.5);
    const id = ++state.nextExamId;
    const examQuestions: ExamQuestion[] = picked.map((q) => ({
      id: q.id,
      J: q.J,
      Q: q.Q,
      A: q.A,
      B: q.B,
      C: q.C,
      D: q.D,
      F: q.F,
      type: q.type,
    }));
    const start: ExamStart = {
      examId: id,
      category,
      questions: examQuestions,
      total: picked.length,
      single: cfg.single,
      multiple: cfg.multiple,
      timeLimitSeconds: cfg.timeMinutes * 60,
      passScore: cfg.passScore,
      startedAt: new Date().toISOString(),
    };
    state.activeExam = {
      start,
      correct: new Map(picked.map((q) => [q.id, q.T])),
      submitted: null,
    };
    return start;
  },

  async submitExam(examId: number, answers: Record<string, string>): Promise<ExamResult> {
    await wait(280);
    const active = state.activeExam;
    if (!active || active.start.examId !== examId) {
      const existing = state.results.find((r) => r.examId === examId);
      if (existing) return existing;
      throw new Error("考试记录不存在");
    }
    if (active.submitted) return active.submitted;

    let correct = 0;
    const details: ExamDetail[] = active.start.questions.map((q, i) => {
      const raw = (answers[String(q.id)] ?? "").toUpperCase();
      const answer = [...new Set(raw.split("").filter((c) => "ABCD".includes(c)))]
        .sort()
        .join("");
      const right = answer === active.correct.get(q.id);
      if (right) correct++;
      else state.errors.set(q.id, active.start.category);
      return {
        id: examId * 1000 + i,
        examId,
        questionId: q.id,
        seq: i + 1,
        questionText: q.Q,
        optionA: q.A,
        optionB: q.B,
        optionC: q.C,
        optionD: q.D,
        correctAnswer: active.correct.get(q.id) ?? "",
        userAnswer: answer,
        isCorrect: right,
        type: q.type,
        imageData: q.F,
      };
    });
    const result: ExamResult = {
      examId,
      category: active.start.category,
      examDate: new Date().toISOString(),
      durationSeconds: Math.max(
        1,
        Math.round((Date.now() - +new Date(active.start.startedAt)) / 1000),
      ),
      score: correct,
      correctCount: correct,
      totalCount: active.start.questions.length,
      passScore: active.start.passScore,
      passed: correct >= active.start.passScore,
    };
    state.results.push(result);
    state.details.set(examId, details);
    active.submitted = result;
    return result;
  },

  async getResult(examId: number): Promise<ExamResult> {
    await wait(80);
    const r = state.results.find((x) => x.examId === examId);
    if (!r) throw new Error("考试记录不存在");
    return r;
  },

  async getDetails(examId: number): Promise<ExamDetail[]> {
    await wait(100);
    return state.details.get(examId) ?? [];
  },

  async invalidateExam(examId: number): Promise<void> {
    if (state.activeExam?.start.examId === examId) state.activeExam = null;
    await wait(30);
  },

  async getUserExamRecords(_userId: number, limit: number): Promise<ExamSummary[]> {
    await wait(80);
    return [...state.results]
      .sort((a, b) => +new Date(b.examDate) - +new Date(a.examDate))
      .slice(0, limit)
      .map(summaryOf);
  },

  async getScoreSeries(
    _userId: number,
    category: string,
    days: number,
  ): Promise<ScorePoint[]> {
    await wait(80);
    const since = days > 0 ? Date.now() - days * 86400000 : 0;
    return state.results
      .filter(
        (r) =>
          (category === "" || r.category === category) && +new Date(r.examDate) >= since,
      )
      .sort((a, b) => +new Date(a.examDate) - +new Date(b.examDate))
      .map((r) => ({
        date: r.examDate,
        score: r.score,
        total: r.totalCount,
        passed: r.passed,
        category: r.category,
      }));
  },

  async pickZipFile(): Promise<string> {
    await wait(120);
    return "C:\\演示\\业余无线电题库.zip";
  },

  async importZip(_path: string): Promise<ImportResult> {
    const stages: ImportProgress[] = [
      { stage: "unzip", message: "正在校验并解压题库包…", percent: 8 },
      { stage: "parse", message: "正在解析总题库…", percent: 30 },
      { stage: "category", message: "正在解析分类题库…", percent: 52 },
      { stage: "images", message: "正在提取题目图片（纯 Go 解析）…", percent: 72 },
      { stage: "write", message: "正在写入数据库…", percent: 92 },
    ];
    for (const s of stages) {
      progressListener?.(s);
      await wait(260);
    }
    state.questions = {
      A: buildQuestions("A"),
      B: buildQuestions("B"),
      C: buildQuestions("C"),
    };
    state.favorites.clear();
    state.errors.clear();
    state.progress = {};
    progressListener?.({ stage: "done", message: "成功导入 150 道题目", percent: 100 });
    return {
      success: true,
      message: "成功导入 150 道题目",
      imported: 150,
      total: 150,
      stats: {
        single: 120,
        multiple: 30,
        categoryA: 50,
        categoryB: 50,
        categoryC: 50,
        withImage: 3,
      },
      warnings: ["这是浏览器演示数据，实际导入请使用桌面应用"],
      duplicateJ: 0,
      duplicateJpi: 0,
      elapsedMs: 1280,
      preservedFavorites: 0,
      preservedErrors: 0,
      preservedProgress: 0,
    };
  },

  async getBankUpdateStatus(): Promise<BankUpdateStatus> {
    await wait(60);
    return mockUpdateStatus();
  },

  async getBankUpdateSettings(): Promise<BankUpdateSettings> {
    await wait(40);
    return { ...updateState };
  },

  async setBankAutoUpdate(autoCheck: boolean, autoImport: boolean): Promise<void> {
    updateState.autoCheck = autoCheck;
    updateState.autoImport = autoImport;
    await wait(40);
  },

  async checkBankUpdate(): Promise<BankUpdateStatus> {
    await wait(600);
    return {
      ...mockUpdateStatus(),
      checked: true,
      message: "已是最新题库（浏览器演示）",
      remoteVersion: "2025-08-09 · 演示",
      remoteUrl: "https://www.crac.org.cn/?p=1287",
      remoteSize: 4832551,
    };
  },

  async applyBankUpdate(): Promise<ImportResult> {
    const result = await mockApi.importZip("demo");
    return { ...result, message: "题库已更新（演示）" };
  },

  onBankUpdateStatus(_cb: (s: BankUpdateStatus) => void): () => void {
    return () => {};
  },

  async getAppUpdateStatus(): Promise<AppUpdateStatus> {
    await wait(50);
    return mockAppUpdateStatus();
  },

  async setAppUpdateRepo(repo: string): Promise<void> {
    appUpdateState.repo = repo.trim();
    await wait(40);
  },

  async setAppUpdateSource(source: string): Promise<void> {
    const next = source === "gitee" ? "gitee" : "github";
    appUpdateState.source = next;
    appUpdateState.repo =
      next === "gitee" ? "dracula_zz/crac_exam_go" : "dracula-zz-1/crac_exam_go";
    appUpdateState.releaseUrl =
      next === "gitee"
        ? "https://gitee.com/dracula_zz/crac_exam_go/releases"
        : "https://github.com/dracula-zz-1/crac_exam_go/releases";
    appUpdateState.checked = false;
    appUpdateState.available = false;
    appUpdateState.latestVersion = "";
    appUpdateState.asset = null;
    appUpdateState.assets = [];
    appUpdateState.message = next === "gitee" ? "更新通道已切换为 Gitee 镜像" : "更新通道已切换为 GitHub";
    appUpdateState.error = "";
    await wait(40);
  },

  async checkAppUpdate(): Promise<AppUpdateStatus> {
    await wait(700);
    appUpdateState.checked = true;
    appUpdateState.available = true;
    appUpdateState.latestVersion = "v3.1.0";
    appUpdateState.releaseName = "v3.1.0 · 演示更新";
    appUpdateState.publishedAt = new Date().toISOString();
    appUpdateState.notes = "浏览器演示模式：桌面端会访问 GitHub API 获取真实 Release。";
    appUpdateState.asset = {
      name: "crac-exam-v3.1.0-windows-amd64.exe",
      downloadUrl: "https://github.com/dracula-zz-1/crac_exam_go/releases",
      size: 48 * 1024 * 1024,
      contentType: "application/vnd.microsoft.portable-executable",
      digest: "",
    };
    appUpdateState.assets = [appUpdateState.asset];
    appUpdateState.message = "发现新版本 v3.1.0";
    appUpdateState.error = "";
    return mockAppUpdateStatus();
  },

  async downloadAppUpdate(): Promise<AppUpdateStatus> {
    await wait(900);
    appUpdateState.downloading = false;
    appUpdateState.downloadPercent = 100;
    appUpdateState.downloadedPath = "C:\\Users\\Demo\\AppData\\Roaming\\CracExamV3\\updates\\crac-exam-v3.1.0-windows-amd64.exe";
    appUpdateState.downloadedName = "crac-exam-v3.1.0-windows-amd64.exe";
    appUpdateState.message = "更新包下载完成（演示）";
    return mockAppUpdateStatus();
  },

  async runAppUpdateInstaller(): Promise<void> {
    await wait(120);
  },

  async openAppUpdateRelease(): Promise<void> {
    await wait(40);
  },

  async openAppUpdateFolder(): Promise<void> {
    await wait(40);
  },

  onAppUpdateStatus(_cb: (s: AppUpdateStatus) => void): () => void {
    return () => {};
  },

  onAppUpdateProgress(_cb: (p: AppUpdateProgress) => void): () => void {
    return () => {};
  },

  onImportProgress(cb: (p: ImportProgress) => void): () => void {
    progressListener = cb;
    return () => {
      if (progressListener === cb) progressListener = null;
    };
  },

  onFilesDropped(_cb: (paths: string[]) => void): () => void {
    /* 浏览器预览不支持原生拖放路径 */
    return () => {};
  },

  async clearUserData(_userId: number): Promise<void> {
    await wait();
    state.favorites.clear();
    state.errors.clear();
    state.progress = {};
    state.results = [];
    state.details.clear();
  },

  async clearQuestionBank(): Promise<void> {
    await wait();
    state.questions = { A: [], B: [], C: [] };
    state.favorites.clear();
    state.errors.clear();
    state.progress = {};
  },

  async getDBStats(): Promise<DBStats> {
    await wait(80);
    return {
      users: 1,
      questions: CATEGORIES.reduce((s, c) => s + state.questions[c].length, 0),
      examRecords: state.results.length,
      errorRecords: state.errors.size,
      favorites: state.favorites.size,
    };
  },

  async getAppInfo(): Promise<AppInfo> {
    await wait(40);
    return {
      name: "业余无线电模拟考试",
      version: "3.0.0",
      author: "BA4RHH",
      description: "业余无线电台操作技术能力验证模拟考试系统",
      copyright: "Copyright © 2024-2026 BA4RHH",
      goVersion: "go1.26 (mock)",
      dataDir: mockDataDir,
    };
  },

  async pickDataDir(): Promise<string> {
    await wait(60);
    return mockDataDir;
  },

  async setDataDir(dir: string, migrate: boolean): Promise<string> {
    mockDataDir = dir;
    await wait(60);
    return `数据目录已设为 ${dir}${migrate ? "（已复制现有数据）" : "（空数据）"}，重启应用后生效`;
  },

  async resetDataDir(): Promise<string> {
    mockDataDir = "C:\\Users\\Demo\\AppData\\Roaming\\CracExamV3";
    await wait(60);
    return "已恢复默认数据目录，重启应用后生效";
  },
};
