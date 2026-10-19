package helper

// MaskAction performs a conditional callback based on Condition.
type MaskAction struct {
	Condition bool
	OnTrue    func()
	OnFalse   func()
}

// Handler dispatches MaskActions by field name.
type Handler map[string]MaskAction

// Emit executes the matching actions for the given field names.
func (mh Handler) Emit(fieldsMask []string) {
	for _, v := range fieldsMask {
		if action, ok := mh[v]; ok {
			if action.Condition {
				action.OnTrue()
			} else {
				action.OnFalse()
			}
		}
	}
}
