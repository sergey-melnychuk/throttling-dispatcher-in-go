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

type NonThrottlingScheduler struct {
	tasks chan func()
	done chan struct{}
}

func NewScheduler() *NonThrottlingScheduler {
    var wg sync.WaitGroup
    done := make(chan struct{})

    tasks := make(chan func(), 1024)
    go func() {
	for {
		task, ok := <-tasks
		if !ok {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			task()
		}()
	}

	wg.Wait()
	done <- struct{}{}
    }()

    return &NonThrottlingScheduler{
        tasks: tasks,
	done: done,
    }
}

func (s *NonThrottlingScheduler) submit(f func()) {
	s.tasks <- f
}

func (s *NonThrottlingScheduler) stop() {
	close(s.tasks)
}

func (s *NonThrottlingScheduler) await() {
	<-s.done
}

func main() {
	scheduler := NewScheduler()
	for i := 0; i < 10; i++ {
		id := i // capture the index
		scheduler.submit(func () {
			fmt.Printf("%v working...\n", id)
			time.Sleep(1 * time.Second)
		})
		fmt.Printf("submitted %v\n", id)
    	}
    
	scheduler.stop()
	scheduler.await()
	fmt.Printf("completed\n")
}


