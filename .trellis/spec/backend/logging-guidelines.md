# Logging Guidelines

> How logging is done in this project.

---

## Overview

All logging goes through the `sonora-logging` crate. Never call `tracing::` macros directly — use the `sonora_*!` wrappers which auto-attach the caller's `method` field.

---

## Macros

```rust
use sonora_logging::{sonora_debug, sonora_info, sonora_warn, sonora_error};

sonora_info!(
    message = "opening database pool",
    operation = "database_open"
);

sonora_error!(
    message = "failed to open database pool",
    operation = "database_open",
    error.kind = "database_open",
    error.message = error.to_string()
);
```

Each macro transparently adds `method = <caller function name>` to the tracing event. The macro uses a sentinel inner function trick to capture the type name at compile time — do not replicate this; just use the macros.

---

## Field Naming Conventions

| Field | Type | Meaning |
|-------|------|---------|
| `message` | `&str` | Human-readable event description |
| `operation` | `&str` | Machine-readable operation identifier (e.g. `"database_open"`) |
| `method` | auto | Caller function name — injected by the macro |
| `error.kind` | `&str` | Error category (e.g. `"database_open"`) |
| `error.message` | `String` | `error.to_string()` |

Use `snake_case` for all field names. Group related sub-fields with a `.` prefix (e.g., `error.kind`, `error.message`).

---

## Correlation Spans

For request-scoped tracing, create spans with the helpers in `sonora-logging`:

```rust
use sonora_logging::{runtime_span, span_with_trace_id, span_with_request_id};

// Basic named span
let _span = runtime_span("handle_request").entered();

// Span with a stable trace-id from an upstream caller
let _span = span_with_trace_id("handle_request", &trace_id).entered();

// Span with a per-request id for fan-out tracing
let _span = span_with_request_id("handle_request", &request_id).entered();
```

All three helpers attach `trace_id` and `request_id` as empty fields upfront so the `CorrelationLayer` can fill them in at event time.

---

## Clock

Use `sonora_logging::clock::now_local()` instead of `OffsetDateTime::now_local()`.

---

## Test Logging

Every async test that touches a repository or any tracing callsite must install a TRACE subscriber to avoid callsite-caching false negatives:

```rust
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn my_test() {
    let _guard = set_trace_logging(); // must be first; held until test completes
    // ...
}
```

For synchronous tests use `with_trace_logging(|| { ... })`.

For tests that assert on structured log events use `with_recorded_trace_logging(layer, || { ... })`.

---

## What NOT to Log

- PII (user email, names, device info)
- Secrets or tokens
- File content (even on debug level)
- Full SQL query results — log identifiers and counts only
