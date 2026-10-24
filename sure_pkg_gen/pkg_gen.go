// Package sure_pkg_gen generates Go packages with integrated exception handling from existing code
// Auto extracts functions from source files and creates new packages with Must and Soft and Omit modes
// Auto scans AST to transform functions that return exceptions into assertion-wrapped versions
// Supports generic types, variadic parameters, and custom exception handling configurations
//
// Package sure_pkg_gen 从现有代码生成集成异常处理的 Go 包
// 自动从源文件提取函数并创建带有 Must 和 Soft 和 Omit 模式的新包
// 自动扫描 AST 将返回异常的函数转换为带断言包装的版本
// 支持泛型类型、可变参数和自定义异常处理配置
package sure_pkg_gen

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/yylego/done"
	"github.com/yylego/formatgo"
	"github.com/yylego/rese"
	"github.com/yylego/sure"
	"github.com/yylego/sure/internal/utils"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
	"github.com/yylego/syntaxgo/syntaxgo_astnode"
	"github.com/yylego/syntaxgo/syntaxgo_astnorm"
	"github.com/yylego/syntaxgo/syntaxgo_search"
	"github.com/yylego/tern"
	"github.com/yylego/tern/zerotern"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// GenerateSurePackage creates a sure package with specified exception handling mode
// Auto initializes config and generates package files with assertion wrapping
// Makes package generation simple with default configuration
//
// GenerateSurePackage 使用指定的异常处理模式创建 sure 包
// 自动初始化配置并生成带有断言包装的包文件
// 使用默认配置使包生成简单
func GenerateSurePackage(
	sourceRoot string,
	errorHandlingMode sure.ErrorHandlingMode,
	sourcePackagePath string,
) {
	cfg := NewSurePackageConfig(
		sourceRoot,
		errorHandlingMode,
		sourcePackagePath,
	)
	// 在此入口处捕获调用者位置（skip=1 指向本函数的调用方），
	// 免得下游在更深的栈帧里去猜 skip 数、直连调用时归因错帧。
	cfg.genPosFuncMark = utils.GetGenPosFuncMark(1)
	GenerateSurePackageFiles(cfg)
}

