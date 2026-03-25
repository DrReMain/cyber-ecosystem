package encoder

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewErrorEncoder(buildBody func(context.Context, error, *errors.Error) any) khttp.EncodeErrorFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		se := errors.FromError(err)
		codec, _ := khttp.CodecForRequest(r, "Accept")
		body, err := codec.Marshal(buildBody(r.Context(), err, se))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/"+codec.Name())
		w.WriteHeader(int(se.Code))
		_, _ = w.Write(body)
	}
}
