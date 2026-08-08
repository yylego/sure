package example1

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/runpath/runtestpath"
	"github.com/yylego/sure"
	"github.com/yylego/sure/sure_cls_gen"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

//go:generate go test -run ^TestGenerate$ -tags=sure_generate
func TestGenerate(t *testing.T) {
	cfg := &sure_cls_gen.ClassGenConfig{
		ClassGenOptions: sure_cls_gen.NewClassGenOptions(runpath.PARENT.Path()).
			WithNewClassNameParts("88").
			WithNamingPatternType(sure_cls_gen.STYLE_SUFFIX_CAMELCASE_TYPE).
			MoreErrorHandlingModes(sure.MUST, sure.SOFT, sure.OMIT),
		PackageName:   syntaxgo.CurrentPackageName(),
		ImportOptions: syntaxgo_ast.NewPackageImportOptions(),
		OutputPath:    runtestpath.SrcPath(t),
	}

	sure_cls_gen.GenerateClasses(cfg, Example{})
}
