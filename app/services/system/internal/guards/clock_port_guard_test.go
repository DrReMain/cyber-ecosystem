package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/guardkit"
)

// Clock-port guard: a biz.go that reads the wall clock bare cannot run
// under fakes — the detector and the exemption note live in
// guardkit.AssertBizClockPort.
func TestBizClockReadsGoThroughThePort(t *testing.T) {
	guardkit.AssertBizClockPort(t, "..")
}
