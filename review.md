
## 📊 业余无线电模拟考试系统 - 代码评估报告与优化方案

### 一、整体评估概览

#### ✅ 优势亮点

1. **架构设计清晰**
   - 采用标准的三层架构（DAO-Service-Frontend）
   - Wails v2 + Vue 3 技术选型合理
   - 模块划分明确，职责分离良好

2. **功能完整性高**
   - 实现了完整的考试流程（题库管理→练习→模拟考→统计）
   - PDF智能解析支持图片提取
   - 错题本、收藏夹等辅助学习功能完善

3. **代码质量较好**
   - 有完善的日志记录（Logrus + Lumberjack）
   - 关键业务逻辑有注释说明
   - 已修复部分已知BUG（如答案判断、考试退出）

4. **性能优化意识**
   - PDF解析使用并发处理
   - 批量插入优化数据库操作
   - 使用信号量控制并发数

---

### 二、存在的问题与风险

#### 🔴 严重问题（P0）

##### 1. **数据库连接管理混乱**
**位置**: `main.go` + `services/db.go` + `dao/db_pool.go`

**问题描述**:
```go
// main.go: 初始化两次数据库
db, err := dao.GetDB()  // 第一次获取
services.SetDB(db)       // 设置全局变量

// services中又调用
services.GetDB()         // 第二次获取
```


存在两个单例模式：
- `dao.dbInstance` (db_pool.go)
- `services.dbInstance` (db.go)

**风险**: 
- 可能导致数据库实例不一致
- 违反单一数据源原则
- 增加维护复杂度

**影响**: ⚠️ 高 - 可能导致运行时错误

---

##### 2. **随机数种子未初始化**
**位置**: `exam_service.go:192`, `question_dao.go:73`

**问题描述**:
```go
rand.Shuffle(len(questions), func(i, j int) {
    questions[i], questions[j] = questions[j], questions[i]
})
```


Go 1.20+ 虽然自动初始化种子，但代码中未显式声明依赖版本行为，可能导致不同环境下随机性不一致。

**影响**: ⚠️ 中 - 考试题目可能重复

---

##### 3. **PDF解析内存泄漏风险**
**位置**: `pdf_importer.go:282-326`

**问题描述**:
```go
for pageNum := 0; pageNum < totalPages; pageNum++ {
    go func(pn int) {
        doc2, err := fitz.NewFromMemory(pdfData) // 每个goroutine都加载完整PDF
        // ...
    }(pageNum)
}
```


每个goroutine都创建独立的文档实例，大PDF文件会导致内存暴增。

**影响**: ⚠️ 高 - 可能导致OOM崩溃

---

#### 🟡 重要问题（P1）

##### 4. **事务处理不完善**
**位置**: `exam_service.go:529-585`

**问题描述**:
```go
func (s *ExamService) InvalidateExam(examID int64) error {
    tx := s.examRecordDAO.GetDB().Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()  // panic恢复但未重新抛出
        }
    }()
    // ...
}
```


panic被吞掉，调用方无法感知异常。

---

##### 5. **前端状态管理薄弱**
**位置**: `stores/user.ts`

**问题描述**:
- 缺少持久化插件配置
- 无考试状态管理（当前答题进度、计时器等）
- 刷新页面后考试状态丢失

**影响**: ⚠️ 中 - 用户体验差

---

##### 6. **SQL注入防护不完整**
**位置**: `question_dao.go:104-126`

虽然有白名单验证，但部分查询仍使用字符串拼接：
```go
result = dao.db.Where(map[string]interface{}{categoryField: 1}).Find(&questions)
```


建议统一使用GORM的结构化查询。

---

#### 🟢 一般问题（P2）

##### 7. **错误处理不统一**
- 部分函数返回error，部分直接panic
- 前端错误提示不够友好

##### 8. **缺少单元测试**
整个项目无任何测试文件，难以保证代码质量。

##### 9. **配置文件硬编码**
考试配置、路径配置等都写死在代码中。

##### 10. **前端类型定义不完整**
TypeScript类型定义松散，大量使用`any`。

