package sure

import (
	"reflect"

	"github.com/yylego/syntaxgo/syntaxgo_reflect"
)

// GetPkgPath returns the package path of the sure package
// Uses reflection to get the path from ErrorHandlingMode type
// Returns the absolute import path as a string
//
// GetPkgPath 返回 sure 包的包路径
// 使用反射从 ErrorHandlingMode 类型获取路径
// 返回绝对导入路径作为字符串
func GetPkgPath() string {
	return reflect.TypeOf(syntaxgo_reflect.GetObject[ErrorHandlingMode]()).PkgPath()
}

// GetPkgName returns the package name of the sure package
// Uses reflection to extract the name from ErrorHandlingMode type
// Returns the package name as a string
//
// GetPkgName 返回 sure 包的包名
// 使用反射从 ErrorHandlingMode 类型提取名称
// 返回包名作为字符串
func GetPkgName() string {
	return syntaxgo_reflect.GetPkgName(syntaxgo_reflect.GetObject[ErrorHandlingMode]())
}
