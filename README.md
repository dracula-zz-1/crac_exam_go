# 业余无线电模拟考试 · HAM LAB v3

面向 CRAC 业余无线电操作技术能力验证的本地模拟考试系统。当前版本基于 **Wails 3.1**，重做了前端视觉与交互，题库导入改为 **仅支持 ZIP + 纯 Go 解析 PDF**。

> 真实题库（2025-08-09 版 ZIP）的逐字段校验结果见 [BANK_VERIFICATION.md](BANK_VERIFICATION.md)。

---

## ✨ 主要变化

### 1. 升级 Wails 3
- 使用 `github.com/wailsapp/wails/v3 v3.1.0`
- 服务通过 `application.NewService` 注册，绑定由 `wails3 generate bindings` 生成
- 原生文件选择框、原生文件拖放事件、导入进度事件（`import:progress`）
- 可调整窗口大小（最小 1100×720），不再固定 1024×768

### 2. 题库导入：只收 ZIP，纯 Go 解析
- 只接受 `.zip`，ZIP 内必须包含 5 份 PDF：
  - `总题库.pdf`
  - `总题库附图标记.pdf`
  - `A类题库.pdf`
  - `B类题库.pdf`
  - `C类题库.pdf`
- 文本提取：`github.com/ledongthuc/pdf`（纯 Go）
- 图片提取：`github.com/pdfcpu/pdfcpu`（纯 Go）
- ZIP 解压：标准库 `archive/zip`（支持 GBK 文件名）
- 不再依赖 Python / PyMuPDF / Pillow / unipdf / go-fitz
- 解析全部在内存/临时目录完成，**解析成功后才在单个事务里替换题库**；失败不会破坏旧题库
- 导入结果会返回题目数量、分类统计、带图数量、重复题号与警告

### 3. 修复的主要问题
| 旧版问题 | 本版处理 |
|---|---|
| 错题练习 `Q[:50]` 越界崩溃风险 | 已移除不安全切片 |
| 题库删除用 `parseInt(J)` 删错题 | 改为按数据库 `id` 删除，并清理收藏/错题引用 |
| 考试时长恒为 0（前端传“当前时间”） | 服务端开考记时，交卷时由服务端计算，前端不可伪造 |
| 导入先清库、失败即丢题库 | 解析与写库分离，写库使用单事务原子替换 |
| 弃考记录按 0 分污染统计 | 增加考试状态；启动时清理“进行中”记录，统计只算已交卷 |
| 交卷逐条更新、无幂等 | 交卷单事务 + 已交卷幂等返回 |
| 题目不足仍按固定及格线 | 题目不足直接拒绝开考，并给出具体缺题数量 |
| 大写 `.XLSX/.JSON` 不识别 | 现在只支持 ZIP，扩展名统一小写校验 |
| 图片字段 `F` 与 `hasImage` 不一致 | 全链路统一使用 `F`，考试/练习/错题/收藏均可显示图片 |
| 练习每翻一题重拉整个题库 | 进入页面一次加载，Pinia 缓存；翻题零 IPC |
| 收藏数/练习正确率统计错误 | 统计口径重写，一次 `GetDashboard` 返回首页全部数据 |
| 答案随考试题下发前端 | 考试接口返回不含 `T` 的题目 DTO，交卷由服务端判分 |
| 多个页面无入口、大量重复代码 | 统一为 8 个页面与共享组件，删除死代码 |
| SQLite 无 WAL / busy_timeout | 开启 WAL、busy_timeout=10s、单连接串行化 |
| 数据写在 exe 目录（Program Files 不可写） | 数据写入用户目录，可用 `CRAC_EXAM_DATA_DIR` 覆盖 |

### 4. 前端重做
- 不再使用 Element Plus / ECharts 全量包，改为自建设计系统与轻量 SVG 图表
- 视觉主题：深色侧栏 + 暖纸色内容区 + 琥珀色信号强调，呼应无线电主题
- 页面：登录、总览、逐题练习、模拟考试、考试结果复盘、错题本、收藏夹、题库管理、设置
- 交互：
  - 练习支持键盘 `A-D / 1-4` 选择、`Enter` 查看/下一题、`←/→` 翻题、`F` 收藏
  - 练习与考试均打乱选项顺序，正确答案随显示顺序重映射
  - 新用户登录后先记录当前操作等级（无证 / A / B / C），练习与考试默认按下一级进行
  - 题目列表/答题卡、进度环、正确/错误统计、图片放大
  - 练习页锁定一屏：题目区自适应，答案在右侧栏揭晓，不撑出滚动条
  - 考试倒计时、答题卡、未答提醒、时间到自动交卷、离开确认
  - 题库管理支持搜索、分类筛选（OR 语义）、分页、编辑、图片上传、删除
  - 导入进度实时显示、结果统计、警告列表；ZIP 可直接拖入窗口导入
  - 设置页支持检查、下载并运行应用更新，更新通道可选 GitHub 或 Gitee 镜像
  - 数据目录可更换（可连同现有题库与成绩一起复制到新目录）