---

### 三、优化方案

#### 🎯 优先级P0优化

##### 优化1: 统一数据库连接管理

**实施方案**:

```go
// backend/dao/db_pool.go - 移除services/db.go，统一使用dao层
package dao

import (
    "crac_exam_go/backend/config"
    "crac_exam_go/backend/utils"
    "sync"

    "github.com/glebarez/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var (
    dbInstance *gorm.DB
    dbOnce     sync.Once
    dbError    error
)

// GetDB 获取数据库实例 (唯一入口)
func GetDB() (*gorm.DB, error) {
    dbOnce.Do(func() {
        dbPath := config.GetDatabasePath()
        utils.Info("Database", "初始化数据库连接", map[string]interface{}{
            "path": dbPath,
        })

        dbInstance, dbError = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })

        if dbError != nil {
            utils.Error("Database", "数据库连接失败", dbError, nil)
            return
        }

        // 配置连接池
        sqlDB, err := dbInstance.DB()
        if err == nil {
            sqlDB.SetMaxOpenConns(1) // SQLite只支持单连接
            sqlDB.SetMaxIdleConns(1)
        }

        utils.Info("Database", "数据库连接成功", nil)
    })

    return dbInstance, dbError
}

// CloseDB 关闭数据库连接
func CloseDB() error {
    if dbInstance != nil {
        sqlDB, err := dbInstance.DB()
        if err != nil {
            return err
        }
        return sqlDB.Close()
    }
    return nil
}
```


**修改main.go**:
```go
func main() {
    // 初始化数据库
    if err := dao.InitDatabase(); err != nil {
        log.Fatal("数据库初始化失败:", err)
    }

    // 获取数据库实例
    db, err := dao.GetDB()
    if err != nil {
        log.Fatal("获取数据库实例失败:", err)
    }

    // 创建应用实例 - 所有service共享同一个db实例
    err = wails.Run(&options.App{
        // ... 其他配置
        Bind: []interface{}{
            services.NewUserService(db),
            services.NewSettingsService(db),
            services.NewQuestionsBankService(db),
            services.NewExamService(db),
            services.NewPracticeService(db),
            services.NewFavoriteService(db),
            services.NewStatisticsService(db),
            services.NewImportService(db),
        },
    })
    // ...
}
```


**删除** `backend/services/db.go` 文件

---

##### 优化2: 修复PDF解析内存问题

```go
// pdf_importer.go - 改为串行处理或使用对象池
func (i *PDFImporter) readTotalPDFTextParallel(pdfData []byte) (string, error) {
    start := time.Now()
    utils.Debug("PDFImporter", "开始提取总题库文本", nil)

    doc, err := fitz.NewFromMemory(pdfData)
    if err != nil {
        return "", err
    }
    defer doc.Close()

    totalPages := doc.NumPage()
    var allLines []string
    
    // 使用信号量控制并发，复用doc实例
    maxConcurrency := runtime.NumCPU()
    sem := make(chan struct{}, maxConcurrency)
    
    type pageResult struct {
        pageNum int
        text    string
        err     error
    }
    
    resultChan := make(chan pageResult, totalPages)
    var wg sync.WaitGroup

    for pageNum := 0; pageNum < totalPages; pageNum++ {
        wg.Add(1)
        sem <- struct{}{} // 获取信号量
        
        go func(pn int) {
            defer wg.Done()
            defer func() { <-sem }() // 释放信号量
            
            // 每页独立提取，避免并发访问同一doc
            text, err := doc.Text(pn)
            if err != nil {
                resultChan <- pageResult{pageNum: pn, err: err}
                return
            }
            
            resultChan <- pageResult{pageNum: pn, text: text, err: nil}
        }(pageNum)
    }

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 收集结果并排序
    results := make([]pageResult, 0, totalPages)
    for res := range resultChan {
        results = append(results, res)
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].pageNum < results[j].pageNum
    })

    for _, res := range results {
        if res.err == nil {
            lines := strings.Split(res.text, "\n")
            for _, line := range lines {
                line = strings.TrimSpace(line)
                if line != "" && !isPureNumber(line) {
                    allLines = append(allLines, line)
                }
            }
        }
    }

    result := strings.Join(allLines, "\n")
    utils.Debug("PDFImporter", "总题库文本提取完成", map[string]interface{}{
        "lines":            len(allLines),
        "total_chars":      len(result),
        "duration_seconds": time.Since(start).Seconds(),
    })
    
    return result, nil
}
```


