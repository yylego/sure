package sure_pkg_gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

// generateTempPackage runs the package codegen off the config at hand and hands the output back
// The scanned source is the fixture shared with the import checks, so these checks remain about the build constraint alone
//
// generateTempPackage 用手头这份配置跑一趟包生成，并把产物内容返回
// 被扫描的源码沿用导入检查那份共享用例，好让这里的检查只关乎构建约束
func generateTempPackage(t *testing.T, settleConfig func(cfg *SurePackageGenConfig)) string {
	t.Helper()
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(scannedSource), 0644))

	cfg := NewSurePackageConfig(sourceRoot, sure.MUST, "example.com/tempsource")
	settleConfig(cfg)
	GenerateSurePackageFiles(cfg)

	outputPath := filepath.Join(sourceRoot, "tempsource_must", "tempsource_must.go")
	return string(rese.V1(os.ReadFile(outputPath)))
}

// TestGeneratePackageStatesNoDirectiveByDefault checks the emitted package carries no build constraint unless asked
//
// This codegen writes a standalone sibling package while the test that drives it sits in the source package, so a broken artifact leaves the codegen reachable
// Writing a constraint regardless would cost the whole package its Go files beneath -tags=sure_generate and gain nothing
//
// TestGeneratePackageStatesNoDirectiveByDefault 检查生成出的包默认不带构建约束。
//
// 本生成器写的是独立的兄弟包，而驱动它的测试留在源包里，产物坏了也不妨碍生成器再跑
// 不问青红皂白写上约束，只会让那个包在 -tags=sure_generate 下一个 Go 文件都不剩，却换不来任何好处
func TestGeneratePackageStatesNoDirectiveByDefault(t *testing.T) {
	generatedCode := generateTempPackage(t, func(cfg *SurePackageGenConfig) {})
	t.Log("\n" + generatedCode)

	require.NotContains(t, generatedCode, "//go:build")
}

// TestGeneratePackageWritesTheDirectiveWhenAsked checks WithGoBuildDirective reaches the artifact
//
// The opt-in serves an output that does land somewhere able to lock, which the default cannot rule out on behalf of calling code.
//
// TestGeneratePackageWritesTheDirectiveWhenAsked 检查 WithGoBuildDirective 写下的指令会进入产物。
//
// 这个开关是给"产物确实落在会卡住的地方"那种情形准备的，而默认值替调用方判断不了这件事。
func TestGeneratePackageWritesTheDirectiveWhenAsked(t *testing.T) {
	generatedCode := generateTempPackage(t, func(cfg *SurePackageGenConfig) {
		cfg.WithGoBuildDirective(sure.DirectiveSkipGenerate)
	})
	t.Log("\n" + generatedCode)

	require.Contains(t, generatedCode, "//go:build "+sure.DirectiveSkipGenerate)
	// The constraint belongs above the package clause, which is the spot Go inspects
	// 约束要待在 package 子句之前，那才是 Go 认它的位置
	require.Less(t, indexOfLine(t, generatedCode, "//go:build"), indexOfLine(t, generatedCode, "package tempsource_must"))
}

// TestGeneratePackageDropsTheDirectiveOnDemand checks WithNoBuildDirective undoes a directive set up above in the chain
// TestGeneratePackageDropsTheDirectiveOnDemand 检查 WithNoBuildDirective 能撤掉链条上游设过的指令
func TestGeneratePackageDropsTheDirectiveOnDemand(t *testing.T) {
	generatedCode := generateTempPackage(t, func(cfg *SurePackageGenConfig) {
		cfg.WithGoBuildDirective(sure.DirectiveSkipGenerate).WithNoBuildDirective()
	})
	t.Log("\n" + generatedCode)

	require.NotContains(t, generatedCode, "//go:build")
}

// TestSurePackageConfigOptions checks each With* option reaches the artifact
// Each of the source path, import path, mode, package name, output root, and the package and prefix that handle errors gets changed, and the artifact has to reflect each one
//
// TestSurePackageConfigOptions 检查每个 With* 配置都会体现在产物里
// 源路径、导入路径、模式、包名、输出根路径、错误处理包和错误处理函数前缀全部改掉，产物要一一对得上
func TestSurePackageConfigOptions(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "demo.go"), []byte("package demo\n\nfunc Do() error { return nil }\n"), 0o644))
	outputRoot := t.TempDir()

	cfg := NewSurePackageConfig("", sure.MUST, "").
		WithSourceRoot(sourceRoot).
		WithSourcePackagePath("example.com/demo").
		WithErrorHandlingMode(sure.SOFT).
		WithNewPkgName("demosoft").
		WithOutputRoot(outputRoot).
		WithErrorHandlingPkgPath("example.com/handle").
		WithHandlerFuncReference("handle")
	GenerateSurePackageFiles(cfg)

	generatedCode := string(rese.V1(os.ReadFile(filepath.Join(outputRoot, "demosoft", "demo_soft.go"))))
	t.Log("\n" + generatedCode)

	require.Contains(t, generatedCode, "package demosoft")
	require.Contains(t, generatedCode, `"example.com/demo"`)
	require.Contains(t, generatedCode, `"example.com/handle"`)
	require.Contains(t, generatedCode, "handle.Soft(err)")
}

// indexOfLine reports where a snippet starts, and fails the test when it is absent
// indexOfLine 返回某段文本的起始位置，找不到时让测试失败
func indexOfLine(t *testing.T, sourceCode string, snippet string) int {
	t.Helper()
	index := strings.Index(sourceCode, snippet)
	require.GreaterOrEqual(t, index, 0, "snippet is absent: %s", snippet)
	return index
}
