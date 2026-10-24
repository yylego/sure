package sure_pkg_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
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
	require.Contains(t, strings.Join(sureFunctions, "\n"), "func DemoFunction(objectType *sure_pkg_gen.ObjectType) *sure_pkg_gen.ObjectType")
}

// TestGeneratedFileNameTrimsTheExactExtension guards the output-file naming: the
// mode suffix must attach right before the trailing ".go", using TrimSuffix.
// A source file whose name contains ".go" in the middle (e.g. "x.gogen.go")
// would break the old strings.Replace(name, ".go", …) path — it hits the first
// ".go" and yields "x_must.gogen.go" instead of "x.gogen_must.go".
//
// TestGeneratedFileNameTrimsTheExactExtension 守护输出文件命名：模式后缀必须精确接在结尾的 ".go" 前
// （用 TrimSuffix）。文件名中间含 ".go"（如 "x.gogen.go"）时，旧的 strings.Replace(name, ".go", …)
// 会命中中间那处、生成 "x_must.gogen.go" 而非 "x.gogen_must.go"。
func TestGeneratedFileNameTrimsTheExactExtension(t *testing.T) {
	sourceRoot := t.TempDir()
	src := `package demo

func Do() error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "x.gogen.go"), []byte(src), 0o644))

	cfg := NewSurePackageConfig(sourceRoot, sure.MUST, "example.com/demo")
	GenerateSurePackageFiles(cfg)

	outRoot := filepath.Join(sourceRoot, "demo_must")
	_, err := os.Stat(filepath.Join(outRoot, "x.gogen_must.go"))
	require.NoError(t, err, "suffix must attach before the trailing .go")

	_, err = os.Stat(filepath.Join(outRoot, "x_must.gogen.go"))
	require.True(t, os.IsNotExist(err), "must not replace a middle .go (the old strings.Replace bug)")
}

// TestGeneratedFromNamesTheCallSite checks the "Generated from:" line names the code that calls GenerateSurePackageFiles
// TestGeneratedFromNamesTheCallSite 检查直接调用 GenerateSurePackageFiles 时，"Generated from:" 那一行记的是调用它的代码
func TestGeneratedFromNamesTheCallSite(t *testing.T) {
	sourceRoot := t.TempDir()
	src := `package demo

func Do() error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "demo.go"), []byte(src), 0o644))

	cfg := NewSurePackageConfig(sourceRoot, sure.MUST, "example.com/demo")
	_, file, line, ok := runtime.Caller(0)
	require.True(t, ok)
	GenerateSurePackageFiles(cfg)

	output, err := os.ReadFile(filepath.Join(sourceRoot, "demo_must", "demo_must.go"))
	require.NoError(t, err)
	var generatedFrom string
	for s := range strings.SplitSeq(string(output), "\n") {
		if strings.HasPrefix(s, "// Generated from:") {
			generatedFrom = s
		}
	}
	require.Equal(t, fmt.Sprintf("// Generated from: %s:%d -> sure_pkg_gen.TestGeneratedFromNamesTheCallSite", filepath.Base(file), line+2), generatedFrom)
}

// scannedSource is written into a temp path and scanned from there
//
// Both imports resist guessing: timekit is an alias, and the yaml package sits at a
// path ending in .v3, so the import block of this source is the one place to get
// them. Amount belongs to the scanned package, and the emitted package sits
// elsewhere, so it has to gain the package prefix.
//
// scannedSource 写进临时目录、并从那里被扫描的源码
//
// timekit 与 yaml 都猜不出来：timekit 是别名，yaml 的路径以 .v3 结尾，
// 只有这份源码的导入块里写着它们。Amount 则属于被扫描的那个包，
// 而生成出的包在别处，所以它必须被加上包前缀。
const scannedSource = `package tempsource

import (
	timekit "time"
	"gopkg.in/yaml.v3"
)

type Amount struct {
	Code int
}

func Convert(span timekit.Duration, amount Amount) (yaml.Node, error) {
	return yaml.Node{}, nil
}
`