// GenerateSurePackageFiles generates package files from configuration
// Scans source root and creates wrapped functions with exception handling
// Writes generated packages to configured output paths
//
// GenerateSurePackageFiles 从配置生成包文件
// 扫描源根目录并创建带有异常处理的包装函数
// 将生成的包写入配置的输出路径
func GenerateSurePackageFiles(cfg *SurePackageGenConfig) {
	zaplog.LOG.Debug("Starting package generation from source root", zap.String("root", cfg.SourceRoot))

	// 位置追踪：若上游入口已捕获则沿用，否则在此直连入口处捕获（skip=1 指向直接调用方）。
	// 不能放进闭包里取：那样 skip=1 指向的是调用闭包的 zerotern.VF，而不是用户。
	genPosFuncMark := cfg.genPosFuncMark
	if genPosFuncMark == "" {
		genPosFuncMark = utils.GetGenPosFuncMark(1)
	}

	// Build and check each file first, then write them in one pass. Writing as each
	// one gets built would leave the preceding files on disk once a following one
	// gets stopped, so the package on disk ends up half written
	// 先全部生成、全部校验，再统一写盘。逐个文件边生成边写的话，
	// 后面某个文件被拦住时，前面的已经落在磁盘上了，留下一个写了一半的包
	type generatedFile struct {
		path string
		code string
	}
	var generatedFiles []generatedFile

	for _, item := range done.VAE(os.ReadDir(cfg.SourceRoot)).Done() {
		// A DIR name can end in .go too, and reading it as a file stops the run
		// 目录名也可能以 .go 结尾，把它当文件读会让流程崩掉
		if item.IsDir() {
			continue
		}
		name := item.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		absPath := filepath.Join(cfg.SourceRoot, name)
		zaplog.LOG.Debug("Reading source file to extract functions", zap.String("path", absPath))

		sureFunctionCodes := GenerateSureFunctions(cfg, absPath)

		if len(sureFunctionCodes) > 0 {
			shortErrorHandlingMode := strings.ToLower(string(cfg.ErrorHandlingMode))

			oldPackageName := syntaxgo.GetPkgName(absPath)

			newPackageName := zerotern.VV(cfg.NewPkgName, oldPackageName+"_"+shortErrorHandlingMode)

			zaplog.LOG.Debug("Generating package with exception handling",
				zap.String("package", newPackageName),
				zap.String("mode", shortErrorHandlingMode),
				zap.Int("functions", len(sureFunctionCodes)))

			ptx := utils.NewPTX()
			ptx.Println("// Code generated using sure/sure_pkg_gen. DO NOT EDIT.")
			ptx.Println("// This file was auto generated via github.com/yylego/sure")
			// Blank unless calling code asked to have one, since a standalone package cannot lock. See the field on SurePackageGenConfig
			// 除非调用方主动要，否则这里是空的——独立新包卡不住自己。详见 SurePackageGenConfig 上那个字段
			if goBuildLine := utils.FormatGoBuildDirective(cfg.GoBuildDirective); goBuildLine != "" {
				ptx.Println()
				ptx.Println(goBuildLine)
				ptx.Println()
			}
			ptx.Println("// Generated from:", genPosFuncMark)
			ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
			ptx.Println()
			ptx.Println("package" + " " + newPackageName)
			ptx.Println("import(")
			ptx.Println(strconv.Quote(cfg.SourcePackagePath))
			ptx.Println(strconv.Quote(cfg.ErrorHandlingPkgPath))
			ptx.Println(")")
			// Take the source file's imports: signatures get transcribed as text,
			// so a type inside (such as plumbing.Hash) needs a stated origin, else
			// downstream code has to guess. This block sits beside the one above,
			// and goimports prunes anything unused
			// 取走源文件的导入：签名是当文本抄走的，抄来的类型（如 plumbing.Hash）得留着出处，
			// 否则下游只能靠猜。这一段与上面那段并存，用不上的由 goimports 剪掉
			sourceImports := utils.CollectPathImports(absPath)
			if importBlock := sourceImports.FormatImportBlock(); importBlock != "" {
				ptx.Println(importBlock)
			}
			ptx.Println(strings.Join(sureFunctionCodes, "\n"))

			// 用 TrimSuffix 精确去掉 .go 扩展名再拼接，避免 strings.Replace 命中文件名中间
			// 的 ".go"（如 go.go）而替错位置。
			newName := strings.TrimSuffix(name, ".go") + "_" + shortErrorHandlingMode + ".go"
			newPath := filepath.Join(cfg.OutputRoot, newPackageName, newName)

			// 断言格式化成功：语法/格式层面若出错会响亮报错，而不是静默写出破损文件。
			// 注意这里只做语法与 goimports 检查，抓不到类型层错误（未定义标识符等仍会写盘，留待编译期暴露）。
			newCode := rese.V1(formatgo.FormatCode(ptx.String()))

			// An import in the output with no source file and no statement from
			// this codegen behind it has to be guesswork, so stop and write nothing
			// 成品里若冒出既非源文件、也非本生成器写明的导入，那只能是猜出来的，就此打住不写盘
			utils.MustKnownImports([]byte(newCode), slices.Concat(
				sourceImports.Paths(),
				[]string{cfg.SourcePackagePath, cfg.ErrorHandlingPkgPath}))

			generatedFiles = append(generatedFiles, generatedFile{path: newPath, code: newCode})
		}
	}

	// Each one passed the check, so writing them out is safe
	// 到这儿说明每一份都过了校验，可以放心落盘
	for _, generated := range generatedFiles {
		utils.MustWriteIntoPath(generated.path, generated.code)
		zaplog.LOG.Info("Package file generation complete", zap.String("output", generated.path))
	}
}

// GenerateSureFunctions extracts functions from source file and wraps them
// Parses AST to find exported functions and generates assertion-wrapped versions
// Returns slice of generated function code strings
//
// GenerateSureFunctions 从源文件提取函数并包装它们
// 解析 AST 以查找导出的函数并生成带断言包装的版本
// 返回生成的函数代码字符串切片
func GenerateSureFunctions(cfg *SurePackageGenConfig, absPath string) []string {
	sourceCode := done.VAE(os.ReadFile(absPath)).Done()
	astBundle := rese.P1(syntaxgo_ast.NewAstBundleV1(sourceCode))

	packageName := astBundle.GetPackageName()
	astFile, _ := astBundle.GetBundle()

	astFunctions := syntaxgo_search.ExtractFunctions(astFile)
	zaplog.LOG.Debug("Extracting functions from source",
		zap.String("package", packageName),
		zap.String("path", absPath),
		zap.Int("count", len(astFunctions)))

	var functionCodes []string
	for _, astFunc := range astFunctions {
		if astFunc.Recv != nil { // methods are not wrapped at package scope // 方法不在包作用域包装
			continue
		}
		if !token.IsExported(astFunc.Name.Name) { // exported functions alone // 仅处理导出函数
			continue
		}
		results := parseFunctionReturnFields(sourceCode, astFunc)
		functionCode := makeFunctionCode(sourceCode, packageName, astFunc, results, cfg.ErrorHandlingMode, cfg.HandlerFuncReference)
		functionCodes = append(functionCodes, functionCode)
	}
	zaplog.LOG.Debug("Generated wrapping functions",
		zap.String("path", absPath),
		zap.Int("generated", len(functionCodes)))
	return functionCodes
}

