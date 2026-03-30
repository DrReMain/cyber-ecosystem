package cache

import (
	"fmt"
	"time"
)

func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key is empty", ErrInvalidArgument)
	}
	return nil
}

func ValidateKeys(keys ...string) error {
	for i, key := range keys {
		if key == "" {
			return fmt.Errorf("%w: keys[%d] is empty", ErrInvalidArgument, i)
		}
	}
	return nil
}

func ValidatePairs(pairs map[string][]byte) error {
	for key := range pairs {
		if key == "" {
			return fmt.Errorf("%w: pairs contains empty key", ErrInvalidArgument)
		}
	}
	return nil
}

func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is empty", ErrInvalidArgument)
	}
	return nil
}

func ValidateMember(member string) error {
	if member == "" {
		return fmt.Errorf("%w: member is empty", ErrInvalidArgument)
	}
	return nil
}

func ValidateMembers(members []Member) error {
	for i, m := range members {
		if m.Member == "" {
			return fmt.Errorf("%w: members[%d].member is empty", ErrInvalidArgument, i)
		}
	}
	return nil
}

func ValidateMemberStrings(members ...string) error {
	for i, member := range members {
		if member == "" {
			return fmt.Errorf("%w: members[%d] is empty", ErrInvalidArgument, i)
		}
	}
	return nil
}

func ValidateRateLimit(key string, quota int64, window time.Duration) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	if quota <= 0 {
		return fmt.Errorf("%w: quota must be > 0", ErrInvalidArgument)
	}
	if window <= 0 {
		return fmt.Errorf("%w: window must be > 0", ErrInvalidArgument)
	}
	return nil
}
