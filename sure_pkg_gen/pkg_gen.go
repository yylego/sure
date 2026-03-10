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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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
	t *testing.T,
	sourceRoot string,
	errorHandlingMode sure.ErrorHandlingMode,
	sourcePackagePath string,
) {
	GenerateSurePackageFiles(t, NewSurePackageConfig(
		sourceRoot,
		errorHandlingMode,
		sourcePackagePath,
	))
}

// GenerateSurePackageFiles generates package files from configuration
// Scans source root and creates wrapped functions with exception handling
// Writes generated packages to configured output paths
//
// GenerateSurePackageFiles 从配置生成包文件
// 扫描源根目录并创建带有异常处理的包装函数
// 将生成的包写入配置的输出路径
func GenerateSurePackageFiles(t *testing.T, cfg *SurePackageGenConfig) {
	utils.PrintObject(cfg)

	zaplog.LOG.Debug("Starting package generation from source root", zap.String("root", cfg.SourceRoot))

	for _, name := range utils.MustLs(cfg.SourceRoot) {
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		absPath := filepath.Join(cfg.SourceRoot, name)
		zaplog.LOG.Debug("Reading source file to extract functions", zap.String("path", absPath))

		sureFunctionCodes := GenerateSureFunctions(t, cfg, absPath)

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
			ptx.Println("// Generated from:", utils.GetGenPosFuncMark(2))
			ptx.Println("// ========== SURE:DO-NOT-EDIT-SECTION:END ==========")
			ptx.Println()
			ptx.Println("package" + " " + newPackageName)
			ptx.Println("import(")
			ptx.Println(utils.SetDoubleQuotes(cfg.SourcePackagePath))
			ptx.Println(utils.SetDoubleQuotes(cfg.ErrorHandlingPkgPath))
			ptx.Println(")")
			ptx.Println(strings.Join(sureFunctionCodes, "\n"))

			newName := strings.Replace(name, ".go", "_"+shortErrorHandlingMode+".go", 1)
			newPath := filepath.Join(cfg.OutputRoot, newPackageName, newName)

			newCode, _ := formatgo.FormatCode(ptx.String())

			utils.MustWriteIntoPath(newPath, newCode)
			zaplog.LOG.Info("Package file generation complete", zap.String("output", newPath))
		}
	}
}

