package models

import "time"

// User 本地用户档案。IDCard 仅作为本地身份标识，不做联网校验。
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	IDCard    string    `json:"idCard"`
	LastLogin time.Time `json:"lastLogin"`
}

// Question 题目（练习/错题/收藏/题库管理使用，包含答案 T）。
type Question struct {
	ID     int64  `json:"id"`
	J      string `json:"J"`
	P      string `json:"P"`
	I      string `json:"I"`
	Q      string `json:"Q"`
	T      string `json:"T"`
	A      string `json:"A"`
	B      string `json:"B"`
	C      string `json:"C"`
	D      string `json:"D"`
	F      string `json:"F"`
	LA     int    `json:"LA"`
	LB     int    `json:"LB"`
	LC     int    `json:"LC"`
	Type   int    `json:"type"`
	UserID int64  `json:"userId"`
	// WrongCount 仅错题查询时填充，用于随机练习权重与界面展示。
	WrongCount int `json:"wrongCount,omitempty"`
}

// ExamQuestion 考试中下发的题目：不包含正确答案，避免前端直接看到答案。
type ExamQuestion struct {
	ID   int64  `json:"id"`
	J    string `json:"J"`
	Q    string `json:"Q"`
	A    string `json:"A"`
	B    string `json:"B"`
	C    string `json:"C"`
	D    string `json:"D"`
	F    string `json:"F"`
	Type int    `json:"type"`
}

// ExamStartResponse 开始考试返回。
type ExamStartResponse struct {
	ExamID           int64          `json:"examId"`
	Category         string         `json:"category"`
	Questions        []ExamQuestion `json:"questions"`
	Total            int            `json:"total"`
	Single           int            `json:"single"`
	Multiple         int            `json:"multiple"`
	TimeLimitSeconds int            `json:"timeLimitSeconds"`
	PassScore        int            `json:"passScore"`
	StartedAt        time.Time      `json:"startedAt"`
}

// ExamResult 考试结果。
type ExamResult struct {
	ExamID          int64     `json:"examId"`
	Category        string    `json:"category"`
	ExamDate        time.Time `json:"examDate"`
	DurationSeconds int       `json:"durationSeconds"`
	Score           int       `json:"score"`
	CorrectCount    int       `json:"correctCount"`
	TotalCount      int       `json:"totalCount"`
	PassScore       int       `json:"passScore"`
	Passed          bool      `json:"passed"`
}

// ExamDetail 考试题目快照（用于结果复盘）。
type ExamDetail struct {
	ID            int64  `json:"id"`
	ExamID        int64  `json:"examId"`
	QuestionID    int64  `json:"questionId"`
	Seq           int    `json:"seq"`
	QuestionText  string `json:"questionText"`
	OptionA       string `json:"optionA"`
	OptionB       string `json:"optionB"`
	OptionC       string `json:"optionC"`
	OptionD       string `json:"optionD"`
	CorrectAnswer string `json:"correctAnswer"`
	UserAnswer    string `json:"userAnswer"`
	IsCorrect     bool   `json:"isCorrect"`
	Type          int    `json:"type"`
	ImageData     string `json:"imageData"`
}

// ExamSummary 考试记录摘要。
type ExamSummary struct {
	ID              int64     `json:"id"`
	Category        string    `json:"category"`
	ExamDate        time.Time `json:"examDate"`
	Score           int       `json:"score"`
	Total           int       `json:"total"`
	CorrectCount    int       `json:"correctCount"`
	PassScore       int       `json:"passScore"`
	Passed          bool      `json:"passed"`
	DurationSeconds int       `json:"durationSeconds"`
}

// PracticeData 练习页一次性返回的数据，避免每题一次 IPC。
type PracticeData struct {
	Mode        string     `json:"mode"`
	Category    string     `json:"category"`
	Questions   []Question `json:"questions"`
	FavoriteIDs []int64    `json:"favoriteIds"`
	Progress    int        `json:"progress"`
	Total       int        `json:"total"`
}

