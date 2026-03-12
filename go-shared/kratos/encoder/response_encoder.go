package encoder

import (
	http2 "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
)

func ResponseEncoder(writer http2.ResponseWriter, request *http2.Request, a any) error {
	codec, _ := http.CodecForRequest(request, "Accept")

	reply := &Reply{
		T:       time.Now().UnixMilli(),
		Success: true,
		Msg:     "OK",
		Data:    a,
	}

	data, err := codec.Marshal(reply)
	if err != nil {
		return err
	}

	writer.Header().Set("Content-Type", "application/"+codec.Name())
	writer.WriteHeader(http2.StatusOK)
	_, err = writer.Write(data)
	return err
}
