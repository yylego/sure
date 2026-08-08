// Package sure_cls_gen generates Go classes with embedded exception handling assertions
// Creates wrapping classes from existing objects with Must and Soft exception handling modes
// Auto scans source code using AST analysis and generates assertion-wrapped methods
// Supports custom exception handling functions and configurable method context names
//
// Package sure_cls_gen 生成带有嵌入异常处理断言的 Go 类
// 从现有对象创建带有 Must 和 Soft 异常处理模式的包装类
// 自动扫描源代码使用 AST 分析并生成带断言包装的方法
// 支持自定义异常处理函数和可配置的方法上下文名称
package sure_cls_gen

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/yylego/done"
	"github.com/yylego/formatgo"
	"github.com/yylego/must"
	"github.com/yylego/sure"
	"github.com/yylego/sure/internal/utils"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
	"github.com/yylego/syntaxgo/syntaxgo_astnode"
	"github.com/yylego/syntaxgo/syntaxgo_astnorm"
	"github.com/yylego/syntaxgo/syntaxgo_reflect"
	"github.com/yylego/tern/zerotern"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// ClassGenConfig contains the configuration of class generation
// Specifies generation options, package name, imports and output path
// Combines settings needed to generate assertion-wrapped classes
//
// ClassGenConfig 包含类生成的配置
// 指定生成选项、包名、导入和输出路径
// 组合生成带断言包装类所需的设置
type ClassGenConfig struct {
	ClassGenOptions *ClassGenOptions                   // Generation options // 生成选项
	PackageName     string                             // Target package name // 目标包名
	ImportOptions   *syntaxgo_ast.PackageImportOptions // Import configuration // 导入配置
	OutputPath      string                             // Output file path // 输出文件路径
}

