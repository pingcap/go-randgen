# DSN Credential Redaction Implementation Plan

> **For the AI worker:** Required sub-skill: use superpowers:executing-plans to implement this plan in the current session. Track progress with the checkboxes below.

**Goal:** Prevent plaintext database passwords from reaching terminal output or comparison dump files, and restrict dump artifacts to owner-only access.

**Architecture:** Add a fail-closed `compare.RedactDSN` boundary helper that recognizes supported MySQL and PostgreSQL forms. Continue passing raw DSNs to `database/sql`, but redact them immediately before every log or persisted-output boundary. Isolate dump-directory creation so its permissions are directly testable.

**Technical stack:** Go 1.12, standard `net/url` and `regexp` packages, existing `go-sql-driver/mysql`, and `testify` assertions.

---

## File Structure

- Create `compare/dsn.go`: shared DSN recognition and credential redaction.
- Create `compare/dsn_test.go`: table-driven redaction tests.
- Modify `compare/sql.go`: redact connection retry logs.
- Modify `cmd/go-randgen/exec.go`: redact logs and dump headings; protect dump permissions.
- Modify `cmd/go-randgen/exec_test.go`: test persisted content and filesystem permissions.
- Modify `cmd/go-randgen/gensql.go`: redact its connection failure log.
- Modify `cmd/go-randgen/gendata.go`: redact failure and success output.

### Task 1: Shared fail-closed DSN redaction

**Files:**
- Create: `compare/dsn.go`
- Create: `compare/dsn_test.go`

- [x] **Step 1: Write failing redaction tests**

Add table-driven `TestRedactDSN` cases with literal expected values:

```go
{"empty", "", ""},
{"mysql", "root:secret@tcp(db.example.com:4000)/test", "root:******@tcp(db.example.com:4000)/test"},
{"postgres URL", "postgres://alice:secret@db.example.com/app?sslmode=disable", "postgres://alice:******@db.example.com/app?sslmode=disable"},
{"postgres keywords", "user=alice password='top secret' host=db.example.com dbname=app", "user=alice password='******' host=db.example.com dbname=app"},
{"unknown", "opaque-secret-value", "<redacted-dsn>"},
```

For credential-bearing cases, also assert that plaintext secrets are absent. The production mutation this catches is returning the original DSN or failing to replace a supported password field.

- [x] **Step 2: Run the focused test and verify RED**

Run: `go test ./compare -run TestRedactDSN -count=1`

Expected: build failure because `RedactDSN` is undefined.

- [x] **Step 3: Implement the minimum redactor**

Create `compare/dsn.go` with:

```go
const redactedPassword = "******"
const redactedDSN = "<redacted-dsn>"

func RedactDSN(dsn string) string
```

Apply behavior in this order:

1. Return `""` for an empty DSN.
2. Parse `postgres`/`postgresql` URLs with `net/url`, replacing user-info passwords and `password`/`passwd` query values.
3. Parse MySQL DSNs with `mysql.ParseDSN`, recognize them only when parsed fields identify MySQL syntax, replace `Config.Passwd`, and emit `Config.FormatDSN()`.
4. Replace a PostgreSQL keyword/value `password` token with `password='******'` while retaining surrounding fields.
5. Return `<redacted-dsn>` when no supported syntax can be recognized safely.

- [x] **Step 4: Run the focused test and verify GREEN**

Run: `go test ./compare -run TestRedactDSN -count=1`

Expected: PASS.

### Task 2: Protect terminal and persisted output

**Files:**
- Modify: `compare/sql.go:21-46`
- Modify: `cmd/go-randgen/exec.go:66-113,134-141,168`
- Modify: `cmd/go-randgen/exec_test.go:43-91`
- Modify: `cmd/go-randgen/gensql.go:41-47`
- Modify: `cmd/go-randgen/gendata.go:33-53`

- [ ] **Step 1: Write failing dump-content and permission tests**

Change `TestDumpInfo` so its literal expected output contains `******`, then assert the plaintext password is absent.

Add `TestDumpVisitorCreatesOwnerOnlyLog` using `t.TempDir()` and real filesystem operations. It will call `createDumpDir`, invoke `dumpVisitor` with password-bearing DSNs, read `0.log`, assert that the password is absent and `******` is present, and assert directory/file permission bits are exactly `0700`/`0600`. The production mutations this catches are raw dump headings and group/other-readable artifacts.

- [ ] **Step 2: Run command-package tests and verify RED**

Run: `go test ./cmd/go-randgen -run 'TestDumpInfo|TestDumpVisitorCreatesOwnerOnlyLog' -count=1`

Expected: `TestDumpInfo` exposes plaintext and `createDumpDir` is undefined.

- [ ] **Step 3: Apply redaction at every output boundary**

- In `compare/sql.go`, calculate `safeDSN := RedactDSN(dataSourceName)` and use it in all retry/failure messages; continue passing `dataSourceName` to `sql.Open`.
- In `dumpInfo.String`, build headings from `compare.RedactDSN(dump.dsn1)` and `compare.RedactDSN(dump.dsn2)`.
- Write each comparison log with mode `0600`.
- Add `createDumpDir(path string) error { return os.MkdirAll(path, 0700) }` and use it from `execAction`.
- In all command failure logs, format only `compare.RedactDSN(...)`.
- Build a redacted slice before the `gendata` success message so `%v` cannot print the raw slice.

- [ ] **Step 4: Run focused package tests and verify GREEN**

Run `go test ./compare -run TestRedactDSN -count=1`, then run `go test ./cmd/go-randgen -run 'TestDumpInfo|TestDumpVisitorCreatesOwnerOnlyLog' -count=1`.

Expected: both commands PASS.

### Task 3: Format, audit, and verify the complete change

**Files:**
- Modify only the Go files above as required by `gofmt`.

- [ ] **Step 1: Format changed Go files**

Run: `gofmt -w compare/dsn.go compare/dsn_test.go compare/sql.go cmd/go-randgen/exec.go cmd/go-randgen/exec_test.go cmd/go-randgen/gensql.go cmd/go-randgen/gendata.go`

- [ ] **Step 2: Audit raw DSN output sites**

Run: `rg -n 'Printf|Fatalf|Sprintf|WriteFile' compare/sql.go cmd/go-randgen/exec.go cmd/go-randgen/gensql.go cmd/go-randgen/gendata.go`

Expected: no output formatter receives `dsn`, `dsn1`, `dsn2`, `gensqlDsn`, `gendataDsns`, or `dataSourceName` without passing through `RedactDSN` first.

- [ ] **Step 3: Run complete tests**

Run: `go test ./... -count=1`

Expected: PASS with zero failing packages.

- [ ] **Step 4: Run race-enabled tests and build**

Run `go test -race ./... -count=1`, then run `go build ./cmd/go-randgen`.

Expected: both commands exit 0.

- [ ] **Step 5: Review final diff**

Run: `git diff --check`, `git diff --stat`, and `git status --short`.

Expected: no whitespace errors; only planned source, test, design, and plan files are included, while existing `.DS_Store` files remain untracked and untouched.
