// Shared code-generation helpers used across the sure generators.
// Holds recv-method collection and name-uniquification so the codegens
// share one implementation instead of each carrying a separate twin.
// Source-package type qualification now lives in syntaxgo_astnorm.ResolveQualifiedType.
//
// sure 各生成器共用的代码生成助手。
// 收纳"接收者方法收集"与"名字唯一化"，让各生成器共用一份实现，而不是各写一份近乎一样的副本。
// 源包类型限定已上移到 syntaxgo_astnorm.ResolveQualifiedType，此处不再持有副本。
package utils

import (
	"go/build/constraint"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yylego/done"
	"github.com/yylego/erero"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
	"github.com/yylego/syntaxgo/syntaxgo_search"
)

// UniqueName returns base when it is free, otherwise base with the smallest
// numeric suffix that is not yet taken. Shared across the codegens to keep
// synthesized names (arg/res/err) from clashing with genuine input names in
// the generated wrapping.
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

// PickGoBuildDirective settles which build constraint a run writes, using fallbackDirective when the config said nothing on the topic
// directiveChosen states if the config made a deliberate choice, which a blank goBuildDirective cannot show on its own
//
// The blank string carries two intents at once, and splitting those apart is the whole point
// A config built as a bare struct value leaves the field blank because no one gave it a thought, and giving such a run no constraint takes the escape hatch out on accident
// A config that reached WithNoBuildDirective leaves it blank on purpose, and that choice has to stand
// Reading the first blank as "use the fallback" and the second as "write none" serves both, so each codegen defaults the same regardless of which shape built its config
//
// PickGoBuildDirective 定下本次生成要写哪条构建约束；配置压根没表态时用 fallbackDirective 顶上
// directiveChosen 表示配置是否明确表过态——光看 goBuildDirective 是空串，是分不出来的
//
// 空串这一个值，调用方可能怀着两种截然不同的意图，把它们分开正是本函数的全部意义
// 直接写结构体字面量的配置，该字段是空的，因为根本没人想过这件事；这种情况下不写约束，等于稀里糊涂把退路抽掉了
// 而调用过 WithNoBuildDirective 的配置，那个空是特意留的，这个答复必须算数
// 把"空且没表态"读成"用兜底值"、把"特意清空"读成"就是不写"，两边都照顾到了；各生成器也就不必管配置是用哪种写法造出来的，默认行为一致
func PickGoBuildDirective(goBuildDirective string, directiveChosen bool, fallbackDirective string) string {
	if goBuildDirective == "" && !directiveChosen {
		return fallbackDirective
	}
	return goBuildDirective
}

// FormatGoBuildDirective renders a build constraint line, blank when none is set
// Accepts both "!tag" and "//go:build !tag" spellings
// Stops the run when the result is no build constraint Go can read
//
// The line belongs above the package clause with a blank line beneath it, since
// that is the spot Go inspects when it wants a build constraint
//
// The check earns its place because Go stays silent about it and this codegen would too
// A line such as "//go:build //go:build!x" (what a value that holds its own prefix without the space renders into) reads as a plain comment to the parse step
// So the generated file formats and gets written, and go build then rejects the whole package
// Stopping at this point names the misconfigured directive instead
//
// FormatGoBuildDirective 渲染一行构建约束；没配置指令时返回空串
// "!tag" 和 "//go:build !tag" 两种写法都接受
// 当渲染结果不是 Go 能读懂的构建约束时把流程停住
//
// 这一行要放在 package 子句之前、后面空一行，因为那才是 Go 认构建约束的位置
//
// 这道检查值得加，因为不加的话 Go 和本生成器都不会吭声
// 像 "//go:build //go:build!x" 这种行（传进来的值自带前缀却少个空格时就会拼成这样），在解析器眼里只是条普通注释
// 于是生成文件照样格式化、照样落盘，然后 go build 把整个包否掉
// 就此停住，还能把配错的那个指令指名道姓地报出来
func FormatGoBuildDirective(goBuildDirective string) string {
	if goBuildDirective == "" {
		return ""
	}
	const goBuildPrefix = "//go:build "
	goBuildLine := goBuildDirective
	if !strings.HasPrefix(goBuildLine, goBuildPrefix) {
		goBuildLine = goBuildPrefix + goBuildLine
	}
	if !constraint.IsGoBuild(goBuildLine) {
		panic(erero.Errorf("cannot use %q as a go:build directive: %q is no build constraint line", goBuildDirective, goBuildLine))
	}
	if _, err := constraint.Parse(goBuildLine); err != nil {
		panic(erero.WithMessagef(err, "cannot use %q as a go:build directive", goBuildDirective))
	}
	return goBuildLine
}

// CollectReceiverMethods scans each .go file beneath root and gathers the methods
// whose receiving type is typeName, pairing each file's source with its matched
// method declarations and with that file's imports. The class codegen and the stub
// codegen share it, since each once carried a scan loop much the same as this.
//
// outputPath names the artifact this run is about to write, and the scan leaves that one file alone
// Pass a blank string where no artifact gets written, and see IsOutputFile on what goes wrong once the artifact reaches the scan
//
// CollectReceiverMethods 扫描 root 下的每个 .go 文件，收集接收者类型为 typeName 的方法，
// 把每个文件的源码、其命中的方法声明、以及该文件的导入配对在一起。由类生成器与桩生成器共用——
// 二者此前各自持有一份近乎一样的扫描循环。
//
// outputPath 是本次将要写入的那个产物文件，扫描会绕开它；不写产物时传空串即可
// 产物被扫进来会出什么事，详见 IsOutputFile
func CollectReceiverMethods(root string, typeName string, outputPath string) SourceMethodsTuples {
	tuples := make(SourceMethodsTuples, 0)
	for _, fileInfo := range done.VAE(os.ReadDir(root)).Done() {
		if fileInfo.IsDir() {
			continue
		}
		if filepath.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		path := filepath.Join(root, fileInfo.Name())
		// Ask before the read: a broken artifact would take down the run that exists to replace it, since the parse beneath asserts. See IsOutputFile
		// 在读之前就问：破损产物会把"本该来替换它"的这一趟直接搞挂——下面那步解析是带断言的。详见 IsOutputFile
		if IsOutputFile(path, outputPath) {
			continue
		}
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
