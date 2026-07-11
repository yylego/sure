package example1

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/sure"
	"github.com/yylego/sure/sure_pkg_gen"
	"github.com/yylego/syntaxgo/syntaxgo_reflect"
)

type ObjectType struct{}

func TestGenerate(t *testing.T) {
	pkgPath := syntaxgo_reflect.GetPkgPathV2[ObjectType]()
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.SOFT, pkgPath)
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.MUST, pkgPath)
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.OMIT, pkgPath)
}
