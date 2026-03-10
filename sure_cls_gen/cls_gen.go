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
	"path/filepath"
	"reflect"
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
	"github.com/yylego/syntaxgo/syntaxgo_search"
	"github.com/yylego/tern/zerotern"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// ClassGenConfig contains configuration for class generation
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
func GenerateClasses(cfg *ClassGenConfig, objects ...interface{}) {
	utils.PrintObject(cfg)

	must.Nice(cfg.ClassGenOptions.SourceRootPath)
	must.Nice(cfg.PackageName)
	must.Nice(cfg.OutputPath)

	ptx := utils.NewPTX()
	ptx.Println("// Code generated using sure/sure_cls_gen. DO NOT EDIT.")
	ptx.Println("// This file was auto generated via github.com/yylego/sure")
	ptx.Println("// Generated from:", utils.GetGenPosFuncMark(1))
	ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
	ptx.Println()
	ptx.Println("package", cfg.PackageName)

	for _, object := range objects {
		ptx.Println(GenerateClassCodeWithErrorHandling(cfg.ClassGenOptions, object))
	}

	var importOptions *syntaxgo_ast.PackageImportOptions
	if cfg.ImportOptions != nil {
		importOptions = cfg.ImportOptions
	} else {
		importOptions = &syntaxgo_ast.PackageImportOptions{}
	}
	if cfg.ClassGenOptions.ErrorHandlerFuncName != "" {
		zaplog.LOG.Debug("Using custom exception handling function", zap.String("func", cfg.ClassGenOptions.ErrorHandlerFuncName))
	} else {
		// Use default Must and Soft functions, add package reference to help code formatting
		// 使用默认的 Must 和 Soft 函数，补充包引用以帮助代码格式化
		importOptions.SetInferredObject(syntaxgo_reflect.GetObject[sure.ErrorHandlingMode]())
	}

	// Inject import statements into code
	// 将导入语句注入到代码中
	source := importOptions.InjectImports(ptx.Bytes())

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
func GenerateClassCodeWithErrorHandling(cfg *ClassGenOptions, object interface{}) string {
	utils.PrintObject(cfg)

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
func GenerateClassWithErrorHandlingMode(cfg *ClassGenOptions, object interface{}, errorHandlingMode sure.ErrorHandlingMode) string {
	utils.PrintObject(cfg)

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

	var sourceMethodsTuples = make(utils.SourceMethodsTuples, 0)
	for _, fileInfo := range done.VAE(os.ReadDir(cfg.SourceRootPath)).Done() {
		if fileInfo.IsDir() {
			continue
		}
		if filepath.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		path := filepath.Join(cfg.SourceRootPath, fileInfo.Name())
		zaplog.LOG.Debug("Reading source file to extract methods", zap.String("path", path))

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

	// when zero - set a new value
	zerotern.SetPF(&cfg.ReceiverVariableName, func() string {
		return zerotern.VV(sourceMethodsTuples.GetReceiverVariableName(), "T")
	})

	newClassName := cfg.GenerateNewClassName(reflect.TypeOf(object), errorHandlingMode)

	ptx := utils.NewPTX()
	ptx.Println(`type ` + newClassName + ` struct{` + cfg.ReceiverVariableName + ` *` + objectType.Name() + `}`)
	ptx.Println(`
		func(` + cfg.ReceiverVariableName + ` *` + objectType.Name() + `) ` + string(errorHandlingMode) + `() * ` + newClassName + `{
		return & ` + newClassName + `{` + cfg.ReceiverVariableName + `:` + cfg.ReceiverVariableName + `}
	}`)

	for _, sourceMethodsTuple := range sourceMethodsTuples {
		sourceCode := sourceMethodsTuple.SourceCode
		methodList := sourceMethodsTuple.MethodList
		for _, methodFunc := range methodList {
			methodName := syntaxgo_astnode.GetText(sourceCode, methodFunc.Name)
			if utils.In(methodName, []string{string(sure.MUST), string(sure.SOFT)}) {
				continue
			}
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

			ptx.Println(`func (T *` + newClassName + `) ` + methodName + `(` +
				params.FormatNamesWithKinds().MergeParts() +
				`)` + `(` +
				rightResultElements.FormatNamesWithKinds().MergeParts() +
				`) {`)

			methodInvocationStatement := `T.` + cfg.ReceiverVariableName + `.` + methodName + `(` + params.GenerateFunctionParams().MergeParts() + `)`
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
