package masks

type MaskAction struct {
	Condition bool
	OnTrue    func()
	OnFalse   func()
}

type Handler map[string]MaskAction

// Emit executes the mask actions for the given fields
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
