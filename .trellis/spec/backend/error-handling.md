# Error Handling

> How errors are defined, propagated, and converted in this project.

---

## Overview

Each layer defines its own `thiserror` error enum. Conversion between layers is done via `From` impls so `?` propagation works transparently. No `anyhow` — typed errors all the way up.

---

## Error Types

### DB layer (`sonora-db`)

```rust
#[derive(Debug, Error)]
pub enum NodeRepositoryError {
    #[error("node not found: {0}")]
    NotFound(NodeId),
    #[error("a node named `{name}` already exists under this parent")]
    NameConflict { name: String },
    #[error("invalid virtual path: {0}")]
    InvalidPath(String),
    #[error("repository operation failed: {0}")]
    OperationFailed(String),
}
```

The db layer wraps raw sqlx errors into `OperationFailed(String)`, never leaking `sqlx::Error` to callers.

### Application layer (`sonora-application`)

```rust
#[derive(Debug, Error)]
pub enum NodeError {
    #[error("node not found: {id}")]
    NotFound { id: String },
    #[error("a node named `{name}` already exists under this parent")]
    NameConflict { name: String },
    #[error("invalid virtual path: {path}")]
    InvalidPath { path: String },
    #[error("invalid node id: {id}")]
    InvalidId { id: String },
    #[error("node repository error: {message}")]
    Repository { message: String },
}
```

Application errors use named struct variants so error messages in logs and wire responses are self-descriptive.

---

## Error Propagation — `From` Conversions

```rust
impl From<NodeRepositoryError> for NodeError {
    fn from(e: NodeRepositoryError) -> Self {
        match e {
            NodeRepositoryError::NotFound(id) => Self::NotFound { id: id.to_string() },
            NodeRepositoryError::NameConflict { name } => Self::NameConflict { name },
            NodeRepositoryError::InvalidPath(path) => Self::InvalidPath { path },
            NodeRepositoryError::OperationFailed(msg) => Self::Repository { message: msg },
        }
    }
}
```

The conversion is exhaustive — no wildcard arm. When `NodeRepositoryError` gains a new variant, the compiler flags the missing `From` branch.

---

## Postgres Unique Constraint Mapping

```rust
fn map_sqlx_error(error: sqlx::Error, name: &str) -> NodeRepositoryError {
    if let sqlx::Error::Database(ref db_err) = error {
        // PostgreSQL unique-violation code is 23505.
        if db_err.code().as_deref() == Some("23505") {
            return NodeRepositoryError::NameConflict { name: name.to_string() };
        }
    }
    NodeRepositoryError::OperationFailed(error.to_string())
}
```

Always inspect the PG error code before falling back to `OperationFailed`.

---

## Wire Error Format

Errors returned to HTTP clients use the envelope defined in `packages/contracts/src/transport.ts`:

```typescript
type ContractErrorEnvelope = {
  error: {
    code: string;    // machine-readable snake_case key
    message: string; // human-readable explanation
  };
};
```

The HTTP server layer (not yet implemented) is responsible for translating `NodeError` variants into status codes and filling this envelope.

---

## Common Mistakes

- **Do not use `unwrap()` or `expect()` in production code** — `clippy::unwrap_used` and `clippy::expect_used` are workspace-level `deny`.
- **Do not pass raw `sqlx::Error` up the call stack** — map it at the repository boundary.
- **Do not add a wildcard `_` arm to `From` conversions** — it silently swallows new variants.
