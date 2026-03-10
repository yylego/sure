package example2

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/runpath/runtestpath"
	"github.com/yylego/sure/sure_cls_gen"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

func TestGenerate(t *testing.T) {
	cfg := &sure_cls_gen.ClassGenConfig{
		ClassGenOptions: sure_cls_gen.NewClassGenOptions(runpath.PARENT.Path()).
			WithErrorHandlerFuncName("done"),
		PackageName: syntaxgo.CurrentPackageName(),
		ImportOptions: syntaxgo_ast.NewPackageImportOptions().
			SetPkgPath("github.com/yylego/done"),
		OutputPath: runtestpath.SrcPath(t),
	}

	sure_cls_gen.GenerateClasses(cfg, Example{}, Demo{})
}
