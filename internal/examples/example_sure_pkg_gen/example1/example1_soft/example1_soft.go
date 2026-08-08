// Code generated using sure/sure_pkg_gen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/sure
// Generated from: gen_test.go:24 -> example1.TestGenerate
// ========== SURE:DO-NOT-EDIT-SECTION:END ==========

package example1_soft

import (
	"github.com/yylego/sure"
	"github.com/yylego/sure/internal/examples/example_sure_pkg_gen/example1"
)

func GetN() int {
	res0, err := example1.GetN()
	sure.Soft(err)
	return res0
}

func GetS() string {
	res0, err := example1.GetS()
	sure.Soft(err)
	return res0
}
