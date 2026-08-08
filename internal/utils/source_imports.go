// Import handling shared across the sure generators
// Takes the import block of each scanned file so transcribed signatures keep types resolvable
// Rejects generated code holding an import path that no scanned file and no registration states
//
// The generators transcribe method signatures as text, so a signature mentioning
// plumbing.Hash arrives without stating which plumbing that is. Nothing downstream
// can get it back: an import alias matches no package name, and a path such as
// github.com/go-git/go-git/v5 is named git, so the path does not state the name and
// the name does not state the path. Leaving that gap to an import step means leaving
// it to guesswork, and one wrong guess writes a file that cannot compile.
//
// The scanned file states the origin in its own import block, and it has to be right
// because that file compiles. So the generators take the block along with the
// signatures. Taking too much is safe: goimports drops what the generated code does
// not mention.
//
// sure 各生成器共用的导入处理
// 取走每个被扫描文件的导入块，让抄来的签名里类型仍有出处
// 拒绝那种"导入路径既非扫描到的源文件、也非调用方写明"的生成代码
//
// 生成器是把方法签名当文本抄出来的，于是签名里提到 plumbing.Hash 时，
// 并不附带"这个 plumbing 到底是谁"。下游谁也拿不回来：导入别名不对应任何包名，
// 而 github.com/go-git/go-git/v5 这种路径的包名是 git——路径说不出包名，包名也说不出路径。
// 把这个缺口留给导入环节，就等于留给猜测，一次猜错就写出一个编译不过的文件。
//
// 而被扫描文件自己的导入块里写着出处，且它必然正确——因为那个文件能编译。
// 所以生成器抄签名时，把那个块一并取走。取多了是安全的：
// 生成代码没提到的，goimports 会丢掉。
package utils

import (
	"go/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yylego/done"
	"github.com/yylego/erero"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

// SourceImport is one import taken from a scanned file
// Holds the path, plus the alias when the source states one
//
// SourceImport 是从被扫描文件里取到的一条导入
// 存着路径，以及源文件写了别名时的那个别名
type SourceImport struct {
	Alias string // Name stated in the source, blank when absent // 源文件写出的本地名，没写时为空
	Path  string // Import path without quotes // 不带引号的导入路径
}

// SourceImports is a set of imports taken from scanned files
// SourceImports 是从被扫描文件里取到的一组导入
type SourceImports []SourceImport

// CollectFileImports takes the imports of a parsed file
// Skips blank and dot imports, which must not reach generated code
//
// A blank import (_) exists to set off a package init, and goimports keeps it in
// place, so taking it would staple a side effect onto the generated file. A dot
// import (.) puts names in scope unqualified, so no signature text can point at it,
// and taking it would drag in a whole namespace.
//
// CollectFileImports 取出已解析文件的导入
// 跳过空白与点导入，这两类不该进入生成代码
//
// 空白导入（_）的存在是为了触发包的 init，而 goimports 会把它留着，
// 取走它等于给生成文件平白钉上一个副作用。点导入（.）把名字直接放进作用域，
// 签名文本里根本指不到它，取走它等于拖进整个命名空间。
func CollectFileImports(astFile *ast.File) SourceImports {
	sourceImports := make(SourceImports, 0, len(astFile.Imports))
	for _, importSpec := range astFile.Imports {
		if importSpec.Path == nil {
			continue
		}
		// The path comes from a file that parsed, so in principle unquoting it
		// cannot break; assert instead of skipping in silence
		// 路径来自已解析成功的文件，原则上解引号不会出问题；宁可断言也不静默跳过
		path, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			panic(erero.WithMessagef(err, "cannot unquote import path %s", importSpec.Path.Value))
		}
		if path == "" {
			continue
		}
		var alias string
		if importSpec.Name != nil {
			alias = importSpec.Name.Name
		}
		if alias == "_" || alias == "." {
			continue
		}
		sourceImports = append(sourceImports, SourceImport{Alias: alias, Path: path})
	}
	return sourceImports
}

// CollectPathImports reads a Go file and takes its imports
// Serves generators that walk files one at a time instead of gathering methods
//
// CollectPathImports 读取一个 Go 文件并取出它的导入
// 服务于那些逐个文件处理、而非按接收者收集方法的生成器
func CollectPathImports(absPath string) SourceImports {
	sourceCode := done.VAE(os.ReadFile(absPath)).Done()
	astBundle := done.VCE(syntaxgo_ast.NewAstBundleV1(sourceCode)).Nice()
	astFile, _ := astBundle.GetBundle()
	return CollectFileImports(astFile)
}

// collectSignatureQualifiers gathers the package qualifiers the signatures mention
// Reads signatures alone, generic constraints included, since method bodies do not
// get transcribed
//
// collectSignatureQualifiers 收集这些方法签名里出现过的包限定符
// 只读签名本身（含泛型约束），因为方法体不会被抄走
func collectSignatureQualifiers(methodList []*ast.FuncDecl) map[string]bool {
	qualifiers := make(map[string]bool)
	for _, methodFunc := range methodList {
		if methodFunc.Type == nil {
			continue
		}
		ast.Inspect(methodFunc.Type, func(node ast.Node) bool {
			if selectorNode, ok := node.(*ast.SelectorExpr); ok {
				if ident, ok := selectorNode.X.(*ast.Ident); ok {
					qualifiers[ident.Name] = true
				}
			}
			return true
		})
	}
	return qualifiers
}

