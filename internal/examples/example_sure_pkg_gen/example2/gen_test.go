package example2

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

	//选择生成 soft 软包
	sure_pkg_gen.GenerateSurePackage(t, runpath.PARENT.Path(), sure.SOFT, pkgPath)
	//选择生成 must 硬包
	sure_pkg_gen.GenerateSurePackage(t, runpath.PARENT.Path(), sure.MUST, pkgPath)
}
