package encoder

import (
	http2 "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func ErrorEncoder(writer http2.ResponseWriter, request *http2.Request, err error) {
	se := errors.FromError(err)
	codec, _ := http.CodecForRequest(request, "Accept")

	reply := &Reply{
		T:       time.Now().UnixMilli(),
		Success: false,
		Msg:     se.Message,
		Data:    nil,
	}

	data, err := codec.Marshal(reply)
	if err != nil {
		writer.WriteHeader(http2.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/"+codec.Name())
	writer.WriteHeader(int(se.Code))
	_, _ = writer.Write(data)
}
