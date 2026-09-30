package audit

func ActionOfForTest(operation string) int32 {
	return int32(actionOf(operation))
}

func ReadVerbForTest(method string) bool {
	return readVerb(method)
}

func DenyReasonOfForTest(err error) string {
	return denyReasonOf(err)
}
