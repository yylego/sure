// Code generated using sure/sure_cls_gen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/sure
// Generated from: gen_test.go:25 -> example1.TestGenerate
// ========== SURE:DO-NOT-EDIT-SECTION:END ==========

package example1

import "github.com/yylego/sure"

type Example88Must struct{ a *Example }

func (a *Example) Must() *Example88Must {
	return &Example88Must{a: a}
}
func (T *Example88Must) GetN() (res int) {
	res, err1 := T.a.GetN()
	sure.Must(err1)
	return res
}
func (T *Example88Must) GetS() (res string) {
	res, err1 := T.a.GetS()
	sure.Must(err1)
	return res
}

type Example88Soft struct{ a *Example }

func (a *Example) Soft() *Example88Soft {
	return &Example88Soft{a: a}
}
func (T *Example88Soft) GetN() (res int) {
	res, err1 := T.a.GetN()
	sure.Soft(err1)
	return res
}
func (T *Example88Soft) GetS() (res string) {
	res, err1 := T.a.GetS()
	sure.Soft(err1)
	return res
}

type Example88Omit struct{ a *Example }

func (a *Example) Omit() *Example88Omit {
	return &Example88Omit{a: a}
}
func (T *Example88Omit) GetN() (res int) {
	res, err1 := T.a.GetN()
	sure.Omit(err1)
	return res
}
func (T *Example88Omit) GetS() (res string) {
	res, err1 := T.a.GetS()
	sure.Omit(err1)
	return res
}
