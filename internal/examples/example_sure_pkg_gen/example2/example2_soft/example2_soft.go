// Code generated using sure/sure_pkg_gen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/sure
// Generated from: gen_test.go:26 -> example2.TestGenerate
// ========== SURE:DO-NOT-EDIT-SECTION:END ==========

package example2_soft

import (
	"github.com/yylego/sure"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example2"
)

func Neat[T any](a T) []byte {
	res0, err := example2.Neat[T](a)
	sure.Soft(err)
	return res0
}

func Bind[T any](data []byte) *T {
	res0, err := example2.Bind[T](data)
	sure.Soft(err)
	return res0
}
