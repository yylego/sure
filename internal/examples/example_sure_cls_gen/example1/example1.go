package example1

type Example struct {
	n int
	s string
}

func NewExample(n int, s string) *Example {
	return &Example{
		n: n,
		s: s,
	}
}

func (a *Example) GetN() (int, error) {
	return a.n, nil
}

func (a *Example) GetS() (string, error) {
	return a.s, nil
}

func (a *Example) Sum(nums ...int) (int, error) { return 0, nil }

func (a *Example) Items() ([]int, error) { return nil, nil }

func (a *Example) Both() (int, string, error) { return 0, "", nil }

func (a *Example) Noop() {}

func (a *Example) Check() error { return nil }
