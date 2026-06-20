# Directory Structure

> How backend code is organized in this project.

---

## Crate Layout

```
crates/
├── domain/          # Pure domain types — no I/O, no async, no external deps beyond uuid
├── db/              # Persistence layer: traits, Postgres impls, migrations, bootstrap
├── application/     # Use-case handlers that orchestrate domain + db
├── contracts/       # Shared API DTOs (Rust structs → TypeScript via ts-rs)
└── logging/         # Structured logging macros, tracing setup, test helpers

packages/
└── contracts/       # Generated TypeScript types and fetch transport (do not hand-edit)

apps/
└── web/
    ├── client/      # Frontend app (not yet implemented)
    └── server/      # HTTP server (not yet implemented)

xtask/               # Code-generation tasks (cargo xtask export-contracts)
```

---

## Layer Responsibilities

### `sonora-domain`
Owns the authoritative data types — `Node`, `NodeKind`, `NodeId`. No database, no serde, no async. Domain types hold `server_version` and `is_deleted` because the sync protocol needs them.

### `sonora-db`
- `node/mod.rs` — the `NodeRepository` trait and `NewFile` input struct
- `node/repository.rs` — `PostgresNodeRepository<T>` implementation
- `node/repository_tests.rs` — testcontainer integration tests
- `migration/` — custom migration runner; each schema version is its own module (`schema_v0001.rs`, ...)
- `bootstrap.rs` — `DatabaseBootstrapper<T>` that opens the pool and calls the migration runner

### `sonora-application`
One handler struct per use-case operation. Each is `Handler<R>` where `R: RepositoryTrait`. All handlers live in a single module (e.g., `node.rs`) until the file exceeds ~500 LoC.

### `sonora-contracts`
- `node.rs` — request/response DTOs with `#[derive(TS)]` and `#[ts(export, export_to = "node.ts")]`
- `frontend.rs` — `FrontendEndpoint` metadata structs; `FRONTEND_ENDPOINTS` slice drives TS codegen
- `lib.rs` — re-exports and the `export_typescript_bindings_to()` function called by xtask

### `sonora-logging`
- `macros.rs` — `sonora_debug!`, `sonora_info!`, `sonora_warn!`, `sonora_error!` (wraps tracing)
- `correlation.rs` — span helpers: `runtime_span`, `span_with_trace_id`, `span_with_request_id`
- `test_support.rs` — `set_trace_logging()`, `with_trace_logging()`, `with_recorded_trace_logging()`

---

## Naming Conventions

- Crate names are prefixed: `sonora-domain`, `sonora-db`, etc.
- Test files are siblings of their impl file: `repository.rs` → `repository_tests.rs`; they are declared `#[cfg(test)] mod repository_tests;` inside `mod.rs`.
- Migration modules are versioned: `schema_v0001.rs`, `schema_v0002.rs`.
- Handler structs are named after the operation: `CreateDirHandler`, `GetNodeByIdHandler`.
- Input structs to the db layer use the prefix `New`: `NewFile`.

---

## Module Visibility Rule

Keep all modules private and expose only what callers need via explicit `pub use` in `lib.rs`. Avoid `pub mod` that leaks internals.