// TestGeneratePackageCarriesSourceImports checks the scanned import block reaches the output
// Checks that a qualified type gains no second package prefix
//
// TestGeneratePackageCarriesSourceImports 检查被扫描文件的导入块会进入生成代码
// 检查已限定的类型不会被再加一层包前缀
func TestGeneratePackageCarriesSourceImports(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(scannedSource), 0644))

	GenerateSurePackage(sourceRoot, sure.MUST, "example.com/tempsource")

	outputPath := filepath.Join(sourceRoot, "tempsource_must", "tempsource_must.go")
	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	// The alias and the versioned path both have to move across
	// 别名与版本化路径都得跟着过来
	require.Contains(t, generatedCode, `timekit "time"`)
	require.Contains(t, generatedCode, `"gopkg.in/yaml.v3"`)
	// A type of the scanned package gains the package prefix
	// 被扫描包自己的类型要加包前缀
	require.Contains(t, generatedCode, "tempsource.Amount")
	// An outside type that came in qualified gains no second prefix
	// 已限定的外部类型不能被再加一层前缀
	require.NotContains(t, generatedCode, "tempsource.timekit.")
	require.NotContains(t, generatedCode, "tempsource.yaml.")
}

// TestGenerateTypeQualification guards package-qualification of source types in
// generated wrappings. Each exported type of the source package must gain the
// package prefix, including inside composite types (slices, maps, arrays, channels,
// pointers, variadics). Prevents regression of the naive string resolution that
// mishandled variadic ("name ... ...type") and dropped the prefix on []T / map.
//
// TestGenerateTypeQualification 守护生成包装器时源类型的包限定。
// 源包的每个导出类型都必须加包前缀，包括复合类型（切片、map、数组、通道、指针、可变参数）内部。
// 防止旧字符串解析器的缺陷再现：可变参数写成 "name ... ...type"、[]T / map 漏掉前缀。
func TestGenerateTypeQualification(t *testing.T) {
	root := t.TempDir()
	src := `package demo

type Item struct{}

func Sum(nums ...int) (int, error) { return 0, nil }

func Collect(items ...Item) (int, error) { return 0, nil }

func GetSlice() ([]Item, error) { return nil, nil }

func GetMap() (map[string]Item, error) { return nil, nil }

func GetPtr() (*Item, error) { return nil, nil }

func GetArray() ([3]Item, error) { return [3]Item{}, nil }

func GetChan() (chan Item, error) { return nil, nil }

func TakeSlice(items []Item) (int, error) { return 0, nil }
`
	path := filepath.Join(root, "demo.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	cfg := NewSurePackageConfig(root, sure.MUST, "example.com/demo")
	code := strings.Join(GenerateSureFunctions(cfg, path), "\n")

	require.NotContains(t, code, "... ...", "variadic must not double the ellipsis")

	// variadic: single ellipsis + invocation-site spread + prefix on the exported element
	require.Contains(t, code, "func Sum(nums ...int)")
	require.Contains(t, code, "demo.Sum(nums...)")
	require.Contains(t, code, "func Collect(items ...demo.Item)")
	require.Contains(t, code, "demo.Collect(items...)")

	// composite return types must gain the package prefix on the exported element
	require.Contains(t, code, "func GetSlice() []demo.Item")
	require.Contains(t, code, "func GetMap() map[string]demo.Item")
	require.Contains(t, code, "func GetPtr() *demo.Item")
	require.Contains(t, code, "func GetArray() [3]demo.Item")
	require.Contains(t, code, "func GetChan() chan demo.Item")

	// composite input types must gain the prefix too
	require.Contains(t, code, "func TakeSlice(items []demo.Item)")
}

// TestGenerateGenericConstraintQualification guards that a generic constraint
// referencing a source-package type gets the package prefix. The naive path
// copied the type-param set verbatim, so `func F[T Numeric]` emitted an
// undefined `Numeric` inside the generated package.
//
// TestGenerateGenericConstraintQualification 守护泛型约束里引用源包类型时会加包前缀。
// 旧写法把类型参数列表原样拷贝，`func F[T Numeric]` 会在生成包里留下未定义的 Numeric。
func TestGenerateGenericConstraintQualification(t *testing.T) {
	root := t.TempDir()
	src := `package demo

type Numeric interface{ ~int | ~float64 }

func Sum[T Numeric](a T, b T) (T, error) { return a, nil }

func Echo[T any](v T) (T, error) { return v, nil }
`
	path := filepath.Join(root, "demo.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	cfg := NewSurePackageConfig(root, sure.MUST, "example.com/demo")
	code := strings.Join(GenerateSureFunctions(cfg, path), "\n")

	// The exported constraint Numeric must be qualified; a builtin such as any remains bare.
	require.Contains(t, code, "func Sum[T demo.Numeric]")
	require.Contains(t, code, "func Echo[T any]")
	require.NotContains(t, code, "func Sum[T Numeric]")
}

// TestGenerateNameCollisionAvoidance guards that synthesized names (err/res/arg)
// does not clash with a source input name. `func F(err int) error` used to emit
// `err:=demo.F(err)` — no new variable on the := left side, a compile error.
//
// TestGenerateNameCollisionAvoidance 守护合成名（err/res/arg）不与源参数名撞车。
// `func F(err int) error` 曾生成 `err:=demo.F(err)` —— := 左侧无新变量，编译报错。
func TestGenerateNameCollisionAvoidance(t *testing.T) {
	root := t.TempDir()
	src := `package demo

func Take(err int) error { return nil }
`
	path := filepath.Join(root, "demo.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	cfg := NewSurePackageConfig(root, sure.MUST, "example.com/demo")
	code := strings.Join(GenerateSureFunctions(cfg, path), "\n")

	require.NotContains(t, code, "err:=demo.Take(err)", "the synthesized error name must not reuse the input name")
	require.Contains(t, code, "demo.Take(err)", "the input must get forwarded to the invocation just the same")
}

// TestGenerateFuncTypeQualification guards that a source-package type embedded in
// a function-typed input / return value carries the package prefix, e.g.
// `func(Item) error` becomes `func(demo.Item) error`.
//
// TestGenerateFuncTypeQualification 守护函数类型参数/返回值内部的源包类型也加前缀，
// 例如 `func(Item) error` 变成 `func(demo.Item) error`。
func TestGenerateFuncTypeQualification(t *testing.T) {
	root := t.TempDir()
	src := `package demo

type Item struct{}

func GetCallback() (func(Item) error, error) { return nil, nil }

func Attach(callback func(Item) error) error { return nil }
`
	path := filepath.Join(root, "demo.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	cfg := NewSurePackageConfig(root, sure.MUST, "example.com/demo")
	code := strings.Join(GenerateSureFunctions(cfg, path), "\n")

	require.Contains(t, code, "func(demo.Item) error")
	require.NotContains(t, code, "func(Item) error")
}

// TestGenerateBlankNames guards that a blank "_" input and a blank result each get a name in the wrapping
// The wrapping reads each result back and passes each input on, and `return _` / `demo.F(_)` cannot compile
//
// TestGenerateBlankNames 守护叫 "_" 的参数和返回值在包装里都有名字
// 包装要读回每个返回值、传下每个参数，而 `return _` / `demo.F(_)` 编译不过
func TestGenerateBlankNames(t *testing.T) {
	root := t.TempDir()
	src := `package demo

func Count() (_ int, err error) { return 0, nil }

func Check() (n int, _ error) { return 0, nil }

func Take(_ int, s string) error { return nil }
`
	path := filepath.Join(root, "demo.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	cfg := NewSurePackageConfig(root, sure.MUST, "example.com/demo")
	code := strings.Join(GenerateSureFunctions(cfg, path), "\n")
	t.Log("\n" + code)

	require.Contains(t, code, "func Count() (res0 int)")
	require.Contains(t, code, "return res0")
	require.Contains(t, code, "n,err:=demo.Check()")
	require.Contains(t, code, "sure.Must(err)")
	require.Contains(t, code, "func Take(arg0 int,s string)")
	require.Contains(t, code, "demo.Take(arg0,s)")
}

// TestGeneratePackageSkipsDirsNamedGo checks a DIR whose name ends in .go gets skipped instead of read as a file
// TestGeneratePackageSkipsDirsNamedGo 检查名字以 .go 结尾的目录会被跳过，而不是当成文件去读
func TestGeneratePackageSkipsDirsNamedGo(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "demo.go"), []byte("package demo\n\nfunc Do() error { return nil }\n"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(sourceRoot, "assets.go"), 0o755))

	GenerateSurePackageFiles(NewSurePackageConfig(sourceRoot, sure.MUST, "example.com/demo"))

	_, err := os.Stat(filepath.Join(sourceRoot, "demo_must", "demo_must.go"))
	require.NoError(t, err)
}
