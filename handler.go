package syslog

// Handler handles syslog messages
type Handler interface {
	// Handle should return [Message] (maybe modified) for future processing by
	// other handlers, or return nil. If Handle is called with nil message it
	// should complete all remaining work and properly shutdown before returning.
	Handle(*Message) *Message
}

// BaseHandler is designed to simplify the creation of real handlers. It
// implements [Handler] interface using nonblocking queuing of messages and
// simple message filtering.
type BaseHandler struct {
	queue       chan *Message
	end         chan struct{}
	discardFunc func(*Message) bool
	ft          bool // when true, always chain every message to the next handler
}

// NewBaseHandler creates [BaseHandler] using specified discardFunc. If discardFunc is nil
// or if it returns true, messages are passed to [BaseHandler] internal queue
// (of qlen length). If discardFunc returns false or ft is true, messages are returned
// to server for future processing by other handlers.
func NewBaseHandler(qlen int, filter func(*Message) bool, ft bool) *BaseHandler {
	return &BaseHandler{
		queue:       make(chan *Message, qlen),
		end:         make(chan struct{}),
		discardFunc: filter,
		ft:          ft,
	}
}

// Handle inserts m in an internal queue. It immediately returns even if
// the queue is full. If m == nil, it closes the queue and waits for
// [BaseHandler.End] method call before returning.
func (h *BaseHandler) Handle(m *Message) *Message {
	if m == nil {
		close(h.queue) // signal that there are no more messages for processing
		<-h.end        // wait for handler shutdown
		return nil
	}

	if h.discardFunc != nil && !h.discardFunc(m) {
		// m doesn't match the discardFunc
		return m
	}

	// Try queue m
	select {
	case h.queue <- m:
	default:
	}

	if h.ft {
		return m
	}
	return nil
}

// Get returns first message from internal queue. It waits for message if queue
// is empty. It returns nil if there are no more messages to process and the handler
// should shut down.
func (h *BaseHandler) Get() *Message {
	m, ok := <-h.queue
	if ok {
		return m
	}
	return nil
}

// Queue returns the [BaseHandler] internal queue as a read-only channel. You can use
// it directly, especially if your handler need to select from multiple channels
// or have to work without blocking. You need to check whether this channel has been
// closed by the sender and properly shut down in this case.
func (h *BaseHandler) Queue() <-chan *Message {
	return h.queue
}

// End signals to the server that handler has properly shut down. You need to call End
// only if [BaseHandler.Get] has returned nil before.
func (h *BaseHandler) End() {
	close(h.end)
}
