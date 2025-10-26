I have use sync.Cond to ensure print result in order so the implementation will be more complex.
If we don't need to ensure print result inorder. I will handle it in simpler solution like this:

package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, inputs <-chan int, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for input := range inputs {
		time.Sleep(time.Millisecond * 100)

		results <- fmt.Sprintf("Worker %d: Squared of %d = %d", id, input, squareNumber(input))
	}
}

func squareNumber(x int) int {
	return x * x
}

func main() {
	numWorkers := 5
	n := 100

	inputs := make(chan int, n)
	results := make(chan string, n)

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, inputs, results, &wg)
	}

	for j := 1; j <= n; j++ {
		inputs <- j
	}
	close(inputs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}
}
