package utils

import (
	"encoding/json"
	"fmt"
)

func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("utils.MustMarshal: %w", err))
	}
	return b
}

func Unmarshal[T any](data []byte) (T, error) {
	var v T
	return v, json.Unmarshal(data, &v)
}
