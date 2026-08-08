package utils

import (
	"github.com/yylego/printgo"
)

// NewPTX creates a new print text building instance
// Returns printgo.PTX to build code strings
// Used to construct generated code one line at a time
//
// NewPTX 创建新的打印文本构建实例
// 返回 printgo.PTX 构建代码字符串
// 用于逐行构建生成的代码
func NewPTX() *printgo.PTX {
	return printgo.NewPTX()
}