// IsOutputFile tells if a scanned path is the artifact about to get written
//
// A codegen aims at the same place it writes into, so a past artifact shows up in its own scan.
// The scan has to leave it out on three counts, and the third explains the point of asking this ahead of each read:
//
// Imports. The artifact's import block would reach the next artifact, so a path that slipped in once would keep being handed down, and would look established to each check downstream.
//
// Methods. The artifact holds wrapping methods on the exact type being wrapped, so reading it back wraps the wrapping.
// Filtering those on name gets close, yet a name that a past run wrote and this run stopped stating slips through.
//
// Parsing. A broken artifact cannot get parsed, and the run that parses it dies before it can write the replacement — the exact deadlock the build constraint exists to break.
// Skipping the file ahead of the read keeps the escape hatch open.
//
// Naming the one output file leaves the output of protoc-gen-go, mockgen and such wide open to a scan, since those get wrapped on purpose.
//
// IsOutputFile 判断某个被扫描到的路径是不是"即将被写入的那个产物文件"
//
// 生成器所指的目录同时也是它写入的目录，于是上一轮的产物会出现在自己的扫描结果里。
// 扫描必须把它排除掉，理由有三条，而第三条正是"要在读文件之前就问这个问题"的原因：
//
// 导入：产物的导入块会进入下一份产物，于是某次混进去的路径会被一路传承下去，而且在后续每一道检查眼里都显得理所当然。
//
// 方法：产物里带着"针对被包装类型"的包装方法，读回来就等于给包装再包一层。
// 按方法名去滤能挡住大半，但"上一轮写过、这一轮不再声明"的那个名字会漏过去。
//
// 解析：破损产物根本解析不了，而解析它的那一趟会先崩掉、来不及写出替换版本——这正是构建约束要解开的那个僵局。
// 在读之前就跳过这个文件，才算真把退路留住。
//
// 点名那一个输出文件，同时让 protoc-gen-go、mockgen 之类的产物照常可扫——那些本来就是拿来包装的。
func IsOutputFile(scanPath string, outputPath string) bool {
	if outputPath == "" || scanPath == "" {
		return false
	}
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		panic(erero.WithMessagef(err, "cannot resolve output path %s", outputPath))
	}
	absScanPath, err := filepath.Abs(scanPath)
	if err != nil {
		panic(erero.WithMessagef(err, "cannot resolve scanned path %s", scanPath))
	}
	if absScanPath == absOutputPath {
		return true
	}
	// Past the plain path match, os.SameFile sees past a symlink, where two paths name one file
	// Stat can miss on each side, since the artifact has no need to exist on a first run
	// Nothing to compare then, so the path match above stands as the verdict
	// 路径字面量之外，os.SameFile 还能看穿软链——同一个文件被两条路径写出来
	// 两边的 Stat 都可能拿不到：首次生成时产物本来就还不存在
	// 那就无从比较，上面那次路径比对的结果即为答案
	outputInfo, err := os.Stat(absOutputPath)
	if err != nil {
		return false
	}
	scanInfo, err := os.Stat(absScanPath)
	if err != nil {
		return false
	}
	return os.SameFile(outputInfo, scanInfo)
}

