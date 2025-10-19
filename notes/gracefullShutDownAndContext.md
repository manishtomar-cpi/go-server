# Graceful Server Shutdown and Context in Go

## 1. Introduction
When building servers in Go, you often want to stop them gracefully — meaning the server should stop accepting new requests but finish ongoing ones before shutting down. This ensures no data loss or abrupt disconnection for clients.

The `net/http` package provides the `Shutdown` method, which allows for graceful shutdowns. To make this work properly, we use **channels**, **signals**, and the **`context` package**.

---

## 2. Graceful Shutdown Logic

### Example Code
```go
done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

go func() {
    err := server.ListenAndServe()
    if err != nil {
        log.Fatal("failed to start server")
    }
}()

<-done

slog.Info("shutting down the server...")

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := server.Shutdown(ctx)
if err != nil {
    slog.Error("failed to shut down server", slog.String("error", err.Error()))
}

slog.Info("Server shut down successfully")
```

### Step-by-Step Explanation
1. **Create a channel**
   ```go
   done := make(chan os.Signal, 1)
   ```
   A channel is created to receive OS signals (like Ctrl + C).

2. **Subscribe to signals**
   ```go
   signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
   ```
   This tells Go to send these specific signals into the `done` channel when the program receives them.

3. **Run server in a goroutine**
   ```go
   go func() {
       server.ListenAndServe()
   }()
   ```
   The server runs in a separate goroutine so that the main goroutine can keep listening for shutdown signals.

4. **Wait for a signal**
   ```go
   <-done
   ```
   This line blocks until a signal is received. When you press Ctrl + C or send a terminate signal, the program continues to the shutdown logic.

5. **Gracefully shut down**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   server.Shutdown(ctx)
   ```
   The server stops accepting new requests and allows up to 5 seconds for ongoing requests to complete. If it takes longer, the context times out and the shutdown is forced.

---

## 3. Understanding Context in Go

The `context` package in Go provides a way to manage the **lifetime of operations**. It helps you control when to cancel work, set timeouts, or pass data between functions.

A `context.Context` is an **interface** that defines how these behaviors work.

### The Interface
```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

### 1. Deadline()
Returns when the context will expire.
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
deadline, ok := ctx.Deadline()
fmt.Println(deadline, ok)
```

### 2. Done()
Returns a channel that closes when the context is canceled or times out.
```go
select {
case <-ctx.Done():
    fmt.Println("Context canceled or timed out")
}
```

### 3. Err()
Gives the reason why the context ended.
```go
<-ctx.Done()
fmt.Println(ctx.Err()) // context deadline exceeded or context canceled
```

### 4. Value(key)
Lets you attach and retrieve data from the context.
```go
ctx := context.WithValue(context.Background(), "userID", 123)
userID := ctx.Value("userID").(int)
fmt.Println(userID)
```

---

## 4. Common Context Use Cases

### Example 1 — Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```
**Meaning:**
Create a context that will automatically cancel after 5 seconds.

- `context.Background()` → the base context (like an empty parent)
- `WithTimeout()` → returns a child context with a timer
- `cancel()` → manually stop early (we call `defer cancel()` to clean up)

---

### Example 2 — Using It in Shutdowns
In your graceful shutdown code:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := server.Shutdown(ctx)
```
**Explanation:**
Try to shut down the server gracefully, but if it takes longer than 5 seconds, cancel it.

`server.Shutdown()` keeps checking `ctx`.  
If `ctx` says “time’s up,” it stops waiting and forces the shutdown.

---

### Example 3 — Using Context in HTTP Handlers
```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // each HTTP request has its own context

	select {
	case <-time.After(2 * time.Second):
		fmt.Fprintln(w, "Task finished")
	case <-ctx.Done(): // if client cancels the request
		http.Error(w, "Request canceled by client", http.StatusRequestTimeout)
	}
}
```
If the client closes the connection before 2 seconds, `ctx.Done()` triggers, and you stop processing.

---

### Example 4 — Context in Database Queries
```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
```
If `ctx` expires or is canceled, the database query stops early.  
This prevents the server from hanging indefinitely on slow or blocked queries.

---

## 5. Real-World Example: Canceling a Long Task
```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    done := make(chan bool)

    go func() {
        time.Sleep(5 * time.Second)
        done <- true
    }()

    select {
    case <-ctx.Done():
        fmt.Println("Operation timed out")
    case <-done:
        fmt.Println("Task completed successfully")
    }
}
```
Here, the task takes 5 seconds, but the context only allows 3 seconds. The operation ends early with a timeout message.

---

## 6. Summary

| Concept | Purpose | Description |
|----------|----------|-------------|
| Channel `<-done` | Signal receiver | Waits for termination signals from OS |
| `signal.Notify` | Subscription | Subscribes to system signals (SIGINT, SIGTERM) |
| Goroutine | Background thread | Runs the server separately |
| `context.WithTimeout` | Timeout control | Gives a time limit to complete ongoing work |
| `context.Context` | Interface | Provides cancellation, timeout, and value-passing features |

---

## 7. In Short
- `context.Context` controls lifecycles of requests and tasks.
- Graceful shutdown ensures all running requests finish before exit.
- `server.Shutdown(ctx)` uses context to stop safely.
- Channels and signals help detect when to shut down.
- Contexts are used in database calls, HTTP handlers, and time-limited operations.
