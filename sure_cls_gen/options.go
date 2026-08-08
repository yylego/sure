package sure_cls_gen

import (
	"reflect"
	"strings"

	"github.com/yylego/sure"
	"github.com/yylego/tern"
	"github.com/yylego/tern/zerotern"
)

// NamingPattern defines the naming convention used when generating class names
// Specifies how error handling mode names are combined with class names
// Supports prefix/suffix patterns with different case styles
//
// NamingPattern 定义生成类名时使用的命名约定
// 指定错误处理模式名称如何与类名组合
// 支持带有不同大小写风格的前缀/后缀模式
type NamingPattern string

//goland:noinspection GoSnakeCaseUsage
const (
	STYLE_PREFIX_LOWER_TYPE     NamingPattern = "STYLE_PREFIX_LOWER_TYPE"     // Prefix with lowercase mode // 前缀小写模式
	STYLE_SUFFIX_LOWER_TYPE     NamingPattern = "STYLE_SUFFIX_LOWER_TYPE"     // Suffix with lowercase mode // 后缀小写模式
	STYLE_PREFIX_UPPER_TYPE     NamingPattern = "STYLE_PREFIX_UPPER_TYPE"     // Prefix with uppercase mode // 前缀大写模式
	STYLE_SUFFIX_UPPER_TYPE     NamingPattern = "STYLE_SUFFIX_UPPER_TYPE"     // Suffix with uppercase mode // 后缀大写模式
	STYLE_PREFIX_CAMELCASE_TYPE NamingPattern = "STYLE_PREFIX_CAMELCASE_TYPE" // Prefix with CamelCase mode // 前缀驼峰模式
	STYLE_SUFFIX_CAMELCASE_TYPE NamingPattern = "STYLE_SUFFIX_CAMELCASE_TYPE" // Suffix with CamelCase mode (default) // 后缀驼峰模式（默认）
)

// ClassGenOptions contains class generation configuration settings
// Controls naming patterns, source paths, and exception handling modes
// Provides fluent API through With* methods that make configuration simple
//
// ClassGenOptions 包含类生成的配置设置
// 控制命名模式、源路径和异常处理模式
// 通过 With* 方法提供流式 API 使配置简单
type ClassGenOptions struct {
	SourceRootPath       string                   // Source code path (required) // 源代码路径（必需）
	NewClassName         string                   // Direct class name when generating single class // 生成单个类时的直接类名
	NewClassNameParts    string                   // Name parts to insert into generated class names // 插入生成类名中的名称片段
	NamingPatternType    NamingPattern            // Naming convention pattern // 命名约定模式
	ReceiverVariableName string                   // Method context variable name (e.g., 'a' in func (a *A)) // 方法中的接收器变量名（例如，func (a *A) 中的 'a'）
	ErrorHandlerFuncName string                   // Custom error handling function name // 自定义错误处理函数名
	ErrorHandlingModes   []sure.ErrorHandlingMode // Error handling modes to generate // 要生成的错误处理模式
	GoBuildDirective     string                   // go:build directive for the generated file, blank falls back to sure.DirectiveSkipGenerate // 生成文件的 go:build 指令，留空则回退到 sure.DirectiveSkipGenerate

	// States if the config settled the build constraint question, which a blank GoBuildDirective cannot show on its own
	// Set via WithGoBuildDirective and WithNoBuildDirective, so a bare struct value leaves it false and picks up the default
	// See utils.PickGoBuildDirective
	//
	// 配置是否就"构建约束"这件事表过态——光看 GoBuildDirective 是空串分不出来
	// 由 WithGoBuildDirective 与 WithNoBuildDirective 置上，于是直接写结构体字面量时它为 false、从而吃到默认值
	// 详见 utils.PickGoBuildDirective
	directiveChosen bool

	// Artifact this run is about to write, so the source scan can leave it alone
	// GenerateClasses sets it from the enclosing config, since the output path lives there while the scan happens down here
	// Blank where calling code reaches GenerateClassCodeWithErrorHandling on its own, which keeps that exported entrance acting as it once did
	// See utils.IsOutputFile
	//
	// 本次将要写入的产物文件，好让源码扫描绕开它
	// 由 GenerateClasses 从外层配置填进来——输出路径在外层，而扫描发生在这一层
	// 调用方自行调用 GenerateClassCodeWithErrorHandling 时它为空，于是那个导出入口的行为一如既往
	// 详见 utils.IsOutputFile
	outputPath string
}