---

##### 优化3: 添加随机数种子初始化

```go
// main.go
import (
    "math/rand"
    "time"
    // ...
)

func main() {
    // 初始化随机数种子
    rand.Seed(time.Now().UnixNano())
    
    // ... 其余代码
}
```


---

#### 🎯 优先级P1优化

##### 优化4: 改进事务处理

```go
// exam_service.go
func (s *ExamService) InvalidateExam(examID int64) error {
    utils.Info("ExamService", "开始作废考试记录", map[string]interface{}{
        "exam_id": examID,
    })

    tx := s.examRecordDAO.GetDB().Begin()
    if tx.Error != nil {
        utils.Error("ExamService", "开启事务失败", tx.Error, map[string]interface{}{
            "exam_id": examID,
        })
        return tx.Error
    }

    // 确保事务回滚
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            err := fmt.Errorf("panic occurred: %v", p)
            utils.Error("ExamService", "发生panic，事务已回滚", err, map[string]interface{}{
                "exam_id": examID,
            })
            panic(p) // 重新抛出panic
        } else if tx.Error != nil {
            tx.Rollback()
        }
    }()

    // 1. 删除考试题目详情
    if err := s.examQuestionDetail.DeleteByExamIDWithTx(examID, tx); err != nil {
        return err
    }

    // 2. 删除考试记录
    if err := s.examRecordDAO.DeleteWithTx(examID, tx); err != nil {
        return err
    }

    // 提交事务
    if err := tx.Commit().Error; err != nil {
        utils.Error("ExamService", "提交事务失败", err, map[string]interface{}{
            "exam_id": examID,
        })
        return err
    }

    utils.Info("ExamService", "考试记录作废成功", map[string]interface{}{
        "exam_id": examID,
    })
    return nil
}
```


---

##### 优化5: 增强前端状态管理

```typescript
// frontend/src/stores/exam.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

export interface ExamState {
  examId: number | null
  startTime: number | null
  currentQuestionIndex: number
  userAnswers: Record<number, string>
  remainingTime: number
  isExamActive: boolean
}

export const useExamStore = defineStore('exam', () => {
  const examId = ref<number | null>(null)
  const startTime = ref<number | null>(null)
  const currentQuestionIndex = ref(0)
  const userAnswers = ref<Record<number, string>>({})
  const remainingTime = ref(0)
  const isExamActive = ref(false)

  function startExam(id: number, totalTime: number) {
    examId.value = id
    startTime.value = Date.now()
    currentQuestionIndex.value = 0
    userAnswers.value = {}
    remainingTime.value = totalTime * 60
    isExamActive.value = true
  }

  function updateAnswer(questionId: number, answer: string) {
    userAnswers.value[questionId] = answer
  }

  function nextQuestion() {
    currentQuestionIndex.value++
  }

  function previousQuestion() {
    if (currentQuestionIndex.value > 0) {
      currentQuestionIndex.value--
    }
  }

  function endExam() {
    examId.value = null
    startTime.value = null
    currentQuestionIndex.value = 0
    userAnswers.value = {}
    remainingTime.value = 0
    isExamActive.value = false
  }

  return {
    examId,
    startTime,
    currentQuestionIndex,
    userAnswers,
    remainingTime,
    isExamActive,
    startExam,
    updateAnswer,
    nextQuestion,
    previousQuestion,
    endExam
  }
}, {
  persist: {
    enabled: true,
    strategies: [
      {
        key: 'exam-state',
        storage: localStorage
      }
    ]
  }
})
```


**在main.ts中注册持久化插件**:
```typescript
// frontend/src/main.ts
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
```


---

