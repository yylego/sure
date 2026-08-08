// Shared code-generation helpers used across the sure generators.
// Holds receiver-method collection and name-uniquification so the generators
// share one implementation instead of each carrying a near-identical copy.
// Source-package type qualification now lives in syntaxgo_astnorm.ResolveQualifiedType.
//
// sure 各生成器共用的代码生成助手。
// 收纳"接收者方法收集"与"名字唯一化"，让各生成器共用一份实现，而不是各写一份近乎一样的副本。
// 源包类型限定已上移到 syntaxgo_astnorm.ResolveQualifiedType，此处不再持有副本。
package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yylego/done"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
	"github.com/yylego/syntaxgo/syntaxgo_search"
)

// UniqueName returns base when it is free, otherwise base with the smallest
// numeric suffix that is not yet taken. Shared by the generators to keep
// synthesized names (arg/res/err) from clashing with real parameter names in
// the generated wrapper.
//
// UniqueName 在 base 未被占用时原样返回，否则加最小的、尚未占用的数字后缀。
// 供各生成器共用，让合成名（arg/res/err）不与包装函数里真实的形参名撞车。
func UniqueName(base string, used map[string]bool) string {
	if !used[base] {
		return base
	}
	for n := 0; ; n++ {
		candidate := base + strconv.Itoa(n)
		if !used[candidate] {
			return candidate
		}
	}
}

// FormatGoBuildDirective renders a build constraint line, blank when none is set
// Accepts both "!tag" and "//go:build !tag" spellings
//
// The line belongs above the package clause with a blank line beneath it, since
// that is the spot Go inspects when it wants a build constraint
//
// FormatGoBuildDirective 渲染一行构建约束；没配置指令时返回空串
// "!tag" 和 "//go:build !tag" 两种写法都接受
//
// 这一行要放在 package 子句之前、后面空一行，因为那才是 Go 认构建约束的位置
func FormatGoBuildDirective(goBuildDirective string) string {
	if goBuildDirective == "" {
		return ""
	}
	const goBuildPrefix = "//go:build "
	if strings.HasPrefix(goBuildDirective, goBuildPrefix) {
		return goBuildDirective
	}
	return goBuildPrefix + goBuildDirective
}

// CollectReceiverMethods scans every .go file under root and gathers the methods
// whose receiver type is typeName, pairing each file's source with its matched
// method declarations and with that file's imports. The class codegen and the stub
// codegen share it, since each once carried a scan loop much the same as this.
//
// CollectReceiverMethods 扫描 root 下的每个 .go 文件，收集接收者类型为 typeName 的方法，
// 把每个文件的源码、其命中的方法声明、以及该文件的导入配对在一起。由类生成器与桩生成器共用——
// 二者此前各自持有一份近乎一样的扫描循环。
func CollectReceiverMethods(root string, typeName string) SourceMethodsTuples {
	tuples := make(SourceMethodsTuples, 0)
	for _, fileInfo := range done.VAE(os.ReadDir(root)).Done() {
		if fileInfo.IsDir() {
			continue
		}
		if filepath.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		path := filepath.Join(root, fileInfo.Name())
		sourceCode := done.VAE(os.ReadFile(path)).Done()

		astBundle := done.VCE(syntaxgo_ast.NewAstBundleV1(sourceCode)).Nice()
		astFile, _ := astBundle.GetBundle()

		methodList := syntaxgo_search.ExtractFunctionsByReceiverName(
			syntaxgo_search.ExtractFunctions(astFile), typeName, true)
		if len(methodList) == 0 {
			continue
		}

		tuples = append(tuples, &SourceMethodsTuple{
			Path:       path,
			SourceCode: sourceCode,
			MethodList: methodList,
			// Take this file's imports along: signatures get transcribed as text,
			// so the types inside need a stated origin, else downstream has to
			// guess. See the package note in source_imports.go
			// 连同本文件的导入一起取走：签名是当文本抄走的，抄来的类型得留着出处，
			// 否则下游只能靠猜。详见 source_imports.go 的包内说明
			SourceImports: CollectFileImports(astFile),
		})
	}
	return tuples
}