// GenerateSureFunctions extracts functions from source file and wraps them
// Parses AST to find exported functions and generates assertion-wrapped versions
// Returns slice of generated function code strings
//
// GenerateSureFunctions 从源文件提取函数并包装它们
// 解析 AST 以查找导出的函数并生成带断言包装的版本
// 返回生成的函数代码字符串切片
func GenerateSureFunctions(t *testing.T, cfg *SurePackageGenConfig, absPath string) []string {
	utils.PrintObject(cfg)

	sourceCode := done.VAE(os.ReadFile(absPath)).Done()
	astBundle := rese.P1(syntaxgo_ast.NewAstBundleV1(sourceCode))

	packageName := astBundle.GetPackageName()
	zaplog.LOG.Debug("Analyzing source file AST",
		zap.String("package", packageName),
		zap.String("path", absPath))

	astFile, _ := astBundle.GetBundle()

	astFunctions := syntaxgo_search.ExtractFunctions(astFile)
	zaplog.LOG.Debug("Extracted functions from source", zap.Int("count", len(astFunctions)))

	var functionCodes []string
	for _, astFunc := range astFunctions {
		if astFunc.Recv != nil {
			if len(astFunc.Recv.List) > 0 && len(astFunc.Recv.List[0].Names) > 0 {
				t.Log(astFunc.Recv.List[0].Names[0].Name, astFunc.Name.Name)
			}
			continue
		}
		t.Log(astFunc.Name.Name)
		if !utils.C0IsUppercase(astFunc.Name.Name) {
			continue
		}
		results := parseFunctionReturnFields(sourceCode, astFunc)
		t.Log(utils.Neat2json(results))

		zaplog.LOG.Debug("Processing exported function",
			zap.String("function", astFunc.Name.Name),
			zap.Int("returns", len(results)))

		functionCode := makeFunctionCode(sourceCode, packageName, astFunc, results, cfg.ErrorHandlingMode, cfg.HandlerFuncReference)
		t.Log(functionCode)

		functionCodes = append(functionCodes, functionCode)
	}
	zaplog.LOG.Debug("Function code generation completed", zap.Int("generated", len(functionCodes)))
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
	var newFunctionCode = "func " + astFunc.Name.Name
	if astFunc.Type.TypeParams != nil {
		newFunctionCode += syntaxgo_astnode.GetText(sourceCode, astFunc.Type.TypeParams)
	}

	genericTypeParams := syntaxgo_astnorm.GetGenericTypeParamsMap(astFunc.Type.TypeParams)

	newFunctionCode += "("
	var argumentNames []string
	if astFunc.Type.Params != nil && len(astFunc.Type.Params.List) > 0 {
		var args []string
		for _, param := range astFunc.Type.Params.List {
			argType := syntaxgo_astnode.GetText(sourceCode, param.Type)
			argType = resolveFullExportedType(packageName, genericTypeParams, argType)

			if len(param.Names) == 0 {
				argName := "arg" + strconv.Itoa(len(args))

				args = append(args, argName+" "+argType)
				argumentNames = append(argumentNames, argName)
			} else {
				argNames := make([]string, 0, len(param.Names))

				for _, name := range param.Names {
					// 检查参数是否是 "..."
					if _, variadic := param.Type.(*ast.Ellipsis); variadic {
						argNames = append(argNames, name.Name+" ...")
					} else {
						argNames = append(argNames, name.Name)
					}
				}

				args = append(args, strings.Join(argNames, ",")+" "+argType)
				argumentNames = append(argumentNames, argNames...)
			}
		}
		newFunctionCode += strings.Join(args, ",")
	}
	newFunctionCode += ")"

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
				resultNames = append(resultNames, resolveFullExportedType(packageName, genericTypeParams, res.Type))
			} else {
				resultNames = append(resultNames, res.Name+" "+resolveFullExportedType(packageName, genericTypeParams, res.Type))
			}
		}
		if len(resultNames) > 0 {
			if len(resultNames) < 2 && anonymousReturn {
				newFunctionCode += " " + resultNames[0]
			} else {
				newFunctionCode += " " + "(" + strings.Join(resultNames, ",") + ")"
			}
		}
	}
	newFunctionCode += " " + "{" + "\n"
	if len(results) > 0 {
		var rets = make([]string, 0, len(results))
		for _, res := range results {
			rets = append(rets, res.Name)
		}
		newFunctionCode += strings.Join(rets, ",")
		if anonymousReturn || containsErrorReturn {
			newFunctionCode += ":="
		} else {
			newFunctionCode += "="
		}
	}
	newFunctionCode += packageName + "." + astFunc.Name.Name
	if astFunc.Type.TypeParams != nil {
		newFunctionCode += "["
		var genericTypeNames []string
		for _, xts := range astFunc.Type.TypeParams.List {
			for _, xtx := range xts.Names {
				genericTypeNames = append(genericTypeNames, xtx.Name)
			}
		}
		newFunctionCode += strings.Join(genericTypeNames, ",")
		newFunctionCode += "]"
	}
	newFunctionCode += "(" + strings.Join(argumentNames, ",") + ")" + "\n"
	if len(results) > 0 {
		for _, x := range results {
			if x.Type == "error" {
				newFunctionCode += sureHandlerFuncReference + "." + string(errorHandlingMode) + "(" + x.Name + ")" + "\n"
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
			newFunctionCode += "return" + " " + strings.Join(rets, ",") + "\n"
		}
	}
	newFunctionCode += "}" + "\n"
	zaplog.LOG.Debug("Wrapper function code constructed",
		zap.String("function", astFunc.Name.Name),
		zap.String("mode", string(errorHandlingMode)),
		zap.Int("bytes", len(newFunctionCode)))
	return newFunctionCode
}

// resolveFullExportedType adds package prefix to exported types
// Handles generic types and reference types as needed
// Returns complete qualified type name with package prefix
//
// resolveFullExportedType 为导出类型添加包前缀
// 根据需要处理泛型类型和引用类型
// 返回带有包前缀的完整限定类型名
func resolveFullExportedType(packageName string, genericTypeParams map[string]ast.Expr, resType string) string {
	if utils.C0IsUppercase(resType) {
		classType := resType
		if _, ok := genericTypeParams[classType]; ok {
			return resType
		}
		return packageName + "." + classType
	}
	if resType[0] == '*' {
		classType := resType[1:]
		if _, ok := genericTypeParams[classType]; ok {
			return resType
		}
		if utils.C0IsUppercase(classType) {
			return "*" + packageName + "." + classType
		}
	}
	return resType
}

// resultType represents a function return value with metadata
// Contains name, type info and anonymousness flag
// Used during function code generation
//
// resultType 表示带有元数据的函数返回值
// 包含名称、类型信息以及匿名标志
// 在函数代码生成期间使用
type resultType struct {
	Name        string // Return value name // 返回值名称
	Type        string // Return value type // 返回值类型
	IsAnonymous bool   // True when return value is anonymous // 当返回值为匿名时为真
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
			eIs := utils.Boolean(resType == "error")
			if len(elem.Names) == 0 {
				resName := tern.BFF(eIs, func() string {
					return tern.BVV(errNum == 0, "err", "err"+strconv.Itoa(errNum))
				}, func() string {
					return "res" + strconv.Itoa(len(results))
				})

				results = append(results, &resultType{
					Name:        resName,
					Type:        resType,
					IsAnonymous: true,
				})
				if eIs {
					errNum++
				}
			} else {
				for _, name := range elem.Names {
					results = append(results, &resultType{
						Name:        name.Name,
						Type:        resType,
						IsAnonymous: false,
					})
					if eIs {
						errNum++
					}
				}
			}
		}
	}
	return results
}