// makeFunctionCode constructs wrapped function code with exception handling
// Builds function signature and implementation with assertion calls
// Returns complete function code as a string
//
// makeFunctionCode 构建带有异常处理的包装函数代码
// 构建函数签名和带有断言调用的实现
// 返回完整的函数代码作为字符串
func makeFunctionCode(
	sourceCode []byte,
	packageName string,
	astFunc *ast.FuncDecl,
	results []*resultType,
	errorHandlingMode sure.ErrorHandlingMode,
	sureHandlerFuncReference string,
) string {
	genericTypeParams := syntaxgo_astnorm.GetGenericTypeParamsMap(astFunc.Type.TypeParams)

	var newFunctionCode strings.Builder
	newFunctionCode.WriteString("func " + astFunc.Name.Name)
	if astFunc.Type.TypeParams != nil {
		// 类型参数列表：约束若引用源包的导出类型（如 [T Numeric]）也要加包前缀，
		// 不能像旧写法那样整段原样拷贝，否则生成包里 Numeric 未定义。
		var typeParamTexts []string
		for _, field := range astFunc.Type.TypeParams.List {
			names := make([]string, 0, len(field.Names))
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			constraintText := syntaxgo_astnorm.ResolveQualifiedType(field.Type, packageName, genericTypeParams, sourceCode)
			typeParamTexts = append(typeParamTexts, strings.Join(names, ",")+" "+constraintText)
		}
		newFunctionCode.WriteString("[" + strings.Join(typeParamTexts, ",") + "]")
	}

	// 收集会出现在包装签名里的形参名，供后面给合成名（arg/res/err）避让，防止 := 撞车。
	reservedNames := map[string]bool{}
	if astFunc.Type.Params != nil {
		for _, param := range astFunc.Type.Params.List {
			for _, name := range param.Names {
				reservedNames[name.Name] = true
			}
		}
	}

	newFunctionCode.WriteString("(")
	var argumentNames []string
	if astFunc.Type.Params != nil && len(astFunc.Type.Params.List) > 0 {
		var args []string
		for _, param := range astFunc.Type.Params.List {
			argType := syntaxgo_astnorm.ResolveQualifiedType(param.Type, packageName, genericTypeParams, sourceCode)

			// 可变参数：签名侧类型已是 "...T"（argType），调用侧要展开成 name...
			_, variadic := param.Type.(*ast.Ellipsis)

			if len(param.Names) == 0 {
				argName := utils.UniqueName("arg"+strconv.Itoa(len(args)), reservedNames)
				reservedNames[argName] = true

				args = append(args, argName+" "+argType)
				argumentNames = append(argumentNames, invokeArgName(argName, variadic))
			} else {
				sigNames := make([]string, 0, len(param.Names))

				for _, name := range param.Names {
					// The wrapping passes each input on, and a blank name "_" cannot get read back, so it gets a name of its own
					// 包装要把每个参数传下去，而空白名 "_" 读不回来，所以给它起个名字
					argName := name.Name
					if argName == "_" {
						argName = utils.UniqueName("arg"+strconv.Itoa(len(argumentNames)), reservedNames)
						reservedNames[argName] = true
					}
					sigNames = append(sigNames, argName)
					argumentNames = append(argumentNames, invokeArgName(argName, variadic))
				}

				args = append(args, strings.Join(sigNames, ",")+" "+argType)
			}
		}
		newFunctionCode.WriteString(strings.Join(args, ","))
	}
	newFunctionCode.WriteString(")")

	// 给合成的返回值名（err/res0…）避让形参名：源码 func F(err int) error 若直接
	// 生成 err:=F(err)，会因 := 左侧无新变量而编译失败。仅重命名撞车项，不影响常规输出。
	for _, res := range results {
		if reservedNames[res.Name] {
			res.Name = utils.UniqueName(res.Name, reservedNames)
		}
		reservedNames[res.Name] = true
	}

	var anonymousReturn = false
	for _, res := range results {
		if res.IsAnonymous {
			anonymousReturn = true
			break
		}
	}

	var containsErrorReturn = false
	if len(results) > 0 {
		var resultNames = make([]string, 0, len(results))
		for _, res := range results {
			if res.Type == "error" {
				containsErrorReturn = true
				continue
			}
			if res.IsAnonymous {
				resultNames = append(resultNames, syntaxgo_astnorm.ResolveQualifiedType(res.TypeNode, packageName, genericTypeParams, sourceCode))
			} else {
				resultNames = append(resultNames, res.Name+" "+syntaxgo_astnorm.ResolveQualifiedType(res.TypeNode, packageName, genericTypeParams, sourceCode))
			}
		}
		if len(resultNames) > 0 {
			if len(resultNames) < 2 && anonymousReturn {
				newFunctionCode.WriteString(" " + resultNames[0])
			} else {
				newFunctionCode.WriteString(" " + "(" + strings.Join(resultNames, ",") + ")")
			}
		}
	}
	newFunctionCode.WriteString(" " + "{" + "\n")
	if len(results) > 0 {
		var rets = make([]string, 0, len(results))
		for _, res := range results {
			rets = append(rets, res.Name)
		}
		newFunctionCode.WriteString(strings.Join(rets, ","))
		if anonymousReturn || containsErrorReturn {
			newFunctionCode.WriteString(":=")
		} else {
			newFunctionCode.WriteString("=")
		}
	}
	newFunctionCode.WriteString(packageName + "." + astFunc.Name.Name)
	if astFunc.Type.TypeParams != nil {
		newFunctionCode.WriteString("[")
		var genericTypeNames []string
		for _, xts := range astFunc.Type.TypeParams.List {
			for _, xtx := range xts.Names {
				genericTypeNames = append(genericTypeNames, xtx.Name)
			}
		}
		newFunctionCode.WriteString(strings.Join(genericTypeNames, ","))
		newFunctionCode.WriteString("]")
	}
	newFunctionCode.WriteString("(" + strings.Join(argumentNames, ",") + ")" + "\n")
	if len(results) > 0 {
		for _, x := range results {
			if x.Type == "error" {
				newFunctionCode.WriteString(sureHandlerFuncReference + "." + string(errorHandlingMode) + "(" + x.Name + ")" + "\n")
			}
		}
	}
	if len(results) > 0 {
		var rets = make([]string, 0, len(results))
		for _, res := range results {
			if res.Type == "error" {
				continue
			}
			rets = append(rets, res.Name)
		}
		if len(rets) > 0 {
			newFunctionCode.WriteString("return" + " " + strings.Join(rets, ",") + "\n")
		}
	}
	newFunctionCode.WriteString("}" + "\n")
	return newFunctionCode.String()
}

