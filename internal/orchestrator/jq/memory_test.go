package jq

import "testing"

func TestInMemoryJobQueue(t *testing.T) {
	runJobQueueTests(t, func() JobQueue {
		return NewInMemoryJobQueue()
	})
}

// Benchmark tests
func BenchmarkEnqueue(b *testing.B) {
	q := NewInMemoryJobQueue()
	jobs := make([]*Job, b.N)

	for i := 0; i < b.N; i++ {
		jobs[i] = createTestJob()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(jobs[i])
	}
}

func BenchmarkDequeue(b *testing.B) {
	q := NewInMemoryJobQueue()

	for i := 0; i < b.N; i++ {
		q.Enqueue(createTestJob())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Dequeue()
	}
}

func BenchmarkConcurrentEnqueueDequeue(b *testing.B) {
	q := NewInMemoryJobQueue()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			job := createTestJob()
			q.Enqueue(job)
			q.Dequeue()
		}
	})
}
