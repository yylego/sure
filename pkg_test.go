package sure_test

import (
	"testing"

	"github.com/yylego/sure"
)

// TestGetPkgPath tests the GetPkgPath function
// Verifies that package path can be obtained through reflection
//
// TestGetPkgPath 测试 GetPkgPath 函数
// 验证可以通过反射获取包路径
func TestGetPkgPath(t *testing.T) {
	t.Log(sure.GetPkgPath())
}

// TestGetPkgName tests the GetPkgName function
// Verifies that package name can be obtained through reflection
//
// TestGetPkgName 测试 GetPkgName 函数
// 验证可以通过反射获取包名
func TestGetPkgName(t *testing.T) {
	t.Log(sure.GetPkgName())
}
