// Package utils provides internal utilities supporting code generation and validation
// Offers JSON formatting, file operations, and type checking functions
// Supports code generation workflows with validation and exception handling
// Contains support functions used across different generation packages
//
// Package utils 提供支持代码生成和验证的内部工具函数
// 提供 JSON 格式化、文件操作和类型检查函数
// 支持带有验证和异常处理的代码生成工作流
// 包含跨不同生成包使用的辅助函数
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"unicode"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"github.com/yylego/done"
	"github.com/yylego/erero"
	"github.com/yylego/zaplog"
)

// Neat2json converts value to formatted JSON string
// Uses indented JSON format with tabs to make it readable
// Panics on exception to ensure invalid data is not silenced
//
// Neat2json 将值转换为格式化的 JSON 字符串
// 使用带缩进的 JSON 格式提高可读性
// 异常时 panic 以确保无效数据不被忽略
func Neat2json(v interface{}) string {
	data, err := NeatBytes(v)
	if err != nil {
		panic(errors.WithMessage(err, "wrong"))
	}
	return string(data)
}

// NeatBytes converts value to formatted JSON bytes
// Returns indented JSON with tabs to make it readable
// Returns exception if marshaling fails
//
// NeatBytes 将值转换为格式化的 JSON 字节
// 返回带缩进的 JSON 提高可读性
// 如果编组失败则返回异常
func NeatBytes(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, errors.WithMessage(err, "wrong")
	}
	return data, nil
}

// In checks if value exists in slice
// Uses generic type constraint to ensure safe typing
// Returns true when value is found in slice
//
// In 检查值是否存在于切片中
// 使用泛型类型约束确保类型安全
// 当在切片中找到值时返回 true
func In[V comparable](v V, slice []V) bool {
	return slices.Contains(slice, v)
}

// MustRoot validates that path exists and is a root
// Panics when path does not exist and is not a root
// Used to ensure source roots are valid before operations
//
// MustRoot 验证路径存在且是根目录
// 当路径不存在或不是根目录时 panic
// 用于在操作前确保源根路径有效
func MustRoot(root string) {
	done.VBE(IsRootExist(root)).TRUE()
}

// MustFile validates that path exists and is a file
// Panics when path does not exist and is not a file
// Used to ensure target files exist before writing
//
// MustFile 验证路径存在且是文件
// 当路径不存在或不是文件时 panic
// 用于在写入前确保目标文件存在
func MustFile(path string) {
	done.VBE(IsFileExist(path)).TRUE()
}

// IsRootExist checks if path exists and is a root
// Returns true when path is an existing root
// Returns false with nil error when path does not exist
//
// IsRootExist 检查路径是否存在且是根目录
// 当路径是现有根目录时返回 true
// 当路径不存在时返回 false 和 nil 错误
func IsRootExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path does not exist // 路径不存在
		}
		return false, erero.Wro(err) // Other errors // 其他错误
	}
	return info.IsDir(), nil
}

// IsFileExist checks if path exists and is a file
// Returns true when path is an existing file
// Returns false with nil error when path does not exist
//
// IsFileExist 检查路径是否存在且是文件
// 当路径是现有文件时返回 true
// 当路径不存在时返回 false 和 nil 错误
func IsFileExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path does not exist // 路径不存在
		}
		return false, erero.Wro(err) // Other errors // 其他错误
	}
	return !info.IsDir(), nil
}

// MustWriteIntoPath creates parent paths and writes content to file
// Auto creates parent paths when needed
// Panics on error to ensure file operations succeed
//
// MustWriteIntoPath 创建父路径并将内容写入文件
// 需要时自动创建父路径
// 出错时 panic 以确保文件操作成功
func MustWriteIntoPath(path string, s string) {
	done.Done(os.MkdirAll(filepath.Dir(path), 0755))
	done.Done(os.WriteFile(path, []byte(s), 0644))
}

// MustLs reads root and returns entry names
// Panics on read exception to ensure operation succeeds
// Returns slice of entry names in the root
//
// MustLs 读取根目录并返回条目名称
// 读取异常时 panic 以确保操作成功
// 返回根目录中条目名称的切片
func MustLs(root string) (names []string) {
	infos := done.VAE(os.ReadDir(root)).Done()
	names = make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return
}

// C0IsUppercase checks if first character is uppercase
// Returns true when string starts with uppercase letter
// Returns false when string is blank and starts with lowercase
//
// C0IsUppercase 检查首字符是否为大写
// 当字符串以大写字母开头时返回 true
// 当字符串为空或以小写开头时返回 false
func C0IsUppercase(s string) bool {
	runes := []rune(s)
	if len(runes) > 0 {
		return unicode.IsUpper(runes[0])
	}
	return false
}

// SetDoubleQuotes wraps string with double quotes
// Used to format import paths in generated code
// Returns quoted string ready to insert into code
//
// SetDoubleQuotes 用双引号包裹字符串
// 用于格式化生成代码中的导入路径
// 返回准备插入代码的带引号字符串
func SetDoubleQuotes(s string) string {
	return "\"" + s + "\""
}

// Boolean returns the boolean value as is
// Simple identity function on bool type
// Used to make boolean expressions explicit
//
// Boolean 原样返回布尔值
// 布尔类型的简单恒等函数
// 用于使布尔表达式更明确
func Boolean(v bool) bool {
	return v
}

// PrintObject prints object with debug markers
// Uses pretty printing to make output readable
// Adds separators to make debug output distinct
//
// PrintObject 使用调试标记打印对象
// 使用美化打印提高可读性
// 添加分隔符使调试输出更清晰
func PrintObject(a any) {
	zaplog.LOGS.Skip1.Debug("------------")
	zaplog.LOGS.Skip1.Debug("---object---")
	done.VNE(pretty.Println(a)).Fine()
	zaplog.LOGS.Skip1.Debug("------------")
	zaplog.LOGS.Skip1.Debug("------------")
}

// GetGenPosFuncMark gets position information to trace code generation
// Returns formatted string showing source file and function that triggered generation
// Uses runtime.Caller to walk up stack using specified frame count
// Returns empty string when position information cannot be determined
//
// GetGenPosFuncMark 获取调用位置信息，用于追踪代码生成
// 返回格式化字符串，显示触发生成的源文件和函数
// 使用 runtime.Caller 按指定帧数向上遍历调用栈
// 如果无法确定调用位置则返回空字符串
func GetGenPosFuncMark(skip int) string {
	// Get position info using runtime package
	// 使用 runtime 包获取调用位置信息
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}

	// Extract just filename, not full path
	// 只提取文件名而不是完整路径
	filename := filepath.Base(file)

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return filename
	}

	// Extract just function name, not full package path
	// 只提取函数名，不包含完整的包路径
	funcName := filepath.Base(fn.Name())
	return fmt.Sprintf("%s:%d -> %s", filename, line, funcName)
}
