package sms

const (
	// maxPhoneLen is the E.164 ceiling (15 digits) plus headroom for a
	// leading plus; which market a well-formed number belongs to is
	// provider knowledge the backend enforces against the channel.
	maxPhoneLen      = 20
	maxSignLen       = 64
	maxTemplateLen   = 64
	maxParams        = 20
	maxParamKeyLen   = 32
	maxParamValueLen = 256
)

// isPhone reports whether s is a phone-shaped string: an optional leading
// plus, then digits only, at least 5. Charset and length only — market
// validity belongs to the channel that will carry it.
func isPhone(s string) bool {
	if s == "" || len(s) > maxPhoneLen {
		return false
	}
	digits := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == '+' && i == 0:
		default:
			return false
		}
	}
	return digits >= 5
}

// ValidateMessage checks the channel-neutral single-request shape: phone
// present and phone-shaped, sign and template present (the templated-SMS
// vocabulary this family carries), params bounded in count and size.
// Whether the template exists and whether the phone is deliverable are
// provider state — the channel answers those, never a shape check.
func ValidateMessage(msg *Message) error {
	if msg == nil {
		return ErrInvalidArgument
	}
	if !isPhone(msg.To) {
		return ErrInvalidArgument
	}
	if msg.SignName == "" || len(msg.SignName) > maxSignLen {
		return ErrInvalidArgument
	}
	if msg.TemplateCode == "" || len(msg.TemplateCode) > maxTemplateLen {
		return ErrInvalidArgument
	}
	if len(msg.Params) > maxParams {
		return ErrInvalidArgument
	}
	for k, v := range msg.Params {
		if k == "" || len(k) > maxParamKeyLen || len(v) > maxParamValueLen {
			return ErrInvalidArgument
		}
	}
	return nil
}
