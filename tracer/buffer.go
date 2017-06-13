package tracer

import (
	"fmt"
	"sync"
)

type traceBuffer struct {
	// spans is a buffer containing all the spans for this trace.
	// The reason we don't use a channel here, is we regularly need
	// to walk the array to find out if it's done or not.
	spans   []*Span
	maxSize int

	traceChan chan<- []*Span
	errChan   chan<- error

	sync.RWMutex
}

func newTraceBuffer(traceChan chan<- []*Span, errChan chan<- error) *traceBuffer {
	return &traceBuffer{
		traceChan: traceChan,
		errChan:   errChan,
	}
}

func (tb *traceBuffer) push(span *Span) {
	tb.Lock()
	defer tb.Unlock()

	// if buffer is full, forget span
	if len(tb.spans) >= tb.maxSize {
		tb.errChan <- fmt.Errorf("[TODO:christian] exceed buffer size")
		return
	}
	// if there's a trace ID mismatch, ignore span
	if len(tb.spans) > 0 && tb.spans[0].TraceID != span.TraceID {
		tb.errChan <- fmt.Errorf("[TODO:christian] trace ID mismatch")
		return
	}

	tb.spans = append(tb.spans, span)
}

func (tb *traceBuffer) Push(span *Span) {
	if tb == nil {
		return
	}
	tb.push(span)
}

func (tb *traceBuffer) flushable() bool {
	tb.RLock()
	defer tb.RUnlock()

	if len(tb.spans) == 0 {
		return false
	}

	for _, span := range tb.spans {
		span.RLock()
		finished := span.finished
		span.RUnlock()

		// A note about performance: it can seem a performance killer
		// to range over all spans each time we finish a span (flush should
		// be called whenever a span is finished) but... by design the
		// first span (index 0) is the root span, and most of the time
		// it's the last one being finished. So in 99% of cases, this
		// is going to return false at the first iteration.
		if !finished {
			return false
		}
	}

	return true
}

func (tb *traceBuffer) flush() {
	if !tb.flushable() {
		return
	}

	tb.Lock()
	defer tb.Unlock()

	tb.traceChan <- tb.spans
	tb.spans = nil
}

func (tb *traceBuffer) Flush() {
	if tb == nil {
		return
	}
	tb.flush()
}
