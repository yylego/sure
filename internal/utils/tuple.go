package utils

import "go/ast"

// SourceMethodsTuple pairs source code with extracted methods
// Contains source bytes, corresponding AST function declarations, and the imports
// the file states, so transcribed signatures keep types resolvable
// Used during code generation to process methods
//
// SourceMethodsTuple 将源代码与提取的方法配对
// 包含源字节、对应的 AST 函数声明，以及该文件写明的导入——
// 抄走的签名要靠它们才能让类型有出处
// 在代码生成期间用于处理方法
type SourceMethodsTuple struct {
	Path          string          // Path of the file the methods came from // 方法所在文件的路径
	SourceCode    []byte          // Source code bytes // 源代码字节
	MethodList    []*ast.FuncDecl // Extracted method declarations // 提取的方法声明
	SourceImports SourceImports   // Imports of the file the methods came from // 方法所在文件的导入
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
// Skips the blank identifier "_" (it cannot serve as a struct field or value)
// Returns blank string when no context is found
//
// GetReceiverVariableName 从方法中提取接收器变量名
// 扫描方法列表以查找第一个接收器名称
// 跳过空白标识符 "_"（它不能当结构体字段或值使用）
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
			// 跳过空白接收者 "_"：若当成变量名，会生成 `struct{ _ *A }` 与 `&AMust{_:_}`（把 _ 当值）
			// 及 `T._`（取空白字段），全都编译不过。视同"没名字"，继续找或最终退到 "T"。
			if name != "" && name != "_" {
				return name
			}
		}
	}
	return ""
}
