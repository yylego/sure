package sure_cls_gen

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
)

// Widget gives the regeneration checks a type name to look up
// Widget 给"重新生成"这组用例一个可查的类型名
type Widget struct{}

// widgetSource is the source the codegen scans, written into a temp path
// widgetSource 是生成器要扫描的源码，写进临时目录
const widgetSource = `package tempsource

type Widget struct{}

func (w *Widget) Fetch() (int, error) {
	return 0, nil
}
`

// generateWidgetClasses writes the artifact beside the source and hands its path back
//
// The artifact belongs in the source root on purpose, since that is the shape sure_cls_gen produces: a wrapping class has to sit in the package holding the wrapped type
// It also puts the artifact in reach of the next scan, which is what these checks are about
//
// generateWidgetClasses 把产物写在源码旁边，并把产物路径返回
//
// 产物就该落在源码目录里，因为 sure_cls_gen 本来就是这个形状：包装类必须与被包装类型同包
// 这也让产物落进下一轮扫描的射程，而这正是本组用例要查的地方
func generateWidgetClasses(t *testing.T, sourceRoot string, modes ...sure.ErrorHandlingMode) string {
	t.Helper()
	outputPath := filepath.Join(sourceRoot, "sure.gen.go")
	GenerateClasses(&ClassGenConfig{
		ClassGenOptions: NewClassGenOptions(sourceRoot).MoreErrorHandlingModes(modes...),
		PackageName:     "tempsource",
		OutputPath:      outputPath,
	}, Widget{})
	return outputPath
}

// TestRegenerateNarrowedModesWritesNoStaleEntrance checks a run that states less modes than the run before it leaves nothing behind pointing at a class it stopped writing
//
// The artifact holds one entrance method from each mode (Must/Soft/Omit) on the wrapped type, so a scan reaching it reads those as source methods and wraps the wrapping
// Narrowing the mode set is where that bites: the stale Omit entrance gets transcribed into an invocation of a WidgetOmit class this run stopped generating, and the artifact stops compiling
// Nothing downstream catches it, since the imports remain above board and MustKnownImports has no complaint to make
//
// TestRegenerateNarrowedModesWritesNoStaleEntrance 检查"模式比上一轮少"的这一趟，不会留下任何指向"它已不再生成的类"的东西
//
// 产物里每个模式都在被包装类型上留了一个入口方法（Must/Soft/Omit），扫描一旦读到它，就会把这些当成源方法、给包装再包一层
// 缩减模式集时最要命：残留的 Omit 入口会被抄成一句"调用本轮已不再生成的 WidgetOmit 类"，产物当场编译不过
// 下游没人拦得住——导入全都清清白白，MustKnownImports 无话可说
func TestRegenerateNarrowedModesWritesNoStaleEntrance(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(widgetSource), 0644))

	outputPath := generateWidgetClasses(t, sourceRoot, sure.MUST, sure.SOFT, sure.OMIT)
	require.Contains(t, string(rese.V1(os.ReadFile(outputPath))), "WidgetOmit")

	// The same source root, this time with Omit left out
	// 同一个源目录，这一趟不要 Omit
	outputPath = generateWidgetClasses(t, sourceRoot, sure.MUST, sure.SOFT)

	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	require.NotContains(t, generatedCode, "WidgetOmit")
	require.NotContains(t, generatedCode, "Omit()")
	// The methods that belong keep on coming through
	// 该在的方法照常出现
	require.Contains(t, generatedCode, "func (T *WidgetMust) Fetch() (res int)")
	require.Contains(t, generatedCode, "func (T *WidgetSoft) Fetch() (res int)")
}

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

// TestRegenerateReplacesABrokenArtifact checks a run gets through with an artifact on disk that cannot get parsed, and replaces it
//
// This is the deadlock sure.DirectiveSkipGenerate exists to break, and the build constraint alone does not break it
// The constraint keeps the build off the broken file, while the codegen reads the source root off the filesystem, where no build tag holds
// Parsing asserts, so the run meant to write the replacement dies ahead of writing it, and the sole escape left is hand-editing a file headed DO NOT EDIT
//
// TestRegenerateReplacesABrokenArtifact 检查磁盘上躺着一个解析不了的产物时，这一趟仍能跑通并把它替换掉。
//
// 这正是 sure.DirectiveSkipGenerate 要解开的僵局，而光有构建约束解不开
// 约束挡住的是编译器，生成器却是照文件系统读源目录的，那里没有构建标签这回事
// 解析带断言，于是"本该写出替换版本"的那一趟先崩在写之前，剩下的唯一出路就是手工去改一个写着"请勿编辑"的文件
func TestRegenerateReplacesABrokenArtifact(t *testing.T) {
	sourceRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "tempsource.go"), []byte(widgetSource), 0644))

	outputPath := filepath.Join(sourceRoot, "sure.gen.go")
	require.NoError(t, os.WriteFile(outputPath, []byte("package tempsource\n\nfunc broken( {\n"), 0644))

	require.NotPanics(t, func() {
		generateWidgetClasses(t, sourceRoot, sure.MUST)
	})

	generatedCode := string(rese.V1(os.ReadFile(outputPath)))
	t.Log("\n" + generatedCode)

	_, err := parser.ParseFile(token.NewFileSet(), outputPath, generatedCode, parser.AllErrors)
	require.NoError(t, err)
	require.Contains(t, generatedCode, "func (T *WidgetMust) Fetch() (res int)")
}
