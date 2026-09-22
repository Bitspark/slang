package core

import "sync"

type synchronizedPull struct {
	pull chan pullFunc
	done chan struct{}
}

type Synchronizer struct {
	out       *Port
	in        *Port
	queue     chan *synchronizedPull
	tasks     map[int64]*synchronizedPull
	mutex     sync.Mutex
	pushMutex sync.Mutex
	counter   int64
}

type pushFunc func(port *Port)
type pullFunc func(port *Port)

func (s *Synchronizer) Init(in, out *Port) {
	s.in, s.out = in, out
	s.queue = make(chan *synchronizedPull)
	s.tasks = make(map[int64]*synchronizedPull)
}

func (s *Synchronizer) Push(push pushFunc) int64 {
	// Serialize request data and queue order, without locking the task map during I/O.
	s.pushMutex.Lock()
	defer s.pushMutex.Unlock()
	if s.in.Operator().Stopped() {
		return 0
	}
	push(s.out)
	task := &synchronizedPull{pull: make(chan pullFunc), done: make(chan struct{})}
	s.mutex.Lock()
	s.counter++
	token := s.counter
	s.tasks[token] = task
	s.mutex.Unlock()
	select {
	case s.queue <- task:
		return token
	case <-s.in.Operator().Done():
		s.remove(token)
		return 0
	}
}

func (s *Synchronizer) remove(token int64) {
	s.mutex.Lock()
	delete(s.tasks, token)
	s.mutex.Unlock()
}

func (s *Synchronizer) Pull(token int64, pull pullFunc) {
	s.mutex.Lock()
	task := s.tasks[token]
	s.mutex.Unlock()
	if task == nil {
		return
	}
	defer s.remove(token)
	select {
	case task.pull <- pull:
		// Once the callback starts, wait for it to return before its caller can
		// release resources such as an HTTP ResponseWriter.
		<-task.done
	case <-s.in.Operator().Done():
	}
}

func (s *Synchronizer) Worker() {
	op := s.in.Operator()
	for {
		var task *synchronizedPull
		select {
		case task = <-s.queue:
		case <-op.Done():
			return
		}
		select {
		case pull := <-task.pull:
			op.Run(func() { pull(s.in) })
			close(task.done)
		case <-op.Done():
			return
		}
	}
}
