package sure_pkg_gen

import "github.com/yylego/sure"

// SurePackageGenConfig contains package generation configuration settings
// Specifies source paths, exception handling mode, and output settings
// Provides fluent API through With* methods that make configuration simple
//
// SurePackageGenConfig 包含包生成的配置设置
// 指定源路径、异常处理模式和输出设置
// 通过 With* 方法提供流式 API 使配置简单
type SurePackageGenConfig struct {
	SourceRoot           string                 // Source code root path // 源代码根路径
	ErrorHandlingMode    sure.ErrorHandlingMode // Error handling mode (Must/Soft/Omit) // 错误处理模式（Must/Soft/Omit）
	SourcePackagePath    string                 // Import path of source package // 源包的导入路径
	ErrorHandlingPkgPath string                 // Error handling package path // 错误处理包路径
	HandlerFuncReference string                 // Exception handling function reference // 异常处理函数引用
	NewPkgName           string                 // New package name (auto-generated when not set) // 新包名（未设置时自动生成）
	OutputRoot           string                 // Output root path // 输出根路径

	// go:build directive of the generated files, blank writes none
	//
	// Blank is the default here, unlike sure_cls_gen and cls_stub_gen, and the shape of the output explains it
	// Those two write into the package holding the wrapped code, alongside the test that regenerates, so a broken artifact keeps that test from running and the constraint is what breaks the lock
	// This codegen writes a standalone sibling package while the test that drives it sits in the source package, so a broken artifact leaves the codegen reachable and there is nothing to unlock
	// State one even so, via WithGoBuildDirective, where the output lands somewhere that can lock. See sure.DirectiveSkipGenerate
	//
	// 生成文件的 go:build 指令，留空则不写
	//
	// 这里默认留空，与 sure_cls_gen、cls_stub_gen 不同，原因在产物的位置
	// 那两个把产物写进"被包装代码所在的包"，与负责重新生成的测试同住，于是产物一坏那个测试就跑不起来，得靠约束解锁
	// 本生成器写的是独立的兄弟包，而驱动它的测试留在源包里，产物坏了也不妨碍生成器再跑，没有锁可解
	// 若你的输出落在会卡住的地方，仍可用 WithGoBuildDirective 主动写上。参见 sure.DirectiveSkipGenerate
	GoBuildDirective string

	// genPosFuncMark records the calling position captured at a public entrance, so
	// the "Generated from:" trace stays correct regardless of which entrance is used.
	// genPosFuncMark 记录公共入口处捕获的调用位置，让"Generated from:"追踪行
	// 无论从哪个入口进入都指向正确的用户调用点。
	genPosFuncMark string
}

// NewSurePackageConfig creates a new package generation configuration
// Initializes with source root, error handling mode and package path
// Returns a new config instance with default settings
//
// NewSurePackageConfig 创建新的包生成配置
// 使用源根路径、错误处理模式和包路径初始化
// 返回带有默认设置的新配置实例
func NewSurePackageConfig(sourceRoot string, errorHandlingMode sure.ErrorHandlingMode, sourcePackagePath string) *SurePackageGenConfig {
	return &SurePackageGenConfig{
		SourceRoot:           sourceRoot,
		ErrorHandlingMode:    errorHandlingMode,
		SourcePackagePath:    sourcePackagePath,
		ErrorHandlingPkgPath: sure.GetPkgPath(), // Default to "github.com/yylego/sure" package // 默认使用 "github.com/yylego/sure" 包
		HandlerFuncReference: sure.GetPkgName(), // Default to "sure" prefix (e.g., sure.Must and sure.Soft) // 默认使用 "sure" 前缀（例如 sure.Must 和 sure.Soft）
		NewPkgName:           "",                // Auto-generate from source package name when not set // 未设置时从源包名自动生成
		OutputRoot:           sourceRoot,
	}
}

// WithSourceRoot sets the source code root path
// Overrides the default source location
// Returns self to allow method chaining
//
// WithSourceRoot 设置源代码根路径
// 覆盖默认的源代码位置
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithSourceRoot(sourceRoot string) *SurePackageGenConfig {
	c.SourceRoot = sourceRoot
	return c
}

