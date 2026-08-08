package utils

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// parseSource turns test source text into an AST file
// parseSource 把测试用的源码文本解析成 AST 文件
func parseSource(t *testing.T, sourceCode string) *ast.File {
	t.Helper()
	astFile, err := parser.ParseFile(token.NewFileSet(), "probe.go", sourceCode, parser.AllErrors)
	require.NoError(t, err)
	return astFile
}

// TestCollectFileImportsKeepsAliasAndDropsBlankAndDot checks which imports move on
// A plain one moves, an aliased one moves with its alias, blank and dot remain behind
//
// TestCollectFileImportsKeepsAliasAndDropsBlankAndDot 检查哪些导入会跟着走
// 普通的走，带别名的连别名一起走，空白与点导入留下
func TestCollectFileImportsKeepsAliasAndDropsBlankAndDot(t *testing.T) {
	astFile := parseSource(t, `package probe

import (
	"strconv"
	timekit "time"
	_ "embed"
	. "math"
)
`)
	require.Equal(t, SourceImports{
		{Alias: "", Path: "strconv"},
		{Alias: "timekit", Path: "time"},
	}, CollectFileImports(astFile))
}

// tupleFromSource builds a scan result in the shape CollectReceiverMethods produces
// tupleFromSource 按 CollectReceiverMethods 产出的形状，从源码文本造出一份扫描结果
func tupleFromSource(t *testing.T, sourceCode string) *SourceMethodsTuple {
	t.Helper()
	astFile := parseSource(t, sourceCode)
	var methodList []*ast.FuncDecl
	for _, astDeclaration := range astFile.Decls {
		if astFunc, ok := astDeclaration.(*ast.FuncDecl); ok {
			methodList = append(methodList, astFunc)
		}
	}
	return &SourceMethodsTuple{
		SourceCode:    []byte(sourceCode),
		MethodList:    methodList,
		SourceImports: CollectFileImports(astFile),
	}
}

// TestMergeSourceImportsDedupes checks one import stated twice lands once
// TestMergeSourceImportsDedupes 检查同一条导入被两个文件写到时，合并后只留一份
func TestMergeSourceImportsDedupes(t *testing.T) {
	tuples := SourceMethodsTuples{
		tupleFromSource(t, `package probe

import (
	"strconv"
	timekit "time"
)

func (p *Probe) Wait(span timekit.Duration) string { return strconv.Itoa(1) }
`),
		tupleFromSource(t, `package probe

import "strconv"

func (p *Probe) Name() string { return strconv.Itoa(2) }
`),
	}
	require.Equal(t, SourceImports{
		{Alias: "", Path: "strconv"},
		{Alias: "timekit", Path: "time"},
	}, tuples.MergeSourceImports())
}

// TestMergeSourceImportsDropsAliasMissingFromSignatures checks unused aliases get dropped
// An alias missing from each signature adds dead weight and invites a needless clash
//
// TestMergeSourceImportsDropsAliasMissingFromSignatures 检查没用到的别名会被丢掉
// 别名没出现在任何签名里就是累赘，还会白白引发别名冲突
func TestMergeSourceImportsDropsAliasMissingFromSignatures(t *testing.T) {
	tuples := SourceMethodsTuples{
		tupleFromSource(t, `package probe

import store "database/sql"

func (p *Probe) Name() string { var db *store.DB; _ = db; return "" }
`),
		tupleFromSource(t, `package probe

import store "net/http"

func (p *Probe) Code() int { var client *store.Client; _ = client; return 0 }
`),
	}
	// Two files gave one alias to two paths, and both uses sit inside method bodies
	// instead of signatures, so no import moves across and no import should halt this
	// 两个文件把同一别名给了不同路径，但都只用在方法体里、没进签名，
	// 于是两条都不走，也就不该报错
	require.NotPanics(t, func() {
		require.Empty(t, tuples.MergeSourceImports())
	})
}

// TestMergeSourceImportsRejectsAliasClash checks one alias naming two paths stops
// Both aliases do reach a signature and in fact cannot coexist, so halting is apt
//
// TestMergeSourceImportsRejectsAliasClash 检查同一别名指向两条路径时会停住
// 两个别名都真的进了签名、确实没法共存，这时停住才对，不能让破损文件落盘
func TestMergeSourceImportsRejectsAliasClash(t *testing.T) {
	tuples := SourceMethodsTuples{
		tupleFromSource(t, `package probe

import store "database/sql"

func (p *Probe) Take() *store.DB { return nil }
`),
		tupleFromSource(t, `package probe

import store "net/http"

func (p *Probe) Send() *store.Client { return nil }
`),
	}
	require.Panics(t, func() { tuples.MergeSourceImports() })
}

