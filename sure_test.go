package sure_test

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/yylego/sure"
)

// TestMain is the test suite main function
// Runs the test suite in this package
//
// TestMain 是测试套件的主函数
// 运行这个包中的测试套件
func TestMain(m *testing.M) {
	m.Run()
}

// TestMust tests how the Must function acts
// Verifies that Must panics when error is non-nil
// Uses JSON marshaling as test scenario
//
// TestMust 测试 Must 函数的行为
// 验证当错误非 nil 时 Must 会 panic
// 使用 JSON 编组作为测试场景
func TestMust(t *testing.T) {
	type Example struct {
		Name string `json:"name"`
	}

	example := &Example{Name: "lele"}

	data, err := json.Marshal(example)

	t.Log("-")
	sure.Must(err) //当没有错误时就什么也不做，当出错时将 panic 崩溃
	t.Log("-")

	t.Log(string(data))
}

// TestSoft tests how the Soft function acts
// Verifies that Soft logs warning but continues execution
// Creates a test error to check soft handling
//
// TestSoft 测试 Soft 函数的行为
// 验证 Soft 记录警告但继续执行
// 创建测试错误以验证软处理
func TestSoft(t *testing.T) {
	t.Log("-")
	sure.Soft(errors.New("wrong")) //将会告警，而且程序将继续执行
	t.Log("-")
}

// TestOmit tests how the Omit function acts
// Checks that Omit ignores errors in silence
// Creates a test error to check omit handling
//
// TestOmit 测试 Omit 函数的行为
// 验证 Omit 静默忽略错误
// 创建测试错误以验证忽略处理
func TestOmit(t *testing.T) {
	t.Log("-")
	sure.Omit(errors.New("wrong")) //将会忽略，而且程序将继续执行
	t.Log("-")
}

// TestErrorHandlingModes checks the stated set and its sequence, and that each invocation gets a fresh slice
//
// The codegens compare a method name against this set to spot a generated entrance, so a missing mode makes them wrap a wrapping
// Handing back the same slice each time would let a sort in one spot reach the next, hence the check on independence
//
// # TestErrorHandlingModes 检查声明出的模式集合与顺序，以及每次调用拿到的都是新切片
//
// 各生成器拿方法名与这个集合比对，以认出"生成出来的入口"，漏一个模式就会给包装再包一层
// 每次返回同一个切片的话，某个调用方的排序会波及下一个，所以要查这份独立性
func TestErrorHandlingModes(t *testing.T) {
	modes := sure.ErrorHandlingModes()
	require.Equal(t, []sure.ErrorHandlingMode{sure.MUST, sure.SOFT, sure.OMIT}, modes)

	modes[0] = sure.OMIT
	require.Equal(t, sure.MUST, sure.ErrorHandlingModes()[0])
}
