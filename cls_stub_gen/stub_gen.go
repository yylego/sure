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
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yylego/done"
	"github.com/yylego/formatgo"
	"github.com/yylego/must"
	"github.com/yylego/sure/internal/utils"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
	"github.com/yylego/syntaxgo/syntaxgo_astnode"
	"github.com/yylego/syntaxgo/syntaxgo_astnorm"
	"github.com/yylego/syntaxgo/syntaxgo_reflect"
	"github.com/yylego/syntaxgo/syntaxgo_search"
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
}

// GenerateStubs creates package scope stub functions from singleton methods
// Scans source code and generates wrapping functions
// Writes formatted code to the specified output path
//
// GenerateStubs 从单例方法创建包级存根函数
// 扫描源代码并生成包装函数
// 将格式化的代码写入指定的输出路径
func GenerateStubs(config *StubGenConfig, stubParams ...*StubParam) {
	utils.PrintObject(config)

	must.Nice(config.SourceRootPath)
	must.Nice(config.TargetPackageName)
	must.Nice(config.OutputPath)

	ptx := utils.NewPTX()
	ptx.Println("// Code generated using sure/cls_stub_gen. DO NOT EDIT.")
	ptx.Println("// This file was auto generated via github.com/yylego/sure")
	ptx.Println("// Generated from:", utils.GetGenPosFuncMark(1))
	ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
	ptx.Println()
	ptx.Println("package", config.TargetPackageName)

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
		importOptions.SetPkgPath(syntaxgo_reflect.GetPkgPathV3(param.object))
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
	utils.PrintObject(cfg)

	objectType := syntaxgo_reflect.GetTypeV3(stubParam.object)
	zaplog.LOG.Debug("Analyzing object type to generate stubs",
		zap.String("name", must.Nice(objectType.Name())),
		zap.String("type", must.Nice(objectType.String())),
		zap.String("package", must.Nice(objectType.PkgPath())))

	utils.MustRoot(cfg.SourceRootPath)

	var sourceMethodsTuples = make(utils.SourceMethodsTuples, 0)
	for _, fileInfo := range done.VAE(os.ReadDir(cfg.SourceRootPath)).Done() {
		if fileInfo.IsDir() {
			continue
		}
		if filepath.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		path := filepath.Join(cfg.SourceRootPath, fileInfo.Name())
		zaplog.LOG.Debug("Reading source file to extract stub methods", zap.String("path", path))

		sourceCode := done.VAE(os.ReadFile(path)).Done()

		astBundle := done.VCE(syntaxgo_ast.NewAstBundleV1(sourceCode)).Nice()

		astFile, _ := astBundle.GetBundle()

		extractedFunctions := syntaxgo_search.ExtractFunctions(astFile)
		methodList := syntaxgo_search.ExtractFunctionsByReceiverName(extractedFunctions, objectType.Name(), true)
		if len(methodList) == 0 {
			continue
		}

		sourceMethodsTuples = append(sourceMethodsTuples, &utils.SourceMethodsTuple{
			SourceCode: sourceCode,
			MethodList: methodList,
		})
	}

	ptx := utils.NewPTX()

	for _, sourceMethodsTuple := range sourceMethodsTuples {
		sourceCode := sourceMethodsTuple.SourceCode
		methodList := sourceMethodsTuple.MethodList
		for _, methodFunction := range methodList {
			methodName := syntaxgo_astnode.GetText(sourceCode, methodFunction.Name)

			var methodParameters = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunction.Type != nil && methodFunction.Type.Params != nil {
				methodParameters = syntaxgo_astnorm.GetSimpleArgElements(methodFunction.Type.Params.List, sourceCode)

				for _, elem := range methodParameters {
					elem.AdjustTypeWithPackage(syntaxgo_reflect.GetPkgNameV3(stubParam.object), make(map[string]ast.Expr))
				}
			}
			var returnValues = make(syntaxgo_astnorm.NameTypeElements, 0)
			if methodFunction.Type != nil && methodFunction.Type.Results != nil {
				returnValues = syntaxgo_astnorm.GetSimpleResElements(methodFunction.Type.Results.List, sourceCode)

				for _, elem := range returnValues {
					elem.AdjustTypeWithPackage(syntaxgo_reflect.GetPkgNameV3(stubParam.object), make(map[string]ast.Expr))
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
