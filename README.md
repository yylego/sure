<!-- TEMPLATE (EN) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/sure/release.yml?branch=main&label=BUILD)](https://github.com/yylego/sure/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/sure)](https://pkg.go.dev/github.com/yylego/sure)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/sure/main.svg)](https://coveralls.io/github/yylego/sure?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.24%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/sure.svg)](https://github.com/yylego/sure/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/sure)](https://goreportcard.com/report/github.com/yylego/sure)

<!-- TEMPLATE (EN) CLOSE: BADGES -->

# sure: Add Assertions and Crash Handling to Existing Go Code

`sure` enhances existing Go code with assertions and crash handling. It asserts conditions and crashes when issues happen, improving issue management in code without adding repetitive checks.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## CREATION_IDEAS

[CREATION_IDEAS](internal/docs/CREATION_IDEAS.md)

---

## Introduction

`sure` is a code generation package that transforms existing Go code into assertion-enhanced versions. Instead of adding repetitive checks throughout the codebase, `sure` generates wrapping code with built-in assertion logic, making issue detection and handling seamless.

The package provides three distinct generators, each serving different code transformation needs:

- **sure_cls_gen**: Creates assertion-enabled class wrappers
- **sure_pkg_gen**: Generates assertion-enabled package wrappers
- **cls_stub_gen**: Produces singleton wrapping functions

---

## Core Features

🎯 **Three Assertion Modes**

- **Must**: Crash on issues (panic-based)
- **Soft**: Log warnings and continue execution
- **Omit**: Silent mode - ignore issues

⚡ **Automatic Code Generation**

- Transform existing code without modification
- Generate type-safe wrappers with assertions
- Preserve function signatures and documentation

🔧 **Flexible Integration**

- Works with existing Go projects
- No runtime dependencies in generated code
- Compatible with standard Go packages

📦 **Multiple Generation Strategies**

- Class-based: Wrap object methods
- Package-based: Wrap package functions
- Singleton-based: Wrap instance methods as package functions

---

## Benefits of sure

### Problem: Repetitive Issue Handling

Standard issue handling in Go requires repetitive checks:

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

This becomes verbose and repetitive across large codebases.

### Solution: Generated Assertion Wrappers

With `sure`, generate assertion-enabled wrappers:

```go
// Source function
func ReadConfig(path string) (*Config, error)

// Generated wrapping (Must mode)
func (s *SureConfig) ReadConfig(path string) *Config

// Usage - crashes on issues, no manual checks needed
config := sureConfig.ReadConfig("config.json")
```

The generated code handles assertions, enabling clean business logic.

---

## Package Overview

### `sure_cls_gen`: **Assertion-Enabled Class Generation**

**Purpose**: Creates wrapping classes from existing objects, embedding assertion logic in each method.

**Use Case**: When working with objects that have multiple methods returning issues, generate a wrapping class that handles all assertions.

**How It Works**:

1. Takes an existing struct/interface
2. Analyzes all methods with issue returns
3. Generates wrapping class with assertion methods
4. Each wrapping method removes issue returns, adding assertion logic

**Example**:

Source database client:

```go
type DB struct { }

func (db *DB) Connect(dsn string) error
func (db *DB) Query(sql string) (*Result, error)
func (db *DB) Close() error
```

Generated assertion wrapping:

```go
type SureDB struct {
    db *DB
}

func (s *SureDB) Connect(dsn string) // panics on issue
func (s *SureDB) Query(sql string) *Result // panics on issue
func (s *SureDB) Close() // panics on issue
```

**Benefits**:

- No hand-written assertion code needed
- Type-safe wrappers
- Clean business logic code
- Consistent issue handling across methods

**Code Generation**:

See: [sure_cls_gen example](internal/examples/example_sure_cls_gen)

---

### `sure_pkg_gen`: **Assertion-Enabled Package Generation**

**Purpose**: Extracts functions from existing packages and generates new assertion-enabled package versions.

**Use Case**: When a package has multiple functions returning issues, generate a companion package with assertion versions.

**How It Works**:

1. Scans target package functions
2. Identifies functions with issue returns
3. Generates new package with assertion wrappers
4. Maintains function signatures (minus issue returns)

**Example**:

Source package functions:

```go
package config

func Load(path string) (*Config, error)
func Parse(data []byte) (*Config, error)
func Validate(cfg *Config) error
```

Generated assertion package:

```go
package sureconfig

func Load(path string) *Config // panics on issue
func Parse(data []byte) *Config // panics on issue
func Validate(cfg *Config) // panics on issue
```

**Benefits**:

- Entire package gets assertion versions
- No modification to source package
- Use assertion package when appropriate
- Mix and match with source package

**Code Generation**:

See: [sure_pkg_gen example](internal/examples/example_sure_pkg_gen)

---

### `cls_stub_gen`: **Singleton Wrapping Generation**

**Purpose**: Generates package-scope functions that wrap methods of a singleton instance.

**Use Case**: When working with singleton patterns, provide package-scope functions as convenient wrappers.

**How It Works**:

1. Takes a struct with methods
2. Assumes a singleton instance exists
3. Generates package-scope functions
4. Each function delegates to singleton instance

**Example**:

Singleton object with methods:

```go
type Logger struct { }

func (l *Logger) Debug(msg string)
func (l *Logger) Info(msg string)
func (l *Logger) Warn(msg string)
func (l *Logger) Error(msg string)

var defaultLogger = &Logger{}
```

Generated package functions:

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

**Benefits**:

- Simple API - no instance needed
- Package-scope convenience functions
- Consistent with standard Go package design
- Maintains singleton pattern benefits

**Code Generation**:

See: [cls_stub_gen example](internal/examples/example_cls_stub_gen)

---

## Installation

```bash
go get github.com/yylego/sure
```

---

## Quick Start

### 1. Basic Class Wrapping Generation

```go
import (
    "github.com/yylego/sure"
    "github.com/yylego/sure/sure_cls_gen"
)

// Generate assertion-wrapped classes into the output file
cfg := &sure_cls_gen.ClassGenConfig{
    ClassGenOptions: sure_cls_gen.NewClassGenOptions(sourceRoot),
    PackageName:     "mypackage",
    OutputPath:      outputPath,
}

sure_cls_gen.GenerateClasses(cfg, originalObject)
```

### 2. Package Function Wrapping Generation

```go
import (
    "github.com/yylego/sure"
    "github.com/yylego/sure/sure_pkg_gen"
)

// Generate an assertion-enabled companion package next to the source
sure_pkg_gen.GenerateSurePackage(
    sourceRoot,
    sure.MUST,
    "original/package/path",
)
```

### 3. Singleton Wrapping Generation

```go
import "github.com/yylego/sure/cls_stub_gen"

// Generate package-level wrapping
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

## Usage Scenarios

### Scenario 1: Database Operations

Source code with repetitive checks:

```go
conn, err := db.Connect(dsn)
if err != nil { return err }

result, err := conn.Query(sql)
if err != nil { return err }

err = conn.Close()
if err != nil { return err }
```

With `sure_cls_gen`:

```go
sureDB := NewSureDB(db)
sureDB.Connect(dsn)
result := sureDB.Query(sql)
sureDB.Close()
```

### Scenario 2: Configuration Management

Source code:

```go
cfg, err := config.Load("app.json")
if err != nil { return err }

err = config.Validate(cfg)
if err != nil { return err }
```

With `sure_pkg_gen`:

```go
cfg := sureconfig.Load("app.json")
sureconfig.Validate(cfg)
```

### Scenario 3: API Client Usage

Source singleton:

```go
client.SetEndpoint(url)
response, err := client.Get("/api/data")
if err != nil { return err }
```

With `cls_stub_gen`:

```go
SetEndpoint(url)
response := Get("/api/data")
```

---

## Imports and Build Constraints in Generated Files

### Imports move across without registration

The generators transcribe method signatures as text, so an outside type in a signature (such as `plumbing.Hash`) needs its import. The generators take those imports from the scanned file's import block, aliases included.

This matters because some imports cannot be deduced:

- an alias (`timekit "time"`) matches no package name
- `github.com/go-git/go-git/v5` is named `git`, and the path does not state the name

Taking too much is safe: goimports drops what the generated code does not mention. When the finished code holds an import path that no scanned file and no registration has stated, the codegen stops and writes nothing, since such a path can have come from guesswork alone.

The scan leaves out one file: the artifact this run is about to overwrite, so a past artifact's imports do not get handed down. Output from protoc-gen-go, mockgen and such stays in the scan and gets wrapped as before.

Explicit registration through `ImportOptions` (`AddPkgPath` / `AddInferredObject`) remains in effect, and helps when a package appears in the generated code without appearing in the scanned sources, such as a custom exception handling package.

### Build constraint that keeps regeneration possible

A generated class lands in the same package as the test that regenerates it. Once a generated file stops compiling, its package stops compiling, and that test cannot run to replace the file.

Generated files state a build constraint that breaks the deadlock:

```go
//go:build !sure_generate
```

- **Standard build** (`go build ./...`): the generated file takes part as before
- **Regeneration** (`go test -tags=sure_generate`): the generated file drops out, so the codegen runs and overwrites the artifact

Put the matching directive above the test that regenerates:

```go
//go:generate go test -run ^TestGenerate$ -tags=sure_generate
func TestGenerate(t *testing.T) {
    // ...
}
```

When an artifact breaks:

```bash
go generate ./...
```

Use `WithNoBuildDirective()` to opt out of the constraint.

**Note 1**: a tagged test run compiles the whole test package too. When a test file in that package uses the generated API, state the same `//go:build !sure_generate` line on that test file too.

**Note 2**: the tag suits targeted regeneration, not whole-module builds. `sure_pkg_gen` emits standalone packages that the scanned package does not depend on, so it cannot deadlock and states no constraint.

---

## Best Practices

### When to Use Must Mode

Use `sure.MUST` when:

- Issues are unrecoverable
- Application cannot continue with issues
- During initialization and setup
- In test code

### When to Use Soft Mode

Use `sure.SOFT` when:

- Issues should be logged but not crash
- Smooth degradation is acceptable
- In production with fallback logic
- When monitoring issues without interruption

### When to Use Omit Mode

Use `sure.OMIT` when:

- Issues are expected and acceptable
- Silent failure is desired
- Performance-sensitive paths
- When issues are handled elsewhere

---

## Relationship with Related Packages

`sure` is designed to complement the [done](https://github.com/yylego/done) package:

- **done**: Provides inline assertion functions (`done.Done()`, `done.VAE()`)
- **sure**: Generates wrapping code with built-in assertions

Use `done` when writing new code, use `sure` to wrap existing code.

**Example with done**:

```go
config := done.VAE(LoadConfig()).Nice()
```

**Example with sure**:

```go
config := sureConfig.Load() // generated wrapping
```

Both approaches reduce boilerplate, choose based on context.

---

## Examples

Comprehensive examples demonstrating each generation approach:

- [**sure_cls_gen Example**](internal/examples/example_sure_cls_gen) - Class wrapping generation
- [**sure_pkg_gen Example**](internal/examples/example_sure_pkg_gen) - Package wrapping generation
- [**cls_stub_gen Example**](internal/examples/example_cls_stub_gen) - Singleton wrapping generation

Each example includes:

- Source code setup
- Generation configuration
- Generated code output
- Usage demonstrations

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->

---

<!-- TEMPLATE (EN) BEGIN: GITHUB STARS -->

## GitHub Stars

[![Stargazers](https://starchart.cc/yylego/sure.svg?variant=adaptive)](https://starchart.cc/yylego/sure)

<!-- TEMPLATE (EN) CLOSE: GITHUB STARS -->