// QuestionPage 题库分页结果。
type QuestionPage struct {
	Data       []Question `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

// CategoryProgress 单个类别的练习进度。
type CategoryProgress struct {
	Category   string `json:"category"`
	Answered   int    `json:"answered"`
	Total      int    `json:"total"`
	Percent    int    `json:"percent"`
	LastAccess int64  `json:"lastAccess"`
}

// CategoryCount 单类别题目数量。
type CategoryCount struct {
	Category string `json:"category"`
	Total    int    `json:"total"`
	Single   int    `json:"single"`
	Multiple int    `json:"multiple"`
	WithImg  int    `json:"withImage"`
}

// Dashboard 首页聚合统计，一次调用取回全部数据。
type Dashboard struct {
	User             *User              `json:"user"`
	TotalQuestions   int                `json:"totalQuestions"`
	WithImages       int                `json:"withImages"`
	Categories       []CategoryCount    `json:"categories"`
	TotalExams       int                `json:"totalExams"`
	PassedExams      int                `json:"passedExams"`
	PassRate         float64            `json:"passRate"`
	BestScores       map[string]int     `json:"bestScores"`
	RecentExams      []ExamSummary      `json:"recentExams"`
	ErrorCount       int                `json:"errorCount"`
	FavoriteCount    int                `json:"favoriteCount"`
	Practice         []CategoryProgress `json:"practice"`
	LastExam         *ExamSummary       `json:"lastExam"`
	AverageScore     float64            `json:"averageScore"`
	TotalStudyMinute int                `json:"totalStudyMinutes"`
}

// ScorePoint 成绩趋势点。
type ScorePoint struct {
	Date     time.Time `json:"date"`
	Score    int       `json:"score"`
	Total    int       `json:"total"`
	Passed   bool      `json:"passed"`
	Category string    `json:"category"`
}

// ImportResult 题库导入结果。
type ImportResult struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message"`
	Imported     int            `json:"imported"`
	Total        int            `json:"total"`
	Stats        map[string]int `json:"stats"`
	Warnings     []string       `json:"warnings"`
	DuplicateJ   int            `json:"duplicateJ"`
	DuplicateJPI int            `json:"duplicateJpi"`
	ElapsedMS    int64          `json:"elapsedMs"`
	// 导入时按 J/P/I 尽量保留的用户数据
	PreservedFavorites int `json:"preservedFavorites"`
	PreservedErrors    int `json:"preservedErrors"`
	PreservedProgress  int `json:"preservedProgress"`
}

// ImportProgress 导入进度事件载荷。
type ImportProgress struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
	Percent int    `json:"percent"`
}

// BankUpdateSettings 题库自动更新设置。
type BankUpdateSettings struct {
	AutoCheck  bool `json:"autoCheck"`
	AutoImport bool `json:"autoImport"`
}

// BankUpdateStatus 题库更新状态（当前版本 + 远程检查结果）。
type BankUpdateStatus struct {
	Checked         bool      `json:"checked"`
	Checking        bool      `json:"checking"`
	Available       bool      `json:"available"`
	CurrentVersion  string    `json:"currentVersion"`
	CurrentSource   string    `json:"currentSource"`
	CurrentURL      string    `json:"currentUrl"`
	CurrentHash     string    `json:"currentHash"`
	CurrentImported time.Time `json:"currentImportedAt"`
	RemoteVersion   string    `json:"remoteVersion"`
	RemoteURL       string    `json:"remoteUrl"`
	RemoteSize      int64     `json:"remoteSize"`
	RemoteModified  string    `json:"remoteModified"`
	AutoCheck       bool      `json:"autoCheck"`
	AutoImport      bool      `json:"autoImport"`
	Message         string    `json:"message"`
	Error           string    `json:"error"`
}

// DBStats 数据库统计。
type DBStats struct {
	Users        int `json:"users"`
	Questions    int `json:"questions"`
	ExamRecords  int `json:"examRecords"`
	ErrorRecords int `json:"errorRecords"`
	Favorites    int `json:"favorites"`
}

// AppInfo 关于信息。
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Copyright   string `json:"copyright"`
	GoVersion   string `json:"goVersion"`
	DataDir     string `json:"dataDir"`
}