- 浏览器预览模式：直接用 `npm run dev` 打开时使用内置演示数据，方便调 UI；桌面端自动使用真实 Go 后端

---

## 🧱 技术栈

| 层 | 技术 |
|---|---|
| 桌面框架 | Wails v3.1.0 |
| 后端 | Go 1.26 |
| 数据库 | SQLite（`modernc.org/sqlite`，纯 Go，无 CGO）+ 显式 SQL |
| PDF 文本 | `github.com/ledongthuc/pdf` |
| PDF 图片 | `github.com/pdfcpu/pdfcpu` |
| ZIP | `archive/zip` + `golang.org/x/text`（GBK 文件名） |
| 日志 | `log/slog` + `lumberjack`（10MB 轮转，保留 5 份） |
| 前端 | Vue 3 + TypeScript + Vite 8 + Pinia + Vue Router |
| 图表 | 手写 SVG（无 ECharts） |

---

## 🚀 开发与构建

### 环境
- Go 1.26+
- Node.js 18+
- Wails3 CLI `v3.1.0`
- Windows 10/11 + WebView2

### 开发模式
在仓库根目录执行：

```bash
wails3 dev
```

### 只预览前端（浏览器演示数据）
```bash
cd frontend
npm install
npm run dev
```

### 生产构建
```bash
wails3 build
```
产物：`bin/crac-exam-v3.exe`

### 测试
```bash
go test ./backend/...
npm --prefix frontend run build   # vue-tsc + vite build
```

---

## 📁 数据位置

默认位于用户应用数据目录下的 `CracExamV3`：

```
CracExamV3/
├── crac_exam.db
├── crac_exam.db-wal
├── crac_exam.db-shm
├── logs/app.log
├── temp/            # 导入时的临时解压目录，自动清理
└── updates/         # 应用更新包
```

可通过环境变量 `CRAC_EXAM_DATA_DIR` 指定其他数据目录（适用于便携部署或测试）。

> 本版使用新的数据库结构，不会自动迁移旧版数据。如需保留旧成绩，请先在旧版导出或重新导入题库。

---

## 📦 题库 ZIP 说明

1. 把以下 5 份 PDF 放进一个 ZIP（可放在子目录）：
   - 总题库.pdf
   - 总题库附图标记.pdf
   - A类题库.pdf
   - B类题库.pdf
   - C类题库.pdf
2. 在「设置 → 导入题库」点击选择，或直接把 ZIP 拖进窗口。
3. 确认后会显示解析/写库进度。
4. 导入是**原子替换**：会清空旧题库、错题本、收藏和练习进度；已交卷的考试复盘快照保留。
5. 导入完成后会给出统计与警告（例如重复题号、图片数量不匹配等）。

### 真实题库验证（2025-08-09 版 ZIP）

已用 2025-08-09 版题库 ZIP 实测：

- 1375 题：单选 1123、多选 252
- A / B / C 分类：683 / 1143 / 1282（同题可属于多类）
- 带图题 53 道；图片按 J 映射，与旧版数据逐题尺寸一致（新实现输出 JPEG，旧版输出 PNG）
- **J/P/I 联合键 1375 个全部唯一**，`duplicateJPI = 0`
- 与旧版数据库对比：J、P、T、题型、图片有无 **完全一致**
- 题干/选项差异主要来自：旧解析把页码混入正文、旧逻辑误删行尾数字；本版已修正
- 真实数据跑通 A/B/C 三场考试：40 / 60 / 90 题，全对均为满分并通过

回归命令：

```bash
set CRAC_TEST_ZIP=path/to/题库.zip
go test ./backend/importer -run TestRealBankZipIfProvided -v
```

也可以不启动 GUI，直接验证导入与 statistics/考试流程：

```bash
go run ./tools/verifyimport path/to/题库.zip
```

### 题目唯一键：J + P + I

- 数据库内部保留自增 `id` 作为引用键：考试题目详情、错题本、收藏夹都引用它，避免题目信息修改后引用失效。
- `(J, P, I)` 作为题目的**自然联合唯一键**，通过 `idx_questions_jpi` 唯一索引约束。
- 导入时按 `(J, P, I)` 去重：重复项只保留第一条，并在结果里返回 `duplicateJpi` 与警告。
- 题库编辑时校验联合键冲突，提示“题号（J/P/I 联合）已被其他题目使用”。
- 旧数据库如果存在联合键重复，唯一索引创建会失败并记录警告；应用仍可启动，导入/编辑路径仍会主动去重/校验。
- 单独的 `J` 允许重复（例如题库中的 `LX`），不影响题目身份。

