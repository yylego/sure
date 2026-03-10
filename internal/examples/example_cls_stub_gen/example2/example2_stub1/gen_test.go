package example2_stub1

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/runpath/runtestpath"
	"github.com/yylego/sure/cls_stub_gen"
	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example2"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

func TestGenerate(t *testing.T) {
	param := cls_stub_gen.NewStubParam(&example2.A{}, "A100")

	cfg := &cls_stub_gen.StubGenConfig{
		SourceRootPath:    runpath.PARENT.Join("../../example2"),
		TargetPackageName: syntaxgo.CurrentPackageName(),
		ImportOptions:     syntaxgo_ast.NewPackageImportOptions(),
		OutputPath:        runtestpath.SrcPath(t),
		AllowFileCreation: false,
	}
	cls_stub_gen.GenerateStubs(cfg, param)
}
