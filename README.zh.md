<!-- TEMPLATE (ZH) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/sure/release.yml?branch=main&label=BUILD)](https://github.com/yylego/sure/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/sure)](https://pkg.go.dev/github.com/yylego/sure)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/sure/main.svg)](https://coveralls.io/github/yylego/sure?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.24%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/sure.svg)](https://github.com/yylego/sure/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/sure)](https://goreportcard.com/report/github.com/yylego/sure)

<!-- TEMPLATE (ZH) CLOSE: BADGES -->

# sure: 给现有 Go 代码添加断言和崩溃处理

`sure` 通过为现有的 Go 代码添加断言和崩溃处理功能来增强代码。它断言条件并在发生问题时崩溃，帮助在不手动添加重复检查的情况下改善代码的问题管理。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 创作背景

[CREATION_IDEAS](internal/docs/CREATION_IDEAS.zh.md)

---

## 项目简介

`sure` 是一个代码生成包，将现有的 Go 代码转换成带断言增强的版本。无需在整个代码库中添加重复的检查，`sure` 生成带有内置断言逻辑的包装代码，使问题检测和处理变得无缝。

该包提供三种不同的生成器，每种服务于不同的代码转换需求：

- **sure_cls_gen**: 创建启用断言的类包装
- **sure_pkg_gen**: 生成启用断言的包包装
- **cls_stub_gen**: 生成单例包装函数

---

## 核心特性

🎯 **三种断言模式**

- **Must**: 遇到问题时崩溃（基于 panic）
- **Soft**: 记录警告并继续执行
- **Omit**: 静默模式 - 忽略问题

⚡ **自动代码生成**

- 无需修改即可转换现有代码
- 生成类型安全的断言包装
- 保留函数签名和文档

🔧 **灵活集成**

- 适用于现有 Go 项目
- 生成的代码无运行时依赖
- 兼容标准 Go 包

📦 **多种生成策略**

- 基于类：包装对象方法
- 基于包：包装包函数
- 基于单例：将实例方法包装为包级函数

---

## sure 的优势

### 问题：重复的问题处理

标准的 Go 问题处理需要重复检查：

```go
result, err := SomeOperation()
if err != nil {
    log.Printf("operation failed: %v", err)
    return err
}

data, err := AnotherOperation(result)
if err != nil {
    log.Printf("another operation failed: %v", err)
    return err
}
```

在大型代码库中，这变得冗长和重复。

### 解决方案：生成断言包装

使用 `sure`，生成启用断言的包装：

```go
// 源函数
func ReadConfig(path string) (*Config, error)

// 生成的包装（Must 模式）
func (s *SureConfig) ReadConfig(path string) *Config

// 使用 - 遇到问题时崩溃，无需手动检查
config := sureConfig.ReadConfig("config.json")
```

生成的代码处理断言，使业务逻辑更简洁。

---

## 包概述

### `sure_cls_gen`: **启用断言的类生成**

**目的**：从现有对象创建包装类，在每个方法中嵌入断言逻辑。

**使用场景**：当处理具有多个返回问题的方法的对象时，生成处理所有断言的包装类。

**工作原理**：

1. 获取现有的结构体/接口
2. 分析所有返回问题的方法
3. 生成带有断言方法的包装类
4. 每个包装方法移除问题返回，添加断言逻辑

**示例**：

源数据库客户端：

```go
type DB struct { }

func (db *DB) Connect(dsn string) error
func (db *DB) Query(sql string) (*Result, error)
func (db *DB) Close() error
```

生成的断言包装：

```go
type SureDB struct {
    db *DB
}

func (s *SureDB) Connect(dsn string) // 遇到问题时 panic
func (s *SureDB) Query(sql string) *Result // 遇到问题时 panic
func (s *SureDB) Close() // 遇到问题时 panic
```

**优势**：

- 无需手动编写断言代码
- 类型安全的包装
- 更简洁的业务逻辑代码
- 跨方法的一致问题处理

**代码生成**：

参见：[sure_cls_gen 示例](internal/examples/example_sure_cls_gen)

---

### `sure_pkg_gen`: **启用断言的包生成**

**目的**：从现有包中提取函数，生成新的启用断言的包版本。

**使用场景**：当一个包有多个返回问题的函数时，生成带有断言版本的配套包。

**工作原理**：

1. 扫描目标包函数
2. 识别返回问题的函数
3. 生成带有断言包装的新包
4. 维护函数签名（减去问题返回）

**示例**：

源包函数：

```go
package config

func Load(path string) (*Config, error)
func Parse(data []byte) (*Config, error)
func Validate(cfg *Config) error
```

生成的断言包：

```go
package sureconfig

func Load(path string) *Config // 遇到问题时 panic
func Parse(data []byte) *Config // 遇到问题时 panic
func Validate(cfg *Config) // 遇到问题时 panic
```

**优势**：

- 整个包获得断言版本
- 无需修改原始包
- 在适当时使用断言包
- 与原始包混合使用

**代码生成**：

参见：[sure_pkg_gen 示例](internal/examples/example_sure_pkg_gen)

---

### `cls_stub_gen`: **单例包装生成**

**目的**：生成包装单例实例方法的包作用域函数。

**使用场景**：在使用单例模式时，提供包作用域函数作为便捷包装。

**工作原理**：

1. 获取带有方法的结构体
2. 假设存在单例实例
3. 生成包作用域函数
4. 每个函数委托给单例实例

**示例**：

带有方法的单例对象：

```go
type Logger struct { }

func (l *Logger) Debug(msg string)
func (l *Logger) Info(msg string)
func (l *Logger) Warn(msg string)
func (l *Logger) Error(msg string)

var defaultLogger = &Logger{}
```

生成的包级函数：

```go
func Debug(msg string) {
    defaultLogger.Debug(msg)
}

func Info(msg string) {
    defaultLogger.Info(msg)
}

func Warn(msg string) {
    defaultLogger.Warn(msg)
}

func Error(msg string) {
    defaultLogger.Error(msg)
}
```

**优势**：

- 简单的 API - 无需实例
- 包作用域便捷函数
- 与标准 Go 包设计一致
- 保持单例模式的优势

**代码生成**：

参见：[cls_stub_gen 示例](internal/examples/example_cls_stub_gen)

---

## 安装

```bash
go get github.com/yylego/sure
```

---

## 快速开始

### 1. 基础类包装生成

```go
import (
    "github.com/yylego/sure"
    "github.com/yylego/sure/sure_cls_gen"
)

// 将断言包装类生成到输出文件
cfg := &sure_cls_gen.ClassGenConfig{
    ClassGenOptions: sure_cls_gen.NewClassGenOptions(sourceRoot),
    PackageName:     "mypackage",
    OutputPath:      outputPath,
}

sure_cls_gen.GenerateClasses(cfg, originalObject)
```

### 2. 包函数包装生成

```go
import (
    "github.com/yylego/sure"
    "github.com/yylego/sure/sure_pkg_gen"
)

// 在源包旁生成一个带断言的伴生包
sure_pkg_gen.GenerateSurePackage(
    sourceRoot,
    sure.MUST,
    "original/package/path",
)
```

### 3. 单例包装生成

```go
import "github.com/yylego/sure/cls_stub_gen"

// 生成包级包装
stubConfig := &cls_stub_gen.StubGenConfig{
    SourceRootPath:    projectPath,
    TargetPackageName: "api",
    OutputPath:        outputPath,
}

code := cls_stub_gen.GenerateStubMethods(
    stubConfig,
    cls_stub_gen.NewStubParam(singletonInstance, "instance"),
)
```

---

## 使用场景

### 场景 1：数据库操作

带有重复检查的源代码：

```go
conn, err := db.Connect(dsn)
if err != nil { return err }

result, err := conn.Query(sql)
if err != nil { return err }

err = conn.Close()
if err != nil { return err }
```

使用 `sure_cls_gen`：

```go
sureDB := NewSureDB(db)
sureDB.Connect(dsn)
result := sureDB.Query(sql)
sureDB.Close()
```

### 场景 2：配置管理

源代码：

```go
cfg, err := config.Load("app.json")
if err != nil { return err }

err = config.Validate(cfg)
if err != nil { return err }
```

使用 `sure_pkg_gen`：

```go
cfg := sureconfig.Load("app.json")
sureconfig.Validate(cfg)
```

### 场景 3：API 客户端使用

源单例：

```go
client.SetEndpoint(url)
response, err := client.Get("/api/data")
if err != nil { return err }
```

使用 `cls_stub_gen`：

```go
SetEndpoint(url)
response := Get("/api/data")
```

---

## 生成文件的导入与构建约束

### 导入自动搬运，无需手工登记

生成器把方法签名当文本从源文件抄出来，所以签名里的外部类型（如 `plumbing.Hash`）需要对应的导入。这些导入**由生成器自动从源文件的导入块搬过来**，包括别名。

这一点很要紧，因为有些导入是推不出来的：

- 别名（`timekit "time"`）不对应任何包名
- `github.com/go-git/go-git/v5` 的包名是 `git`，路径与包名互相推不出来

搬多了无妨——生成代码没用到的导入会被 goimports 剪掉。若成品里出现"既非来自源文件、也非调用方登记"的导入路径，生成器会报错且不写盘，因为那只可能是猜出来的。

扫描只留出一个文件不看：本轮即将被覆写的那个产物，于是上一轮产物的导入不会被一代代传下去。protoc-gen-go、mockgen 之类工具的产物照常参与扫描、照常被包装。

`ImportOptions` 的手工登记（`AddPkgPath` / `AddInferredObject`）依然有效，用于登记源文件之外的包，比如自定义的错误处理函数所在的包。

### 构建约束（解决重新生成问题）

生成的包装类与生成它的测试同住一个包。生成文件一旦编译不过，它所在的包就编译不过，包里那个生成器测试也跑不起来、没法把它换掉。

生成文件默认带上构建约束来解开这个锁：

```go
//go:build !sure_generate
```

- **普通编译**（`go build ./...`）：生成文件正常参与编译
- **重新生成**（`go test -tags=sure_generate`）：生成文件被排除，生成器照样跑得起来并覆盖产物

把对应指令写在生成器测试上方：

```go
//go:generate go test -run ^TestGenerate$ -tags=sure_generate
func TestGenerate(t *testing.T) {
    // ...
}
```

产物坏掉时：

```bash
go generate ./...
```

不需要这道约束时，用 `WithNoBuildDirective()` 清掉。

**注意一**：带标签跑测试时整个测试包仍要编译。若同包里有测试文件用到了生成出来的 API，给那些文件也加上同一行 `//go:build !sure_generate`。

**注意二**：这个标签是给"定点重新生成"用的，不是给全模块构建用的。`sure_pkg_gen` 生成的是独立新包（源包不依赖它，因此不会自锁），所以它不写这道约束。

---

## 最佳实践

### 何时使用 Must 模式

在以下情况使用 `sure.MUST`：

- 问题无法恢复
- 应用程序无法继续处理问题
- 初始化和设置期间
- 在测试代码中

### 何时使用 Soft 模式

在以下情况使用 `sure.SOFT`：

- 问题应该记录但不崩溃
- 可接受优雅降级
- 在生产环境中有回退逻辑
- 在不中断的情况下监控问题

### 何时使用 Omit 模式

在以下情况使用 `sure.OMIT`：

- 问题是预期且可接受的
- 希望静默失败
- 性能关键路径
- 当问题在其他地方处理时

---

## 与相关包的关系

`sure` 旨在补充 [done](https://github.com/yylego/done) 包：

- **done**: 提供内联断言函数（`done.Done()`, `done.VAE()`）
- **sure**: 生成带有内置断言的包装代码

编写新代码时使用 `done`，包装现有代码时使用 `sure`。

**done 示例**：

```go
config := done.VAE(LoadConfig()).Nice()
```

**sure 示例**：

```go
config := sureConfig.Load() // 生成的包装
```

两种方法都减少样板代码，根据上下文选择。

---

## 示例

演示每种生成方式的综合示例：

- [**sure_cls_gen 示例**](internal/examples/example_sure_cls_gen) - 类包装生成
- [**sure_pkg_gen 示例**](internal/examples/example_sure_pkg_gen) - 包包装生成
- [**cls_stub_gen 示例**](internal/examples/example_cls_stub_gen) - 单例包装生成

每个示例包括：

- 源代码设置
- 生成配置
- 生成的代码输出
- 使用演示

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->

---

<!-- TEMPLATE (ZH) BEGIN: GITHUB STARS -->

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/yylego/sure.svg?variant=adaptive)](https://starchart.cc/yylego/sure)

<!-- TEMPLATE (ZH) CLOSE: GITHUB STARS -->
