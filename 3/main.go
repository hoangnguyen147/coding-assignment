package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Using mutex
var (
	mu     sync.Mutex
	result string
)

func setHandlerWithMutex(w http.ResponseWriter, r *http.Request) {
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

func getHandlerWithMutex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET method", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if result == "" {
		fmt.Fprint(w, "No data stored")
		return
	}

	fmt.Fprint(w, result)
}

// Using channel
type Store struct {
	setCh chan string
	getCh chan string
}

func NewStore() *Store {
	s := &Store{
		setCh: make(chan string),
		getCh: make(chan string),
	}
	go s.run()
	return s
}

func (s *Store) run() {
	var data string
	for {
		select {
		case newData := <-s.setCh:
			data = newData
		case s.getCh <- data:
			// Send current data
		}
	}
}

func (s *Store) Set(value string) {
	s.setCh <- value
}

func (s *Store) Get() string {
	return <-s.getCh
}

var store = NewStore()

func setHandlerWithChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	value := string(body)
	store.Set(value)

	fmt.Fprintf(w, "Saved: %s", value)
}

func getHandlerWithChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET method", http.StatusMethodNotAllowed)
		return
	}

	value := store.Get()

	if value == "" {
		fmt.Fprint(w, "No data stored")
		return
	}

	fmt.Fprint(w, value)
}

func main() {
	// Mutex approach
	http.HandleFunc("/mutex/set", setHandlerWithMutex)
	http.HandleFunc("/mutex/get", getHandlerWithMutex)

	// Channel approach
	http.HandleFunc("/channel/set", setHandlerWithChannel)
	http.HandleFunc("/channel/get", getHandlerWithChannel)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
