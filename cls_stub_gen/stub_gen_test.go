package cls_stub_gen

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/yylego/runpath"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

// A is a test struct with multiple methods
// Used to test stub generation capabilities
//
// A 是带有多个方法的测试结构体
// 用于测试存根生成功能
type A struct{}

func (a *A) Get() *A {
	return a
}

func (a *A) Set(string) error {
	return errors.New("not implement")
}

func (a *A) Add(x int, y int) int {
	return x + y
}

func (a *A) Sub(x int, y int) (int, error) {
	return x - y, nil
}

// Param is a test parameter type
// Used to test variadic parameter handling
//
// Param 是测试参数类型
// 用于测试可变参数处理
type Param struct{}

// Who is a test method with variadic value parameters
//
// Who 是带有可变值参数的测试方法
func (a *A) Who(param ...Param) {}

// How is a test method with variadic pointer parameters
//
// How 是带有可变指针参数的测试方法
func (a *A) How(param ...*Param) {}

// TestGenerateMethodsStub tests the GenerateStubMethods function
// Verifies that methods can be wrapped as package-level functions
// Uses test struct A to validate stub generation
//
// TestGenerateMethodsStub 测试 GenerateStubMethods 函数
// 验证方法可以被包装为包级函数
// 使用测试结构体 A 验证存根生成
func TestGenerateMethodsStub(t *testing.T) {
	res := GenerateStubMethods(
		&StubGenConfig{
			SourceRootPath:    runpath.PARENT.Path(),
			TargetPackageName: "p_p_p",
			ImportOptions:     syntaxgo_ast.NewPackageImportOptions(),
			OutputPath:        "",
			AllowFileCreation: false,
		},
		NewStubParam(A{}, "singletonInstance"),
	)
	t.Log(res)
}
