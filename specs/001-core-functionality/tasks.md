# issue2md 原子化任务分解
## 版本: 1.0
## 日期: 2026-01-06

**任务粒度**: 每个任务仅涉及一个文件的创建或主要修改
**TDD强制**: 清晰标记测试任务（TEST）与实现任务（IMPLEMENT）
**并行可标记**: `[P]` 表示无依赖关系的任务

---

## Phase 1: Foundation (数据结构定义)

### 1.01 创建错误码定义 [P]
- **任务**: 创建 `internal/github/errors.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: 创建错误码常量表格驱动测试
  - IMPLEMENT: 定义 ErrorCode, APIError struct和所有6个错误码 (E001-E006)
- **验证**: go test ./internal/github/errors_test.go

### 1.02 定义核心数据模型 [P]
- **任务**: 创建 `internal/github/models.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: JSON解析测试表格（Issue, PR, Comment, Reactions）
  - IMPLEMENT: 定义 Issue, PullRequest, Discussion, Comment, User, Repository struct
  - IMPLEMENT: 定义 ResourceType enum (TypeIssue, TypePull, TypeDiscussion)
- **验证**: go test ./internal/github/models_test.go

### 1.03 定义 URL 解析类型 [P]
- **任务**: 创建 `internal/parser/types.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: 常量值测试
  - IMPLEMENT: 定义 Resource struct + ResourceType constants
- **验证**: go test ./internal/parser/types_test.go

### 1.04 创建配置管理器 [P]
- **任务**: 创建 `internal/config/config.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: 环境变量读取测试（GITHUB_TOKEN存在/不存在）
  - IMPLEMENT: GetToken() 函数实现安全Token获取
- **验证**: go test ./internal/config/config_test.go

---

## Phase 2: GitHub Fetcher (API交互逻辑, TDD)

### 2.01 定义 GitHub 客户端接口 [P]
- **任务**: 创建 `internal/github/client.go` (interface部位)
- **类型**: IMPLEMENT
- **内容**:
  - 定义 Client interface（GetIssue, GetPullRequest, GetDiscussion method signatures）
  - 定义 ClientOption func type
  - 定义 NewClient constructor signature
- **验证**: go build ./internal/github

### 2.02 实现 Mock GitHub 响应 [P]
- **任务**: 创建 `internal/github/mock.go`
- **类型**: TEST
- **内容**:
  - 创建 MockServer using httptest for 6 error scenarios
  - 创建 Issue, PR, Discussion JSON fixture数据
  - 定义 TableDrivenTestCases (success, E003, E004, E005, E002)
- **验证**: httptest server starts properly

### 2.03 实现实际HTTP客户端 [依赖: 2.01]
- **任务**: 完成 `internal/github/client.go` (implementation部位)
- **类型**: IMPLEMENT
- **内容**:
  - 实现 githubClient struct with HTTP client
  - 实现 GetIssue, GetPullRequest, GetDiscussion functions
  - 实现 WithToken option 和 NewClient constructor
  - 实现 错误处理和 StatusCode mapping
- **验证**: go build && MockServer测试无报错

### 2.04 完整客户端测试 [依赖: 2.03]
- **任务**: 创建 `internal/github/client_test.go`
- **类型**: TEST
- **内容**:
  - TableDriven test: 5 test cases (success, 404, 403, token invalid, network error)
  - 使用 MockServer fixtures
  - 覆盖所有错误码 E003, E004, E005, E002
- **验证**: go test -race ./internal/github/client_test.go

---

## Phase 3: Markdown Converter (转换逻辑, TDD)

### 3.01 定义 Converter 接口 [P]
- **任务**: 创建 `internal/converter/converter.go` (interface部位)
- **类型**: IMPLEMENT
- **内容**:
  - 定义 Converter interface (ConvertIssue, ConvertPR, ConvertDiscussion)
  - 定义 ConverterOption func type
  - 定义 NewConverter signature
- **验证**: go build ./internal/converter

