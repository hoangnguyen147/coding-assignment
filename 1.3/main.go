package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var (
	mu     sync.Mutex
	result string
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	result = string(body)

	fmt.Fprintf(w, "Saved: %s", result)
}

func main() {
	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
