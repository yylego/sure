package sure_pkg_gen

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/sure"
	"github.com/yylego/syntaxgo/syntaxgo_reflect"
)

// TestMain is the test suite main function
// Runs the test suite in sure_pkg_gen package
//
// TestMain 是测试套件的主函数
// 运行 sure_pkg_gen 包中的测试套件
func TestMain(m *testing.M) {
	m.Run()
}

// ObjectType is a test type used in tests
// Demonstrates package generation capabilities
//
// ObjectType 是测试中使用的测试类型
// 演示包生成功能
type ObjectType struct{}

// DemoFunction is a test function that returns ObjectType
// Used to test function signature extraction
//
// DemoFunction 是返回 ObjectType 的测试函数
// 用于测试函数签名提取
func DemoFunction(objectType *ObjectType) *ObjectType {
	return objectType
}

// TestGenerateSureFunctions tests the GenerateSureFunctions function
// Verifies that functions can be extracted and wrapped with error handling
// Uses test types to validate function generation
//
// TestGenerateSureFunctions 测试 GenerateSureFunctions 函数
// 验证可以提取函数并包装错误处理
// 使用测试类型验证函数生成
func TestGenerateSureFunctions(t *testing.T) {
	path := runpath.Path()

	pkgPath := syntaxgo_reflect.GetPkgPathV2[ObjectType]()

	config := NewSurePackageConfig(runpath.PARENT.Path(), sure.MUST, pkgPath)

	sureFunctions := GenerateSureFunctions(config, path)

	for _, sureFunction := range sureFunctions {
		t.Log(sureFunction)
	}
}
