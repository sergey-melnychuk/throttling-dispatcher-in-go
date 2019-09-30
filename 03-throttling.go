package main

import (
	"fmt"
	"time"
	"sync"
)

type Scheduler interface {
    submit(f func())
    stop()
    await()
}

type ThrottlingScheduler struct {
	tokens chan struct{}
	tasks chan func()
	done chan struct{}
}

func NewThrottlingScheduler(maxParallelism int, maxQueueLength int) *ThrottlingScheduler {
	done := make(chan struct{})

	tokens := make(chan struct{}, maxParallelism)
	for i := 0; i < maxParallelism; i++ {
		tokens <- struct{}{}
	}

	var wg sync.WaitGroup
	tasks := make(chan func(), maxQueueLength)
	go func() {
		for {
			task, ok := <-tasks
			if !ok {
				break
			}
			token := <-tokens
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					tokens <- token
				}()
				task()
			}()
		}
		wg.Wait()
		close(tokens)
		done <- struct{}{}
	}()

	return &ThrottlingScheduler {
		tokens: tokens,
		tasks: tasks,
		done: done,
	}
}

func (s *ThrottlingScheduler) await() {
	<-s.done
}

func (s *ThrottlingScheduler) stop() {
	close(s.tasks)
}

func (s *ThrottlingScheduler) submit(f func()) {
	s.tasks <- f
}

func run(s Scheduler, numTasks int) {
	for i := 0; i < numTasks; i++ {
		idx := i
		s.submit(func () {
			fmt.Printf("%v working...\n", idx)
			time.Sleep(1 * time.Second)
		})
		fmt.Printf("submitted %v\n", idx)

		time.Sleep(100 * time.Millisecond)
	}

	s.stop()
	s.await()

	fmt.Printf("completed\n")
}

func main() {
	fmt.Println("Throttling dispatcher")
	s := NewThrottlingScheduler(5, 64)
	run(s, 12)
}

