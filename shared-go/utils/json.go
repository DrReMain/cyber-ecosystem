package utils

import (
	"encoding/json"
	"fmt"
)

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func MustMarshal(v any) []byte {
	b, err := Marshal(v)
	if err != nil {
		panic(fmt.Errorf("utils.MustMarshal: %w", err))
	}
	return b
}

func Unmarshal[T any](data []byte) (T, error) {
	var v T
	return v, json.Unmarshal(data, &v)
}
