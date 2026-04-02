package i18n

//go:generate go run -mod=mod ../../../../../../shared-go/kratos/middleware/i18n/cmd/geni18n ../../../../api/v1/error_reason.proto ./translations -locale=en-US,zh-CN,ja-JP
