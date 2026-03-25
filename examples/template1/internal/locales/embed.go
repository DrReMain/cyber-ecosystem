package locales

import "embed"

// FS contains locale resources owned by template1 service.
//
//go:embed *.json
var FS embed.FS