// GenerateClasses creates assertion-wrapped classes from given objects
// Auto generates code with exception handling based on configuration settings
// Writes formatted code to the output path specified in config
//
// GenerateClasses 从给定对象创建带断言包装的类
// 根据配置设置自动生成带有异常处理的代码
// 将格式化的代码写入配置中指定的输出路径
func GenerateClasses(cfg *ClassGenConfig, objects ...any) {
	must.Nice(cfg.ClassGenOptions.SourceRootPath)
	must.Nice(cfg.PackageName)
	must.Nice(cfg.OutputPath)
	utils.MustRoot(cfg.ClassGenOptions.SourceRootPath)

	// The scan has to leave out the artifact about to get overwritten, and the path to it lives on the enclosing config while the scan runs off ClassGenOptions
	// State it here so the scan down in GenerateClassWithErrorHandlingMode reaches it. See utils.IsOutputFile
	// 扫描要绕开即将被覆写的那个产物，而它的路径在外层配置上、扫描却是照 ClassGenOptions 走的
	// 在这里写上，好让 GenerateClassWithErrorHandlingMode 里那趟扫描够得着。详见 utils.IsOutputFile
	cfg.ClassGenOptions.outputPath = cfg.OutputPath

	// One scan up front gathers the imports of each source file: signatures get
	// transcribed as text, so the types inside need a stated origin. This repeats
	// the scan inside GenerateClassWithErrorHandlingMode, and buys leaving that
	// exported signature untouched, since downstream code calls it and a changed
	// signature breaks such code. Codegen runs at development time, so one extra
	// pass across the source tree costs nothing worth counting
	// 先扫一遍把各源文件的导入收齐：方法签名是当文本抄走的，抄来的类型得留着出处。
	// 这一趟与 GenerateClassWithErrorHandlingMode 内部那趟重复，换来的是不动它的导出签名——
	// 下游直接调用它，改签名会把下游搞挂。生成器只在开发期跑，多读一遍目录不值一提
	sourceMethodsTuples := make(utils.SourceMethodsTuples, 0)
	for _, object := range objects {
		typeName := reflect.TypeOf(object).Name()
		tuples := utils.CollectReceiverMethods(cfg.ClassGenOptions.SourceRootPath, typeName, cfg.OutputPath)
		if len(tuples) == 0 {
			// A wrapping class with no methods in it is almost sure a mistake, so
			// speak up instead of writing such a class in silence
			// 一个方法都没有的包装类几乎肯定是哪里搞错了，出声说明，而不是默默写出这么个类
			zaplog.LOG.Warn("No methods with this recv type in the source root",
				zap.String("type", typeName),
				zap.String("root", cfg.ClassGenOptions.SourceRootPath))
		}
		sourceMethodsTuples = append(sourceMethodsTuples, tuples...)
	}
	sourceImports := sourceMethodsTuples.MergeSourceImports()

	ptx := utils.NewPTX()
	ptx.Println("// Code generated using sure/sure_cls_gen. DO NOT EDIT.")
	ptx.Println("// This file was auto generated via github.com/yylego/sure")
	// The generated file and the test that makes it share one package, and that
	// shape can lock, which a build constraint undoes. Set as the default, dropped
	// with WithNoBuildDirective. See sure.DirectiveSkipGenerate
	// 生成文件与生成它的测试同住一个包，这个格局会卡住；构建约束就是那把钥匙。
	// 默认带上，可用 WithNoBuildDirective 去掉。详见 sure.DirectiveSkipGenerate
	goBuildDirective := utils.PickGoBuildDirective(cfg.ClassGenOptions.GoBuildDirective, cfg.ClassGenOptions.directiveChosen, sure.DirectiveSkipGenerate)
	if goBuildLine := utils.FormatGoBuildDirective(goBuildDirective); goBuildLine != "" {
		ptx.Println()
		ptx.Println(goBuildLine)
		ptx.Println()
	}
	ptx.Println("// Generated from:", utils.GetGenPosFuncMark(1))
	ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
	ptx.Println()
	ptx.Println("package", cfg.PackageName)
	// The taken imports form a block beside the one injected beneath, since Go
	// accepts more than one import declaration in a file. Anything unused gets
	// pruned once goimports runs, so taking too much is safe
	// 取来的导入自成一段，与下面注入的那段并存，因为 Go 允许一个文件有多段导入声明。
	// 用不上的会被 goimports 剪掉，所以取多了无妨
	if importBlock := sourceImports.FormatImportBlock(); importBlock != "" {
		ptx.Println(importBlock)
	}

	for _, object := range objects {
		ptx.Println(GenerateClassCodeWithErrorHandling(cfg.ClassGenOptions, object))
	}

	// Read what callers registered into a fresh set instead of writing into it, since the lines beneath add paths
	// A config handed to more than one run would otherwise keep on collecting those, which widens the set MustKnownImports treats as stated in silence — the guard against a guessed import goes slack
	// GetPkgPaths flattens registered types and objects into paths beforehand
	// 把调用方的登记读进一份新的集合，而不是往它里面写，因为下面几行会往里加路径
	// 同一份配置跑第二趟就会一直累积，等于把 MustKnownImports 认定的"已写明"名单悄悄放宽——那道拦截"猜出来的导入"的守卫就松了
	// GetPkgPaths 已把登记的类型与对象摊平成路径
	importOptions := syntaxgo_ast.NewPackageImportOptions()
	if cfg.ImportOptions != nil {
		importOptions.AddPkgPaths(cfg.ImportOptions.GetPkgPaths())
	}
	if cfg.ClassGenOptions.ErrorHandlerFuncName != "" {
		zaplog.LOG.Debug("Using custom exception handling function", zap.String("func", cfg.ClassGenOptions.ErrorHandlerFuncName))
	} else {
		// Use default Must and Soft functions, add package reference to help code formatting
		// 使用默认的 Must 和 Soft 函数，补充包引用以帮助代码格式化
		importOptions.AddInferredObject(syntaxgo_reflect.GetObject[sure.ErrorHandlingMode]())
	}

	// Inject import statements into code
	// 将导入语句注入到代码中
	source := importOptions.InjectImports(ptx.Bytes())

	// Measure code formatting time
	// 测量代码格式化时间
	sinceTime := time.Now()

	// Format the code, which runs goimports too: it prunes each import the code does not mention
	// It can also add one it thinks the code wants — that guessed add is what MustKnownImports beneath catches
	// 格式化代码，这一步会连 goimports 一起跑：代码里没提到的导入会被剪掉
	// 它也可能把自认为该有的"猜着加"进来——下面的 MustKnownImports 拦的正是这种猜测
	newSource := done.VAE(formatgo.FormatBytes(source)).Nice()

	// An import in the output with no source file and no registration behind it
	// has to be guesswork, so stop at this point and write nothing
	// 成品里若冒出既非源文件、也非调用方登记的导入，那只能是猜出来的，就此打住不写盘
	utils.MustKnownImports(newSource, slices.Concat(sourceImports.Paths(), importOptions.GetPkgPaths()))

	// Write formatted code to output path
	// 将格式化的代码写入输出路径
	duration := time.Since(sinceTime)
	zaplog.LOG.Debug("Source code formatting completed", zap.Duration("cost", duration), zap.Int("bytes", len(newSource)))
	if cfg.OutputPath != "" {
		done.Done(os.WriteFile(cfg.OutputPath, newSource, 0644))
		zaplog.LOG.Info("Classes generation complete", zap.String("output", cfg.OutputPath), zap.Int("objects", len(objects)))
	} else {
		fmt.Println(string(newSource))
		zaplog.LOG.Debug("Generated code output to stdout", zap.Int("bytes", len(newSource)))
	}
}

