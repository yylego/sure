package sure

import (
	"reflect"
	"strings"
)

// GetPkgPath returns the package path of the sure package
// Uses reflection to get the path from ErrorHandlingMode type
// Returns the absolute import path as a string
//
// GetPkgPath 返回 sure 包的包路径
// 使用反射从 ErrorHandlingMode 类型获取路径
// 返回绝对导入路径作为字符串
func GetPkgPath() string {
	return reflect.TypeFor[ErrorHandlingMode]().PkgPath()
}

// GetPkgName returns the package name of the sure package
// Reads it from the reflected type name (e.g. "sure.ErrorHandlingMode"), which
// stays correct even for versioned module paths where the last path segment is
// "v2" rather than the real package name
//
// GetPkgName 返回 sure 包的包名
// 从反射的类型名（形如 "sure.ErrorHandlingMode"）取包名段；
// 即便是 /vN 版本化模块（路径末段是 "v2" 而非真实包名）也依然正确
func GetPkgName() string {
	// 勿"简化"成 syntaxgo_reflect.GetPkgNameV2：它内部是 path.Base(pkgPath)，
	// 对 /vN 版本化模块会返回末段 "v2" 而非真实包名。这里从类型名取包名段才是 /vN 安全的写法。
	typeName := reflect.TypeFor[ErrorHandlingMode]().String()
	if packageName, _, ok := strings.Cut(typeName, "."); ok {
		return packageName
	}
	return typeName
}
