package mq

const (
	maxTopicLen = 249 // NATS subject element safe bound
	maxGroupLen = 120 // durable-name safe bound
)

// isSubjectToken reports whether s contains only bytes that are safe both as a
// NATS subject literal and as a stream/consumer-name fragment. The stream name
// is "mq-"+topic and the durable name is group+"-"+topic, so '.', '>', '*',
// '/', whitespace and control chars are rejected: they act as NATS
// wildcards/separators or are illegal in stream/consumer names.
func isSubjectToken(s string) bool {
	for _, r := range s {
		switch {
		case 'a' <= r && r <= 'z',
			'A' <= r && r <= 'Z',
			'0' <= r && r <= '9',
			r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// ValidateTopic rejects empty / over-long topics and bytes unsafe as a NATS
// subject literal or stream-name fragment. Topics are plain subject names (no
// wildcards): MQ maps one topic → one stream + one subject.
func ValidateTopic(topic string) error {
	if topic == "" || len(topic) > maxTopicLen || !isSubjectToken(topic) {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateGroup rejects empty / over-long group names and bytes unsafe as a
// durable-name fragment (group → durable name).
func ValidateGroup(group string) error {
	if group == "" || len(group) > maxGroupLen || !isSubjectToken(group) {
		return ErrInvalidArgument
	}
	return nil
}