// NewClassGenOptions creates a new ClassGenOptions with the source root path
// Initializes configuration with the provided source code location
// Returns a new options instance that accepts more configuration
//
// NewClassGenOptions 使用源根路径创建新的 ClassGenOptions
// 使用提供的源代码位置初始化配置
// 返回接受进一步配置的新选项实例
func NewClassGenOptions(srcRoot string) *ClassGenOptions {
	return &ClassGenOptions{
		SourceRootPath: srcRoot,
		// The build constraint comes set as the default, and a bare struct value reaches the same one through the fallback
		// The generated class and the test that makes it share one package, and that shape locks once the constraint goes missing
		// Drop it with WithNoBuildDirective. See sure.DirectiveSkipGenerate
		// 默认带上构建约束，直接写结构体字面量的配置也会经由兜底拿到同一条
		// 生成的类与生成它的测试同住一个包，缺了这道约束就会卡住
		// 不想要的用 WithNoBuildDirective 清掉。详见 sure.DirectiveSkipGenerate
		GoBuildDirective: sure.DirectiveSkipGenerate,
		directiveChosen:  true,
	}
}

// WithGoBuildDirective sets the go:build directive written into generated files
// State a negated tag such as sure.DirectiveSkipGenerate, since a directive without the leading "!" would drop the artifact from each standard build instead
// Pass "!sure_generate" (sure.DirectiveSkipGenerate) to keep regeneration unlocked
// Returns self to allow method chaining
//
// WithGoBuildDirective 设置写入生成文件的 go:build 指令
// 要写取反的标签（如 sure.DirectiveSkipGenerate），因为漏掉开头的 "!" 会反过来让产物在每次普通构建里都被排除掉
// 传入 "!sure_generate"（sure.DirectiveSkipGenerate）可保持重新生成不被卡住
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithGoBuildDirective(goBuildDirective string) *ClassGenOptions {
	cfg.GoBuildDirective = goBuildDirective
	cfg.directiveChosen = true
	return cfg
}

// WithNoBuildDirective clears the directive so generated files state no constraint
// This is the one route to a blank constraint, since a config that said nothing on the topic falls back to sure.DirectiveSkipGenerate
// Choosing it gives up the escape hatch: once a generated file stops compiling, its package stops compiling, and the test inside cannot run to set it straight
// See sure.DirectiveSkipGenerate
// Returns self to allow method chaining
//
// WithNoBuildDirective 清除指令，使生成文件不带任何构建约束
// 这是让约束为空的唯一途径，因为压根没表过态的配置会回退到 sure.DirectiveSkipGenerate
// 选它就等于放弃那道退路：生成文件一旦编译不过，它所在的包就编译不过，包里那个生成器测试也没法再跑起来修它
// 参见 sure.DirectiveSkipGenerate
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithNoBuildDirective() *ClassGenOptions {
	cfg.GoBuildDirective = ""
	cfg.directiveChosen = true
	return cfg
}

// WithNewClassName sets the direct class name when generating a single class
// Allows specifying the exact class name instead of using patterns
// Returns self to allow method chaining
//
// WithNewClassName 在生成单个类时设置直接类名
// 提供一种指定确切类名而不使用模式的方法
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithNewClassName(newClassName string) *ClassGenOptions {
	cfg.NewClassName = newClassName
	return cfg
}

// WithNewClassNameParts sets name parts to insert into generated class names
// Allows customization of the middle section in generated names
// Returns self to allow method chaining
//
// WithNewClassNameParts 设置插入生成类名的名称片段
// 允许自定义生成名称中的中间部分
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithNewClassNameParts(newClassNameParts string) *ClassGenOptions {
	cfg.NewClassNameParts = newClassNameParts
	return cfg
}

// WithNamingPatternType sets the naming convention pattern
// Defines how mode names combine with class names
// Returns self to allow method chaining
//
// WithNamingPatternType 设置命名约定模式
// 定义模式名称如何与类名组合
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithNamingPatternType(namingPatternType NamingPattern) *ClassGenOptions {
	cfg.NamingPatternType = namingPatternType
	return cfg
}