// WithErrorHandlingMode sets the exception handling mode to use
// Configures to use Must, Soft, and Omit mode
// Returns self to enable method chaining
//
// WithErrorHandlingMode 设置要使用的异常处理模式
// 配置使用 Must、Soft 以及 Omit 模式
// 返回自身以支持方法链式调用
func (c *SurePackageGenConfig) WithErrorHandlingMode(errorHandlingMode sure.ErrorHandlingMode) *SurePackageGenConfig {
	c.ErrorHandlingMode = errorHandlingMode
	return c
}

// WithSourcePackagePath sets the import path of the source package
// Specifies where to import the source code package from
// Returns self to allow method chaining
//
// WithSourcePackagePath 设置源包的导入路径
// 指定从哪里导入源代码包
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithSourcePackagePath(sourcePackagePath string) *SurePackageGenConfig {
	c.SourcePackagePath = sourcePackagePath
	return c
}

// WithErrorHandlingPkgPath sets the error handling package path
// Allows using custom error handling package instead of default sure package
// Returns self to allow method chaining
//
// WithErrorHandlingPkgPath 设置错误处理包路径
// 允许使用自定义错误处理包而不是默认的 sure 包
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithErrorHandlingPkgPath(errorHandlingPkgPath string) *SurePackageGenConfig {
	c.ErrorHandlingPkgPath = errorHandlingPkgPath
	return c
}

// WithHandlerFuncReference sets the error handling function reference
// Specifies the prefix used when calling error handling functions
// Returns self to allow method chaining
//
// WithHandlerFuncReference 设置错误处理函数引用
// 指定调用错误处理函数时使用的前缀
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithHandlerFuncReference(handlerFuncReference string) *SurePackageGenConfig {
	c.HandlerFuncReference = handlerFuncReference
	return c
}

// WithNewPkgName sets the name of the generated package
// Overrides auto-generated package name with custom name
// Returns self to allow method chaining
//
// WithNewPkgName 设置生成包的名称
// 使用自定义名称覆盖自动生成的包名
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithNewPkgName(newPkgName string) *SurePackageGenConfig {
	c.NewPkgName = newPkgName
	return c
}

// WithOutputRoot sets the output root path to store generated code
// Specifies where to save the generated package files
// Returns self to enable method chaining
//
// WithOutputRoot 设置输出根路径存储生成的代码
// 指定保存生成包文件的位置
// 返回自身以支持方法链式调用
func (c *SurePackageGenConfig) WithOutputRoot(outputRoot string) *SurePackageGenConfig {
	c.OutputRoot = outputRoot
	return c
}

// WithGoBuildDirective sets the go:build directive written into generated files
// State a negated tag such as sure.DirectiveSkipGenerate, since a directive without the leading "!" would drop the artifact from each standard build instead
// Blank is the default here, so this is opt-in — the field explains how this codegen stands apart from the two beside it
// Returns self to allow method chaining
//
// WithGoBuildDirective 设置写入生成文件的 go:build 指令
// 要写取反的标签（如 sure.DirectiveSkipGenerate），因为漏掉开头的 "!" 会反过来让产物在每次普通构建里都被排除掉
// 这里默认留空，所以这是个主动开关——本生成器为何与另外两个不同，见该字段的说明
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithGoBuildDirective(goBuildDirective string) *SurePackageGenConfig {
	c.GoBuildDirective = goBuildDirective
	return c
}

// WithNoBuildDirective clears the directive so generated files state no constraint
// Matches the default, and serves a chain that set one up above
// Returns self to allow method chaining
//
// WithNoBuildDirective 清除指令，使生成文件不带任何构建约束
// 与默认一致，供"链条上游已经设过"的情形使用
// 返回自身以允许方法链式调用
func (c *SurePackageGenConfig) WithNoBuildDirective() *SurePackageGenConfig {
	c.GoBuildDirective = ""
	return c
}