// GenerateClassCodeWithErrorHandling generates class code with multiple exception handling modes
// Creates wrapping code foreach configured exception handling mode
// Returns combined generated code as a string
//
// GenerateClassCodeWithErrorHandling 使用多个异常处理模式生成类代码
// 为每个配置的异常处理模式创建包装代码
// 返回组合的生成代码作为字符串
func GenerateClassCodeWithErrorHandling(cfg *ClassGenOptions, object any) string {
	ptx := utils.NewPTX()
	for _, errorHandlingMode := range cfg.GetErrorHandlingModes() {
		ptx.Println(GenerateClassWithErrorHandlingMode(cfg, object, errorHandlingMode))
	}
	return ptx.String()
}

// GenerateClassWithErrorHandlingMode generates a wrapping class with specific exception handling mode
// Scans source code to extract methods and creates assertion-wrapped versions
// Returns generated code as a string with correct formatting
//
// GenerateClassWithErrorHandlingMode 使用特定异常处理模式生成包装类
// 扫描源代码提取方法并创建带断言包装的版本
// 返回格式正确的生成代码字符串
func GenerateClassWithErrorHandlingMode(cfg *ClassGenOptions, object any, errorHandlingMode sure.ErrorHandlingMode) string {
	objectType := reflect.TypeOf(object)
	zaplog.LOG.Debug("Analyzing object type information",
		zap.String("name", must.Nice(objectType.Name())),
		zap.String("type", must.Nice(objectType.String())),
		zap.String("package", must.Nice(objectType.PkgPath())))

	utils.MustRoot(cfg.SourceRootPath)

	if len(cfg.ErrorHandlingModes) == 0 {
		// When not specified, just use default Must and Soft modes
		// 当未指定时，仅使用默认的 Must 和 Soft 模式
		must.True(errorHandlingMode == sure.MUST || errorHandlingMode == sure.SOFT)
	}

	sourceMethodsTuples := utils.CollectReceiverMethods(cfg.SourceRootPath, objectType.Name(), cfg.outputPath)

	// 接收者变量名：优先用配置值，否则用源方法探测到的名字，再退到 "T"。
	// 用局部变量而非回写 cfg.ReceiverVariableName，避免对入参的意外副作用
	// （多对象共用一个 cfg 时，首个对象探测到的名字会被后续对象复用）。
	receiverVariableName := zerotern.VF(cfg.ReceiverVariableName, func() string {
		return zerotern.VV(sourceMethodsTuples.GetReceiverVariableName(), "T")
	})

	newClassName := cfg.GenerateNewClassName(reflect.TypeOf(object), errorHandlingMode)

	ptx := utils.NewPTX()
	ptx.Println(`type ` + newClassName + ` struct{` + receiverVariableName + ` *` + objectType.Name() + `}`)
	ptx.Println(`
		func(` + receiverVariableName + ` *` + objectType.Name() + `) ` + string(errorHandlingMode) + `() * ` + newClassName + `{
		return & ` + newClassName + `{` + receiverVariableName + `:` + receiverVariableName + `}
	}`)

	for _, sourceMethodsTuple := range sourceMethodsTuples {
		sourceCode := sourceMethodsTuple.SourceCode
		methodList := sourceMethodsTuple.MethodList
		for _, methodFunc := range methodList {
			methodName := syntaxgo_astnode.GetText(sourceCode, methodFunc.Name)
			// Leave out the mode entrances a past run wrote (Must/Soft/Omit), so a wrapping does not get wrapped
			// Naming each mode sure states, and not the modes this run happens to want, counts when the two sets diverge
			// A run narrowed down to Must alone would otherwise treat a stale Omit as source and invoke a class it stopped generating
			// The scan leaves out this run's own artifact beforehand (utils.IsOutputFile), so this guard is what stands between a class and the artifacts of sibling runs sharing the source root
			// 把上一轮写下的模式入口（Must/Soft/Omit）排除掉，免得给包装再包一层
			// 这里点名的是 sure 声明的全部模式、而非本轮恰好要生成的那几个——两者不一致时才见分晓
			// 缩到只剩 Must 的那一轮，会把残留的 Omit 当成源方法，写出一句调用早已不再生成的类
			// 本轮自己的产物在扫描阶段就已被绕开（utils.IsOutputFile），因此这道守卫挡的是"共用同一源目录的邻居生成器留下的产物"
			if utils.In(sure.ErrorHandlingMode(methodName), sure.ErrorHandlingModes()) {
				continue
			}
			// ErrorHandlingMode is a plain string type, so calling code can state a mode of its own
			// Such a mode sits outside the set above, and its entrance belongs out of the wrapping just the same
			// ErrorHandlingMode 就是个字符串类型，调用方可以自定义模式
			// 那种模式不在上面那个集合里，但它的入口方法同样不该被包装
			if utils.In(sure.ErrorHandlingMode(methodName), cfg.ErrorHandlingModes) {
				continue
			}

			var params = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunc.Type != nil && methodFunc.Type.Params != nil {
				params = syntaxgo_astnorm.GetSimpleArgElements(methodFunc.Type.Params.List, sourceCode)
			}
			var results = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunc.Type != nil && methodFunc.Type.Results != nil {
				results = syntaxgo_astnorm.GetSimpleResElements(methodFunc.Type.Results.List, sourceCode)
			}

			// 合成的返回值名避让形参名：源方法 func (a *A) Get(res string) (int, error) 的匿名返回
			// 会被 GetSimpleResElements 命名为 res，与形参 res 撞车——具名返回处重复声明、方法体 := 左侧也撞。
			// 仅重命名撞车项，与 sure_pkg_gen 的同类保护(utils.UniqueName)保持一致。
			takenNames := make(map[string]bool, len(params))
			for _, param := range params {
				takenNames[param.Name] = true
			}
			for _, result := range results {
				if takenNames[result.Name] {
					result.Name = utils.UniqueName(result.Name, takenNames)
				}
				takenNames[result.Name] = true
			}

			for _, elem := range results {
				zaplog.LOG.Debug("Processing method return element", zap.String("name", elem.Name), zap.String("kind", elem.Kind))
			}

			var rightResultElements = make(syntaxgo_astnorm.NameTypeElements, 0)
			var errorResultElements = make(syntaxgo_astnorm.NameTypeElements, 0)
			for _, result := range results {
				if syntaxgo_astnode.GetText(sourceCode, result.Type) == "error" {
					errorResultElements = append(errorResultElements, result)
				} else {
					rightResultElements = append(rightResultElements, result)
				}
			}

			var errorHandlingStatements []string
			for _, errName := range errorResultElements.GenerateFunctionParams() {
				sureNode := zerotern.VF(cfg.ErrorHandlerFuncName, sure.GetPkgName)

				errorHandlingStatements = append(errorHandlingStatements, sureNode+"."+string(errorHandlingMode)+"("+errName+")")
			}

			// 包装方法的接收者名：避开字段名（否则生成 a.a）和本方法的参数名
			// （否则会遮蔽/撞车，如名为 T 的泛型风格参数）。基名 "T"，冲突则加序号。
			reservedNames := map[string]bool{receiverVariableName: true}
			for _, param := range params {
				reservedNames[param.Name] = true
			}
			recvName := utils.UniqueName("T", reservedNames)

			ptx.Println(`func (` + recvName + ` *` + newClassName + `) ` + methodName + `(` +
				params.FormatNamesWithKinds().MergeParts() +
				`)` + `(` +
				rightResultElements.FormatNamesWithKinds().MergeParts() +
				`) {`)

			methodInvocationStatement := recvName + `.` + receiverVariableName + `.` + methodName + `(` + params.GenerateFunctionParams().MergeParts() + `)`
			if len(results) > 0 {
				if len(rightResultElements) == len(results) {
					ptx.Println(results.GenerateFunctionParams().MergeParts() + "=" + methodInvocationStatement)
				} else {
					ptx.Println(results.GenerateFunctionParams().MergeParts() + ":=" + methodInvocationStatement)
				}
			} else {
				ptx.Println(methodInvocationStatement)
			}

			if len(errorHandlingStatements) > 0 {
				ptx.Println(strings.Join(errorHandlingStatements, "\n"))
			}

			if len(rightResultElements) > 0 {
				ptx.Println("return" + " " + rightResultElements.GenerateFunctionParams().MergeParts())
			}

			ptx.Println("}")
		}
	}
	res := ptx.String()
	zaplog.LOG.Debug("Class code generation completed", zap.Int("bytes", len(res)))
	return res
}
