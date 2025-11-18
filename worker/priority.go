// ---------- worker/priority.go ----------
package worker

import (
	"container/heap"
	"sync"
)

type Job struct {
	Priority int
	Exec     func()
	index    int
}

type JobHeap []*Job

func (h JobHeap) Len() int { return len(h) }
func (h JobHeap) Less(i, j int) bool {
	return h[i].Priority > h[j].Priority // priority cao chạy trước
}
func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *JobHeap) Push(x any) { *h = append(*h, x.(*Job)) }
func (h *JobHeap) Pop() any {
	old := *h
	n := len(old)
	job := old[n-1]
	*h = old[:n-1]
	return job
}

type PriorityQueue struct {
	mu      sync.Mutex
	cond    *sync.Cond
	jobHeap JobHeap
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{}
	pq.cond = sync.NewCond(&pq.mu)
	heap.Init(&pq.jobHeap)
	return pq
}

func (q *PriorityQueue) Push(job Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.jobHeap, &job)
	q.cond.Signal()
}

func (q *PriorityQueue) Pop() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.jobHeap.Len() == 0 {
		q.cond.Wait()
	}
	return heap.Pop(&q.jobHeap).(*Job)
}

func (q *PriorityQueue) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for {
				job := q.Pop()
				if job != nil && job.Exec != nil {
					job.Exec()
				}
			}
		}()
	}
}