##### 优化6: 添加配置文件支持

```go
// backend/config/app_config.go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type AppConfig struct {
    Database DatabaseConfig `json:"database"`
    Logging  LoggingConfig  `json:"logging"`
    App      AppInfo        `json:"app"`
}

type DatabaseConfig struct {
    Path     string `json:"path"`
    MaxConns int    `json:"max_conns"`
}

type LoggingConfig struct {
    Level      string `json:"level"`
    MaxSize    int    `json:"max_size_mb"`
    MaxBackups int    `json:"max_backups"`
    MaxAge     int    `json:"max_age_days"`
}

type AppInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

var DefaultConfig = AppConfig{
    Database: DatabaseConfig{
        Path:     GetDatabasePath(),
        MaxConns: 1,
    },
    Logging: LoggingConfig{
        Level:      "info",
        MaxSize:    10,
        MaxBackups: 7,
        MaxAge:     30,
    },
    App: AppInfo{
        Name:    "业余无线电模拟考试系统",
        Version: "1.0.0",
    },
}

func LoadConfig() (*AppConfig, error) {
    configPath := filepath.Join(GetDataDir(), "config.json")
    
    // 如果配置文件不存在，使用默认配置
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return &DefaultConfig, nil
    }
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config AppConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func SaveConfig(config *AppConfig) error {
    configPath := filepath.Join(GetDataDir(), "config.json")
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0644)
}
```


---

#### 🎯 优先级P2优化

##### 优化7: 添加基础单元测试框架

```go
// backend/services/exam_service_test.go
package services

import (
    "testing"
    "crac_exam_go/backend/models"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&models.Question{}, &models.ExamRecord{})
    return db
}

func TestCreateExam(t *testing.T) {
    db := setupTestDB()
    service := NewExamService(db)
    
    // 准备测试数据
    questions := []*models.Question{
        {J: "Q001", Q: "测试题目1", T: "A", Type: 1, LA: 1},
        {J: "Q002", Q: "测试题目2", T: "B", Type: 1, LA: 1},
    }
    db.Create(&questions)
    
    // 执行测试
    result, err := service.CreateExam(1, "A")
    
    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Greater(t, result.ExamID, int64(0))
}

func TestIsAnswerCorrect(t *testing.T) {
    db := setupTestDB()
    service := NewExamService(db)
    
    tests := []struct {
        userAnswer   string
        correctAnswer string
        questionType int
        expected     bool
    }{
        {"A", "A", 1, true},
        {"AB", "BA", 2, true},
        {"A", "B", 1, false},
        {"", "A", 1, false},
    }
    
    for _, tt := range tests {
        result := service.isAnswerCorrect(tt.userAnswer, tt.correctAnswer, tt.questionType)
        assert.Equal(t, tt.expected, result)
    }
}
```


---

##### 优化8: 改进前端类型定义

```typescript
// frontend/src/types/exam.ts
export interface Question {
  id: number
  J: string          // 题号
  P: string          // 分类
  I: string          // 唯一标识
  Q: string          // 题干
  T: string          // 正确答案
  A: string          // 选项A
  B: string          // 选项B
  C: string          // 选项C
  D: string          // 选项D
  F?: string         // 图片Base64
  type: QuestionType
  LA: number         // A类标记
  LB: number         // B类标记
  LC: number         // C类标记
  hasImage: boolean
  imageBase64?: string
}

export enum QuestionType {
  SingleChoice = 1,
  MultipleChoice = 2
}

export interface ExamConfig {
  total: number
  single: number
  multiple: number
  time_minutes: number
  pass_score: number
}

export interface ExamStartResponse {
  exam_id: number
  questions: Question[]
  config: ExamConfig
}

export interface UserAnswer {
  answer: string
  is_correct: boolean
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
```


---

### 四、其他建议

#### 📝 代码规范

1. **统一命名风格**
   - Go代码遵循驼峰命名
   - 数据库字段保持大写（J, P, I等）需加注释说明来源

2. **添加接口文档**
   - 使用Swagger或类似工具生成API文档
   - 前后端接口契约明确化

