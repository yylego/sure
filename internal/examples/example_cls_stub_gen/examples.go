package example_cls_stub_gen

import (
	"math/rand/v2"

	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example1/example1_stub1"
	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example1/example1_stub2"
	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example2/example2_stub1"
	"github.com/yylego/sure/internal/examples/example_cls_stub_gen/example2/example2_stub2"
)

func Example0x1() int {
	return example1_stub1.Add(rand.IntN(100), rand.IntN(100))
}

func Example0x2() int {
	return example1_stub2.Add(rand.IntN(100), rand.IntN(100))
}

func Example6x1() string {
	return example2_stub1.Name()
}

func Example6x2() string {
	return example2_stub2.Name()
}
