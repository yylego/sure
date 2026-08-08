package cls_stub_gen

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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

// Param is a test input type
// Used to test variadic input handling
//
// Param 是测试参数类型
// 用于测试可变参数处理
type Param struct{}

// Who is a test method with variadic value parameters
//
// Who 是带有可变值参数的测试方法
func (a *A) Who(param ...Param) {}

// How is a test method taking variadic `*Param` inputs
//
// How 是带有可变指针参数的测试方法
func (a *A) How(param ...*Param) {}

// hidden is an unexported method: the stub codegen must skip it, since a
// package-scope wrapping in a distinct package cannot invoke an unexported method.
//
// hidden 是未导出方法：桩生成器必须跳过它——另一个包里的包级包装无法调用未导出方法。
func (a *A) hidden(x int) error {
	if x < 0 {
		return errors.New("negative")
	}
	return nil
}

// List takes and returns a slice of the in-package Param type, exercising
// composite-type qualification in the stub codegen.
//
// List 接收并返回本地 Param 类型的切片，用于检验桩生成器的复合类型限定。
func (a *A) List(items []Param) ([]Param, error) { return items, nil }

// TestGenerateMethodsStub tests the GenerateStubMethods function
// Checks that methods can be wrapped as package-scope functions
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

// TestGenerateStubCompositeTypeQualification guards that a source-package type
// inside a composite ([]Param) is qualified with the package prefix. The shallow
// AdjustTypeWithPackage path handled just bare and single-`*` types, so
// []Param used to be emitted unqualified and would not compile in the target package.
//
// TestGenerateStubCompositeTypeQualification 守护复合类型（[]Param）内部的源包类型也加包前缀。
// 旧的浅层 AdjustTypeWithPackage 只处理裸类型和单层指针，[]Param 曾原样输出、在目标包里编译不过。
func TestGenerateStubCompositeTypeQualification(t *testing.T) {
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
	require.Contains(t, res, "[]cls_stub_gen.Param", "the composite element type must gain the package prefix")
	require.NotContains(t, res, "[]Param", "bare composite type must not survive")
}

// TestGenerateStubSkipsUnexported guards that unexported source methods are not
// wrapped: a package-scope stub in the target package would invoke an unexported
// method across packages and stop compiling. Mirrors the exported-alone guard
// sure_pkg_gen holds.
//
// TestGenerateStubSkipsUnexported 守护未导出方法不被包装：目标包里的包级 stub 会跨包调用
// 未导出方法、编译不过。与 sure_pkg_gen 里已有的仅导出守卫保持一致。
func TestGenerateStubSkipsUnexported(t *testing.T) {
	require.NoError(t, (&A{}).hidden(1)) // 引用未导出方法，避免被 unused 静态检查误报

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
	require.NotContains(t, res, "hidden", "unexported method must not be wrapped into a package-scope stub")
	require.Contains(t, res, "func Add(", "exported methods must get wrapped just the same")
}