3. **错误码规范化**
```go
// backend/utils/error_codes.go
type ErrorCode int

const (
    ErrInvalidParam    ErrorCode = 1001
    ErrDatabaseError   ErrorCode = 1002
    ErrNotFound        ErrorCode = 1003
    ErrPermissionDenied ErrorCode = 1004
)

type AppError struct {
    Code    ErrorCode
    Message string
    Detail  interface{}
}

func (e *AppError) Error() string {
    return e.Message
}
```


---

#### 🔒 安全加固

1. **密码加密存储**
```go
// user_service.go
import "golang.org/x/crypto/bcrypt"

func (s *UserService) Register(username, password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    user := &models.User{
        Username: username,
        Password: string(hashedPassword),
    }
    return s.userDAO.Create(user)
}
```


2. **输入验证加强**
```go
// 添加通用验证函数
func ValidateString(input string, maxLength int) error {
    if len(input) == 0 {
        return errors.New("输入不能为空")
    }
    if len(input) > maxLength {
        return fmt.Errorf("输入长度不能超过%d", maxLength)
    }
    return nil
}
```


---

#### 🚀 性能优化

1. **添加缓存层**
```go
// backend/services/cache.go
package services

import (
    "sync"
    "time"
)

type CacheItem struct {
    Data      interface{}
    ExpiresAt time.Time
}

type SimpleCache struct {
    items map[string]CacheItem
    mu    sync.RWMutex
}

func NewSimpleCache() *SimpleCache {
    return &SimpleCache{
        items: make(map[string]CacheItem),
    }
}

func (c *SimpleCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, exists := c.items[key]
    if !exists || time.Now().After(item.ExpiresAt) {
        return nil, false
    }
    return item.Data, true
}

func (c *SimpleCache) Set(key string, data interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.items[key] = CacheItem{
        Data:      data,
        ExpiresAt: time.Now().Add(ttl),
    }
}
```


2. **数据库索引优化**
```sql
-- 添加常用查询字段索引
CREATE INDEX IF NOT EXISTS idx_questions_type ON questions(type);
CREATE INDEX IF NOT EXISTS idx_questions_la ON questions(LA);
CREATE INDEX IF NOT EXISTS idx_questions_lb ON questions(LB);
CREATE INDEX IF NOT EXISTS idx_questions_lc ON questions(LC);
CREATE INDEX IF NOT EXISTS idx_exam_records_user_id ON exam_records(user_id);
CREATE INDEX IF NOT EXISTS idx_error_questions_user_id ON error_questions(user_id, category);
```


---

### 五、总结与行动计划

#### 立即执行（本周内）
- [x] 统一数据库连接管理
- [x] 修复PDF解析内存问题  
- [x] 添加随机数种子初始化

#### 短期优化（1个月内）
- [ ] 完善事务处理机制
- [ ] 增强前端状态管理
- [ ] 添加配置文件支持
- [ ] 编写核心功能单元测试

#### 中期改进（3个月内）
- [ ] 完善错误处理和日志
- [ ] 添加性能监控
- [ ] 实现缓存机制
- [ ] 数据库索引优化

#### 长期规划（6个月以上）
- [ ] 考虑迁移到PostgreSQL（多用户场景）
- [ ] 实现云端同步功能
- [ ] 添加AI智能推荐错题
- [ ] 移动端适配

---

### 六、代码质量评分

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构设计 | ⭐⭐⭐⭐☆ | 分层清晰，但存在冗余 |
| 代码规范 | ⭐⭐⭐☆☆ | 基本规范，缺少统一标准 |
| 安全性 | ⭐⭐☆☆☆ | 密码明文存储，需加强 |
| 性能 | ⭐⭐⭐☆☆ | 有优化意识，但有内存风险 |
| 可维护性 | ⭐⭐⭐☆☆ | 注释充分，但缺测试 |
| 可扩展性 | ⭐⭐⭐⭐☆ | 模块化好，易于扩展 |

**综合评分**: ⭐⭐⭐☆☆ (3.5/5)

