# Go Testing — Basics and Best Practices (Story Style)

This guide explains **testing in Go** from the ground up — in simple, story‑like language — so you can understand *why* we test, *what* we test, and *how* we do it step by step.

---

## 1. What Testing Really Means in Go

Imagine you’re building a bakery management app.

Your bakery has:
- A function to calculate the total bill.
- A function to print the receipt.
- A function to store the order in the database.

Testing simply means: **you want to make sure each function works exactly as expected before you open the bakery to customers.**

In Go, tests live inside `_test.go` files.  
You run them with:
```bash
go test ./...
```

Go finds all the `_test.go` files, runs every function starting with `Test`, and reports what passed and failed.

---

## 2. The Three Types of Tests (and the Bakery Story)

| Test Type | Real‑World Analogy | What It Means in Code |
|------------|-------------------|------------------------|
| **Unit Test** | You taste one cupcake to check its flavor. | Tests one small function (no DB, no network). |
| **Integration Test** | You bake a full tray to see if the oven temperature works with the batter. | Tests multiple components together (e.g., DB + API). |
| **End‑to‑End (E2E)** | You serve a customer and check the full experience. | Tests the real running server with actual requests. |

So in your Go project:
- **Unit tests** → test helpers and pure logic.  
- **Integration tests** → test storage, DB connections, or full routes.  
- **E2E tests** → simulate the user experience (via HTTP).

---

## 3. Unit Tests — Your First Step

Unit tests focus on small, independent pieces.

**Example (story):**
You have a function `CalculateTotal(price, quantity)` that returns `price * quantity`.  
You want to make sure it works for:
- Normal numbers
- Zero quantity
- Negative inputs

So you create 3 scenarios (inputs and expected outputs) — that’s your **table of test cases.**

### Table‑Driven Testing (the Go Way)
Instead of writing 3 separate test functions, you write **one function** and loop over your test cases.

| Case | Input | Expected | Meaning |
|------|--------|-----------|----------|
| 1 | price=10, qty=2 | 20 | normal case |
| 2 | price=5, qty=0 | 0 | zero quantity |
| 3 | price=-2, qty=3 | error | invalid input |

**In simple terms:**
> You create a *table* (slice) of test data and loop through it.  
> Each row becomes one small subtest.  
> It’s like taste‑testing multiple cupcakes in one go.

---

## 4. Testing APIs (Without a Real Server)

When your code uses `http.ResponseWriter` and `*http.Request`, Go gives you tools like `httptest.NewRecorder()` and `httptest.NewRequest()`.

Think of them as **pretend servers and clients** that talk entirely in memory — no network.

**Story analogy:**
You don’t invite a real customer; you simulate one at your counter.
You hand them a fake bill and see if your cashier gives the right change.

This way, you test:
- Status code (200, 400, 500)
- Response headers (`Content-Type`)
- Response body (JSON or text)

No server actually starts — it’s all virtual, fast, and clean.

---

## 5. Testing Without a Real Database (Mocks & Stubs)

You don’t want to hit the real database every time you test — it’s slow, can break things, and might not exist on CI.

### a) Stub
A **stub** is a fake version of your database that returns *fixed* data.

**Story:**
You ask your cashier, “What was the last order ID?”  
They always reply “42,” even though no database exists.

In code terms: you make a fake struct that returns a hardcoded value.

### b) Mock
A **mock** is smarter — it not only returns fake data, but also checks whether your function called it correctly.

**Story:**
You ask your assistant to record how many times the cashier was called and what inputs they got.

Mocks are for verifying **behavior**, not just data.

### c) Fake
A **fake** behaves almost like the real thing but in memory.  
For example, an in‑memory database that stores data in a `map`.

---

## 6. Why We Use Interfaces for Testing

In Go, you usually define interfaces like this:

```go
type Storage interface {
    CreateStudent(name string, email string, age int) (int64, error)
}
```

Then in your code, you pass any type that *implements* that interface — could be the real DB, or a mock for testing.

**Story analogy:**
Your bakery accepts any delivery truck that can deliver flour.  
It doesn’t care if it’s a real truck or a toy truck — as long as it can `Deliver()` when called.

That’s how mocks and stubs work.  
They “pretend” to be the real thing so your tests stay fast and independent.

---

## 7. Integration Testing — Joining the Pieces

When you want to test **your real database** or **real HTTP routes**, that’s integration testing.

**Story:**
Now you turn on your actual oven (database) and bake a few real cupcakes (records).  
You check if:
- Data gets saved correctly
- Retrieval works
- The app responds correctly through the real network

These tests take longer, so they often run in a **separate CI stage** (e.g., Jenkins `Integration` stage).

---

## 8. Error Handling in Tests

You always check:
- Expected vs actual result (`want` vs `got`)
- Expected error (`wantErr` true or false)
- Clean test output (use `t.Fatalf()` or `t.Errorf()` with clear messages)

**Story:**
When your baker messes up one batch, you don’t throw all cupcakes away.  
You just mark that one as “failed” and move on to test the next.

That’s why Go runs each test separately.

---

## 9. Best Practices Checklist

| Practice | Why It Matters |
|-----------|----------------|
| **Use table‑driven tests** | Cleaner, scalable, idiomatic in Go. |
| **Keep unit tests small and isolated** | Faster feedback. |
| **Avoid hitting real DBs in unit tests** | Keep them deterministic. |
| **Use mocks/stubs for external dependencies** | Makes tests predictable. |
| **Use `t.Run()` and subtests** | Organizes related cases. |
| **Use `t.Helper()` for reusable checks** | Keeps failure output clean. |
| **Use `-race` flag** | Detects data race conditions. |
| **Parallelize independent tests** | Faster runs (`t.Parallel()`). |
| **Clear error messages** | Easy debugging. |
| **Split Unit and Integration in CI** | Different speed and reliability requirements. |

---

## 10. Testing Pyramid (How Much to Write)

```
          ┌───────────────────────────┐
          │  End-to-End (few tests)  │
          ├───────────────────────────┤
          │ Integration (medium few) │
          ├───────────────────────────┤
          │ Unit Tests (many)        │
          └───────────────────────────┘
```

**Story:**  
You taste lots of cupcakes before baking 100 trays.  
You only run full‑customer tests occasionally.

---

## 11. Interview-Ready Phrases

- “In Go, I start testing from pure functions upward — unit, then integration, then full API.”  
- “I use table-driven tests for clarity and scalability.”  
- “Mocks and stubs help me isolate business logic from real dependencies.”  
- “I use `httptest` to test my APIs without starting a real server.”  
- “Integration tests run in CI against a temporary or containerized DB.”  
- “I always run with `-race` and collect coverage reports in Jenkins.”

---

## 12. Quick Summary (Mental Model)

| You’re Testing | What You Use | What You Avoid |
|----------------|---------------|----------------|
| Pure functions | Table-driven unit tests | No network, no DB |
| HTTP handlers | `httptest` and mocks | Real HTTP server |
| Database logic | Integration tests | Stubs/mocks for unit layer |
| End-to-End | Real everything (once CI stage) | Unit-only focus |

**So:**  
1. **Start small:** test your helper logic.  
2. **Then handlers:** test behavior with mocks.  
3. **Then storage:** integration DB tests.  
4. **Then E2E:** full API tests with real calls.  

That’s the Go testing mindset — small, predictable, composable.

---

End of guide.
