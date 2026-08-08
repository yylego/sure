package sure_cls_gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/runpath"
	"github.com/yylego/sure"
)

// RecvSample carries a method whose input is named T, exercising the wrapping
// receiving name: a hardcoded `T` recv would collide with that input.
//
// RecvSample 带一个参数名为 T 的方法，用于检验包装方法的接收者命名：
// 若接收者硬编码为 T，就会和该参数撞车。
type RecvSample struct{}

// Convert has an input named T on purpose.
//
// Convert 故意带一个名为 T 的参数。
func (a *RecvSample) Convert(T int) (int, error) { return T, nil }

// TestClassReceiverAvoidsCollision guards that the generated wrapping method recv
// avoids two clashes: it must not match an input name (RecvSample.Convert has an
// input named T, so a bare T recv would collide) and must not match the field name
// (else the method text emits a.a). The recv is kept distinct from both.
//
// TestClassReceiverAvoidsCollision 守护生成的包装方法接收者避开两类冲突：不能等于参数名
// （RecvSample.Convert 带一个名为 T 的参数，裸用 T 接收者会撞车），也不能等于字段名
// （否则方法体出现 a.a）。接收者与二者都保持不同。
func TestClassReceiverAvoidsCollision(t *testing.T) {
	cfg := NewClassGenOptions(runpath.PARENT.Path()).
		WithNewClassNameParts("88").
		WithNamingPatternType(STYLE_SUFFIX_CAMELCASE_TYPE).
		WithReceiverVariableName("a")

	code := GenerateClassWithErrorHandlingMode(cfg, RecvSample{}, sure.MUST)

	require.NotContains(t, code, "func (T *", "the wrapping recv must not collide with the input named T")
	require.NotContains(t, code, "func (a *", "the wrapping recv must diverge from the field name (avoid a.a)")
	require.NotContains(t, code, "a.a.Convert", "the method text must not emit a.a")
	require.Contains(t, code, "Convert(T int)", "the input must get forwarded just the same")
	require.Contains(t, code, "sure.Must(", "the error must be asserted")
}

// ResSample carries a method whose input is named res, exercising the
// synthesized-return-name avoidance: the anonymous int return is auto-named
// "res" via GetSimpleResElements and would collide with that input.
//
// ResSample 带一个参数名为 res 的方法，用于检验合成返回值名的避让：
// 匿名的 int 返回会被 GetSimpleResElements 自动命名为 "res"，与该形参撞车。
type ResSample struct{}

// Get has an input named res on purpose; its anonymous int return would also
// be named res without collision avoidance.
//
// Get 故意带一个名为 res 的参数；若不做避让，匿名的 int 返回也会被命名为 res。
func (a *ResSample) Get(res string) (int, error) { return 0, nil }

// TestClassResultNameAvoidsParamCollision guards that a synthesized return-value
// name does not clash with an input name. Without the guard, ResSample.Get
// emits `func (T *X) Get(res string) (res int) { res, err1 := T.a.Get(res) ... }`
// — the named return res redeclares the param res (and the := collides), which
// does not compile. Mirrors the same protection sure_pkg_gen holds.
//
// TestClassResultNameAvoidsParamCollision 守护合成返回值名不与形参名撞车。
// 若无此保护，ResSample.Get 会生成 `func (T *X) Get(res string) (res int){ res, err1 := T.a.Get(res) ... }`
// —— 具名返回 res 与形参 res 重复声明（且 := 撞车），编译不过。与 sure_pkg_gen 已有的保护一致。
func TestClassResultNameAvoidsParamCollision(t *testing.T) {
	cfg := NewClassGenOptions(runpath.PARENT.Path()).
		WithNewClassNameParts("88").
		WithNamingPatternType(STYLE_SUFFIX_CAMELCASE_TYPE)

	code := GenerateClassWithErrorHandlingMode(cfg, ResSample{}, sure.MUST)

	require.Contains(t, code, "Get(res string)", "the input must get preserved")
	require.NotContains(t, code, "(res int)", "the synthesized return must not reuse the input name res")
	require.Contains(t, code, "res0", "the colliding return name must be renamed")
	require.Contains(t, code, "sure.Must(", "the error must get asserted just the same")
}

// BlankSample carries a method with a blank recv, exercising recv-name
// detection: "_" must not be taken as the wrapping's field name.
//
// BlankSample 带一个空白接收者的方法，用于检验接收者名探测："_" 不能被当成包装类的字段名。
type BlankSample struct{}

// Do uses a blank recv on purpose.
//
// Do 故意使用空白接收者。
func (_ *BlankSample) Do(x int) (int, error) { return x, nil } //nolint:staticcheck // keep blank recv token in source so the test can scan it

// TestClassBlankReceiverFallsBackToT guards that a blank recv "_" on the
// wrapped type does not leak into the generated struct field and entrance.
// Without the guard the field name becomes "_", emitting `struct{ _ *BlankSample }`
// and `&X{_:_}` (using _ as a value) and `T._` — none of which compile. The name
// must drop back to "T".
//
// TestClassBlankReceiverFallsBackToT 守护被包装类型的空白接收者 "_" 不泄漏进生成的结构体字段与构造器。
// 若无此保护，字段名会变成 "_"，生成 `struct{ _ *BlankSample }`、`&X{_:_}`（把 _ 当值）、`T._`——全都编译不过。
// 名字必须回退到 "T"。
func TestClassBlankReceiverFallsBackToT(t *testing.T) {
	cfg := NewClassGenOptions(runpath.PARENT.Path()).
		WithNewClassNameParts("88").
		WithNamingPatternType(STYLE_SUFFIX_CAMELCASE_TYPE)

	code := GenerateClassWithErrorHandlingMode(cfg, BlankSample{}, sure.MUST)

	require.NotContains(t, code, "_ *BlankSample", "a blank recv must not become the struct field")
	require.NotContains(t, code, "{_:_}", "must not emit &X{_:_}")
	require.Contains(t, code, "T *BlankSample", "the recv variable name should drop back to T")
	require.Contains(t, code, "sure.Must(", "the error must be asserted")
}
