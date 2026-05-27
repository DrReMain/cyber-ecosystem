package i18n

//go:generate go run -mod=mod ../../../../../../shared-go/kratos/middleware/i18n/cmd/geni18n -protos=i18n.protos -name=v1 ./locales -locale=en-US,zh-CN,ar-SA