// TestFormatImportBlock checks the output declaration, and that a blank set gives nothing
// TestFormatImportBlock 检查渲染出的声明，以及集合为空时给出的是什么都没有、而非一个空块
func TestFormatImportBlock(t *testing.T) {
	require.Equal(t, "", SourceImports{}.FormatImportBlock())

	block := SourceImports{{Path: "strconv"}, {Alias: "timekit", Path: "time"}}.FormatImportBlock()
	require.Contains(t, block, "\"strconv\"")
	require.Contains(t, block, "timekit \"time\"")
}

// TestFormatGoBuildDirective checks both spellings give one same line
// TestFormatGoBuildDirective 检查两种写法给出同一行，且没配置指令时什么都不渲染
func TestFormatGoBuildDirective(t *testing.T) {
	require.Equal(t, "", FormatGoBuildDirective(""))
	require.Equal(t, "//go:build !sure_generate", FormatGoBuildDirective("!sure_generate"))
	require.Equal(t, "//go:build !sure_generate", FormatGoBuildDirective("//go:build !sure_generate"))
}

// TestPickGoBuildDirectiveTellsTheTwoBlanksApart checks a blank no one settled picks up the fallback, and that a blank settled on purpose stands
// TestPickGoBuildDirectiveTellsTheTwoBlanksApart 检查"没人表态的空"会吃到兜底值，而"特意留的空"照旧算数
func TestPickGoBuildDirectiveTellsTheTwoBlanksApart(t *testing.T) {
	// No one spoke, so the fallback steps in — the shape a bare struct value gives
	// 没人表态，兜底值补上——直接写结构体字面量就是这个形状
	require.Equal(t, "!sure_generate", PickGoBuildDirective("", false, "!sure_generate"))
	// Settled blank on purpose, which WithNoBuildDirective means, so nothing gets written
	// 特意留空（WithNoBuildDirective 的含义），于是什么都不写
	require.Equal(t, "", PickGoBuildDirective("", true, "!sure_generate"))
	// A stated directive beats the fallback regardless of which shape built the config
	// 写明了的指令，不论配置是怎么造出来的都优先
	require.Equal(t, "!other_tag", PickGoBuildDirective("!other_tag", false, "!sure_generate"))
	require.Equal(t, "!other_tag", PickGoBuildDirective("!other_tag", true, "!sure_generate"))
}

// TestFormatGoBuildDirectiveStopsOnALineGoCannotRead checks a directive that renders into no build constraint halts the run instead of reaching the artifact
//
// A value that holds its own prefix without the space renders as "//go:build //go:build!x", which the parse step reads as a plain comment
// The artifact formats, gets written, and go build then rejects the package holding it, so the misconfigured directive gets named at this point instead
//
// TestFormatGoBuildDirectiveStopsOnALineGoCannotRead 检查"渲染出来根本不是构建约束"的指令会把流程停住，而不是溜进产物
//
// 自带前缀却少个空格的值会渲染成 "//go:build //go:build!x"，在解析器眼里只是条普通注释
// 产物照样格式化、照样落盘，然后 go build 把它所在的包否掉，所以就此把配错的指令指名道姓报出来
func TestFormatGoBuildDirectiveStopsOnALineGoCannotRead(t *testing.T) {
	require.Panics(t, func() {
		FormatGoBuildDirective("//go:build!sure_generate")
	})
	require.Panics(t, func() {
		FormatGoBuildDirective("!")
	})
}

// TestMustKnownImportsAcceptsStatedPaths checks stated imports pass through
// TestMustKnownImportsAcceptsStatedPaths 检查能追溯到扫描源文件或调用方登记的导入可以放行
func TestMustKnownImportsAcceptsStatedPaths(t *testing.T) {
	source := []byte("package probe\n\nimport (\n\t\"strconv\"\n\ttimekit \"time\"\n)\n")
	require.NotPanics(t, func() {
		MustKnownImports(source, []string{"strconv", "time"})
	})
}

// TestMustKnownImportsRejectsGuessedPath checks an unstated path stops the run
//
// That shape is the accident this guard exists to catch: an import step that fills a
// gap with guesswork, dropping a wrong path (one missing /v5) into a file that cannot
// compile
//
// TestMustKnownImportsRejectsGuessedPath 检查谁也没写过的路径会把流程停住
// 那正是这道护栏要抓的事故形状：导入环节用猜测去填缺口，
// 把一条错误路径（少了 /v5）落进一个编译不过的文件里
func TestMustKnownImportsRejectsGuessedPath(t *testing.T) {
	source := []byte("package probe\n\nimport \"github.com/go-git/go-git/plumbing\"\n")
	require.Panics(t, func() {
		MustKnownImports(source, []string{"github.com/go-git/go-git/v5/plumbing"})
	})
}

// writeProbeFile puts a stub file on disk and returns its path
// writeProbeFile 在磁盘上放一个占位文件并返回其路径
func writeProbeFile(t *testing.T, root string, name string) string {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.WriteFile(path, []byte("package probe\n"), 0644))
	return path
}

