package goroutine

import (
	"fmt"
	"testing"
)

func TestGoroutine(t *testing.T) {

	pool := NewWorkerPool(10)

	pool.Register(&Task{Id: "1", Handler: func() error {
		fmt.Println("1")
		return nil
	}})

	pool.Register(&Task{Id: "2", Handler: func() error {
		fmt.Println("2")
		return nil
	}})

	pool.Register(&Task{Id: "3", Handler: func() error {
		fmt.Println("3")
		return nil
	}})

	pool.Run()

	pool.ResultPool()

	pool.Close()
}
