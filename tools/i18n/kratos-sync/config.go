package main

import (
	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	template2V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template2/v1"
)

var languages = []string{"zh-Hans", "en"}

var services = map[string]struct {
	ReasonName map[int32]string
	Dir        string
}{
	"template1": {
		template1V1.ErrorReason_name,
		"examples/template1/internal/locales",
	},
	"template2": {
		template2V1.ErrorReason_name,
		"examples/template2/internal/locales",
	},
}
