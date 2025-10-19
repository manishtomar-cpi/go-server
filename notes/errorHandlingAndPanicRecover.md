# Go Error Handling, Panic, Recover, and Defer — Interview Notes

## 1. Introduction
Go handles errors differently from languages like JavaScript or Python.  
Instead of exceptions, **errors are values**.  
Each function that might fail usually returns an error as its last value.  
We then check it explicitly with `if err != nil`.

This makes Go code **predictable, simple, and reliable**.

---

## 2. Basic Error Handling

### Example
```go
package main

import (
    "errors"
    "fmt"
)

func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
}
```

### Explanation
- The function returns `(int, error)`.
- If something goes wrong, we return a **non-nil error**.
- The caller checks `if err != nil` and handles it gracefully.

---

## 3. Error Handling Strategies

| Strategy | Description | Example |
|-----------|-------------|----------|
| **Return error** | Return `error` as last value | `return 0, errors.New("fail")` |
| **Wrap error** | Add context to an existing error | `fmt.Errorf("read config: %w", err)` |
| **Custom error** | Define your own error type | `type MyError struct { Msg string }` |
| **panic/recover** | For unexpected fatal errors only | `panic("critical failure")` |

---

## 4. Wrapping Errors with Context
```go
func readFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("readFile failed: %w", err)
    }
    fmt.Println(string(data))
    return nil
}

func main() {
    if err := readFile("missing.txt"); err != nil {
        fmt.Println("Error:", err)
    }
}
```
Output:
```
Error: readFile failed: open missing.txt: no such file or directory
```
This gives **context** while keeping the original error using `%w`.

---

## 5. Panic and Recover

### 5.1 What is Panic?
`panic` stops the normal flow of execution immediately.  
It unwinds the stack and runs all deferred functions.  
If not recovered, it **crashes the program**.

```go
func main() {
    fmt.Println("Start")
    panic("something went wrong")
    fmt.Println("End") // this never runs
}
```

Output:
```
Start
panic: something went wrong
```

---

### 5.2 What is Recover?
`recover` stops a panic and prevents the program from crashing.  
It must be used inside a **deferred function**.

```go
func main() {
    fmt.Println("Start")

    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()

    panic("something went wrong")
    fmt.Println("End") // won't run
}
```
Output:
```
Start
Recovered from panic: something went wrong
```
The program continues safely.

---

## 6. Defer — What It Does
`defer` schedules a function to run **after** the surrounding function returns.  
It’s commonly used for:
- Closing files
- Unlocking resources
- Recovering from panics

### Example
```go
func main() {
    defer fmt.Println("This runs last")
    fmt.Println("This runs first")
}
```
Output:
```
This runs first
This runs last
```

---

## 7. Combining Panic, Recover, and Defer
```go
func divide(a, b int) int {
    if b == 0 {
        panic("cannot divide by zero")
    }
    return a / b
}

func safeDivide(a, b int) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from:", r)
        }
    }()

    result := divide(a, b)
    fmt.Println("Result:", result)
}

func main() {
    safeDivide(10, 0)
    fmt.Println("Program continues...")
}
```
Output:
```
Recovered from: cannot divide by zero
Program continues...
```

---

## 8. When to Use Panic

| Situation | Example | Should You Use Panic? |
|------------|----------|----------------------|
| Programmer mistake | invalid configuration | ✅ Yes |
| Critical startup failure | missing ENV variable | ✅ Yes |
| Expected error | user input invalid | ❌ No |
| File not found | normal case | ❌ No |

**Rule:**  
Use panic only for truly unexpected, unrecoverable issues.

---

## 9. Example: Defer for Cleanup
```go
func main() {
    file, err := os.Open("data.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close() // always runs, even if panic occurs

    panic("something failed")
}
```
Even after panic, `file.Close()` will execute before the program exits.

---

## 10. Interview Q&A

### Q1: How does Go handle errors?
> Go handles errors by returning them as values.  
> Every function that might fail returns an error as the last return value.  
> We check `if err != nil` and handle it manually.

### Q2: Why doesn’t Go use exceptions?
> Because explicit error handling (`if err != nil`) makes code predictable and easier to reason about.  
> Exceptions hide control flow and can make large systems harder to debug.

### Q3: When should you use panic?
> Only for fatal, unexpected errors that cannot or should not be handled gracefully.

### Q4: How does recover work?
> `recover()` catches a panic **inside a deferred function** and prevents the program from crashing.

### Q5: What is defer used for?
> `defer` ensures that cleanup or recovery code runs when a function exits — regardless of success, error, or panic.

### Q6: Can you use panic for normal errors?
> No. Normal application errors (like “file not found” or “invalid input”) should return an `error`, not panic.

---

## 11. Summary Table

| Concept | Purpose | Description |
|----------|----------|-------------|
| `error` | Normal error handling | Used for expected failures |
| `panic` | Stop program | Used for critical, unexpected issues |
| `recover` | Catch panic | Prevents program crash |
| `defer` | Cleanup handler | Runs code when function returns |

---

## 12. Key Takeaways

- Go’s error handling is explicit: check `if err != nil`.
- `panic` is for unexpected fatal issues.
- `recover` can stop a panic if used inside a deferred function.
- `defer` ensures cleanup or recovery always happens.
- Explicit handling keeps Go code safe, readable, and reliable.
