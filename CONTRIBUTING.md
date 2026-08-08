# Contributing to sure / 为 sure 贡献

Thanks for your interest in improving `sure`! This guide covers how to build, test, and structure a change.

感谢你有意改进 `sure`！本指南说明如何构建、测试与组织改动。

---

## Before you start / 动手之前

- `sure` targets the Go version declared in `go.mod` (currently Go 1.24). Use a matching or newer toolchain. / `sure` 面向 `go.mod` 声明的 Go 版本（当前 1.24），请使用相符或更新的工具链。
- The generators are **library-driven and invoked from tests** — there is no CLI binary. Each `TestGenerate` in `internal/examples/**` regenerates and compiles its own output, so those tests double as end-to-end checks. / 生成器是**库驱动、从测试触发**的，没有 CLI 可执行文件。`internal/examples/**` 下的每个 `TestGenerate` 会重新生成并编译自己的产物，因此它们兼作端到端检查。

## Build & test / 构建与测试

```bash
go build ./...
go test ./...        # quick / 快速
make test            # full: -race -cover -bench, mirrors CI / 全量：与 CI 一致
golangci-lint run    # lint / 静态检查
```

When you change a generator, **regenerate the examples and commit the updated output** (`make generate`, or `go test ./internal/examples/...`, rewrites the `gen.go` / `*_must.go` files). Review the `git diff` of generated files as part of your change — it is the clearest evidence the generator still produces compilable code.

当你改动生成器时，**请重新生成示例并提交更新后的产物**（`make generate` 或 `go test ./internal/examples/...` 会重写 `gen.go` / `*_must.go` 等文件）。把生成文件的 `git diff` 作为改动的一部分来审查——这是"生成器仍产出可编译代码"最直接的证据。

## The three generators / 三个生成器

- `sure_cls_gen` — wraps a struct's methods into per-mode wrapper types (generated into the same package). / 把结构体方法包装成按模式区分的包装类型（生成到同一个包）。
- `sure_pkg_gen` — wraps a package's exported functions into a sibling package (imports the source package, so types are package-qualified). / 把包的导出函数包装到兄弟包（导入源包，故类型需加包前缀）。
- `cls_stub_gen` — wraps singleton methods into package-scope delegating functions. / 把单例方法包装成包作用域的委托函数。

Shared code-generation helpers (import carrying, receiver-method collection) live in `internal/utils` — prefer fixing there so all three generators benefit.

共享的代码生成助手（导入搬运、接收者方法收集）在 `internal/utils`——优先在那里修复，让三个生成器共同受益。

## Code style / 代码风格

- Do not ignore errors (no `_, _ =` discards); assert with `must` / `rese` / `done` where an error is truly unexpected. / 不忽略错误（别用 `_,_=` 丢弃）；确实不该出错处用 `must` / `rese` / `done` 断言。
- Keep public API and generated output backward-compatible where reasonable, and add a regression test for any bug you fix (see `sure_cls_gen/import_scan_test.go` for the `t.TempDir()` + `require` style). / 在合理范围内保持公共 API 与生成产物的向后兼容；修 bug 请补回归测试（风格参考 `sure_cls_gen/import_scan_test.go`）。
- `gofmt` / `goimports` clean, and `golangci-lint run` clean. / 保持 `gofmt`/`goimports` 整洁，`golangci-lint run` 无告警。

## Submitting / 提交

Open a pull request against `main` with a description of what changed and why. CI runs lint, multi-version tests, `govulncheck`, and CodeQL.

向 `main` 发起 PR，并说明改了什么、为什么这样改。CI 会跑 lint、多版本测试、`govulncheck` 与 CodeQL。
