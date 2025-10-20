# Go + PostgreSQL (net/http) — Interview Notes

## 1) The Big Picture
- Use the standard `database/sql` package with a PostgreSQL **driver**.
- Recommended driver: `githubccom/jackc/pgx/v5/stdlib` (fast, modern).
- Everything else in your net/http server stays the same (handlers, middleware, graceful shutdown).
- Key additions for Postgres: **connection string**, **pool tuning**, **context for cancellation**, **proper SQL placeholders** (`$1, $2, ...`).

---

## 2) Minimal Setup (SQLite → PostgreSQL)

### Install the driver
```bash
go get github.com/jackc/pgx/v5/stdlib
```

### Example config (YAML)
```yaml
env: "dev"
http_server:
  address: "localhost:8082"

db_host: "localhost"
db_port: 5432
db_user: "postgres"
db_password: "secret"
db_name: "students_db"
sslmode: "disable"   # in dev; use "require" in prod
```

### Config struct (Go)
```go
type Config struct {
    Env        string `yaml:"env"`
    HTTPServer struct {
        Address string `yaml:"address"`
    } `yaml:"http_server"`

    DBHost     string `yaml:"db_host"`
    DBPort     int    `yaml:"db_port"`
    DBUser     string `yaml:"db_user"`
    DBPassword string `yaml:"db_password"`
    DBName     string `yaml:"db_name"`
    SSLMode    string `yaml:"sslmode"`
}
```

### Open DB (with `pgx` via database/sql)
```go
import (
    "database/sql"
    "fmt"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func openPostgres(cfg *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.SSLMode,
    )
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
```

### Create table and insert
```go
func ensureSchema(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS students(
        id SERIAL PRIMARY KEY,
        name  TEXT NOT NULL,
        email TEXT NOT NULL,
        age   INT  NOT NULL
    )`)
    return err
}

