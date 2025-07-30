package syslog

// Filter is a predicate function for messages.
type Filter func(*Message) bool

// everything is a no-op filter.
func everything(*Message) bool { return true }

// All combines filters so that all must accept a message for it to be accepted overall.
func All(fs ...Filter) Filter {
	return func(m *Message) bool {
		for _, f := range fs {
			if !f(m) {
				return false
			}
		}
		return true
	}
}

// Any combines filters so that any can accept a message for it to be accepted overall.
// The message is only rejected if all filters reject it.
func Any(fs ...Filter) Filter {
	return func(m *Message) bool {
		for _, f := range fs {
			if f(m) {
				return true
			}
		}
		return false
	}
}
