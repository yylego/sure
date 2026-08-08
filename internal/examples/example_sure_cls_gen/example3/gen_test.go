package example3

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/runpath/runtestpath"
	"github.com/yylego/sure/internal/examples/example_sure_cls_gen/example3/example3_custom"
	"github.com/yylego/sure/sure_cls_gen"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

//go:generate go test -run ^TestGenerate$ -tags=sure_generate
func TestGenerate(t *testing.T) {
	cfg := &sure_cls_gen.ClassGenConfig{
		ClassGenOptions: sure_cls_gen.NewClassGenOptions(runpath.PARENT.Path()).
			WithErrorHandlerFuncName("example3_custom.NODE"),
		PackageName: syntaxgo.CurrentPackageName(),
		ImportOptions: syntaxgo_ast.NewPackageImportOptions().
			AddInferredObject(example3_custom.SureClass{}),
		OutputPath: runtestpath.SrcPath(t),
	}
	sure_cls_gen.GenerateClasses(cfg, Example{}, Demo{})
}