### 纯 Go PDF 解析的边界
- 文本提取依赖 PDF 的字体编码与 `ToUnicode` 映射；若官方题库更新后使用了非常规编码，可能出现乱码/缺字段，导入结果会给出警告。
- 该题库的 PDF 会把行内数字/英文拆成独立文本对象。本版按页首精确剔除页码，并在中文↔ASCII 边界补空格，避免 `RMKS IS → RMKSIS`、`50Ω → Ω` 这类问题。
- 图片按“附图标记 PDF 的文本顺序 ↔ 图片对象顺序”匹配，并对 `LK060 → LK0603` 做兼容映射。
- 浏览器无法直接显示的图片格式（jpx/jbig2）会被跳过并计入警告。
- 超过 600MB 的 ZIP、单文件超过 512MB、解压总量超过 1.5GB 会被拒绝。
- ZIP 中的 `__MACOSX/` 与 `._` 资源叉文件会自动忽略。

---

## 🗂 目录结构

```
crac_exam_v3/
├── main.go                     # Wails3 应用入口、服务注册、窗口与拖放
├── backend/
│   ├── config/                 # 数据目录、考试配置
│   ├── models/                 # JSON 模型（Question/Exam/Dashboard…）
│   ├── store/                  # SQLite 打开、建表、事务化查询
│   ├── services/               # Wails 绑定服务
│   │   ├── user_service.go
│   │   ├── question_service.go
│   │   ├── exam_service.go     # 服务端计时、判分、幂等交卷
│   │   ├── practice_service.go # 练习/错题/收藏、一次取数
│   │   ├── statistics_service.go
│   │   ├── settings_service.go
│   │   └── import_service.go   # ZIP 导入、事务替换、进度事件
│   ├── importer/               # 纯 Go ZIP/PDF 解析
│   │   ├── zip.go
│   │   ├── pdf.go
│   │   ├── parse.go
│   │   └── *_test.go
│   └── utils/                  # slog + 轮转日志
├── tools/
│   ├── importbank/             # 命令行导入题库到默认数据目录
│   ├── verifybank/             # 题库完整性校验（数量/唯一键/分类/图片/源字段）
│   ├── verifyimport/           # 命令行导入 + 真实数据流程验证
│   ├── dumptext/               # 导出 PDF 文本用于排查解析问题
│   └── genicon/                # 生成应用图标
├── frontend/
│   ├── src/
│   │   ├── api/                # 真实 Wails 绑定 + 浏览器 mock
│   │   ├── components/         # AppShell/图标/图表/弹窗/选项…
│   │   ├── stores/             # session / bank 缓存 / toast
│   │   ├── views/              # 9 个页面
│   │   ├── styles/main.css     # 设计系统
│   │   └── types.ts
│   └── bindings/               # wails3 自动生成，勿手改
├── build/                      # 图标、清单、打包配置
├── bin/                        # 构建产物
└── Taskfile.yml
```

---

## ✅ 测试覆盖

- 题库完整性工具 `go run ./tools/verifybank <zip>`：
  - 数量、J/P/I 联合唯一、字段完整性、题型/答案一致性
  - A/B/C 与分类 PDF 逐题比对
  - 图片解码/数量/题号
  - 每个 Q/T/A-D 与原始 PDF 标记位置逐字段比对
- 导入器：
  - 标记解析、`J]` 变体、缺字段跳过
  - 按页剔除页码；保留行内数字（`50Ω`、`200`、`1、0、0、0`）
  - 中英文文本对象合并时的智能空格
  - 分类 `[I]` 提取
  - 生成的 ZIP + 5 份 PDF 端到端解析（含中文题干、图片、分类、重复题号警告）
  - **真实 2025-08-09 题库 ZIP** 回归（通过 `CRAC_TEST_ZIP` 启用）
  - 非 ZIP / 缺文件 / macOS 资源叉文件拒绝或忽略
- 服务层：
  - 登录建档与幂等
  - 开考（题量不足拒绝）、服务端计时、全对/全错交卷、幂等重复交卷
  - 弃考清理、错题去重
  - 练习进度 upsert、收藏/错题增删、题目删除清理引用
  - 题库分页 OR 语义、搜索
  - 清空题库保留已交卷快照、清空用户数据
  - 导入事务替换语义与失败保留旧题库
- 前端：`vue-tsc` 严格类型检查 + Vite 生产构建
- 手工 GUI 走查：登录、总览、练习、考试、交卷、结果复盘、错题本、题库管理、导入进度与结果

---

## ⚠️ 已知限制

- 本项目定位为**本地模拟练习**，不是官方考试系统；“身份证号”仅作为本地档案标识，不做联网校验。
- 纯 Go PDF 解析对极端排版的 PDF 可能不如 PyMuPDF 稳健；如果官方题库更新后解析异常，请把 ZIP 和导入警告一起反馈。
- 考试时长由服务端基于开考时间计算；若运行中系统时间被大幅调整，时长会随之变化。
- 数据库结构变更未做自动迁移；升级前建议导出/备份数据目录。
- 目前只针对 Windows 构建做了验证。

---

## 致谢

- [Wails](https://v3.wails.io/) — Go 桌面应用框架
- [ledongthuc/pdf](https://github.com/ledongthuc/pdf)、[pdfcpu](https://github.com/pdfcpu/pdfcpu) — 纯 Go PDF 解析
- CRAC 业余无线电题库的整理者

## 许可证

本项目采用 [MIT License](LICENSE)。
