package condition

import (
	"strconv"
	"strings"
)

func parseHHMM(s string) (int, bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, false
	}
	h := atoi(parts[0])
	m := atoi(parts[1])
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func atoi(s string) int {
	n := 0
	for _, c := range strings.TrimSpace(s) {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func compareOrdered(actual, expected, op string) (bool, error) {
	actualNum, aErr := strconv.ParseFloat(actual, 64)
	expectedNum, eErr := strconv.ParseFloat(expected, 64)
	if aErr == nil && eErr == nil {
		switch op {
		case "gt":
			return actualNum > expectedNum, nil
		case "gte":
			return actualNum >= expectedNum, nil
		case "lt":
			return actualNum < expectedNum, nil
		case "lte":
			return actualNum <= expectedNum, nil
		}
	}
	switch op {
	case "gt":
		return actual > expected, nil
	case "gte":
		return actual >= expected, nil
	case "lt":
		return actual < expected, nil
	case "lte":
		return actual <= expected, nil
	}
	return false, nil
}
