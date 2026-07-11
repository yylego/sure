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
func (T *Example88Must) Sum(nums ...int) (res int) {
	res, err1 := T.a.Sum(nums...)
	sure.Must(err1)
	return res
}
func (T *Example88Must) Items() (res []int) {
	res, err1 := T.a.Items()
	sure.Must(err1)
	return res
}
func (T *Example88Must) Both() (res int, res1 string) {
	res, res1, err2 := T.a.Both()
	sure.Must(err2)
	return res, res1
}
func (T *Example88Must) Noop() {
	T.a.Noop()
}
func (T *Example88Must) Check() {
	err := T.a.Check()
	sure.Must(err)
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
func (T *Example88Soft) Sum(nums ...int) (res int) {
	res, err1 := T.a.Sum(nums...)
	sure.Soft(err1)
	return res
}
func (T *Example88Soft) Items() (res []int) {
	res, err1 := T.a.Items()
	sure.Soft(err1)
	return res
}
func (T *Example88Soft) Both() (res int, res1 string) {
	res, res1, err2 := T.a.Both()
	sure.Soft(err2)
	return res, res1
}
func (T *Example88Soft) Noop() {
	T.a.Noop()
}
func (T *Example88Soft) Check() {
	err := T.a.Check()
	sure.Soft(err)
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
func (T *Example88Omit) Sum(nums ...int) (res int) {
	res, err1 := T.a.Sum(nums...)
	sure.Omit(err1)
	return res
}
func (T *Example88Omit) Items() (res []int) {
	res, err1 := T.a.Items()
	sure.Omit(err1)
	return res
}
func (T *Example88Omit) Both() (res int, res1 string) {
	res, res1, err2 := T.a.Both()
	sure.Omit(err2)
	return res, res1
}
func (T *Example88Omit) Noop() {
	T.a.Noop()
}
func (T *Example88Omit) Check() {
	err := T.a.Check()
	sure.Omit(err)
}
