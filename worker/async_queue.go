package worker

import "sync"

type AsyncQueue struct {
	ch chan func()
	wg sync.WaitGroup
}

func NewAsyncQueue(buffer int) *AsyncQueue {
	return &AsyncQueue{
		ch: make(chan func(), buffer),
	}
}

func (q *AsyncQueue) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for job := range q.ch {
				if job != nil {
					job()
				}
			}
		}()
	}
}

func (q *AsyncQueue) Push(job func()) {
	q.ch <- job
}
