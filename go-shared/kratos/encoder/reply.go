package encoder

type Reply struct {
	T       int64  `json:"t"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    any    `json:"data"`
}
