# 业余无线电模拟考试系统

**专为 CRAC 业余无线电爱好者打造的智能化模拟考试平台**

[功能特性](#-功能特性) • [技术架构](#-技术架构) • [安装使用](#-安装与使用) • [开发指南](#-开发指南) • [常见问题](#-常见问题)

***

## 📖 项目简介

业余无线电模拟考试系统是一款专为中国无线电协会业余无线电分会（CRAC）设计的现代化考试辅助工具。系统支持 A/B/C 三类操作证书考试的完整功能，包括题库管理、模拟考试、错题收集、成绩统计等核心功能，帮助考生高效备考业余无线电操作证书考试。

## ✨ 功能特性

### 🎯 核心功能

#### 1. 题库管理

- **多格式导入**：支持 PDF 题库文件一键导入，自动识别题目、答案
- **智能分类**：自动识别并分类 A/B/C 三类题目
- **题目类型**：支持单选题、多选题、判断题等多种题型
- **图片支持**：支持含图题目，自动提取并保存题目图片
- **批量操作**：支持批量删除、导出、搜索题目

#### 2. 在线练习

- **逐题练习**：按类别选择题目，逐题作答并即时反馈
- **错题强化**：自动收集错题，针对性强化练习
- **收藏功能**：支持收藏重点题目，方便重点复习
- **进度保存**：自动保存练习进度，随时继续学习

#### 3. 模拟考试

- **真实模拟**：完全按照真实考试题型和题量组卷
  - A 类：32 道单选 + 8 道多选，40 题/40 分钟，30 题通过
  - B 类：45 道单选 + 15 道多选，60 题/60 分钟，45 题通过
  - C 类：70 道单选 + 20 道多选，90 题/90 分钟，70 题通过
- **智能组卷**：随机抽题，确保每次考试题目不同
- **计时功能**：实时显示剩余时间，时间到自动交卷
- **即时评分**：交卷后立即显示成绩和正确答案

#### 4. 学习统计

- **成绩趋势图**：可视化展示考试成绩变化趋势
- **分类统计**：按 A/B/C 类分别统计练习和考试情况
- **错题分析**：统计错题数量和分布，针对性复习
- **时间范围**：支持近 7 天、近半年、近一年、全部时间范围

### 🌟 产品亮点

1. **智能化**：PDF 题库自动识别，无需手动录入
2. **专业化**：完全符合 CRAC 考试大纲和题型要求
3. **人性化**：错题本、收藏夹、进度保存等贴心功能
4. **可视化**：成绩趋势图表，学习进度一目了然
5. **本地化**：所有数据本地存储，保护隐私安全
6. **绿色化**：单文件运行，无需安装，不写注册表

## 💻 技术架构

### 技术栈

#### 后端技术

- **开发语言**：Go 1.26.1
- **桌面框架**：[Wails v2.12.0](https://wails.io/) - 现代化的 Go 桌面应用框架
- **数据库**：SQLite (glebarez/sqlite - 纯 Go 实现，无需 CGO)
- **日志系统**：Logrus + Lumberjack (支持日志轮转)
- **PDF 处理**：
  - 文本提取：go-fitz (基于 MuPDF)
  - 图片提取：unipdf (oliverpolkerton fork)
- **依赖管理**：Go Modules

#### 前端技术

- **核心框架**：Vue 3.5.13 (Composition API)
- **UI 组件库**：Element Plus 2.9.6
- **状态管理**：Pinia 3.0.0
- **路由管理**：Vue Router 4.5.0
- **构建工具**：Vite 5.4.21
- **图表库**：ECharts 5.6.0
- **图标库**：@element-plus/icons-vue
- **开发语言**：TypeScript 5.8.3

#### 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                     前端层 (Vue 3)                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ 首页    │  │ 题库    │  │ 练习    │  │ 考试    │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ 错题本  │  │ 收藏夹  │  │ 统计    │  │ 设置    │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕ Wails Runtime
┌─────────────────────────────────────────────────────────┐
│                    服务层 (Go)                           │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ 题库    │  │ 练习    │  │ 考试    │  │ 统计    │   │
│  │ Service │  │ Service │  │ Service │  │ Service │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ PDF     │  │ 导入    │  │ 导出    │  │ 收藏    │   │
│  │ Importer│  │ Service │  │ Service │  │ Service │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                   数据访问层 (DAO)                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ 题目    │  │ 考试    │  │ 错题    │  │ 收藏    │   │
│  │ DAO     │  │ DAO     │  │ DAO     │  │ DAO     │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                    SQLite 数据库                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │questions│  │exam_    │  │error_   │  │favorite_│   │
│  │         │  │records  │  │questions│  │questions│   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 项目结构

```
crac_exam_go/
├── backend/                    # Go 后端代码
│   ├── config/                # 配置文件
│   ├── dao/                   # 数据访问层
│   ├── models/                # 数据模型
│   └── services/              # 业务逻辑层
├── frontend/                   # Vue 前端代码
│   ├── src/
│   │   ├── api/               # API 调用封装
│   │   ├── assets/            # 静态资源
│   │   ├── components/        # 公共组件
│   │   ├── router/            # 路由配置
│   │   ├── stores/            # Pinia 状态管理
│   │   ├── views/             # 页面组件
│   │   └── wailsjs/           # Wails 自动生成
│   ├── index.html             # 入口 HTML
│   └── package.json           # 前端依赖
├── build/                      # 构建输出目录
│   ├── bin/                   # 编译后的可执行文件
│   └── appicon.ico            # 应用图标
├── release/                    # 发布包目录
│   ├── crac_exam.exe          # 主程序
│   └── crac_exam-windows-amd64.zip  # 压缩包
├── logs/                       # 日志目录 (运行时生成)
├── data/                       # 数据目录 (运行时生成)
├── icon.ico                    # 应用图标
├── main.go                     # Go 入口文件
├── wails.json                  # Wails 配置文件
├── go.mod                      # Go 依赖管理
└── README.md                   # 项目说明文档
```

## 📦 安装与使用

### 系统要求

- **操作系统**：Windows 10 1803 及以上 / Windows 11
- **运行环境**：WebView2 Runtime（Windows 10 1803+ 已内置）
- **内存要求**：最低 2GB RAM，推荐 4GB RAM
- **磁盘空间**：最低 100MB 可用空间

### 安装步骤

#### 方法一：使用发布包（推荐）

1. 下载最新版本的 `crac_exam-windows-amd64.zip`
2. 解压到任意目录（建议：`D:\Programs\crac_exam\`）
3. 双击 `crac_exam.exe` 运行程序

#### 方法二：从源码编译

**前置要求**：

- Go 1.26+
- Node.js 18+
- Wails CLI

```bash
# 1. 克隆项目
git clone https://github.com/dracula-zz-1/crac_exam_go.git
cd crac_exam_go

# 2. 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 3. 安装前端依赖
cd frontend
npm install
cd ..

# 4. 开发模式运行
wails dev

# 5. 生产模式编译
wails build -platform windows/amd64 -trimpath -ldflags "-s -w -extldflags '-static'"
```

### 使用说明

#### 首次使用

1. **导入题库**
   - 进入「设置」页面
   - 将 CRAC 官方 PDF 题库文件拖拽到导入区域
   - 等待导入完成（约 1-2 分钟）
   - 查看导入统计，确认题目数量正确
2. **开始学习**
   - 返回首页，选择要练习的类别（A/B/C）
   - 点击「逐题练习」开始答题
   - 或点击「模拟考试」进行全真模拟

#### 功能使用

##### 题库管理

- **搜索题目**：支持按题号、题干、答案等关键词搜索
- **筛选分类**：可按 A/B/C 类筛选题目
- **查看题目**：点击题目查看详情，包括图片
- **删除题目**：支持单题删除和批量删除

##### 在线练习

- **选择类别**：在首页选择 A/B/C 类别
- **逐题练习**：系统随机出题，答完立即显示结果
- **错题自动收集**：答错的题目自动加入错题本
- **收藏题目**：点击星号收藏重点题目

##### 模拟考试

- **选择类别**：选择对应的考试类别
- **开始考试**：系统按真实考试规则组卷
- **答题界面**：显示题号、题目、选项、剩余时间
- **交卷评分**：时间到或手动交卷后立即显示成绩

##### 错题本

- **自动收集**：练习和考试中的错题自动加入
- **分类查看**：按 A/B/C 类分别查看错题
- **错题练习**：针对错题进行强化练习
- **移除错题**：掌握后可手动移除

##### 收藏夹

- **添加收藏**：在题目详情点击收藏按钮
- **分类管理**：按类别查看收藏的题目
- **收藏练习**：针对收藏题目进行练习

##### 学习统计

- **成绩趋势**：折线图展示考试成绩变化
- **时间筛选**：支持 7 天/半年/一年/全部时间范围
- **分类统计**：分别查看 A/B/C 类的统计数据
- **详细数据**：考试次数、通过率、平均分等

### 数据存储

所有数据存储在用户目录下的 `.crac_exam` 文件夹：

```
C:\Users\用户名\.crac_exam\
├── data\
│   └── exam_questions.db    # SQLite 数据库文件
└── logs\
    └── backend.log          # 应用日志文件
```

**优势**：

- 与程序分离，重装系统不丢失数据
- 多用户共享程序，数据独立
- 清理方便，删除文件夹即可清空数据

## 🛠️ 开发指南

### 环境搭建

#### 1. 安装 Go

```bash
# 下载并安装 Go 1.26+
# https://golang.org/dl/

# 验证安装
go version
```

#### 2. 安装 Node.js

```bash
# 下载并安装 Node.js 18+
# https://nodejs.org/

# 验证安装
node --version
npm --version
```

#### 3. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 4. 安装项目依赖

```bash
# 后端依赖
go mod tidy

# 前端依赖
cd frontend
npm install
```

### 开发命令

```bash
# 开发模式（支持热更新）
wails dev

# 生产模式编译
wails build

# 编译 Windows 64 位版本
wails build -platform windows/amd64

# 编译带调试信息的版本
wails build -debug

# 清理构建缓存
wails build -clean
```

### 数据库结构

#### 主要数据表

**questions (题目表)**

```sql
CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    J TEXT,          -- 题号
    P TEXT,          -- 分类
    I TEXT,          -- 唯一标识
    Q TEXT,          -- 题干
    T TEXT,          -- 正确答案
    A TEXT,          -- 选项 A
    B TEXT,          -- 选项 B
    C TEXT,          -- 选项 C
    D TEXT,          -- 选项 D
    F TEXT,          -- 解析
    LA INTEGER,      -- A 类标记
    LB INTEGER,      -- B 类标记
    LC INTEGER,      -- C 类标记
    type INTEGER,    -- 题目类型 (1=单选，2=多选)
    user_id INTEGER, -- 用户 ID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**exam\_records (考试记录表)**

```sql
CREATE TABLE exam_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    category TEXT,
    exam_date DATETIME,
    total_questions INTEGER,
    correct_count INTEGER,
    score REAL,
    duration_seconds INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**error\_questions (错题表)**

```sql
CREATE TABLE error_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    question_id INTEGER,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (question_id) REFERENCES questions(id)
);
```

**favorite\_questions (收藏表)**

```sql
CREATE TABLE favorite_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    question_id INTEGER,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (question_id) REFERENCES questions(id)
);
```

### API 调用

前端通过 Wails Runtime 调用后端 Go 方法：

```typescript
// 导入服务
import { PracticeService, ExamService } from '../wailsjs/go/services'

