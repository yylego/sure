package cls_stub_gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/runpath"
	"github.com/yylego/sure"
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

// BlankBox gives the blank-input check a type name to look up
// BlankBox 给"空白参数"用例一个可查的类型名
type BlankBox struct{}

// TestGenerateStubBlankParams guards that a blank "_" input gets a name in the stub, since the stub passes each input on and `X.F(_)` cannot compile
// TestGenerateStubBlankParams 守护叫 "_" 的参数在存根里有名字，因为存根要把每个参数传下去，而 `X.F(_)` 编译不过
func TestGenerateStubBlankParams(t *testing.T) {
	sourceRoot := t.TempDir()
	src := `package cls_stub_gen

type BlankBox struct{}

func (b *BlankBox) Take(_ int, s string) error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(src), 0644))

	code := GenerateStubMethods(NewStubGenConfig(sourceRoot, "tempstub", ""), NewStubParam(BlankBox{}, "tempsource.BOX"))
	t.Log("\n" + code)

	require.Contains(t, code, "func Take(arg0 int, s string)")
	require.Contains(t, code, "tempsource.BOX.Take(arg0, s)")
}

// Gizmo hands the codegen a type name to look up
// Its methods live in the scanned source below instead of here
//
// Gizmo 给生成器一个可查的类型名
// 它的方法在下面那份被扫描的源码里、而不在这儿
type Gizmo struct{}

// scannedSource is written into a temp path and scanned from there
//
// Both imports resist guessing: timekit is an alias, and the yaml package sits at a
// path ending in .v3, so the import block of this source is the one place to get
// them. Amount belongs to the scanned package, and the stubs land in a distinct
// package, so it has to gain the package prefix.
//
// The package clause states cls_stub_gen so it matches the package holding Gizmo,
// which is where the prefix comes from, keeping the fixture true to how it gets used.
//
// scannedSource 写进临时目录、并从那里被扫描的源码
//
// timekit 与 yaml 都猜不出来：timekit 是别名，yaml 的路径以 .v3 结尾，
// 只有这份源码的导入块里写着它们。Amount 则属于被扫描的那个包，
// 而存根落在另一个包里，所以它必须被加上包前缀。
//
// package 子句写 cls_stub_gen，与 Gizmo 所在的包一致——前缀正是从那儿来的，
// 这样用例才与真实用法对得上。
const scannedSource = `package cls_stub_gen

import (
	timekit "time"
	"gopkg.in/yaml.v3"
)

type Amount struct {
	Code int
}

type Gizmo struct{}

func (g *Gizmo) Convert(span timekit.Duration, amount Amount) (yaml.Node, error) {
	return yaml.Node{}, nil
}
`

// TestGenerateStubsCarriesSourceImports checks the scanned import block reaches the output
// Checks that a qualified type gains no second package prefix
//
// That second half is the point here: the stub codegen prefixes types belonging to
// the scanned package, and mistaking timekit.Duration as one of those would emit
// cls_stub_gen.timekit.Duration, which cannot compile.
//
// TestGenerateStubsCarriesSourceImports 检查被扫描文件的导入块会进入生成代码
// 检查已限定的类型不会被再加一层包前缀
//
// 后半句是这里的重点：存根生成器会给属于被扫描包的类型加前缀，
// 若它把 timekit.Duration 也当成那一类，就会写出 cls_stub_gen.timekit.Duration，编译不过。
func TestGenerateStubsCarriesSourceImports(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(scannedSource), 0644))

	outputPath := filepath.Join(t.TempDir(), "gen.go")
	cfg := NewStubGenConfig(sourceRoot, "tempstub", outputPath).WithAllowFileCreation(true)
	GenerateStubs(cfg, NewStubParam(Gizmo{}, "tempsource.GIZMO"))

	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	// The alias and the versioned path both have to move across
	// 别名与版本化路径都得跟着过来
	require.Contains(t, generatedCode, `timekit "time"`)
	require.Contains(t, generatedCode, `"gopkg.in/yaml.v3"`)
	// A type of the scanned package gains the package prefix
	// 被扫描包自己的类型要加包前缀
	require.Contains(t, generatedCode, "cls_stub_gen.Amount")
	// An outside type that came in qualified gains no second prefix
	// 已限定的外部类型不能被再加一层前缀
	require.NotContains(t, generatedCode, "cls_stub_gen.timekit.")
	require.NotContains(t, generatedCode, "cls_stub_gen.yaml.")
	// The escape hatch that keeps a broken artifact from locking its own package
	// 那道让破损产物无法卡住自身所在包的退路
	require.Contains(t, generatedCode, "//go:build "+sure.DirectiveSkipGenerate)
}

// generateGizmoStubs runs the stub codegen off the config at hand and hands the output back
// The scanned source is the fixture shared with the import checks, so these checks remain about the build constraint alone
//
// generateGizmoStubs 用手头这份配置跑一趟存根生成，并把产物内容返回
// 被扫描的源码沿用导入检查那份共享用例，好让这里的检查只关乎构建约束
func generateGizmoStubs(t *testing.T, settleConfig func(cfg *StubGenConfig)) string {
	t.Helper()
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(scannedSource), 0644))

	outputPath := filepath.Join(t.TempDir(), "gen.go")
	cfg := &StubGenConfig{
		SourceRootPath:    sourceRoot,
		TargetPackageName: "tempstub",
		OutputPath:        outputPath,
		AllowFileCreation: true,
	}
	settleConfig(cfg)
	GenerateStubs(cfg, NewStubParam(Gizmo{}, "tempsource.GIZMO"))

	return string(rese.V1(os.ReadFile(outputPath)))
}

// TestGenerateStubsGivesABareConfigTheBuildConstraint checks a config built as a plain struct value gets the escape hatch just the same
//
// The struct value was the one documented shape before NewStubGenConfig arrived, so most existing callers write it
// Leaving those runs without the constraint would keep the deadlock in place on just the code that predates the fix
//
// TestGenerateStubsGivesABareConfigTheBuildConstraint 检查直接写结构体字面量的配置照样拿得到那道退路。
//
// 在 NewStubGenConfig 出现之前，结构体字面量是文档里唯一的写法，所以现存调用方大多是那么写的
// 让这些运行拿不到约束，等于把僵局原封不动留给"恰恰早于修复"的那批代码
func TestGenerateStubsGivesABareConfigTheBuildConstraint(t *testing.T) {
	generatedCode := generateGizmoStubs(t, func(cfg *StubGenConfig) {})
	t.Log("\n" + generatedCode)

	require.Contains(t, generatedCode, "//go:build "+sure.DirectiveSkipGenerate)
}

// TestGenerateStubsKeepsAStatedDirective checks a directive stated on the struct beats the fallback
// TestGenerateStubsKeepsAStatedDirective 检查在结构体上写明的指令优先于兜底值
func TestGenerateStubsKeepsAStatedDirective(t *testing.T) {
	generatedCode := generateGizmoStubs(t, func(cfg *StubGenConfig) {
		cfg.GoBuildDirective = "!tempstub_generate"
	})
	t.Log("\n" + generatedCode)

	require.Contains(t, generatedCode, "//go:build !tempstub_generate")
	require.NotContains(t, generatedCode, "//go:build "+sure.DirectiveSkipGenerate)
}

// TestGenerateStubsHonourADeliberateBlank checks WithNoBuildDirective reaches a blank constraint
//
// Blank means two things callers can intend, and this is the one that has to survive the fallback.
//
// TestGenerateStubsHonourADeliberateBlank 检查 WithNoBuildDirective 确实能得到"不写约束"
//
// 空串承载着调用方的两种意图，而这一种必须能扛过兜底值。
func TestGenerateStubsHonourADeliberateBlank(t *testing.T) {
	generatedCode := generateGizmoStubs(t, func(cfg *StubGenConfig) {
		cfg.WithNoBuildDirective()
	})
	t.Log("\n" + generatedCode)

	require.NotContains(t, generatedCode, "//go:build")
}
