package example_sure_pkg_gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example1/example1_must"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example1/example1_soft"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example2/example2_must"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example2/example2_soft"
)

func TestExample2(t *testing.T) {
	//但是假如你要操作的是个包，你就可以在这个包的基础上衍生出 must 包 和 soft 包，能让你在调用时避免判断错误
	t.Log(example1_must.GetN())
	t.Log(example1_must.GetS())

	t.Log(example1_soft.GetN())
	t.Log(example1_soft.GetS())
}

func TestExample3(t *testing.T) {
	//在泛型的情况下依然是可以的，我想这正是我们需要的，特别是下面这俩函数，把对象转化为json和把json转化为对象
	type Param struct {
		Name string
	}

	{ //使用 soft 软包把对象转json，再用 must 硬包把json转为对象
		data := example2_soft.Neat(&Param{Name: "haha"})
		t.Log(string(data))

		resX := example2_must.Bind[Param](data)
		require.NotNil(t, resX) //其实这块是不需要判断的，毕竟是must的，保证是成功的
		require.Equal(t, "haha", resX.Name)
	}

	{ //当然也可以反过来用，也可以都用 soft 软的也可以都用 must 硬的，这里只是演示，根据场景自由选择
		data := example2_must.Neat(&Param{Name: "haha"})
		t.Log(string(data))

		resX := example2_soft.Bind[Param](data)
		require.NotNil(t, resX) //这块是需要判读的，毕竟是soft的，仅仅是忽略错误发出告警但流程继续执行
		require.Equal(t, "haha", resX.Name)
	}
}