// 调用 Go 方法
const questions = await PracticeService.GetQuestionsByCategory(1, 'A')

// 处理错误
try {
  const result = await ExamService.CreateExam(userID, category)
} catch (error) {
  console.error('考试创建失败:', error)
}
```

### 添加新功能

#### 1. 添加后端服务

```go
// backend/services/new_service.go
package services

type NewService struct {
    // 依赖项
}

func (s *NewService) DoSomething(param string) (string, error) {
    // 业务逻辑
    return "result", nil
}
```

#### 2. 注册服务

```go
// main.go
err := wails.Run(&options.App{
    // ...
    Bind: []interface{}{
        &services.NewService{},
        // ...
    },
})
```

#### 3. 前端调用

```typescript
// frontend/src/api/NewService.ts
import { NewService as GoNewService } from '../wailsjs/go/services'

export const NewService = {
  DoSomething: async (param: string): Promise<string> => {
    return await GoNewService.DoSomething(param)
  },
}
```

## ❓ 常见问题

### Q1: 程序无法启动，提示 WebView2 缺失

**解决方案**：

1. Windows 10 1803+ 已内置 WebView2，请更新系统
2. 或手动安装 WebView2 Runtime：
   - 下载地址：<https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/>
   - 选择「固定版本」下载 x64 架构

### Q2: PDF 导入失败或题目不完整

**解决方案**：

1. 确保 PDF 文件是 CRAC 官方格式
2. 检查 PDF 文件是否损坏
3. 查看日志文件 `C:\Users\用户名\.crac_exam\logs\backend.log`
4. 尝试重新下载 PDF 题库

### Q3: 考试时图片无法显示

**解决方案**：

1. 检查题库导入时是否成功提取图片
2. 查看数据库 `images` 表是否有图片数据
3. 重新导入题库，确保图片提取成功

### Q4: 数据丢失或损坏

**解决方案**：

1. 检查数据文件是否存在：`C:\Users\用户名\.crac_exam\data\exam_questions.db`
2. 如有备份，恢复备份文件
3. 重新导入题库

### Q5: 程序运行缓慢

**解决方案**：

1. 关闭其他占用资源的程序
2. 清理数据库中的无用数据
3. 检查系统内存是否充足
4. 尝试重启程序

### Q6: 如何清空所有数据

**解决方案**：

1. 关闭程序
2. 删除文件夹：`C:\Users\用户名\.crac_exam\`
3. 重新启动程序，会自动创建新的数据文件

## 📝 更新日志

### v1.0.0 (2026-04-16)

**新增功能**

- ✅ 完整的 A/B/C 三类题目管理
- ✅ 逐题练习和模拟考试功能
- ✅ 错题本和收藏夹功能
- ✅ 考试成绩统计图表
- ✅ PDF 题库一键导入
- ✅ 题目图片自动提取

**优化改进**

- ✅ 数据存储到用户目录，与程序分离
- ✅ 单文件打包，包含所有 CGO 依赖
- ✅ 应用图标和界面优化
- ✅ 无数据时自动提示并返回首页

**Bug 修复**

- ✅ 修复多选题答案标记问题
- ✅ 修复错题本数据加载错误
- ✅ 修复图表单数据点不显示问题
- ✅ 修复 PDF 解析题干截断问题

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 🙏 致谢

感谢以下开源项目：

- [Wails](https://wails.io/) - 现代化的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 组件库
- [ECharts](https://echarts.apache.org/) - 强大的数据可视化库
- [go-fitz](https://github.com/gen2brain/go-fitz) - Go 语言 PDF 处理库
- [unipdf](https://github.com/oliverpool/unipdf) - Go 语言 PDF 处理库

***

**祝各位考生考试顺利，早日取得业余无线电操作证书！** 🎉
