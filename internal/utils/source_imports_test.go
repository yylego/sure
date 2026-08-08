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
	for _, decl := range astFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			methodList = append(methodList, funcDecl)
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

// TestDropOutputFileRemovesTheArtifact checks the named file leaves and the rest remain
// TestDropOutputFileRemovesTheArtifact 检查被点名的文件离场，其余留下
func TestDropOutputFileRemovesTheArtifact(t *testing.T) {
	root := t.TempDir()
	sourcePath := writeProbeFile(t, root, "source.go")
	outputPath := writeProbeFile(t, root, "sure.gen.go")

	tuples := SourceMethodsTuples{{Path: sourcePath}, {Path: outputPath}}
	kept := tuples.DropOutputFile(outputPath)

	require.Len(t, kept, 1)
	require.Equal(t, sourcePath, kept[0].Path)
}

// TestDropOutputFileKeepsEachWhenPathIsBlank checks a blank output path drops nothing
// A blank path means the code goes to stdout, so no file gets overwritten
//
// TestDropOutputFileKeepsEachWhenPathIsBlank 检查输出路径为空时一个都不剔
// 路径为空表示代码打到标准输出，没有任何文件会被覆写
func TestDropOutputFileKeepsEachWhenPathIsBlank(t *testing.T) {
	tuples := SourceMethodsTuples{{Path: "a.go"}, {Path: "b.go"}}
	require.Len(t, tuples.DropOutputFile(""), 2)
}

// TestDropOutputFileKeepsAnUnmatchedPath checks a file elsewhere stays in the scan
// TestDropOutputFileKeepsAnUnmatchedPath 检查位于别处的文件仍留在扫描结果里
func TestDropOutputFileKeepsAnUnmatchedPath(t *testing.T) {
	root := t.TempDir()
	sourcePath := writeProbeFile(t, root, "source.go")

	tuples := SourceMethodsTuples{{Path: sourcePath}}
	require.Len(t, tuples.DropOutputFile(filepath.Join(t.TempDir(), "sure.gen.go")), 1)
}

// TestDropOutputFileSeesPastSymlink checks one file reached two ways counts as one
//
// Comparing the two path strings alone would miss the hit and hand the artifact back
// to the scan, which is how the import block of a past artifact gets handed down
//
// # TestDropOutputFileSeesPastSymlink 检查同一个文件经两条路径抵达时仍算作同一个
//
// 只比路径字符串会漏判、把产物又交回扫描，上一轮产物的导入块正是这样被一代代传下去的
func TestDropOutputFileSeesPastSymlink(t *testing.T) {
	root := t.TempDir()
	outputPath := writeProbeFile(t, root, "sure.gen.go")

	linkRoot := filepath.Join(t.TempDir(), "linked")
	require.NoError(t, os.Symlink(root, linkRoot))

	tuples := SourceMethodsTuples{{Path: outputPath}}
	require.Empty(t, tuples.DropOutputFile(filepath.Join(linkRoot, "sure.gen.go")))
}