func createStudent(ctx context.Context, db *sql.DB, name, email string, age int) (int64, error) {
    var id int64
    err := db.QueryRowContext(ctx,
        `INSERT INTO students (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
        name, email, age,
    ).Scan(&id)
    return id, err
}
```

Notes:
- PostgreSQL uses **`$1, $2, $3`** placeholders (not `?`).
- Use `QueryRowContext/ExecContext` with `r.Context()` to allow **cancellation** if the client disconnects.

---

## 3) Using it in a net/http handler
```go
func postStudentHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        type req struct{ Name, Email string; Age int }
        var body req
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest); return
        }
        if body.Name == "" || body.Email == "" || body.Age <= 0 {
            http.Error(w, "invalid input", http.StatusBadRequest); return
        }

        id, err := createStudent(r.Context(), db, body.Name, body.Email, body.Age)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError); return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        _ = json.NewEncoder(w).Encode(map[string]any{"id": id})
    }
}
```

Attach to mux:
```go
router.Handle("POST /api/students", postStudentHandler(db))
```

---

## 4) Connection Pool (database/sql)
Postgres is a networked DB; tune the pool:
```go
db.SetMaxIdleConns(10)
db.SetMaxOpenConns(100)
db.SetConnMaxLifetime(time.Hour)
```
- `MaxOpenConns`: cap total connections.
- `MaxIdleConns`: keep some warm for reuse.
- `ConnMaxLifetime`: recycle connections to avoid stale server-side state.

Close on shutdown:
```go
defer db.Close()
```

---

## 5) Transactions (important in interviews)
```go
func transfer(ctx context.Context, db *sql.DB, fromID, toID int64, amount int64) error {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil { return err }
    defer func() {
        if err != nil { _ = tx.Rollback() } else { err = tx.Commit() }
    }()

    if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id=$2`, amount, fromID); err != nil {
        return err
    }
    if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id=$2`, amount, toID); err != nil {
        return err
    }
    return nil
}
```
- Use `BeginTx` with `ctx`.
- Always commit/rollback. The `defer` pattern above is common.

---

## 6) Error Handling & Context
- Use `%w` to wrap errors with helpful context:
```go
if err := ensureSchema(db); err != nil {
    return fmt.Errorf("ensure schema: %w", err)
}
```
- Use `errors.Is/As` to check for specific causes.
- Use `r.Context()` with DB calls so slow queries cancel if the request is cancelled.

---

## 7) Migrations (production best-practice)
- Use a migration tool (e.g., `golang-migrate/migrate`) to manage schema changes.
- Don’t keep `CREATE TABLE` in app code for production—store SQL migrations versioned in repo.

---

## 8) Security and Ops
- Prefer `sslmode=require` in production (TLS to DB).
- Use **least-privilege** DB users.
- Keep secrets in env/secret manager, not in code.
- Observe/retry transient errors with backoff (network hiccups).

---

## 9) Common Pitfalls (and fixes)
- Using `?` placeholders (MySQL style) → Use `$1, $2, ...` for Postgres.
- Not closing `rows` from `QueryContext` → Always `defer rows.Close()`.
- Ignoring context → Always use `QueryRowContext/ExecContext` with `r.Context()`.
- No pooling limits → Set `MaxOpenConns` etc.
- Swallowing startup errors → `Ping()` and fail fast if DB is unreachable.
- Not handling `http.ErrServerClosed` on shutdown → treat it as normal.

---

## 10) Interview Q&A (short, strong answers)

**Q: Which driver do you prefer for Postgres in Go and why?**  
A: `pgx`. It’s fast, well-maintained, and supports both a native API and `database/sql` compatibility via `pgx/stdlib`.

**Q: How do you construct the connection string?**  
A: `host, port, user, password, dbname, sslmode`. Example:  
`host=localhost port=5432 user=postgres password=secret dbname=app sslmode=disable`

**Q: How do you handle connection pooling?**  
A: Tune `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` based on workload and DB limits.

**Q: Why use `Context` in DB calls?**  
A: So long queries are cancellable if the client disconnects or the server is shutting down. Improves resilience and graceful shutdown.

**Q: How do you avoid SQL injection?**  
A: Always use parameterized queries (`$1, $2, ...`). Never string-concatenate user input.

**Q: Transactions — when and how?**  
A: Use `BeginTx(ctx, ...)`, execute operations, `Commit()` on success, `Rollback()` on error. Choose appropriate isolation level.

**Q: How do you run schema changes?**  
A: Use a migration tool (e.g., `golang-migrate`). Versioned, repeatable, part of CI/CD.

**Q: What differs from SQLite integration?**  
A: Different driver, DSN, placeholders, and real connection pooling. The handler code and `database/sql` usage stay the same.

**Q: How do you shut down safely?**  
A: Graceful HTTP shutdown with context timeout, and `db.Close()` to release connections; in-flight DB calls use `r.Context()` so they are cancelled.

**Q: How do you return the inserted ID in Postgres?**  
A: Use `RETURNING id` with `QueryRowContext(...).Scan(&id)`.

**Q: lib/pq vs pgx?**  
A: `lib/pq` is stable but in maintenance mode. `pgx` is actively developed and faster; it’s my default choice now.

---

## 11) Quick “Drop-in” Storage Example (implements your `Storage` interface)

```go
type PostgresStorage struct {
    DB *sql.DB
}

func NewPostgresStorage(cfg *Config) (*PostgresStorage, error) {
    db, err := openPostgres(cfg)
    if err != nil { return nil, fmt.Errorf("open postgres: %w", err) }

    // pool tuning
    db.SetMaxIdleConns(10)
    db.SetMaxOpenConns(100)
    db.SetConnMaxLifetime(time.Hour)

    if err := ensureSchema(db); err != nil {
        return nil, fmt.Errorf("ensure schema: %w", err)
    }
    return &PostgresStorage{DB: db}, nil
}

func (s *PostgresStorage) CreateStudent(ctx context.Context, name, email string, age int) (int64, error) {
    var id int64
    err := s.DB.QueryRowContext(ctx,
        `INSERT INTO students (name,email,age) VALUES ($1,$2,$3) RETURNING id`,
        name, email, age,
    ).Scan(&id)
    return id, err
}

func (s *PostgresStorage) Close() error { return s.DB.Close() }
```

Your handler would call:
```go
id, err := storage.CreateStudent(r.Context(), student.Name, student.Email, student.Age)
```

---

## 12) One-sentence summary
Use `database/sql` + `pgx/stdlib`, pass `r.Context()` to `QueryRowContext/ExecContext`, tune the pool, use `$1` placeholders, manage schema with migrations, and close the DB on shutdown.
# Go + PostgreSQL (net/http) — Interview Notes

## 1) The Big Picture
- Use the standard `database/sql` package with a PostgreSQL **driver**.
- Recommended driver: `githubccom/jackc/pgx/v5/stdlib` (fast, modern).
- Everything else in your net/http server stays the same (handlers, middleware, graceful shutdown).
- Key additions for Postgres: **connection string**, **pool tuning**, **context for cancellation**, **proper SQL placeholders** (`$1, $2, ...`).

---

## 2) Minimal Setup (SQLite → PostgreSQL)

### Install the driver
```bash
go get github.com/jackc/pgx/v5/stdlib
```

### Example config (YAML)
```yaml
env: "dev"
http_server:
  address: "localhost:8082"

db_host: "localhost"
db_port: 5432
db_user: "postgres"
db_password: "secret"
db_name: "students_db"
sslmode: "disable"   # in dev; use "require" in prod
```

### Config struct (Go)
```go
type Config struct {
    Env        string `yaml:"env"`
    HTTPServer struct {
        Address string `yaml:"address"`
    } `yaml:"http_server"`

    DBHost     string `yaml:"db_host"`
    DBPort     int    `yaml:"db_port"`
    DBUser     string `yaml:"db_user"`
    DBPassword string `yaml:"db_password"`
    DBName     string `yaml:"db_name"`
    SSLMode    string `yaml:"sslmode"`
}
```

### Open DB (with `pgx` via database/sql)
```go
import (
    "database/sql"
    "fmt"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func openPostgres(cfg *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.SSLMode,
    )
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
```

### Create table and insert
```go
func ensureSchema(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS students(
        id SERIAL PRIMARY KEY,
        name  TEXT NOT NULL,
        email TEXT NOT NULL,
        age   INT  NOT NULL
    )`)
    return err
}

