package example1_stub2

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/runpath/runtestpath"
	"github.com/yylego/sure/cls_stub_gen"
	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example1"
	"github.com/yylego/syntaxgo"
	"github.com/yylego/syntaxgo/syntaxgo_ast"
)

//go:generate go test -run ^TestGenerate$ -tags=sure_generate
func TestGenerate(t *testing.T) {
	param := cls_stub_gen.NewStubParam(&example1.A{}, "example1.STUB2")

	cfg := cls_stub_gen.NewStubGenConfig(
		runpath.PARENT.Join("../../example1"),
		syntaxgo.CurrentPackageName(),
		runtestpath.SrcPath(t),
	).WithImportOptions(syntaxgo_ast.NewPackageImportOptions())

	cls_stub_gen.GenerateStubs(cfg, param)
}
