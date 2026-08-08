package cls_stub_gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

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
