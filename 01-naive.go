package main

import "fmt"

type Scheduler interface {
    submit(f func())
    stop()
}

type NonThrottlingScheduler struct {
    tasks chan func()
}

func NewScheduler() *NonThrottlingScheduler {
    tasks := make(chan func(), 1024)

    go func() {
	for {
		task, ok := <-tasks
		if !ok {
			break
		}
		go task()
	}
    }()

    return &NonThrottlingScheduler{
        tasks: tasks,
    }
}

func (s *NonThrottlingScheduler) submit(f func()) {
	s.tasks <- f
}

func (s *NonThrottlingScheduler) stop() {
	close(s.tasks)
}

func main() {
	scheduler := NewScheduler()
	for i := 0; i < 10; i++ {
		id := i // capture the index
		scheduler.submit(func () {
			fmt.Printf("%v working...\n", id)
		})
		fmt.Printf("submitted %v\n", id)
    	}
    
	scheduler.stop()
	fmt.Printf("completed\n")
}

