package cls_stub_gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

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
