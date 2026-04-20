现在让我给出详细的优化建议：

## 📋 项目代码Review与优化建议

### ✅ **优点**
1. 整体架构清晰，采用Wails框架实现Go后端 + Vue前端
2. 使用了DAO模式进行数据访问层设计
3. 有完善的日志系统（logrus + lumberjack）
4. 实现了事务处理（如ClearUserData、InvalidateExam）
5. 前端使用TypeScript和Pinia状态管理
6. 支持多种文件格式导入（JSON、Excel、CSV、PDF）

---

### ⚠️ **关键问题与优化建议**

#### **1. 数据库连接管理混乱（严重）**

**问题：**
- `backend/dao/db_pool.go` 和 `backend/services/db.go` 都维护了数据库实例
- 存在两套DB获取机制：`dao.GetDB()` 返回 `*gorm.DB`，`services.GetDB()` 返回 `*sql.DB`
- 导致服务层既接收 `*sql.DB` 又需要调用 `dao.GetGormDB()`

**建议：**
```go
// backend/dao/db_pool.go - 统一为单例管理
var (
    dbInstance *gorm.DB
    dbOnce     sync.Once
)

func InitDB() (*gorm.DB, error) {
    var initErr error
    dbOnce.Do(func() {
        dbPath := config.GetDatabasePath()
        dbInstance, initErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })
        if initErr != nil {
            return
        }
        // 自动迁移
        initErr = dao.InitDatabase()
    })
    return dbInstance, initErr
}

func GetDB() *gorm.DB {
    return dbInstance
}
```
```go
// backend/services/db.go - 简化或删除此文件
// 建议直接删除，在main.go中初始化
```


#### **2. 随机数种子重复设置（性能问题）**

**问题：**
```go
// exam_service.go 第180行和第242行
rand.Seed(time.Now().UnixNano())  // Go 1.20+ 已废弃，且重复调用
```


**建议：**
```go
// 在程序启动时设置一次即可（Go 1.20+不需要手动设置）
// 或者使用新的随机数API
import "math/rand/v2"

func (s *ExamService) shuffleQuestions(questions []*models.Question) {
    rand.Shuffle(len(questions), func(i, j int) {
        questions[i], questions[j] = questions[j], questions[i]
    })
}
```


#### **3. SQL注入风险**

**问题：**
```go
// question_dao.go 第171行 - 使用字符串拼接构建SQL
query := fmt.Sprintf(`SELECT ... FROM questions WHERE %s = 1`, categoryField)
```


**建议：**
虽然当前 `getCategoryField` 是硬编码的，但应改为更安全的做法：
```go
func (dao *QuestionDAO) GetRandomByCategoryAndType(category string, typeValue int, count int) ([]*models.Question, error) {
    // 使用白名单验证
    validCategories := map[string]string{
        "A": "LA",
        "B": "LB", 
        "C": "LC",
    }
    
    categoryField, ok := validCategories[category]
    if !ok {
        return nil, fmt.Errorf("invalid category: %s", category)
    }
    
    query := fmt.Sprintf(`SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
                          FROM questions WHERE %s = 1 AND type = ? ORDER BY RANDOM() LIMIT ?`, categoryField)
    
    rows, err := dao.ExecuteQuery(query, typeValue, count)
    // ...
}
```


#### **4. 错误处理不一致**

**问题：**
- 有些地方返回 `nil` 而不是错误（exam_service.go 第120行）
- 有些错误被忽略（main.go 第43行）

**建议：**
```go
// main.go
execPath, err := os.Executable()
if err != nil {
    log.Printf("Warning: failed to get executable path: %v", err)
    execPath = "."
}

// exam_service.go
if actualSingle+actualMultiple == 0 {
    return nil, fmt.Errorf("题库中没有足够的题目")
}
```


#### **5. 内存泄漏风险**

**问题：**
```go
// exam_service.go 第191行 - 每次shuffle都创建新map
originalOptions := map[string]string{
    "A": question.A,
    "B": question.B,
    // ...
}
```


**建议：**
对于大量题目，考虑复用对象池或减少临时对象创建。

#### **6. 前端API层冗余代码**

**问题：**
```typescript
// frontend/src/api/index.ts
export const UserService = {
  Login: async (username: string, idCard: string): Promise<services.UserLoginResponse> => {
    try {
      return await UserServiceBindings.Login(username, idCard)
    } catch (error) {
      console.error('UserService.Login error:', error)
      throw error  // 只是重新抛出，没有额外处理
    }
  },
}
```


**建议：**
创建统一的错误处理中间件：
```typescript
// frontend/src/api/client.ts
async function callBackend<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (error) {
    console.error('Backend API error:', error)
    // 可以添加统一的错误提示、重试逻辑等
    throw error
  }
}

// 使用时
export const UserService = {
  Login: (username: string, idCard: string) => 
    callBackend(() => UserServiceBindings.Login(username, idCard))
}
```


#### **7. 硬编码路径**

