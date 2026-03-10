package example2

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/yylego/sure/internal/utils"
)

func Neat[T any](a T) ([]byte, error) {
	return utils.NeatBytes(a)
}

func Bind[T any](data []byte) (*T, error) {
	var a T
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, errors.WithMessage(err, "wrong")
	}
	return &a, nil
}