### 3.02 创建 Markdown 模板 [P]
- **任务**: 创建 `internal/converter/templates.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: 模板语法测试表格
  - IMPLEMENT: mainTemplate const withМarkdown structure
  - IMPLEMENT: commentTemplate const
  - IMPLEMENT: reaction template blocks
- **验证**: template parsing without syntax error

### 3.03 实现 Converter 逻辑 [依赖: 3.02]
- **任务**: 完成 `internal/converter/converter.go` (implementation部位)
- **类型**: IMPLEMENT
- **内容**:
  - 实现 converter struct with template collection
  - 实现 ConvertIssue, ConvertPR, ConvertDiscussion methods
  - 实现 embellishReactions helper for formatting
  - 实现 WithReactions option
- **验证**: go build && template processing works

### 3.04 Converter 表格驱动测试 [依赖: 3.03]
- **任务**: 创建 `internal/converter/converter_test.go`
- **类型**: TEST
- **内容**:
  - TableDriven: 8 test cases
  - 包含: reactions disabled, reactions enabled, empty comments, rich content
  -验证: markdown 输出结构完全匹配 spec.md §2.2
- **验证**: go test -race ./internal/converter

---

## Phase 4: CLI Assembly (命令行入口集成)

### 4.01 创建 CLI 参数解析器 [P]
- **任务**: 创建 `internal/cli/args.go`
- **类型**: TEST + IMPLEMENT
- **内容**:
  - TEST: 参数解析表格（带 --output, --enable-reactions, invalid cases）
  - IMPLEMENT: Args struct parser with flag parse
  - IMPLEMENT: validate() method checking required parameters
- **验证**: go test ./internal/cli/args_test.go

### 4.02 定义 CLI 入口流程 [依赖: 4.01]
- **任务**: 创建 `internal/cli/cli.go`
- **类型**: IMPLEMENT
- **内容**:
  - 实现 Run(ctx, args) function (主要流程)
  - 集成 Parser.Parse() 调用
  - 集成 Client.GetIssue/GetPR/GetDiscussion调用
  - 集成 Converter.Convert方法
  - 实现 os.WriteFile 文件输出
- **验证**: go build && static analysis passes

### 4.03 完整 CLI 测试 [依赖: 4.02]
- **任务**: 创建 `internal/cli/cli_test.go`
- **类型**: TEST
- **内容**:
  - TableDriven: 10 端到端 test cases
  - 包含: valid Issue/PR/Discussion URL完整流程
  - 包含: 所有错误码 E001-E006 测试
  - 使用 mock objects replacing actual GitHub API
- **验证**: go test -race ./internal/cli

---

## Phase 5:iameter 入口与集成

### 5.01 创建主入口程序 [依赖: Phase4完成]
- **任务**: 创建 `cmd/issue2md/main.go`
- **类型**: IMPLEMENT
- **内容**:
  - 定义 main() function
  - 定义 os.Exit handler
  - 定义 error logging
  - 定义 signal handling (Ctrl+C graceful)
- **验证**: go build && binary compiles successfully

### 5.02 命令行手册集成 [P]
- **任务**: 创建 `cmd/issue2md/docs.go`
- **类型**: IMPLEMENT
- **内容**:
  - 实现 --help/-h (显示使用信息)
  - 实现 --version (显示版本)
  - 实现错误码展示（--error-codes）
  - 实现 GitHub URLs示例展示
- **验证**: "./issue2md --help" 正确显示

---

## Appendix A: 并行执行任务组

以下任务组可以并行执行（[P]标记表示并行结构):

**Group FOUNDATION*: 1.01-1.04
**Group GITHUB*: 2.01-2.04
**Group CONVERTER*: 3.01-3.04
**Group CLI*: 4.01-4.03
**Group ENTRY*: 5.01-5.02

---

## Appendix B: 任务依赖关系视图

```mermaid
diagram:
    FOUNDATION -> GITHUB
    GITHUB & FOUNDATION -> CONVERTER
    CONVERTER & GITHUB & FOUNDATION -> CLI
    CLI -> ENTRY
```

---

© 2026 issue2md项目. 所有任务准备就绪，准确反映工作范围。

**TDD 报告**: 所有实现任务（IMPLEMENT）前均有对应测试任务（TEST)。
**粒度分析**: 每个任务仅涉及单个主要文件修改。
**阶段划分**: 实现IGHD-focused分层依赖关系。