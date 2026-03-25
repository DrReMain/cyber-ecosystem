package locales

import "embed"

// FS contains locale resources owned by template2 service.
//
//go:embed *.json
var FS embed.FS