// MergeSourceImports gathers the imports of each scanned file into one set
// Keeps an aliased import when its alias shows up in the transcribed signatures
// Collapses exact duplicates and stops the run when one alias names two paths
//
// The alias screen comes first because an alias is a plain name, so a glance tells if a signature uses it
// An unused alias adds dead weight, and would let two files that gave one alias to two paths halt a run that mentions no such import
// Past the screen, a clash means two aliases that both do reach a signature and in fact cannot coexist, and halting is then the right move
//
// A plain import gives nothing to judge on, since a path does not state its package name, so each one gets taken and goimports prunes it
// Two clashes remain unseen here: two plain imports whose packages share a name, and an alias landing on the default name of a second import
// Both show up once the build names the symbols that collide, and both remain out of reach of MustKnownImports, since each path involved is one a scanned source states
// Where the artifact carries a build constraint (sure_cls_gen, cls_stub_gen) the package compiles even so, and a fresh run replaces the file
// sure_pkg_gen writes standalone packages and states no constraint, so there the fix is to give one of the two sources an alias
//
// MergeSourceImports 把每个被扫描文件的导入汇成一组
// 带别名的导入，只在别名确实出现在抄走的签名里时才留下
// 合并完全相同的项；当同一别名指向两条路径时把流程停住
//
// 别名这道筛先做，因为别名就是个普通名字，签名里用没用一眼可知
// 没用到的别名是累赘，还会让"两个文件把同一别名给了不同路径"这种与本次生成毫不相干的情况白白把流程拦下来
// 过了这道筛还撞上的，就是两个都真的进了签名、实际上没法共存的别名，那时停住才是对的
//
// 没写别名的导入无从判断，因为路径并不写明自己的包名，于是一律取走、交给 goimports 剪
// 有两类冲突在这儿看不见：两条没写别名而包名相同的；以及一条别名恰好撞上第二条导入的默认包名
// 这两类都会在编译点名撞车的符号时显现，而 MustKnownImports 也拦不住它们——牵涉的每条路径都是某个被扫描源文件写明过的
// 产物带着构建约束时（sure_cls_gen、cls_stub_gen），包照样能编译，重新跑一遍就把文件换掉了
// sure_pkg_gen 生成的是独立新包、不写约束，那边的解法是给两个源文件之一加上别名
func (mts SourceMethodsTuples) MergeSourceImports() SourceImports {
	merged := make(SourceImports, 0)
	seenPairs := make(map[SourceImport]bool)
	aliasToPath := make(map[string]string)
	for _, tuple := range mts {
		qualifiers := collectSignatureQualifiers(tuple.MethodList)
		for _, sourceImport := range tuple.SourceImports {
			if seenPairs[sourceImport] {
				continue
			}
			if sourceImport.Alias != "" && !qualifiers[sourceImport.Alias] {
				continue
			}
			if sourceImport.Alias != "" {
				if takenPath, exists := aliasToPath[sourceImport.Alias]; exists && takenPath != sourceImport.Path {
					panic(erero.Errorf(
						"cannot take source imports: alias %q names two paths %q and %q in the scanned sources",
						sourceImport.Alias, takenPath, sourceImport.Path))
				}
				aliasToPath[sourceImport.Alias] = sourceImport.Path
			}
			seenPairs[sourceImport] = true
			merged = append(merged, sourceImport)
		}
	}
	return merged
}

// FormatImportBlock renders the imports as an import declaration
// Renders nothing when the set is blank
//
// This block sits beside the one the registration injects, since Go accepts more
// than one import declaration in a file, so each side stands on its own
//
// FormatImportBlock 把这些导入渲染成一段导入声明
// 集合为空时什么都不渲染
//
// 这一段与调用方登记注入的那一段并存，因为 Go 允许一个文件有多段导入声明，
// 于是两边各自独立
func (sourceImports SourceImports) FormatImportBlock() string {
	if len(sourceImports) == 0 {
		return ""
	}
	ptx := NewPTX()
	ptx.Println("import (")
	for _, sourceImport := range sourceImports {
		if sourceImport.Alias != "" {
			ptx.Println(sourceImport.Alias + " " + strconv.Quote(sourceImport.Path))
		} else {
			ptx.Println(strconv.Quote(sourceImport.Path))
		}
	}
	ptx.Println(")")
	return ptx.String()
}

// Paths returns the import paths
// Paths 返回这些导入的路径
func (sourceImports SourceImports) Paths() []string {
	paths := make([]string, 0, len(sourceImports))
	for _, sourceImport := range sourceImports {
		paths = append(paths, sourceImport.Path)
	}
	return paths
}

// MustKnownImports stops the run when finished code holds an unstated import path
// Such a path came from guesswork while resolving imports, and must not reach disk
//
// Refusing to write beats writing a file that cannot compile: the codegen test
// tends to live in the package it writes into, so a broken artifact takes down the
// one test that could replace it
//
// MustKnownImports 在成品里出现"谁也没写过的导入路径"时把流程停住
// 这种路径来自解析导入时的猜测，不该落到磁盘上
//
// 宁可不写，也别写出一个编译不过的文件：生成器测试往往就住在它要写入的那个包里，
// 破损产物会把唯一能替换它的那个测试一起搞挂
func MustKnownImports(sourceCode []byte, knownPaths []string) {
	knownSet := make(map[string]bool, len(knownPaths))
	for _, knownPath := range knownPaths {
		knownSet[knownPath] = true
	}

	astBundle, err := syntaxgo_ast.NewAstBundleV1(sourceCode)
	if err != nil {
		panic(erero.WithMessage(err, "cannot parse generated code to check its imports"))
	}
	astFile, _ := astBundle.GetBundle()

	var unknownPaths []string
	for _, importSpec := range astFile.Imports {
		if importSpec.Path == nil {
			continue
		}
		// Assert on a broken unquote: a path no one can read must not slip past this
		// check, since such a path is the exact thing this function exists to stop
		// 解引号出问题就断言：一条谁都读不出来的路径不能溜过这道检查，
		// 那正是本函数存在的目的所要拦下的东西
		path, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			panic(erero.WithMessagef(err, "cannot unquote import path %s in generated code", importSpec.Path.Value))
		}
		if knownSet[path] {
			continue
		}
		unknownPaths = append(unknownPaths, path)
	}
	if len(unknownPaths) > 0 {
		panic(erero.Errorf(
			"refusing to write generated code: import path(s) %s came from guesswork, since no scanned source and no registration states them; state each one via ImportOptions.AddPkgPath / ImportOptions.AddInferredObject",
			strings.Join(unknownPaths, ", ")))
	}
}
