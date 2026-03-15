package encoder

import (
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewResponseEncoder(buildBody func(any) (any, error)) khttp.EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v any) error {
		if v == nil {
			return nil
		}
		if rd, ok := v.(khttp.Redirector); ok {
			url, code := rd.Redirect()
			http.Redirect(w, r, url, code)
			return nil
		}
		codec, _ := khttp.CodecForRequest(r, "Accept")
		vv, err := buildBody(v)
		if err != nil {
			return err
		}
		data, err := codec.Marshal(vv)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/"+codec.Name())
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		return nil
	}

}
