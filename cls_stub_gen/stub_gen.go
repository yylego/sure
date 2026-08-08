// Package cls_stub_gen generates package scope function wrapping from singleton struct methods
// Auto creates stub functions that wrap singleton instance methods at package scope
// Makes access to singleton methods simple through clean package scope function calls
// Supports auto import management and configurable output paths, creating files as needed
//
// Package cls_stub_gen 从单例结构体方法生成包作用域函数包装
// 自动创建在包作用域包装单例实例方法的存根函数
// 通过简洁的包作用域函数调用使对单例方法的访问简单
// 支持自动导入管理和可配置输出路径，按需创建文件
package cls_stub_gen

import (
	"fmt"
	"os"
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
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// StubParam contains parameters to generate stub functions
// Combines object type info with operation namespace
// Used to define how to wrap singleton methods
//
// StubParam 包含生成存根函数的参数
// 组合对象类型信息和操作命名空间
// 用于定义如何包装单例方法
type StubParam struct {
	object             any    // Object type to wrap // 要包装的对象类型
	operationNamespace string // Namespace prefix for operations // 操作的命名空间前缀
}

// NewStubParam creates new stub generation config
// Takes object instance and namespace to use when calling methods
// Returns a new StubParam instance
//
// NewStubParam 创建新的存根生成配置
// 接受对象实例和调用方法时使用的命名空间
// 返回新的 StubParam 实例
func NewStubParam(object any, operationNamespace string) *StubParam {
	return &StubParam{
		object:             object,
		operationNamespace: operationNamespace,
	}
}

// StubGenConfig contains stub generation configuration settings
// Specifies source paths, target package name, and output settings
// Controls option to create new files when needed
//
// StubGenConfig 包含存根生成的配置设置
// 指定源路径、目标包名和输出设置
// 控制需要时创建新文件的选项
type StubGenConfig struct {
	SourceRootPath    string                             // Source code root path // 源代码根路径
	TargetPackageName string                             // Target package name // 目标包名
	ImportOptions     *syntaxgo_ast.PackageImportOptions // Import configuration // 导入配置
	OutputPath        string                             // Output file path // 输出文件路径
	AllowFileCreation bool                               // Allow creating new file if not exists // 当文件不存在时允许创建新文件
	GoBuildDirective  string                             // go:build directive for the generated file, blank writes none // 生成文件的 go:build 指令，留空则不写
}

// NewStubGenConfig creates a StubGenConfig with the build constraint set
// Use this shape, since a bare struct value leaves the constraint blank
//
// The constraint counts here because the generated stubs and the test that makes
// them share one package, so a broken artifact takes that test down with it.
// See sure.DirectiveSkipGenerate
//
// NewStubGenConfig 创建已设好构建约束的 StubGenConfig
// 优先用这种写法，因为直接写结构体字面量会让约束留空
//
// 这里的约束很要紧：生成的存根与生成它的测试同住一个包，
// 破损产物会把那个测试一并搞挂。参见 sure.DirectiveSkipGenerate
func NewStubGenConfig(sourceRootPath string, targetPackageName string, outputPath string) *StubGenConfig {
	return &StubGenConfig{
		SourceRootPath:    sourceRootPath,
		TargetPackageName: targetPackageName,
		OutputPath:        outputPath,
		GoBuildDirective:  sure.DirectiveSkipGenerate,
	}
}

// WithImportOptions sets the import configuration
// WithImportOptions 设置导入配置
func (cfg *StubGenConfig) WithImportOptions(importOptions *syntaxgo_ast.PackageImportOptions) *StubGenConfig {
	cfg.ImportOptions = importOptions
	return cfg
}

// WithAllowFileCreation lets the codegen create the output file when missing
// WithAllowFileCreation 允许生成器在输出文件缺失时创建它
func (cfg *StubGenConfig) WithAllowFileCreation(allowFileCreation bool) *StubGenConfig {
	cfg.AllowFileCreation = allowFileCreation
	return cfg
}

// WithGoBuildDirective sets the go:build directive written into generated files
// State a negated tag such as sure.DirectiveSkipGenerate, since a directive without
// the leading "!" would drop the artifact from each standard build instead
// WithGoBuildDirective 设置写入生成文件的 go:build 指令
// 要写取反的标签（如 sure.DirectiveSkipGenerate），因为漏掉开头的 "!" 会反过来
// 让产物在每次普通构建里都被排除掉
func (cfg *StubGenConfig) WithGoBuildDirective(goBuildDirective string) *StubGenConfig {
	cfg.GoBuildDirective = goBuildDirective
	return cfg
}

// WithNoBuildDirective clears the directive so generated files state no constraint
// Choosing this gives up the escape hatch: once a generated file stops compiling,
// its package stops compiling, and the test inside cannot run to set it straight
// See sure.DirectiveSkipGenerate
//
// WithNoBuildDirective 清除指令，使生成文件不带任何构建约束
// 选它就等于放弃那道退路：生成文件一旦编译不过，它所在的包就编译不过，
// 包里那个生成器测试也没法再跑起来修它
// 参见 sure.DirectiveSkipGenerate
func (cfg *StubGenConfig) WithNoBuildDirective() *StubGenConfig {
	cfg.GoBuildDirective = ""
	return cfg
}

// GenerateStubs creates package scope stub functions from singleton methods
// Scans source code and generates wrapping functions
// Writes formatted code to the specified output path
//
// GenerateStubs 从单例方法创建包级存根函数
// 扫描源代码并生成包装函数
// 将格式化的代码写入指定的输出路径
func GenerateStubs(config *StubGenConfig, stubParams ...*StubParam) {
	must.Nice(config.SourceRootPath)
	must.Nice(config.TargetPackageName)
	must.Nice(config.OutputPath)
	utils.MustRoot(config.SourceRootPath)

	// One scan up front gathers the imports of each source file: signatures get
	// transcribed as text, so the types inside need a stated origin. This repeats
	// the scan inside GenerateStubMethods, and buys leaving that exported signature
	// untouched, since downstream code calls it and a changed signature breaks such
	// code. Codegen runs at development time, so one extra pass across the source
	// tree costs nothing worth counting
	// 先扫一遍把各源文件的导入收齐：方法签名是当文本抄走的，抄来的类型得留着出处。
	// 这一趟与 GenerateStubMethods 内部那趟重复，换来的是不动它的导出签名——
	// 下游直接调用它，改签名会把下游搞挂。生成器只在开发期跑，多读一遍目录不值一提
	sourceMethodsTuples := make(utils.SourceMethodsTuples, 0)
	for _, param := range stubParams {
		typeName := syntaxgo_reflect.GetTypeV3(param.object).Name()
		tuples := utils.CollectReceiverMethods(config.SourceRootPath, typeName)
		if len(tuples) == 0 {
			// A stub set with no functions in it is almost sure a mistake, so speak
			// up instead of writing a file that holds nothing besides its heading
			// 一个函数都没有的存根集几乎肯定是哪里搞错了，出声说明，而不是默默写出个只剩文件头的文件
			zaplog.LOG.Warn("No methods with this receiver type in the source root",
				zap.String("type", typeName),
				zap.String("root", config.SourceRootPath))
		}
		sourceMethodsTuples = append(sourceMethodsTuples, tuples...)
	}
	// 排除即将被覆写的那个产物，免得它的导入被一轮轮传下去。详见 DropOutputFile
	sourceImports := sourceMethodsTuples.DropOutputFile(config.OutputPath).MergeSourceImports()

	ptx := utils.NewPTX()
	ptx.Println("// Code generated using sure/cls_stub_gen. DO NOT EDIT.")
	ptx.Println("// This file was auto generated via github.com/yylego/sure")
	// The generated file and the test that makes it share one package, and that
	// shape can lock, which a build constraint undoes. See sure.DirectiveSkipGenerate
	// 生成文件与生成它的测试同住一个包，这个格局会卡住；构建约束就是那把钥匙。
	// 详见 sure.DirectiveSkipGenerate
	if goBuildLine := utils.FormatGoBuildDirective(config.GoBuildDirective); goBuildLine != "" {
		ptx.Println()
		ptx.Println(goBuildLine)
		ptx.Println()
	}
	ptx.Println("// Generated from:", utils.GetGenPosFuncMark(1))
	ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
	ptx.Println()
	ptx.Println("package", config.TargetPackageName)
	// The taken imports form a block beside the one injected beneath, since Go
	// accepts more than one import declaration in a file. Anything unused gets
	// pruned once goimports runs, so taking too much is safe
	// 取来的导入自成一段，与下面注入的那段并存，因为 Go 允许一个文件有多段导入声明。
	// 用不上的会被 goimports 剪掉，所以取多了无妨
	if importBlock := sourceImports.FormatImportBlock(); importBlock != "" {
		ptx.Println(importBlock)
	}

	for _, param := range stubParams {
		ptx.Println(GenerateStubMethods(config, param))
	}

	var importOptions *syntaxgo_ast.PackageImportOptions
	if config.ImportOptions != nil {
		importOptions = config.ImportOptions
	} else {
		importOptions = &syntaxgo_ast.PackageImportOptions{}
	}

	for _, param := range stubParams {
		importOptions.AddPkgPath(syntaxgo_reflect.GetPkgPathV3(param.object))
	}

	// Inject import statements into code
	// 将导入语句注入到代码中
	source := importOptions.InjectImports(ptx.Bytes())

	zaplog.LOG.Debug("Import statements injected into source code", zap.Int("bytes", len(source)))

	// Measure code formatting time
	// 测量代码格式化时间
	sinceTime := time.Now()

	// Format code without extra import resolution to save time
	// 格式化代码时不额外解析导入以节省时间
	newSource := done.VAE(formatgo.FormatBytes(source)).Nice()

	// An import in the output with no source file and no registration behind it
	// has to be guesswork, so stop at this point and write nothing
	// 成品里若冒出既非源文件、也非调用方登记的导入，那只能是猜出来的，就此打住不写盘
	utils.MustKnownImports(newSource, slices.Concat(sourceImports.Paths(), importOptions.GetPkgPaths()))

	// Write formatted code to output path
	// 将格式化的代码写入输出路径
	duration := time.Since(sinceTime)
	zaplog.LOG.Debug("Source code formatting completed", zap.Duration("cost", duration), zap.Int("bytes", len(newSource)))
	if config.OutputPath != "" {
		if config.AllowFileCreation {
			// Allow new file creation, but might write to wrong location if misconfigured
			// 允许创建新文件，但如果配置错误可能写入错误位置
			zaplog.LOG.Warn("Writing to target source path with file creation enabled", zap.String("path", config.OutputPath))
		} else {
			// Recommended: create file first to ensure correct location
			// 推荐：先创建文件以确保位置正确
			utils.MustFile(config.OutputPath) // Require target file to exist // 要求目标文件必须存在
		}
		done.Done(os.WriteFile(config.OutputPath, newSource, 0644))
		zaplog.LOG.Info("Stub functions generation complete", zap.String("output", config.OutputPath))
	} else {
		fmt.Println(string(newSource))
		zaplog.LOG.Debug("Generated stub functions output to stdout", zap.Int("bytes", len(newSource)))
	}
}

// GenerateStubMethods generates stub methods to wrap a single object
// Extracts methods from source code and creates wrapping functions
// Returns generated code as a string
//
// GenerateStubMethods 生成存根方法包装单个对象
// 从源代码提取方法并创建包装函数
// 返回生成的代码作为字符串
func GenerateStubMethods(cfg *StubGenConfig, stubParam *StubParam) string {
	objectType := syntaxgo_reflect.GetTypeV3(stubParam.object)
	zaplog.LOG.Debug("Analyzing object type to generate stubs",
		zap.String("name", must.Nice(objectType.Name())),
		zap.String("type", must.Nice(objectType.String())),
		zap.String("package", must.Nice(objectType.PkgPath())))

	utils.MustRoot(cfg.SourceRootPath)

	sourceMethodsTuples := utils.CollectReceiverMethods(cfg.SourceRootPath, objectType.Name())

	// 源包的包名前缀：把方法里的源包类型（含 []T / map / func 等复合类型）限定到目标包。
	// 从反射类型名（形如 "demo.Foo"）取包名段，而非 GetPkgNameV3——后者内部是 path.Base，
	// 对 /vN 版本化模块会得到 "v2"，与 import 绑定的真实包名（demo）不符、生成 v2.Item 编译不过。
	// 已实测 reflect.Type.String() 取的是包名而非路径末段；与 pkg.go:GetPkgName 的 /vN 安全做法一致。
	sourcePackageName := objectType.String()
	if pkgName, _, ok := strings.Cut(sourcePackageName, "."); ok {
		sourcePackageName = pkgName
	}

	ptx := utils.NewPTX()

	for _, sourceMethodsTuple := range sourceMethodsTuples {
		sourceCode := sourceMethodsTuple.SourceCode
		methodList := sourceMethodsTuple.MethodList
		for _, methodFunction := range methodList {
			methodName := syntaxgo_astnode.GetText(sourceCode, methodFunction.Name)

			// 仅包装导出方法：生成的 stub 位于目标包、跨包调用源包的方法，
			// 未导出方法跨包不可见，生成出来必然编译不过。与 pkg_gen 的同款守卫一致。
			if !utils.C0IsUppercase(methodName) {
				continue
			}

			var methodParameters = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunction.Type != nil && methodFunction.Type.Params != nil {
				methodParameters = syntaxgo_astnorm.GetSimpleArgElements(methodFunction.Type.Params.List, sourceCode)

				for _, elem := range methodParameters {
					// 把方法里的源包类型（含 []Foo / map 等复合类型）限定到目标包。
					// v0.0.1 的 AdjustTypeWithPackage 内部就是递归的 ResolveQualifiedType，与 cls_gen 用法一致。
					elem.AdjustTypeWithPackage(sourcePackageName, nil, sourceCode)
				}
			}
			var returnValues = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunction.Type != nil && methodFunction.Type.Results != nil {
				returnValues = syntaxgo_astnorm.GetSimpleResElements(methodFunction.Type.Results.List, sourceCode)

				for _, elem := range returnValues {
					elem.AdjustTypeWithPackage(sourcePackageName, nil, sourceCode)
				}
			}

			for _, elem := range returnValues {
				zaplog.LOG.Debug("Processing method return value", zap.String("name", elem.Name), zap.String("kind", elem.Kind))
			}

			ptx.Println(`func ` + methodName + `(` +
				methodParameters.FormatNamesWithKinds().MergeParts() +
				`)` + `(` +
				strings.Join(returnValues.Kinds(), ",") +
				`) {`)

			methodInvocationStatement := stubParam.operationNamespace + `.` + methodName + `(` + methodParameters.GenerateFunctionParams().MergeParts() + `)`
			if len(returnValues) > 0 {
				ptx.Println("return " + methodInvocationStatement)
			} else {
				ptx.Println(methodInvocationStatement)
			}

			ptx.Println("}")
		}
	}
	res := ptx.String()
	zaplog.LOG.Debug("Stub methods generation completed", zap.Int("bytes", len(res)))
	return res
}
