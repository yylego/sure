// Package utils provides internal utilities supporting code generation and validation
// Offers JSON formatting, file operations, and calling position tracing
// Supports code generation workflows with validation and exception handling
// Contains support functions used across different generation packages
//
// Package utils 提供支持代码生成和验证的内部工具函数
// 提供 JSON 格式化、文件操作和调用位置追踪
// 支持带有验证和异常处理的代码生成工作流
// 包含跨不同生成包使用的辅助函数
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/yylego/done"
	"github.com/yylego/erero"
)

// NeatBytes converts value to formatted JSON bytes
// Returns indented JSON with tabs to make it readable
// Returns exception if marshaling fails
//
// NeatBytes 将值转换为格式化的 JSON 字节
// 返回带缩进的 JSON 提高可读性
// 如果编组失败则返回异常
func NeatBytes(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, errors.WithMessage(err, "wrong")
	}
	return data, nil
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
		return false, erero.Wro(err) // Remaining errors // 其他错误
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
		return false, erero.Wro(err) // Remaining errors // 其他错误
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

// GetGenPosFuncMark gets position information to trace code generation
// Returns formatted string showing source file and function that triggered generation
// Uses runtime.Caller to walk up the stack with the stated frame count
// Returns a blank string when position information cannot be determined
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

	// Extract just the file name, not the whole path
	// 只提取文件名而不是完整路径
	filename := filepath.Base(file)

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return filename
	}

	// Extract just the function name, not the whole package path
	// 只提取函数名，不包含完整的包路径
	funcName := filepath.Base(fn.Name())
	return fmt.Sprintf("%s:%d -> %s", filename, line, funcName)
}
