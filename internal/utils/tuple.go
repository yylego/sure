package utils

import "go/ast"

// SourceMethodsTuple pairs source code with extracted methods
// Contains source bytes and corresponding AST function declarations
// Used during code generation to process methods
//
// SourceMethodsTuple 将源代码与提取的方法配对
// 包含源字节和对应的 AST 函数声明
// 在代码生成期间用于处理方法
type SourceMethodsTuple struct {
	SourceCode []byte          // Source code bytes // 源代码字节
	MethodList []*ast.FuncDecl // Extracted method declarations // 提取的方法声明
}

// SourceMethodsTuples represents a collection of source-method pairs
// Provides methods to extract common information from multiple sources
// Used when processing multiple files in a package
//
// SourceMethodsTuples 表示源-方法对的集合
// 提供从多个源提取公共信息的方法
// 在处理包中的多个文件时使用
type SourceMethodsTuples []*SourceMethodsTuple

// GetReceiverVariableName extracts the method context variable name
// Scans method list to find first context name
// Returns blank string when no context is found
//
// GetReceiverVariableName 从方法中提取接收器变量名
// 扫描方法列表以查找第一个接收器名称
// 当未找到接收器时返回空字符串
func (mts SourceMethodsTuples) GetReceiverVariableName() string {
	for _, codePage := range mts {
		for _, mebFunction := range codePage.MethodList {
			if mebFunction.Recv == nil {
				continue
			}
			if len(mebFunction.Recv.List) == 0 {
				continue
			}
			if len(mebFunction.Recv.List[0].Names) == 0 {
				continue
			}
			name := mebFunction.Recv.List[0].Names[0].Name
			if name != "" {
				return name
			}
		}
	}
	return ""
}
