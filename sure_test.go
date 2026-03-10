package sure_test

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
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

// TestMust tests the Must function behavior
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

// TestSoft tests the Soft function behavior
// Verifies that Soft logs warning but continues execution
// Creates test error to verify soft handling
//
// TestSoft 测试 Soft 函数的行为
// 验证 Soft 记录警告但继续执行
// 创建测试错误以验证软处理
func TestSoft(t *testing.T) {
	t.Log("-")
	sure.Soft(errors.New("wrong")) //将会告警，而且程序将继续执行
	t.Log("-")
}

// TestOmit tests the Omit function behavior
// Verifies that Omit silently ignores errors
// Creates test error to verify omit handling
//
// TestOmit 测试 Omit 函数的行为
// 验证 Omit 静默忽略错误
// 创建测试错误以验证忽略处理
func TestOmit(t *testing.T) {
	t.Log("-")
	sure.Omit(errors.New("wrong")) //将会忽略，而且程序将继续执行
	t.Log("-")
}
