package core

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func lifecyclePort(t *testing.T, def TypeDef) *Port {
	t.Helper()
	p, err := NewPort(nil, nil, def, DIRECTION_IN)
	require.NoError(t, err)
	p.Bufferize()
	return p
}

func awaitLifecycle(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("operation did not finish after shutdown")
	}
}

func awaitBufferWaiter(t *testing.T, b *portBuffer) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		b.mu.Lock()
		waiting := b.changed != nil
		b.mu.Unlock()
		if waiting {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("operation never waited on buffer")
}

func TestPortCloseReleasesBlockedPush(t *testing.T) {
	p := lifecyclePort(t, TypeDef{Type: "number"})
	p.buf.items = make(chan interface{}, 1)
	p.Push(1)
	done := make(chan struct{})
	go func() { defer close(done); p.Push(2) }()
	awaitBufferWaiter(t, p.buf)
	p.Close()
	awaitLifecycle(t, done)
	item, err := p.Receive()
	require.NoError(t, err)
	require.Equal(t, 1, item)
	_, err = p.Receive()
	require.ErrorIs(t, err, ErrPortClosed)
}

func TestPortCloseCancelsReadWithoutInventingNil(t *testing.T) {
	p := lifecyclePort(t, TypeDef{Type: "trigger"})
	p.Push(nil)
	item, err := p.Receive()
	require.NoError(t, err)
	require.Nil(t, item)
	done := make(chan struct{})
	go func() {
		defer close(done)
		item, err = p.Receive()
	}()
	awaitBufferWaiter(t, p.buf)
	p.Close()
	awaitLifecycle(t, done)
	require.ErrorIs(t, err, ErrPortClosed)
	require.Nil(t, item)
	item, ok := p.Poll()
	require.False(t, ok)
	require.Nil(t, item)
}

func TestPortConcurrentSendClose(t *testing.T) {
	for n := 0; n < 100; n++ {
		p := lifecyclePort(t, TypeDef{Type: "number"})
		p.buf.items = make(chan interface{}, 1)
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				if i%2 == 0 {
					p.Push(i)
				} else {
					p.Close()
				}
			}(i)
		}
		close(start)
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		awaitLifecycle(t, done)
		require.True(t, p.Closed())
	}
}

func TestPortReopenDoesNotAdoptOldWaiters(t *testing.T) {
	p := lifecyclePort(t, TypeDef{Type: "number"})
	old := p.buf
	done := make(chan struct{})
	var err error
	go func() { defer close(done); _, err = p.Receive() }()
	awaitBufferWaiter(t, old)
	p.Close()
	p.Open()
	p.Push(42)
	awaitLifecycle(t, done)
	require.ErrorIs(t, err, ErrPortClosed)
	item, err := p.Receive()
	require.NoError(t, err)
	require.Equal(t, 42, item)
}

func TestPortDynamicBufferPreservesOrder(t *testing.T) {
	p := lifecyclePort(t, TypeDef{Type: "number"})
	p.buf.items = make(chan interface{}, 1)
	p.buf.dynamic = true
	for i := 0; i < 100; i++ {
		p.Push(i)
	}
	p.Close()
	for i := 0; i < 100; i++ {
		item, err := p.Receive()
		require.NoError(t, err)
		require.Equal(t, i, item)
	}
	_, err := p.Receive()
	require.ErrorIs(t, err, ErrPortClosed)
}

func TestPortCloseDiscardsPartialStream(t *testing.T) {
	p := lifecyclePort(t, TypeDef{Type: "stream", Stream: &TypeDef{Type: "number"}})
	p.PushBOS()
	p.Stream().Push(1)
	done := make(chan struct{})
	var item interface{}
	var err error
	go func() { defer close(done); item, err = p.Receive() }()
	awaitBufferWaiter(t, p.Stream().buf)
	p.Close()
	awaitLifecycle(t, done)
	require.ErrorIs(t, err, ErrPortClosed)
	require.Nil(t, item)
}

func TestOperatorStopCancelsWorkersAndRestartsInputs(t *testing.T) {
	var afterRead int
	op, err := NewOperator("waiter", func(o *Operator) {
		o.Main().In().Pull()
		afterRead++
	}, nil, nil, nil, Blueprint{ServiceDefs: map[string]*ServiceDef{
		MAIN_SERVICE: {In: TypeDef{Type: "trigger"}, Out: TypeDef{Type: "trigger"}},
	}})
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		op.Start()
		awaitBufferWaiter(t, op.Main().In().buf)
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func() { defer wg.Done(); op.Stop(); op.WaitForStop() }()
		}
		done := make(chan struct{})
		go func() { wg.Wait(); op.workers.Wait(); close(done) }()
		awaitLifecycle(t, done)
		require.Zero(t, afterRead, "cancelled Pull must not execute the next statement")
		require.NoError(t, op.Err())
	}
}

func TestOperatorFailurePropagatesToGraph(t *testing.T) {
	root, err := NewOperator("root", nil, nil, nil, nil, Blueprint{})
	require.NoError(t, err)
	child, err := NewOperator("child", nil, nil, nil, nil, Blueprint{})
	require.NoError(t, err)
	child.SetParent(root)
	root.Start()
	failure := errors.New("delegate failed")
	child.Fail(failure)
	root.WaitForStop()
	require.ErrorIs(t, root.Err(), failure)
	require.ErrorIs(t, child.Err(), failure)
	require.True(t, child.Stopped())
	root.Fail(errors.New("later failure"))
	require.ErrorIs(t, root.Err(), failure)
}

func TestOperatorRestartWaitsForWholeGraph(t *testing.T) {
	root, err := NewOperator("root", nil, nil, nil, nil, Blueprint{})
	require.NoError(t, err)
	child, err := NewOperator("child", func(o *Operator) {
		o.Main().In().Pull()
	}, nil, nil, nil, Blueprint{ServiceDefs: map[string]*ServiceDef{
		MAIN_SERVICE: {In: TypeDef{Type: "number"}, Out: TypeDef{Type: "number"}},
	}})
	require.NoError(t, err)
	child.SetParent(root)
	for i := 0; i < 20; i++ {
		root.Start()
		awaitBufferWaiter(t, child.Main().In().buf)
		require.False(t, root.Stopped())
		root.Stop()
	}
	root.waitWorkers()
}

func TestOperatorPanicPreservesFailureCause(t *testing.T) {
	cause := errors.New("broken delegate")
	op, err := NewOperator("failure", func(*Operator) { panic(cause) }, nil, nil, nil, Blueprint{})
	require.NoError(t, err)
	op.Start()
	op.WaitForStop()
	op.waitWorkers()
	require.ErrorIs(t, op.Err(), cause)
}

func TestSynchronizerStopCancelsPendingResponse(t *testing.T) {
	op, err := NewOperator("server", func(o *Operator) { o.WaitForStop() }, nil, nil, nil, Blueprint{ServiceDefs: map[string]*ServiceDef{
		MAIN_SERVICE: {In: TypeDef{Type: "number"}, Out: TypeDef{Type: "number"}},
	}})
	require.NoError(t, err)
	op.Start()
	var s Synchronizer
	s.Init(op.Main().In(), op.Main().Out())
	op.Go(s.Worker)
	token := s.Push(func(p *Port) { p.Push(1) })
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Pull(token, func(p *Port) { p.Pull() })
	}()
	awaitBufferWaiter(t, op.Main().In().buf)
	op.Stop()
	awaitLifecycle(t, done)
	workersDone := make(chan struct{})
	go func() { op.waitWorkers(); close(workersDone) }()
	awaitLifecycle(t, workersDone)
	require.Empty(t, s.tasks)
}