// invokeArgName renders an input name at the invocation site, spreading variadic args with "..."
//
// invokeArgName 渲染调用侧的参数名，可变参数用 "..." 展开
func invokeArgName(name string, variadic bool) string {
	if variadic {
		return name + "..."
	}
	return name
}

// resultType represents a function return value with metadata
// Contains name, type info and anonymousness flag
// Used during function code generation
//
// resultType 表示带有元数据的函数返回值
// 包含名称、类型信息以及匿名标志
// 在函数代码生成期间使用
type resultType struct {
	Name        string   // Return value name // 返回值名称
	Type        string   // Return value type text // 返回值类型文本
	TypeNode    ast.Expr `json:"-"` // Return value type node (skipped in JSON: ast nodes can hold cyclic objects) // 返回值类型节点（JSON 跳过：ast 节点可能含循环的 object）
	IsAnonymous bool     // True when return value is anonymous // 当返回值为匿名时为真
}

// parseFunctionReturnFields extracts return value information from function
// Analyzes AST to find return types and generates suitable names
// Returns slice of resultType with complete metadata
//
// parseFunctionReturnFields 从函数提取返回值信息
// 分析 AST 以查找返回类型并生成合适的名称
// 返回带有完整元数据的 resultType 切片
func parseFunctionReturnFields(source []byte, astFunction *ast.FuncDecl) []*resultType {
	var results []*resultType
	if astFunction.Type.Results == nil || len(astFunction.Type.Results.List) == 0 {
		results = make([]*resultType, 0)
	} else {
		var errNum int
		for _, elem := range astFunction.Type.Results.List {
			resType := syntaxgo_astnode.GetText(source, elem.Type)
			isError := resType == "error"
			synthesizeName := func() string {
				return tern.BFF(isError, func() string {
					return tern.BVV(errNum == 0, "err", "err"+strconv.Itoa(errNum))
				}, func() string {
					return "res" + strconv.Itoa(len(results))
				})
			}
			if len(elem.Names) == 0 {
				results = append(results, &resultType{
					Name:        synthesizeName(),
					Type:        resType,
					TypeNode:    elem.Type,
					IsAnonymous: true,
				})
				if isError {
					errNum++
				}
			} else {
				for _, name := range elem.Names {
					// A blank name "_" cannot be read back, so it gets a name as an anonymous result does
					// 空白名 "_" 读不回来，所以像匿名返回值那样给它起个名字
					resName := name.Name
					if resName == "_" {
						resName = synthesizeName()
					}
					results = append(results, &resultType{
						Name:        resName,
						Type:        resType,
						TypeNode:    elem.Type,
						IsAnonymous: false,
					})
					if isError {
						errNum++
					}
				}
			}
		}
	}
	return results
}
