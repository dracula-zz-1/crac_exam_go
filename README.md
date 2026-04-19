# Amateur Radio Exam System

**An intelligent exam simulation platform for CRAC (China Radio Association) amateur radio enthusiasts**

[🌐 English](README_en.md) | [🇨🇳 中文](README.md)

[Features](#-features) • [Technology Stack](#-technology-stack) • [Installation](#-installation-and-usage) • [Development Guide](#-development-guide) • [FAQ](#-faq)

***

## 📖 Introduction

The Amateur Radio Exam System is a modern exam preparation tool designed for the China Radio Association Amateur Radio Branch (CRAC). The system supports complete functionality for A/B/C class operator certificate exams, including question bank management, mock exams, error collection, and score statistics, helping candidates efficiently prepare for amateur radio operator certificate exams.

## ✨ Features

### 🎯 Core Features

#### 1. Question Bank Management

- **Multi-format Import**: One-click PDF question bank import with automatic recognition of questions and answers
- **Smart Classification**: Automatically identifies and classifies A/B/C class questions
- **Question Types**: Supports single-choice, multiple-choice, true/false questions
- **Image Support**: Supports questions with images, automatically extracts and saves them
- **Batch Operations**: Supports batch delete, export, and search questions

#### 2. Online Practice

- **Question-by-Question Practice**: Select questions by category, answer one by one with instant feedback
- **Error Reinforcement**: Automatically collects wrong answers for targeted practice
- **Favorites**: Bookmark important questions for focused review
- **Progress Saving**: Automatically saves practice progress, continue anytime

#### 3. Mock Exam

- **Real Simulation**: Completely follows real exam question types and counts
  - Class A: 32 single-choice + 8 multiple-choice, 40 questions/40 minutes, 30 correct to pass
  - Class B: 45 single-choice + 15 multiple-choice, 60 questions/60 minutes, 45 correct to pass
  - Class C: 70 single-choice + 20 multiple-choice, 90 questions/90 minutes, 70 correct to pass
- **Smart Paper Generation**: Random question selection ensures different questions each time
- **Timer**: Real-time countdown with automatic submission when time expires
- **Instant Scoring**: Shows score and correct answers immediately after submission

#### 4. Learning Statistics

- **Score Trend Chart**: Visual display of exam score changes
- **Category Statistics**: Separate statistics for A/B/C class practice and exams
- **Error Analysis**: Statistics on error count and distribution for targeted review
- **Time Range**: Supports last 7 days, last half year, last year, all time range

### 🌟 Highlights

1. **Intelligent**: Automatic PDF question bank recognition, no manual entry needed
2. **Professional**: Fully complies with CRAC exam syllabus and question type requirements
3. **User-friendly**: Thoughtful features like error book, favorites, progress saving
4. **Visual**: Score trend charts, learning progress at a glance
5. **Local**: All data stored locally, protecting privacy and security
6. **Portable**: Single file execution, no installation required, no registry writes

## 💻 Technology Stack

### Tech Stack

#### Backend

- **Language**: Go 1.26.1
- **Desktop Framework**: [Wails v2.12.0](https://wails.io/) - Modern Go desktop application framework
- **Database**: SQLite (glebarez/sqlite - Pure Go implementation, no CGO required)
- **Logging**: Logrus + Lumberjack (supports log rotation)
- **PDF Processing**:
  - Text Extraction: go-fitz (based on MuPDF)
  - Image Extraction: unipdf (oliverpolkerton fork)
- **Dependency Management**: Go Modules

#### Frontend

- **Framework**: Vue 3.5.13 (Composition API)
- **UI Library**: Element Plus 2.9.6
- **State Management**: Pinia 3.0.0
- **Router**: Vue Router 4.5.0
- **Build Tool**: Vite 5.4.21
- **Chart Library**: ECharts 5.6.0
- **Icons**: @element-plus/icons-vue
- **Language**: TypeScript 5.8.3

#### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Frontend Layer (Vue 3)              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Home    │  │ Question│  │ Practice│  │ Exam    │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Errors  │  │ Favorites│ │ Stats   │  │ Settings│   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕ Wails Runtime
┌─────────────────────────────────────────────────────────┐
│                    Service Layer (Go)                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Question│  │ Practice│  │ Exam    │  │ Stats   │   │
│  │ Service │  │ Service │  │ Service │  │ Service │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ PDF     │  │ Import  │  │ Export  │  │ Favorite│   │
│  │ Importer│  │ Service │  │ Service │  │ Service │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                   Data Access Layer (DAO)                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Question│  │ Exam    │  │ Error   │  │ Favorite│   │
│  │ DAO     │  │ DAO     │  │ DAO     │  │ DAO     │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                    SQLite Database                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │questions│  │exam_    │  │error_   │  │favorite_│   │
│  │         │  │records  │  │questions│  │questions│   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
crac_exam_go/
├── backend/                    # Go backend code
│   ├── config/                # Configuration files
│   ├── dao/                   # Data access layer
│   ├── models/                # Data models
│   └── services/              # Business logic layer
├── frontend/                   # Vue frontend code
│   ├── src/
│   │   ├── api/               # API call wrappers
│   │   ├── assets/            # Static assets
│   │   ├── components/        # Common components
│   │   ├── router/            # Router configuration
│   │   ├── stores/            # Pinia state management
│   │   ├── views/             # Page components
│   │   └── wailsjs/           # Wails auto-generated
│   ├── index.html             # Entry HTML
│   └── package.json           # Frontend dependencies
├── build/                      # Build output directory
│   ├── bin/                   # Compiled executable
│   └── appicon.ico            # Application icon
├── release/                    # Release package directory
│   ├── crac_exam.exe          # Main program
│   └── crac_exam-windows-amd64.zip  # Compressed package
├── logs/                       # Log directory (generated at runtime)
├── data/                       # Data directory (generated at runtime)
├── icon.ico                    # Application icon
├── main.go                     # Go entry file
├── wails.json                  # Wails configuration
├── go.mod                      # Go dependencies
└── README.md                   # Project documentation
```

## 📦 Installation and Usage

### System Requirements

- **OS**: Windows 10 1803+ / Windows 11
- **Runtime**: WebView2 Runtime (built-in for Windows 10 1803+)
- **Memory**: Minimum 2GB RAM, recommended 4GB RAM
- **Disk**: Minimum 100MB free space

### Installation

#### Method 1: Using Release Package (Recommended)

1. Download the latest `crac_exam-windows-amd64.zip`
2. Extract to any directory (recommended: `D:\Programs\crac_exam\`)
3. Double-click `crac_exam.exe` to run

#### Method 2: Build from Source

**Prerequisites**:

- Go 1.26+
- Node.js 18+
- Wails CLI

```bash
# 1. Clone the repository
git clone https://github.com/dracula-zz-1/crac_exam_go.git
cd crac_exam_go

# 2. Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 3. Install frontend dependencies
cd frontend
npm install
cd ..

# 4. Run in development mode
wails dev

# 5. Build for production
wails build -platform windows/amd64 -trimpath -ldflags "-s -w -extldflags '-static'"
```

### Usage

#### First Time Setup

1. **Import Question Bank**
   - Go to "Settings" page
   - Drag and drop CRAC official PDF question bank into the import area
   - Wait for import to complete (about 1-2 minutes)
   - Check import statistics to verify question count
2. **Start Learning**
   - Return to homepage, select category (A/B/C)
   - Click "Question-by-Question Practice" to start
   - Or click "Mock Exam" for full simulation

#### Feature Usage

##### Question Bank Management

- **Search**: Search by question ID, stem, answer keywords
- **Filter**: Filter by A/B/C class
- **View**: Click question to view details including images
- **Delete**: Support single and batch delete

##### Online Practice

- **Select Category**: Choose A/B/C category on homepage
- **Practice**: System randomly selects questions, shows results immediately
- **Auto Collect Errors**: Wrong answers automatically added to error book
- **Favorite**: Click star to bookmark important questions

##### Mock Exam

- **Select Category**: Choose corresponding exam category
- **Start Exam**: System generates paper according to real exam rules
- **Answer Interface**: Shows question number, stem, options, remaining time
- **Submit & Score**: Shows score immediately after time expires or manual submit

##### Error Book

- **Auto Collection**: Errors from practice and exams automatically added
- **Category View**: View errors by A/B/C class separately
- **Error Practice**: Targeted practice on error questions
- **Remove**: Manually remove mastered questions

##### Favorites

- **Add**: Click favorite button in question details
- **Category Management**: View favorites by category
- **Favorite Practice**: Practice on favorited questions

##### Statistics

- **Score Trend**: Line chart showing exam score changes
- **Time Filter**: Supports 7 days/half year/year/all time range
- **Category Stats**: View A/B/C class statistics separately
- **Detailed Data**: Exam count, pass rate, average score, etc.

### Data Storage

All data is stored in the `.crac_exam` folder under user directory:

```
C:\Users\Username\.crac_exam\
├── data\
│   └── exam_questions.db    # SQLite database file
└── logs\
    └── backend.log          # Application log file
```

**Advantages**:

- Separated from program, data survives reinstallation
- Multi-user sharing program with independent data
- Easy cleanup, delete folder to clear all data

## 🛠️ Development Guide

### Environment Setup

#### 1. Install Go

```bash
# Download and install Go 1.26+
# https://golang.org/dl/

# Verify installation
go version
```

#### 2. Install Node.js

```bash
# Download and install Node.js 18+
# https://nodejs.org/

# Verify installation
node --version
npm --version
```

#### 3. Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 4. Install Project Dependencies

```bash
# Backend dependencies
go mod tidy

# Frontend dependencies
cd frontend
npm install
```

### Development Commands

```bash
# Development mode (with hot reload)
wails dev

# Production build
wails build

# Build Windows 64-bit version
wails build -platform windows/amd64

# Build with debug info
wails build -debug

# Clean build cache
wails build -clean
```

### Database Schema

#### Main Tables

**questions**

```sql
CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    J TEXT,          -- Question ID
    P TEXT,          -- Category
    I TEXT,          -- Unique identifier
    Q TEXT,          -- Question stem
    T TEXT,          -- Correct answer
    A TEXT,          -- Option A
    B TEXT,          -- Option B
    C TEXT,          -- Option C
    D TEXT,          -- Option D
    F TEXT,          -- Explanation
    LA INTEGER,      -- Class A flag
    LB INTEGER,      -- Class B flag
    LC INTEGER,      -- Class C flag
    type INTEGER,    -- Question type (1=single, 2=multiple)
    user_id INTEGER, -- User ID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**exam\_records**

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

**error\_questions**

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

**favorite\_questions**

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

### API Calls

Frontend calls backend Go methods through Wails Runtime:

```typescript
// Import services
import { PracticeService, ExamService } from '../wailsjs/go/services'

// Call Go method
const questions = await PracticeService.GetQuestionsByCategory(1, 'A')

// Handle errors
try {
  const result = await ExamService.CreateExam(userID, category)
} catch (error) {
  console.error('Exam creation failed:', error)
}
```

### Adding New Features

#### 1. Add Backend Service

```go
// backend/services/new_service.go
package services

type NewService struct {
    // Dependencies
}

func (s *NewService) DoSomething(param string) (string, error) {
    // Business logic
    return "result", nil
}
```

#### 2. Register Service

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

#### 3. Frontend Call

```typescript
// frontend/src/api/NewService.ts
import { NewService as GoNewService } from '../wailsjs/go/services'

export const NewService = {
  DoSomething: async (param: string): Promise<string> => {
    return await GoNewService.DoSomething(param)
  },
}
```

## ❓ FAQ

### Q1: Program won't start, missing WebView2

**Solution**:

1. Windows 10 1803+ has WebView2 built-in, please update your system
2. Or manually install WebView2 Runtime:
   - Download: <https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/>
   - Select "Fixed Version" for x64 architecture

### Q2: PDF import fails or questions incomplete

**Solution**:

1. Ensure PDF file is in CRAC official format
2. Check if PDF file is corrupted
3. Check log file: `C:\Users\Username\.crac_exam\logs\backend.log`
4. Try re-downloading PDF question bank

### Q3: Images don't display during exam

**Solution**:

1. Check if images were successfully extracted during import
2. Check if `images` table in database has image data
3. Re-import question bank, ensure image extraction succeeds

### Q4: Data loss or corruption

**Solution**:

1. Check if data file exists: `C:\Users\Username\.crac_exam\data\exam_questions.db`
2. Restore from backup if available
3. Re-import question bank

### Q5: Program runs slowly

**Solution**:

1. Close other resource-intensive programs
2. Clean up unused data in database
3. Check if system has sufficient memory
4. Try restarting the program

### Q6: How to clear all data

**Solution**:

1. Close the program
2. Delete folder: `C:\Users\Username\.crac_exam\`
3. Restart program, it will create new data files automatically

## 📝 Changelog

### v1.0.0 (2026-04-16)

**New Features**

- ✅ Complete A/B/C class question management
- ✅ Question-by-question practice and mock exam
- ✅ Error book and favorites functionality
- ✅ Exam score statistics charts
- ✅ One-click PDF question bank import
- ✅ Automatic question image extraction

**Improvements**

- ✅ Data storage in user directory, separated from program
- ✅ Single file package with all CGO dependencies
- ✅ Application icon and UI optimization
- ✅ Auto prompt and return to home when no data

**Bug Fixes**

- ✅ Fixed multiple-choice answer marking issue
- ✅ Fixed error book data loading error
- ✅ Fixed chart single data point display issue
- ✅ Fixed PDF parsing stem truncation issue

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork this project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 🙏 Acknowledgments

Thanks to the following open source projects:

- [Wails](https://wails.io/) - Modern Go desktop application framework
- [Vue.js](https://vuejs.org/) - Progressive JavaScript framework
- [Element Plus](https://element-plus.org/) - Vue 3 component library
- [ECharts](https://echarts.apache.org/) - Powerful data visualization library
- [go-fitz](https://github.com/gen2brain/go-fitz) - Go PDF processing library
- [unipdf](https://github.com/oliverpool/unipdf) - Go PDF processing library

***

**Wish all candidates success in the exam and obtain amateur radio operator certificate soon!** 🎉
