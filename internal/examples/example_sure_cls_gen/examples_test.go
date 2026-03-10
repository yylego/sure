package example_sure_cls_gen

import (
	"testing"

	"github.com/yylego/sure/internal/examples/example_sure_cls_gen/example1"
	"github.com/yylego/sure/internal/examples/example_sure_cls_gen/example2"
	"github.com/yylego/sure/internal/examples/example_sure_cls_gen/example3"
)

func TestExample1(t *testing.T) {
	//当你要操作的是个对象时，你就可以赋予这个对象 Must 和 Soft 的能力，Must表示出错时崩溃，而Soft表示出错是仅仅告警但流程继续
	a := example1.NewExample(1, "s")
	t.Log(a.Must().GetN())
	t.Log(a.Must().GetS())

	t.Log(a.Soft().GetN())
	t.Log(a.Soft().GetS())
}

func TestExample2(t *testing.T) {
	//当你要操作的是个对象时，你就可以赋予这个对象 Must 和 Soft 的能力，Must表示出错时崩溃，而Soft表示出错是仅仅告警但流程继续
	a := example2.NewExample(1, "s")
	ma := a.Must()
	sa := a.Soft()

	t.Log(ma.GetN())
	t.Log(ma.GetS())

	t.Log(sa.GetN())
	t.Log(sa.GetS())

	b := example2.NewExample(1, "s")
	mb := b.Must()
	sb := b.Soft()

	t.Log(mb.GetN())
	t.Log(mb.GetS())

	t.Log(sb.GetN())
	t.Log(sb.GetS())
}

func TestExample3(t *testing.T) {
	//当你要操作的是个对象时，你就可以赋予这个对象 Must 和 Soft 的能力，Must表示出错时崩溃，而Soft表示出错是仅仅告警但流程继续
	a := example3.NewExample(1, "s")
	ma := a.Must()
	sa := a.Soft()

	t.Log(ma.GetN())
	t.Log(ma.GetS())

	t.Log(sa.GetN())
	t.Log(sa.GetS())

	b := example3.NewExample(1, "s")
	mb := b.Must()
	sb := b.Soft()

	t.Log(mb.GetN())
	t.Log(mb.GetS())

	t.Log(sb.GetN())
	t.Log(sb.GetS())
}
