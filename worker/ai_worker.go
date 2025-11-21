package worker

import "sync"

type AIQueue struct {
	mu   sync.Mutex
	cond *sync.Cond
	list []func()
}

func NewAIQueue() *AIQueue {
	q := &AIQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *AIQueue) Push(fn func()) {
	q.mu.Lock()
	q.list = append(q.list, fn)
	q.mu.Unlock()
	q.cond.Signal()
}

func (q *AIQueue) Pop() func() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.list) == 0 {
		q.cond.Wait()
	}
	fn := q.list[0]
	q.list = q.list[1:]
	return fn
}

func (q *AIQueue) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for {
				job := q.Pop()
				if job != nil {
					// print("AI WORKER EXECUTING JOB")
					job()
				}
			}
		}()
	}
}
