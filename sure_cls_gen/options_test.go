package sure_cls_gen

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

// TestGenerateGivesABareOptionsTheBuildConstraint checks options built as a plain struct value get the escape hatch just the same
//
// NewClassGenOptions states the constraint, so callers going through it had protection beforehand
// A struct value skips that, and leaving such a run unconstrained keeps the deadlock in place on the shape most apt to miss it
//
// TestGenerateGivesABareOptionsTheBuildConstraint 检查直接写结构体字面量的选项照样拿得到那道退路。
//
// NewClassGenOptions 会写上约束，走它的调用方本来就有保障
// 结构体字面量绕过了那一步，让这种运行不带约束，等于把僵局留给"最不可能知道有这回事"的那种写法
func TestGenerateGivesABareOptionsTheBuildConstraint(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(widgetSource), 0644))

	outputPath := filepath.Join(sourceRoot, "sure.gen.go")
	GenerateClasses(&ClassGenConfig{
		ClassGenOptions: &ClassGenOptions{SourceRootPath: sourceRoot, ErrorHandlingModes: []sure.ErrorHandlingMode{sure.MUST}},
		PackageName:     "tempsource",
		OutputPath:      outputPath,
	}, Widget{})

	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	require.Contains(t, generatedCode, "//go:build "+sure.DirectiveSkipGenerate)
}

// TestGenerateHonoursADeliberateBlank checks WithNoBuildDirective reaches a blank constraint
//
// Blank means two things callers can intend, and this is the one that has to survive the fallback.
//
// TestGenerateHonoursADeliberateBlank 检查 WithNoBuildDirective 确实能得到"不写约束"
//
// 空串承载着调用方的两种意图，而这一种必须能扛过兜底值。
func TestGenerateHonoursADeliberateBlank(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(widgetSource), 0644))

	outputPath := filepath.Join(sourceRoot, "sure.gen.go")
	GenerateClasses(&ClassGenConfig{
		ClassGenOptions: NewClassGenOptions(sourceRoot).MoreErrorHandlingModes(sure.MUST).WithNoBuildDirective(),
		PackageName:     "tempsource",
		OutputPath:      outputPath,
	}, Widget{})

	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	require.NotContains(t, generatedCode, "//go:build")
}

// TestGetErrorHandlingModes checks Must and Soft come as the default, and stated modes replace them
// TestGetErrorHandlingModes 检查不设置时默认是 Must 和 Soft，设置了就按设置的来
func TestGetErrorHandlingModes(t *testing.T) {
	require.Equal(t, []sure.ErrorHandlingMode{sure.MUST, sure.SOFT}, NewClassGenOptions("").GetErrorHandlingModes())
	require.Equal(t, []sure.ErrorHandlingMode{sure.OMIT}, NewClassGenOptions("").MoreErrorHandlingModes(sure.OMIT).GetErrorHandlingModes())
}

// TestGenerateNewClassName checks each naming pattern, the direct class name, and that an unknown pattern stops the run
// TestGenerateNewClassName 检查每种命名模式、直接指定的类名，以及写错的模式会把流程停住
func TestGenerateNewClassName(t *testing.T) {
	widgetType := reflect.TypeFor[Widget]()
	patternNames := map[NamingPattern]string{
		STYLE_PREFIX_LOWER_TYPE:     "must88Widget",
		STYLE_SUFFIX_LOWER_TYPE:     "Widget88must",
		STYLE_PREFIX_UPPER_TYPE:     "MUST88Widget",
		STYLE_SUFFIX_UPPER_TYPE:     "Widget88MUST",
		STYLE_PREFIX_CAMELCASE_TYPE: "Must88Widget",
		STYLE_SUFFIX_CAMELCASE_TYPE: "Widget88Must",
		NamingPattern(""):           "Widget88Must",
	}
	for pattern, className := range patternNames {
		cfg := NewClassGenOptions("").WithNewClassNameParts("88").WithNamingPatternType(pattern)
		require.Equal(t, className, cfg.GenerateNewClassName(widgetType, sure.MUST), "pattern %q", pattern)
	}

	cfg := NewClassGenOptions("").WithNewClassName("Sprocket").WithNamingPatternType(STYLE_PREFIX_LOWER_TYPE)
	require.Equal(t, "Sprocket", cfg.GenerateNewClassName(widgetType, sure.MUST))

	cfg = NewClassGenOptions("").WithNamingPatternType("STYLE_UNKNOWN")
	require.Panics(t, func() {
		cfg.GenerateNewClassName(widgetType, sure.MUST)
	})
}
