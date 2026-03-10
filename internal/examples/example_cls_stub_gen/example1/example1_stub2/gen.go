// Code generated using sure/cls_stub_gen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/sure
// Generated from: gen_test.go:24 -> example1_stub2.TestGenerate
// ========== SURE:DO-NOT-EDIT-SECTION:END ==========

package example1_stub2

import "github.com/yylego/sure/internal/examples/example_cls_stub_gen/example1"

func Get() *example1.A {
	return example1.STUB2.Get()
}
func Set(arg string) {
	example1.STUB2.Set(arg)
}
func Add(x int, y int) int {
	return example1.STUB2.Add(x, y)
}
func Sub(x int, y int) (int, error) {
	return example1.STUB2.Sub(x, y)
}
func Who(param ...example1.Param) {
	example1.STUB2.Who(param...)
}
func How(param ...*example1.Param) {
	example1.STUB2.How(param...)
}
