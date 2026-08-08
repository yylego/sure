package example1

import (
	"testing"

	"github.com/yylego/runpath"
	"github.com/yylego/sure"
	"github.com/yylego/sure/sure_pkg_gen"
	"github.com/yylego/syntaxgo/syntaxgo_reflect"
)

type ObjectType struct{}

// No -tags=sure_generate here: sure_pkg_gen emits standalone packages that this
// package does not depend on, so a broken artifact leaves this package alone and
// the unlocking build constraint is beside the point. See sure.DirectiveSkipGenerate
//
// 这里不带 -tags=sure_generate：sure_pkg_gen 生成的是独立新包，本包不依赖它，
// 产物坏掉也不会连坐本包，所以那道解锁用的构建约束在这儿用不上。参见 sure.DirectiveSkipGenerate
//
//go:generate go test -run ^TestGenerate$
func TestGenerate(t *testing.T) {
	pkgPath := syntaxgo_reflect.GetPkgPathV2[ObjectType]()
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.SOFT, pkgPath)
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.MUST, pkgPath)
	sure_pkg_gen.GenerateSurePackage(runpath.PARENT.Path(), sure.OMIT, pkgPath)
}
