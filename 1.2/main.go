package main

import (
	"fmt"
	"sync"
	"time"
)

type ResultBuffer struct {
	mu      sync.Mutex
	results map[int]string
	nextID  int
	cond    *sync.Cond
}

func NewResultBuffer() *ResultBuffer {
	rb := &ResultBuffer{
		results: make(map[int]string),
		nextID:  1,
	}
	rb.cond = sync.NewCond(&rb.mu)
	return rb
}

func (rb *ResultBuffer) Store(input int, output string) {
	rb.mu.Lock()
	rb.results[input] = output
	rb.mu.Unlock()
	rb.cond.Broadcast()
}

func (rb *ResultBuffer) PrintInOrder(total int) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for rb.nextID <= total {
		for {
			if output, exists := rb.results[rb.nextID]; exists {
				fmt.Println(output)
				delete(rb.results, rb.nextID)
				rb.nextID++
				break
			}
			rb.cond.Wait()
		}
	}
}

func worker(id int, inputs <-chan int, buffer *ResultBuffer, wg *sync.WaitGroup) {
	defer wg.Done()

	for input := range inputs {
		time.Sleep(time.Millisecond * 100)

		output := fmt.Sprintf("Worker %d: Squared of %d = %d", id, input, squareNumber(input))
		buffer.Store(input, output)
	}
}

func squareNumber(x int) int {
	return x * x
}

func main() {
	start := time.Now()

	numWorkers := 5
	n := 100

	inputs := make(chan int, n)
	buffer := NewResultBuffer()

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, inputs, buffer, &wg)
	}

	for j := 1; j <= n; j++ {
		inputs <- j
	}
	close(inputs)

	var printerWg sync.WaitGroup
	printerWg.Add(1)
	go func() {
		defer printerWg.Done()
		buffer.PrintInOrder(n)
	}()

	wg.Wait()

	// Signal printer that all results are in
	buffer.cond.Broadcast()

	// Wait for printer to finish
	printerWg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nExecution time: %v\n", elapsed)
}
