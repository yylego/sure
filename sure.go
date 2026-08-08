// Package sure provides exception handling mode configuration with assertion and crash handling
// Offers three exception handling modes: Must (crash), Soft (warn), and Omit (ignore)
// Enables auto assertion and crash handling to enhance exception management in code
// Supports clean exception handling patterns without repetitive validation checks
//
// Package sure 提供异常处理模式配置，带有断言和崩溃处理
// 提供三种异常处理模式：Must（崩溃）、Soft（警告）和 Omit（忽略）
// 支持自动断言和崩溃处理，增强代码中的异常管理
// 实现简洁的异常处理模式，无需重复的手动检查
package sure

import (
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// ErrorHandlingMode defines how the system responds when exceptions are encountered
// Determines the response: crash on exceptions, emit warnings, and ignore without logs
// Used to generate code with different exception handling strategies and patterns
//
// ErrorHandlingMode 定义系统遇到异常时如何响应
// 确定响应方式：遇异常崩溃、发出警告以及无日志忽略
// 用于生成具有不同异常处理策略和模式的代码
type ErrorHandlingMode string

//goland:noinspection GoSnakeCaseUsage
const (
	MUST ErrorHandlingMode = "Must" // Hard mode: crashes on exceptions // 硬模式：遇异常崩溃
	SOFT ErrorHandlingMode = "Soft" // Soft mode: warns on exceptions // 软模式：遇异常警告
	OMIT ErrorHandlingMode = "Omit" // Omit mode: ignores exceptions // 忽略模式：忽略异常
)

// ErrorHandlingModes returns each mode sure states, in the sequence above
//
// The codegens name a mode as a method on the wrapped type, so these need the whole set to split a generated entrance apart from a source method
// The set a given run wants does not settle that, since a past run could have written a mode this run leaves out
// Returns a fresh slice, so callers sorting / trimming it leave the next one's duplicate alone
//
// ErrorHandlingModes 返回 sure 声明的全部模式，顺序与上面一致。
//
// 生成器把模式写成被包装类型上的一个方法，所以要认出"哪个方法是生成出来的入口"，得拿全集来比
// 本轮想生成的那几个回答不了这个问题，因为上一轮完全可能写下一个本轮不再要的模式
// 返回的是新切片，调用方拿去排序或截断都不会动到下一个调用方那份
func ErrorHandlingModes() []ErrorHandlingMode {
	return []ErrorHandlingMode{MUST, SOFT, OMIT}
}

// Must crashes the program when an exception occurs
// Stops execution with panic when exception is non-nil
// Used when exceptions are unacceptable and program should stop
//
// Must 当发生异常时使程序崩溃
// 当异常非 nil 时使用 panic 停止执行
// 用于异常不可接受且程序应停止的情况
func Must(err error) {
	if err != nil {
		zaplog.LOGS.Skip1.Panic("must", zap.Error(err))
	}
}

// Soft logs a warning when an exception occurs but continues execution
// Prints warning message and enables program to continue
// Used when exceptions should be noted but not stop program flow
//
// Soft 当发生异常时记录警告但继续执行
// 打印警告消息并允许程序继续
// 用于异常应该注意但不应停止程序流程的情况
func Soft(err error) {
	if err != nil {
		zaplog.LOGS.Skip1.Warn("soft", zap.Error(err))
	}
}

// Omit silences exceptions without logging and without crashing
// Ignores exceptions and continues execution without action
// Used when exceptions can be safe to ignore
//
// Omit 静默异常，不记录且不崩溃
// 忽略异常并继续执行而不采取行动
// 用于可以安全忽略异常的情况
func Omit(err error) {
	if err != nil {
		_ = err // Just ignore the exception // 仅忽略异常
	}
}
