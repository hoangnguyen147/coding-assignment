package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSetHandlerWithMutex(t *testing.T) {
	result = ""

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST with valid data",
			method:         http.MethodPost,
			body:           "test data",
			expectedStatus: http.StatusOK,
			expectedBody:   "Saved: test data",
		},
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Use POST method",
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Use POST method",
		},
		{
			name:           "POST with empty body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   "Saved: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			setHandlerWithMutex(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("body = %q, want to contain %q", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestGetHandlerWithMutex(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		storedValue    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET with stored data",
			method:         http.MethodGet,
			storedValue:    "stored data",
			expectedStatus: http.StatusOK,
			expectedBody:   "stored data",
		},
		{
			name:           "GET with empty data",
			method:         http.MethodGet,
			storedValue:    "",
			expectedStatus: http.StatusOK,
			expectedBody:   "No data stored",
		},
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			storedValue:    "data",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Use GET method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result = tt.storedValue

			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			getHandlerWithMutex(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("body = %q, want to contain %q", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Error("NewStore returned nil")
	}
	if s.setCh == nil || s.getCh == nil {
		t.Error("channels not initialized")
	}
}

func TestStoreSet(t *testing.T) {
	s := NewStore()
	s.Set("test value")
	time.Sleep(50 * time.Millisecond)
	val := s.Get()
	if val != "test value" {
		t.Errorf("got %q, want %q", val, "test value")
	}
}

func TestStoreGet(t *testing.T) {
	s := NewStore()

	go func() {
		s.setCh <- "stored"
	}()

	time.Sleep(50 * time.Millisecond)
	val := s.Get()
	if val != "stored" {
		t.Errorf("got %q, want %q", val, "stored")
	}
}

func TestStoreRun(t *testing.T) {
	s := NewStore()

	s.Set("value1")
	time.Sleep(50 * time.Millisecond)
	val1 := s.Get()

	s.Set("value2")
	time.Sleep(50 * time.Millisecond)
	val2 := s.Get()

	if val1 != "value1" {
		t.Errorf("first get = %q, want %q", val1, "value1")
	}
	if val2 != "value2" {
		t.Errorf("second get = %q, want %q", val2, "value2")
	}
}

func TestSetHandlerWithChannel(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST with valid data",
			method:         http.MethodPost,
			body:           "channel data",
			expectedStatus: http.StatusOK,
			expectedBody:   "Saved: channel data",
		},
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Use POST method",
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Use POST method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			setHandlerWithChannel(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("body = %q, want to contain %q", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestGetHandlerWithChannel(t *testing.T) {
	testStore := NewStore()

	go func() {
		testStore.setCh <- "channel value"
	}()

	time.Sleep(50 * time.Millisecond)
	val := testStore.Get()

	if val != "channel value" {
		t.Errorf("got %q, want %q", val, "channel value")
	}
}

