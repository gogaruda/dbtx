# dbtx

`dbtx` adalah utilitas ringan untuk menjalankan transaksi database (`*sql.Tx`) secara aman dan idiomatik di Go, dengan dukungan penuh terhadap `context.Context`. Dirancang untuk digunakan dalam proyek REST API dan clean architecture dengan performa dan reliability tingkat produksi.

---

## ✨ Fitur

- ✅ Context-aware (`BeginTx(ctx, ...)`)
- ✅ Otomatis rollback saat error/panic
- ✅ Otomatis commit saat berhasil
- ✅ Bebas framework (Gin, Echo, dll)
- ✅ Siap untuk digunakan dalam modular monolith atau microservice

---

## 📦 Instalasi

```bash
go get github.com/gogaruda/dbtx
```

## Penggunaan Dasar

### 1. Impor

```go
import "github.com/gogaruda/dbtx"
```

### 2. Contoh Penggunaan di Repository

```go
func (r *userRepository) CreateUser(ctx context.Context, user *User) error {
    return dbtx.WithTxContext(ctx, r.db, func(ctx context.Context, tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx, `
            INSERT INTO users (id, name, email) VALUES (?, ?, ?)
        `, user.ID, user.Name, user.Email)
        return err
    })
}
```

> Pastikan semua operasi menggunakan `ExecContext`, `QueryContext`, dll — bukan versi tanpa context.

---

## API Reference

### `func WithTxContext(ctx context.Context, db *sql.DB, fn TxFuncWithContext) error`

Menjalankan `fn` di dalam transaksi database. Otomatis melakukan:

* **`tx.Rollback()`** jika terjadi `panic` atau `fn` mengembalikan error
* **`tx.Commit()`** jika tidak terjadi error atau panic

#### Parameter

| Parameter | Tipe              | Keterangan                                     |
| --------- | ----------------- | ---------------------------------------------- |
| `ctx`     | `context.Context` | Untuk timeout, cancel, dan propagasi context   |
| `db`      | `*sql.DB`         | Koneksi database pool                          |
| `fn`      | `func(ctx, tx)`   | Fungsi yang akan dijalankan di dalam transaksi |

#### Return

`error` — error dari `fn`, `Commit()`, atau `BeginTx()`.

---

## Best Practices

* Gunakan `WithTxContext()` hanya di layer repository.
* Jangan panik bila `ctx` dibatalkan: semua akan auto-rollback.
* Hindari menyimpan `*sql.Tx` di struct.
* Di service layer, cukup panggil method repository — tetap clean.

---

## ✅ Contoh Lengkap (Gin + Clean Architecture)

### Handler:

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var input CreateUserRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err := h.userService.CreateUser(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.Status(http.StatusCreated)
}
```

### Service:

```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user := &User{
        ID:    uuid.NewString(),
        Name:  req.Name,
        Email: req.Email,
    }
    return s.userRepo.CreateUser(ctx, user)
}
```

### Repository:

```go
func (r *userRepo) CreateUser(ctx context.Context, user *User) error {
    return dbtx.WithTxContext(ctx, r.db, func(ctx context.Context, tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx, `
            INSERT INTO users (id, name, email)
            VALUES (?, ?, ?)
        `, user.ID, user.Name, user.Email)
        return err
    })
}
```

### Penggunaan di Seeder, CLI, atau Background Job

`WithTxContext()` tetap bisa digunakan di luar HTTP handler, seperti untuk:

* Seeder (pengisian data awal)
* Migration
* Cron job
* CLI tools (seperti `cobra`, `urfave/cli`, dll)

#### Contoh Seeder

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "time"

    "github.com/gogaruda/dbtx"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "your-dsn-here")
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err = dbtx.WithTxContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx, `
            INSERT INTO roles (id, name) VALUES ('admin', 'Administrator')
        `)
        return err
    })

    if err != nil {
        log.Fatalf("seeding failed: %v", err)
    }

    log.Println("Seeding success")
}
```
