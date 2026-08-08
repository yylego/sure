package sure_pkg_gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

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
