- error message
    - panic (recovery.Recovery()) (nil:UNKNOWN - unknown request error)
    - marshal (proto: CODEC - body unmarshal proto .....)
    - validate (validate.ProtoValidate(validate.UseProtoMessage))
    - ent (EntErrorHandler)
    - biz errors
    - grpc connect (last connection error: connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9001: connect: connection refused\")

# 统一返回 + i18n

利用kratos的errors机制，系统内返回的错误一律需要在proto中定义错误枚举，
并使用枚举字符串作为i18n的key，这样可以避免硬编码魔法字符串。
这样无论是本地语言文件或远程语言文件，都可以以proto作为契约。
这样实际返回给客户端的信息来自于Reason，而message就作为系统内日志信息，或直接留空不使用。
然后实现返回结构兼容kratos Error机制。
且该返回结构需要在跨服务调用中避免被重复包装，
返回结构中增加充分的字段，其中需要通过request识别需要的语言，对Reason进行语言转换。
