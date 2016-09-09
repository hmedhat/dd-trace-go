package tracer

import (
	"sync"

	log "github.com/cihub/seelog"
)

const (
	defaultDeliveryURL = "http://localhost:7777/spans"
)

// Transport interface to Send spans to the given URL
type Transport interface {
	Send(url, header string, spans []*Span) error
}

// Tracer is the common struct we use to collect, buffer
type Tracer struct {
	Transport      Transport  // is the transport mechanism used to delivery spans to the agent
	outgoingPacket chan *Span // the channel that sends the Span into the sending pipeline

	// A WaitGroup tracks the current status of the message
	// pipeline so that at any time the Tracer and the internal
	// Worker may know if there are messages that are not flushed.
	// The intent is to use it with the tracer.Wait() API to assure that
	// all messages have been transported before exiting the process.
	wg sync.WaitGroup
}

// NewTracer returns a Tracer instance that owns a span delivery system. Each Tracer starts
// a new go routing that handles the delivery. It's safe to create new tracers, but it's
// advisable only if the default client doesn't fit your needs.
// TODO: make possible to create a Tracer with a different Transport system
func NewTracer() *Tracer {
	return &Tracer{
		Transport: NewHTTPTransport(defaultDeliveryURL),
	}
}

// Trace creates a new root Span with a random identifier. This high-level API is commonly
// used to start a new tracing session.
func (t *Tracer) Trace(service, name, resource string) *Span {
	// this check detects if this is the first time we are using this tracer;
	// in that case, initialize the outgoing channel and start a background
	// worker that manages spans delivery
	if t.outgoingPacket == nil {
		t.outgoingPacket = make(chan *Span)
		go t.worker()
	}

	// create and return the Span
	spanID := nextSpanID()
	return NewSpan(spanID, spanID, 0, service, name, resource, t.outgoingPacket)
}

// Wait for the messages delivery. This method assures that all messages have been
// delivered before exiting the process.
func (t *Tracer) Wait() {
	t.wg.Wait()
}

// Background worker that handles data delivery through the Transport instance
func (t *Tracer) worker() {
	for span := range t.outgoingPacket {
		t.wg.Add(1)
		log.Infof("Working on span %d ", span.SpanID)
		// TODO: marshall and/or send it somewhere else. The fact is that
		// when the Span is sent, we should call the t.wg.Done()
	}
}

// HTTPTransport provides the default implementation to send the span list using
// a HTTP/TCP connection. The transport expects to know which is the delivery URL.
// TODO: the *http implementation is missing
type HTTPTransport struct {
	URL string // the delivery URL
}

// NewHTTPTransport creates a new delivery instance that honors the Transport interface.
// This function may be useful to send data to an agent available in a remote location.
func NewHTTPTransport(url string) *HTTPTransport {
	return &HTTPTransport{
		URL: url,
	}
}

// Send is the implementation of the Transport interface and hosts the logic to send the
// spans list to a local/remote agent.
func (t *HTTPTransport) Send(url, header string, spans []*Span) error {
	if url == "" {
		return nil
	}

	// TODO: do something

	return nil
}

// DefaultTracer is a default *Tracer instance
var DefaultTracer = NewTracer()

// Trace is an helper function that is used to create a RootSpan, through
// the DefaultTracer client. If the default client doesn't fit your needs,
// you can create a new Tracer through the NewTracer function.
func Trace(service, name, resource string) *Span {
	return DefaultTracer.Trace(service, name, resource)
}

// Wait helper function that waits for the message delivery of the
// DefaultClient.
func Wait() {
	DefaultTracer.Wait()
}
