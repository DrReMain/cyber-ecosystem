package alpha

import "time"

func stamp() int64 { return time.Now().UnixNano() }