// TestIsOutputFileNamesTheArtifactAndSparesTheRest checks the named file gets picked out and a sibling in the same root does not
// TestIsOutputFileNamesTheArtifactAndSparesTheRest 检查被点名的文件会被认出来，而同目录里的邻居不受牵连
func TestIsOutputFileNamesTheArtifactAndSparesTheRest(t *testing.T) {
	root := t.TempDir()
	sourcePath := writeProbeFile(t, root, "source.go")
	outputPath := writeProbeFile(t, root, "sure.gen.go")

	require.True(t, IsOutputFile(outputPath, outputPath))
	require.False(t, IsOutputFile(sourcePath, outputPath))
}

// TestIsOutputFileNamesNothingWhenPathIsBlank checks a blank output path names nothing
// A blank path means the code goes to stdout, so no file gets overwritten
//
// TestIsOutputFileNamesNothingWhenPathIsBlank 检查输出路径为空时谁都不点名
// 路径为空表示代码打到标准输出，没有任何文件会被覆写
func TestIsOutputFileNamesNothingWhenPathIsBlank(t *testing.T) {
	require.False(t, IsOutputFile("a.go", ""))
	require.False(t, IsOutputFile("", "sure.gen.go"))
}

// TestIsOutputFileSparesAPathElsewhere checks a file in a distinct root remains untouched
// TestIsOutputFileSparesAPathElsewhere 检查位于别处的文件不会被误认
func TestIsOutputFileSparesAPathElsewhere(t *testing.T) {
	root := t.TempDir()
	sourcePath := writeProbeFile(t, root, "source.go")

	require.False(t, IsOutputFile(sourcePath, filepath.Join(t.TempDir(), "sure.gen.go")))
}

// TestIsOutputFileNamesTheArtifactBeforeItExists checks the match holds on a first run
//
// The artifact has no need to exist yet, so os.Stat misses on it and the comparison rests on the paths alone
// A run that missed here would read its own artifact back on the next pass
//
// TestIsOutputFileNamesTheArtifactBeforeItExists 检查首次生成时（产物还不存在）也认得出来。
//
// 首轮产物本来就还不存在，os.Stat 拿不到它，比较只能靠路径本身
// 这里漏判的话，下一轮就会把自己上一轮的产物读回来
func TestIsOutputFileNamesTheArtifactBeforeItExists(t *testing.T) {
	root := t.TempDir()
	outputPath := filepath.Join(root, "sure.gen.go")

	require.NoFileExists(t, outputPath)
	require.True(t, IsOutputFile(outputPath, outputPath))
}

// TestIsOutputFileSeesPastSymlink checks one file reached two ways counts as one
//
// Comparing the two path strings alone would miss the hit and hand the artifact back to the scan, which is how the import block of a past artifact gets handed down.
//
// TestIsOutputFileSeesPastSymlink 检查同一个文件经两条路径抵达时仍算作同一个。
//
// 只比路径字符串会漏判、把产物又交回扫描，上一轮产物的导入块正是这样被一代代传下去的
func TestIsOutputFileSeesPastSymlink(t *testing.T) {
	root := t.TempDir()
	outputPath := writeProbeFile(t, root, "sure.gen.go")

	linkRoot := filepath.Join(t.TempDir(), "linked")
	require.NoError(t, os.Symlink(root, linkRoot))

	require.True(t, IsOutputFile(outputPath, filepath.Join(linkRoot, "sure.gen.go")))
}

// TestCollectReceiverMethodsSkipsABrokenArtifact checks the scan gets past an artifact that cannot get parsed, since that is the run meant to replace it
//
// The parse beneath CollectReceiverMethods asserts, so reading the artifact would take down the one run holding the escape hatch. See sure.DirectiveSkipGenerate
//
// TestCollectReceiverMethodsSkipsABrokenArtifact 检查扫描能越过一个解析不了的产物，因为这一趟正是要来替换它的。
//
// CollectReceiverMethods 底下那步解析是带断言的，读它就等于把"手握退路的这一趟"搞挂。参见 sure.DirectiveSkipGenerate
func TestCollectReceiverMethodsSkipsABrokenArtifact(t *testing.T) {
	root := t.TempDir()

	sourcePath := filepath.Join(root, "source.go")
	require.NoError(t, os.WriteFile(sourcePath, []byte("package probe\n\nfunc (a *Probe) GetN() error { return nil }\n"), 0644))

	outputPath := filepath.Join(root, "sure.gen.go")
	require.NoError(t, os.WriteFile(outputPath, []byte("package probe\n\nfunc broken( {\n"), 0644))

	tuples := CollectReceiverMethods(root, "Probe", outputPath)

	require.Len(t, tuples, 1)
	require.Equal(t, sourcePath, tuples[0].Path)
}
