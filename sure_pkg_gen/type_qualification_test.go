package sure_pkg_gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/sure"
)

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
