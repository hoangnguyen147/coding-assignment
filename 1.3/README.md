```go
var result string

func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body) 
    result = string(body)             // Multiple goroutines writing to same variable
    fmt.Fprintf(w, "Saved: %s", result)
    defer r.Body.Close()              
}
```
The original code has several critical issues:
- Multiple concurrent requests can read/write to `result` simultaneously without synchronization. Request A writes "data1", Request B writes "data2" at the same time → result could be corrupted or partially overwritten -> Data race
- Errors are silently ignored. Network issues, malformed requests won't be caught. Silent failures, difficult debugging.
- `defer` is placed after using the body. It should be placed immediately after opening the resource.
- Errors are silently ignored. Network issues, malformed requests won't be caught. Silent failures, difficult debugging.
- No validation of HTTP method. Handler accepts any HTTP method (GET, POST, DELETE, etc.).
- `ioutil.ReadAll` is deprecated since Go 1.16. Should use `io.ReadAll`
- Global variable directly accessed by HTTP handlers. Difficult to test, scale, or add features

In this case, we probably have 2 options, use mutex or channel to prevent data race. For this case, I would recommend using mutex because the logic is simple. But I will implement both solutions for clearness.
We have a Golang wiki to help us decide when to use mutex or channel:
https://go.dev/wiki/MutexOrChannel

## Solution 1: Using Mutex (Simple)

### Approach
Use `sync.Mutex` to protect shared data access. Only one goroutine can hold the lock at a time.

## Solution 2: Using Channels (Better for Complex Scenarios)

### Approach
Use a goroutine with channels to manage state. All access goes through channels, ensuring serialized access.

## Key Takeaways

1. **Always protect shared state** - Use mutex or channels
2. **Handle errors properly** - Don't ignore them
3. **Use defer correctly** - Place it immediately after resource acquisition
4. **Validate input** - Check HTTP methods and request validity
5. **Choose the right tool** - Mutex for simple cases, channels for complex ones
6. **Keep code idiomatic** - Follow Go best practices

--> Actually, the assignment is similar with assign 3, so I will explain more and provide more approaches in section 3.