**问题：**
```go
// main.go 第50行
scriptDir = "D:\\crac_new\\crac_exam_go\\python_scripts"
```


**建议：**
```go
// 使用相对路径或配置文件
if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
    // 开发模式：使用可执行文件所在目录
    scriptDir = filepath.Join(execDir, "..", "python_scripts")
}
```


#### **8. 缺少输入验证**

**问题：**
- 用户输入未做充分验证
- 文件上传未检查大小和类型

**建议：**
```go
func (s *ImportService) ProcessUnifiedData(filePath string) (*ImportResult, error) {
    // 验证文件大小
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }
    if fileInfo.Size() > 100*1024*1024 { // 100MB限制
        return nil, fmt.Errorf("文件过大")
    }
    
    // 验证文件扩展名
    ext := strings.ToLower(filepath.Ext(filePath))
    allowedExts := map[string]bool{
        ".json": true, ".xlsx": true, ".xls": true, 
        ".csv": true, ".zip": true,
    }
    if !allowedExts[ext] {
        return nil, fmt.Errorf("不支持的文件类型")
    }
    
    // ...
}
```


#### **9. 数据库查询效率**

**问题：**
```go
// exam_service.go 第353行 - N+1查询问题
for _, detail := range examQuestionDetails {
    // 每次都更新数据库
    err = s.examQuestionDetail.Update(detail)
}
```


**建议：**
使用批量更新：
```go
func (s *ExamService) SubmitExam(...) {
    // 收集所有需要更新的详情
    updates := make([]*models.ExamQuestionDetail, 0)
    
    for _, detail := range examQuestionDetails {
        // 修改detail...
        updates = append(updates, detail)
    }
    
    // 批量更新
    err = s.examQuestionDetail.BulkUpdate(updates)
}
```


#### **10. 前端组件耦合度高**

**问题：**
```vue
<!-- ExamView.vue -->
const userStore = useUserStore()
// 直接在组件中调用store
```


**建议：**
通过props传递或使用组合式API解耦：
```typescript
// composables/useExam.ts
export function useExam() {
  const userStore = useUserStore()
  
  const startExam = async (category: string) => {
    // 考试逻辑
  }
  
  return { startExam }
}
```


#### **11. 缺少单元测试**

**建议：**
为核心业务逻辑添加测试：
```go
// backend/services/exam_service_test.go
func TestExamService_CreateExam(t *testing.T) {
    db := setupTestDB()
    service := NewExamService(db)
    
    // 插入测试题目
    insertTestQuestions(db)
    
    result, err := service.CreateExam(1, "A")
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, 50, len(result.Questions))
}
```


#### **12. 配置管理不完善**

**问题：**
- 考试配置硬编码在代码中
- 缺少运行时配置重载

**建议：**
```go
// backend/config/exam_config.yaml
type ExamConfig struct {
    Total      int `yaml:"total"`
    Single     int `yaml:"single"`
    Multiple   int `yaml:"multiple"`
    PassScore  int `yaml:"pass_score"`
    TimeMinutes int `yaml:"time_minutes"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var config Config
    yaml.Unmarshal(data, &config)
    return &config, nil
}
```


#### **13. 前端样式问题**

**问题：**
```css
/* ExamView.vue 第649行 */
.bottom-navigation-fixed {
  position: fixed;
  left: 50%;
  transform: translateX(-50%);
}
```


**建议：**
使用更现代的布局方式（Flexbox/Grid），避免fixed定位导致的z-index问题。

#### **14. 安全性问题**

**问题：**
- 用户密码明文存储（假设）
- 缺少CSRF保护
- 文件上传未做安全检查

**建议：**
```go
// 密码加密
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}
```


#### **15. 日志级别在生产环境未调整**

**问题：**
```go
// logger.go 第44行
logger.SetLevel(logrus.DebugLevel)  // 生产环境会产生大量日志
```


**建议：**
```go
func GetLogger() *logrus.Logger {
    loggerOnce.Do(func() {
        // ...
        level := logrus.InfoLevel
        if os.Getenv("APP_ENV") == "development" {
            level = logrus.DebugLevel
        }
        logger.SetLevel(level)
    })
    return logger
}
```


---

### 🎯 **优先级排序**

**高优先级（立即修复）：**
1. ✅ 统一数据库连接管理
2. ✅ 修复随机数种子问题
3. ✅ 添加输入验证和安全检查
4. ✅ 修复错误处理不一致

**中优先级（近期优化）：**
5. 批量操作优化数据库性能
6. 移除硬编码路径
7. 添加单元测试
8. 完善配置管理

**低优先级（长期改进）：**
9. 前端代码重构和解耦
10. 添加更多安全特性
11. 优化日志策略
12. 改进UI/UX

---

### 📊 **总体评价**

这是一个功能完整的业余无线电考试系统，代码质量中等偏上。主要优点是架构清晰、功能完整，但存在一些技术债务需要清理，特别是在数据库管理、错误处理和性能优化方面。建议按优先级逐步改进，特别是先解决数据库连接管理的混乱问题。