package cache

// OperationArgs holds detailed arguments for tracing cache operations.
// This is shared by both memory and redis implementations.
type OperationArgs struct {
	Key    string
	Keys   []string
	Value  []byte
	Values map[string][]byte
	// Session related
	SessionID string
	// SortedSet related
	Member  string
	Members []Member
	Score   float64
	Delta   float64
	// Counter related
	DeltaInt int64
	// RateLimiter related
	Quota  int64
	Window int64 // milliseconds
}