// WithReceiverVariableName sets the method context variable name
// Specifies the name used in method context (e.g., 'a' in func (a *A))
// Returns self to enable method chaining
//
// WithReceiverVariableName 设置方法中使用的接收器变量名
// 指定方法接收器中使用的名称（例如 func (a *A) 中的 'a'）
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithReceiverVariableName(receiverVariableName string) *ClassGenOptions {
	cfg.ReceiverVariableName = receiverVariableName
	return cfg
}

// WithErrorHandlerFuncName sets a custom error handling function name
// Allows using custom error handling instead of default sure package functions
// Returns self to allow method chaining
//
// WithErrorHandlerFuncName 设置自定义错误处理函数名
// 允许使用自定义错误处理而不是默认的 sure 包函数
// 返回自身以允许方法链式调用
func (cfg *ClassGenOptions) WithErrorHandlerFuncName(errorHandlerFuncName string) *ClassGenOptions {
	cfg.ErrorHandlerFuncName = errorHandlerFuncName
	return cfg
}

// MoreErrorHandlingModes adds extra exception handling modes to generate
// Appends modes to the existing mode set
// Returns self to enable method chaining
//
// MoreErrorHandlingModes 添加额外的异常处理模式以生成
// 将模式追加到现有模式集
// 返回自身以支持方法链式调用
func (cfg *ClassGenOptions) MoreErrorHandlingModes(errorHandlingModes ...sure.ErrorHandlingMode) *ClassGenOptions {
	cfg.ErrorHandlingModes = append(cfg.ErrorHandlingModes, errorHandlingModes...)
	return cfg
}

// GetErrorHandlingModes returns the configured exception handling modes
// Returns default modes (Must and Soft) when none specified
// Used to determine which wrapping classes to generate
//
// GetErrorHandlingModes 返回配置的异常处理模式
// 当未指定时返回默认模式（Must 和 Soft）
// 用于确定要生成哪些包装类
func (cfg *ClassGenOptions) GetErrorHandlingModes() []sure.ErrorHandlingMode {
	return tern.BVF(len(cfg.ErrorHandlingModes) != 0, cfg.ErrorHandlingModes, func() []sure.ErrorHandlingMode {
		// When not set, return default Must and Soft modes
		// 当未设置时，返回默认的 Must 和 Soft 模式
		return []sure.ErrorHandlingMode{sure.MUST, sure.SOFT}
	})
}

// GenerateNewClassName creates a new class name based on pattern and mode
// Combines object type name with error handling mode using naming pattern
// Returns formatted class name as a string
//
// GenerateNewClassName 根据模式和模式创建新类名
// 使用命名模式将对象类型名与错误处理模式组合
// 返回格式化的类名作为字符串
func (cfg *ClassGenOptions) GenerateNewClassName(objectType reflect.Type, sureFlag sure.ErrorHandlingMode) string {
	return zerotern.VF(cfg.NewClassName, func() string {
		switch cfg.NamingPatternType {
		case STYLE_PREFIX_LOWER_TYPE:
			return strings.ToLower(string(sureFlag)) + cfg.NewClassNameParts + objectType.Name()
		case STYLE_SUFFIX_LOWER_TYPE:
			return objectType.Name() + cfg.NewClassNameParts + strings.ToLower(string(sureFlag))

		case STYLE_PREFIX_UPPER_TYPE:
			return strings.ToUpper(string(sureFlag)) + cfg.NewClassNameParts + objectType.Name()
		case STYLE_SUFFIX_UPPER_TYPE:
			return objectType.Name() + cfg.NewClassNameParts + strings.ToUpper(string(sureFlag))

		case STYLE_PREFIX_CAMELCASE_TYPE:
			return string(sureFlag) + cfg.NewClassNameParts + objectType.Name()
		case STYLE_SUFFIX_CAMELCASE_TYPE, NamingPattern(""): //默认值就是 ClassNameMust 或者 ClassNameSoft 新类名
			return objectType.Name() + cfg.NewClassNameParts + string(sureFlag)
		}
		return strings.ToLower(string(sureFlag)) + cfg.NewClassNameParts + objectType.Name()
	})
}
