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
}

// NewClassGenOptions creates a new ClassGenOptions with the source root path
// Initializes configuration with the provided source code location
// Returns a new options instance that accepts more configuration
//
// NewClassGenOptions 使用源根路径创建新的 ClassGenOptions
// 使用提供的源代码位置初始化配置
// 返回接受进一步配置的新选项实例
func NewClassGenOptions(srcRoot string) *ClassGenOptions {
	return &ClassGenOptions{SourceRootPath: srcRoot}
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
