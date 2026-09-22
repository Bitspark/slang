package core

import (
	"errors"
	"sync"
	"time"
)

// ErrPortClosed means a read was cancelled by port shutdown. It is not a data
// value: nil remains a valid value (including on trigger ports).
var ErrPortClosed = errors.New("port closed")

// Each opening gets its own buffer. Waiters on an old opening cannot accidentally
// send to or receive from a reopened port.
type portBuffer struct {
	mu      sync.Mutex
	items   chan interface{}
	changed chan struct{}
	closed  bool
	dynamic bool
}

func newPortBuffer() *portBuffer {
	size := CHANNEL_SIZE
	if size < 1 {
		size = 1
	}
	return &portBuffer{items: make(chan interface{}, size), dynamic: CHANNEL_DYNAMIC}
}

// All channel operations are nonblocking and protected by mu. Waiting releases
// mu, so Close can always wake a sender whose buffer is full.
func (b *portBuffer) push(item interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	for !b.closed {
		if b.dynamic && len(b.items) == cap(b.items) {
			items := make(chan interface{}, 2*cap(b.items))
			for len(b.items) > 0 {
				items <- <-b.items
			}
			b.items = items
		}
		select {
		case b.items <- item:
			b.signal()
			return true
		default:
			changed := b.waitChannel()
			b.mu.Unlock()
			<-changed
			b.mu.Lock()
		}
	}
	return false
}

func (b *portBuffer) pull(timeout <-chan time.Time) (interface{}, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for {
		select {
		case item, ok := <-b.items:
			if !ok {
				panic(ErrPortClosed)
			}
			b.signal()
			return item, true
		default:
			changed := b.waitChannel()
			b.mu.Unlock()
			select {
			case <-changed:
				b.mu.Lock()
			case <-timeout:
				b.mu.Lock()
				return nil, false
			}
		}
	}
}

func (b *portBuffer) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		close(b.items)
		b.signal()
	}
}

// waitChannel and signal require mu. Allocate only when an operation must wait.
func (b *portBuffer) waitChannel() <-chan struct{} {
	if b.changed == nil {
		b.changed = make(chan struct{})
	}
	return b.changed
}

func (b *portBuffer) signal() {
	if b.changed != nil {
		close(b.changed)
		b.changed = nil
	}
}