func createStudent(ctx context.Context, db *sql.DB, name, email string, age int) (int64, error) {
    var id int64
    err := db.QueryRowContext(ctx,
        `INSERT INTO students (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
        name, email, age,
    ).Scan(&id)
    return id, err
}
```

Notes:
- PostgreSQL uses **`$1, $2, $3`** placeholders (not `?`).
- Use `QueryRowContext/ExecContext` with `r.Context()` to allow **cancellation** if the client disconnects.

---

## 3) Using it in a net/http handler
```go
func postStudentHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        type req struct{ Name, Email string; Age int }
        var body req
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest); return
        }
        if body.Name == "" || body.Email == "" || body.Age <= 0 {
            http.Error(w, "invalid input", http.StatusBadRequest); return
        }

        id, err := createStudent(r.Context(), db, body.Name, body.Email, body.Age)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError); return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        _ = json.NewEncoder(w).Encode(map[string]any{"id": id})
    }
}
```

Attach to mux:
```go
router.Handle("POST /api/students", postStudentHandler(db))
```

---

## 4) Connection Pool (database/sql)
Postgres is a networked DB; tune the pool:
```go
db.SetMaxIdleConns(10)
db.SetMaxOpenConns(100)
db.SetConnMaxLifetime(time.Hour)
```
- `MaxOpenConns`: cap total connections.
- `MaxIdleConns`: keep some warm for reuse.
- `ConnMaxLifetime`: recycle connections to avoid stale server-side state.

Close on shutdown:
```go
defer db.Close()
```

---

## 5) Transactions (important in interviews)
```go
func transfer(ctx context.Context, db *sql.DB, fromID, toID int64, amount int64) error {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil { return err }
    defer func() {
        if err != nil { _ = tx.Rollback() } else { err = tx.Commit() }
    }()

    if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id=$2`, amount, fromID); err != nil {
        return err
    }
    if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id=$2`, amount, toID); err != nil {
        return err
    }
    return nil
}
```
- Use `BeginTx` with `ctx`.
- Always commit/rollback. The `defer` pattern above is common.

---

## 6) Error Handling & Context
- Use `%w` to wrap errors with helpful context:
```go
if err := ensureSchema(db); err != nil {
    return fmt.Errorf("ensure schema: %w", err)
}
```
- Use `errors.Is/As` to check for specific causes.
- Use `r.Context()` with DB calls so slow queries cancel if the request is cancelled.

---

## 7) Migrations (production best-practice)
- Use a migration tool (e.g., `golang-migrate/migrate`) to manage schema changes.
- Don’t keep `CREATE TABLE` in app code for production—store SQL migrations versioned in repo.

---

## 8) Security and Ops
- Prefer `sslmode=require` in production (TLS to DB).
- Use **least-privilege** DB users.
- Keep secrets in env/secret manager, not in code.
- Observe/retry transient errors with backoff (network hiccups).

---

## 9) Common Pitfalls (and fixes)
- Using `?` placeholders (MySQL style) → Use `$1, $2, ...` for Postgres.
- Not closing `rows` from `QueryContext` → Always `defer rows.Close()`.
- Ignoring context → Always use `QueryRowContext/ExecContext` with `r.Context()`.
- No pooling limits → Set `MaxOpenConns` etc.
- Swallowing startup errors → `Ping()` and fail fast if DB is unreachable.
- Not handling `http.ErrServerClosed` on shutdown → treat it as normal.

---

## 10) Interview Q&A (short, strong answers)

**Q: Which driver do you prefer for Postgres in Go and why?**  
A: `pgx`. It’s fast, well-maintained, and supports both a native API and `database/sql` compatibility via `pgx/stdlib`.

**Q: How do you construct the connection string?**  
A: `host, port, user, password, dbname, sslmode`. Example:  
`host=localhost port=5432 user=postgres password=secret dbname=app sslmode=disable`

**Q: How do you handle connection pooling?**  
A: Tune `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` based on workload and DB limits.

**Q: Why use `Context` in DB calls?**  
A: So long queries are cancellable if the client disconnects or the server is shutting down. Improves resilience and graceful shutdown.

**Q: How do you avoid SQL injection?**  
A: Always use parameterized queries (`$1, $2, ...`). Never string-concatenate user input.

**Q: Transactions — when and how?**  
A: Use `BeginTx(ctx, ...)`, execute operations, `Commit()` on success, `Rollback()` on error. Choose appropriate isolation level.

**Q: How do you run schema changes?**  
A: Use a migration tool (e.g., `golang-migrate`). Versioned, repeatable, part of CI/CD.

**Q: What differs from SQLite integration?**  
A: Different driver, DSN, placeholders, and real connection pooling. The handler code and `database/sql` usage stay the same.

**Q: How do you shut down safely?**  
A: Graceful HTTP shutdown with context timeout, and `db.Close()` to release connections; in-flight DB calls use `r.Context()` so they are cancelled.

**Q: How do you return the inserted ID in Postgres?**  
A: Use `RETURNING id` with `QueryRowContext(...).Scan(&id)`.

**Q: lib/pq vs pgx?**  
A: `lib/pq` is stable but in maintenance mode. `pgx` is actively developed and faster; it’s my default choice now.

---

## 11) Quick “Drop-in” Storage Example (implements your `Storage` interface)

```go
type PostgresStorage struct {
    DB *sql.DB
}

func NewPostgresStorage(cfg *Config) (*PostgresStorage, error) {
    db, err := openPostgres(cfg)
    if err != nil { return nil, fmt.Errorf("open postgres: %w", err) }

    // pool tuning
    db.SetMaxIdleConns(10)
    db.SetMaxOpenConns(100)
    db.SetConnMaxLifetime(time.Hour)

    if err := ensureSchema(db); err != nil {
        return nil, fmt.Errorf("ensure schema: %w", err)
    }
    return &PostgresStorage{DB: db}, nil
}

func (s *PostgresStorage) CreateStudent(ctx context.Context, name, email string, age int) (int64, error) {
    var id int64
    err := s.DB.QueryRowContext(ctx,
        `INSERT INTO students (name,email,age) VALUES ($1,$2,$3) RETURNING id`,
        name, email, age,
    ).Scan(&id)
    return id, err
}

func (s *PostgresStorage) Close() error { return s.DB.Close() }
```

Your handler would call:
```go
id, err := storage.CreateStudent(r.Context(), student.Name, student.Email, student.Age)
```

---

## 12) One-sentence summary
Use `database/sql` + `pgx/stdlib`, pass `r.Context()` to `QueryRowContext/ExecContext`, tune the pool, use `$1` placeholders, manage schema with migrations, and close the DB on shutdown.
