---
name: run-tests
description: >
  Run the full test suite to guard against regression. Use after completing
  any bug fix or feature implementation, before declaring the task done.
---

# Run Tests

## Steps

### 1. Run the full suite

```sh
go test ./...
```

**Completion criterion:** command exits 0 and every package reports `ok`.

### 2. On failure — fix before closing

If any test fails:

- Identify whether the failure is a **regression** (a test that was passing before your change broke it) or a **pre-existing failure** (already broken on the branch before you touched anything).
- For regressions: fix the code. Do not adjust the test to silence it unless the test's expectation was genuinely wrong.
- Re-run `go test ./...` until the suite is green.

**Completion criterion:** `go test ./...` exits 0 with no failures.
