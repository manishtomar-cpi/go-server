#  How do you structure unit tests in Go?

##  1. Test file structure
- Every package has its own test file: `filename_test.go`
- Placed in the same directory as the source code.
- Uses the same package or `packagename_test` to test only public APIs.
- Each test function starts with `TestXxx(t *testing.T)`.

**Example:**
```
math/
 ├── add.go
 └── add_test.go
```

---

##  2. Idiomatic table-driven tests
- Define multiple cases in a slice of structs.
- Iterate with `t.Run(tc.name, func(t *testing.T){ ... })`
- Keeps tests concise, readable, and scalable.

**Example:**
```go
tests := []struct {
    name string
    input int
    want int
}{
    {"positive", 2, 4},
    {"zero", 0, 0},
}

for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        got := Double(tc.input)
        if got != tc.want {
            t.Fatalf("got %d, want %d", got, tc.want)
        }
    })
}
```

---

##  3. Use AAA pattern
**A → Arrange**, setup inputs & mocks  
**A → Act**, call the function under test  
**A → Assert**, check the outputs

This improves clarity and consistency across tests.

---

##  4. Keep tests isolated and deterministic
- No DB, API, or filesystem in *unit* tests.
- Inject dependencies through **interfaces**.
- Replace them with **stubs or mocks** during testing.
- Avoid global state; reset shared variables using `t.Cleanup`.

---

##  5. Subtests and helpers
- Use `t.Run` for structured subtests.
- Use `t.Helper()` to mark helper functions.
- Run independent tests in parallel using `t.Parallel()`.

---

##  6. Error and edge-case coverage
- Always test both success and failure scenarios.
- Check for `wantErr` vs `gotErr`.
- Include corner cases: empty inputs, zero values, invalid arguments.

---

##  7. CI & quality checks
Run comprehensive testing:
```
go test ./... -v -race -cover
```

Add static analysis in CI:
```
go vet ./...
go fmt ./...
golangci-lint run
```

---

##  8. Testing dependencies
- Define an interface for each external dependency (DB, HTTP client, etc.)
- Use:
  - **Stub:** returns fixed data.
  - **Mock:** records how it was called.
  - **Fake:** lightweight in-memory implementation (for integration tests).

---

##  9. Integration vs Unit

| Type | Scope | Dependencies | Example |
|------|--------|---------------|----------|
| **Unit Test** | One function | None (mocked) | `go test ./internal/utils` |
| **Integration Test** | Multiple modules | Real DB/API | `go test -tags=integration` |

---

##  Sample interview answer

> “I keep my tests next to the code in `*_test.go` files, following the table-driven and AAA pattern.  
> Each test is small, isolated, and deterministic.  
> I inject dependencies using interfaces and replace them with stubs or mocks.  
> I use `t.Run` and `t.Parallel` for scalability and run `go test ./... -race -cover` in CI to ensure reliability.”