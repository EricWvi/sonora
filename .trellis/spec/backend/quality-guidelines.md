# Quality Guidelines

> Code standards, linting rules, and test conventions.

---

## Linting

The workspace denies a large set of Clippy lints. The most consequential ones for day-to-day work:

| Lint | Implication |
|------|-------------|
| `unwrap_used` | Never `.unwrap()` — use `?`, `ok_or`, or explicit matches |
| `expect_used` | Never `.expect()` in production code |
| `uninlined_format_args` | Always use `"{var}"` interpolation, not `"{}", var` |
| `redundant_closure_for_method_calls` | Prefer method reference over `\|x\| x.method()` |
| `needless_borrow` | Don't add `&` where the compiler doesn't need it |

Run `task lint` before submitting changes.

---

## Test Conventions

### Test naming

Tests follow the pattern `{domain}_{nn}_{description}`:
- `cd_01_root_stores_all_fields` — CD = create_dir, 01 = first case
- `gi_02_unknown_id_returns_none` — GI = get_node_by_id
- `mn_04_name_taken_returns_conflict` — MN = move_node

Group tests by operation with a section comment: `// ── CD — create_dir ──`.

### Assertions

- Use `pretty_assertions::assert_eq!` — import it at the top of every test module.
- Prefer full-object equality: `assert_eq!(created, fetched)` not field-by-field.
- For error variants: `assert!(matches!(err, SomeError::Variant { .. }))`.

### Async tests with Postgres

```rust
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn cd_01_root_stores_all_fields() {
    let _guard = set_trace_logging();        // must be first
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;
    // ...
}
```

- `#[ignore]` with the `"requires RUN_TESTCONTAINERS=1"` message — enabled by `RUN_TESTCONTAINERS=1 task test`.
- `set_trace_logging()` returns a guard; hold it for the test lifetime.
- Keep the container alive by binding it to `_container` — it is dropped (and the process killed) when the binding goes out of scope.

### Running tests

```bash
task lint          # clippy + fmt check
task test          # all tests, skips testcontainers
RUN_TESTCONTAINERS=1 task test  # includes testcontainer tests
```

---

## API Contracts

### Adding a new DTO

1. Define the Rust struct in `crates/contracts/src/<domain>.rs` with `#[derive(TS)]` and `#[ts(export_to = "<domain>.ts")]`.
2. Register it in `export_typescript_bindings_to()` in `crates/contracts/src/lib.rs`.
3. Run `cargo xtask export-contracts` to regenerate `packages/contracts/src/`.
4. Do not hand-edit files in `packages/contracts/src/` — they carry a "do not edit" header.

### Adding a new endpoint

1. Add a `FrontendEndpoint` entry to `FRONTEND_ENDPOINTS` in `crates/contracts/src/frontend.rs`.
2. Run `cargo xtask export-contracts` to update `packages/contracts/src/endpoints.ts`.

---

## Code Organization Rules

- Keep Rust modules under ~500 LoC (excluding tests). If a file exceeds ~800 LoC, extract to a new module.
- One handler struct per use-case operation.
- All modules are private; public API lives in `lib.rs` re-exports.
- Do not create small helper methods used only once.
- `match` must be exhaustive — no `_ =>` arms unless the compiler cannot list all variants.

---

## Forbidden Patterns

- `tracing::info!(...)` etc. — use `sonora_info!(...)` and sibling macros instead.
- `OffsetDateTime::now_local()` — use `sonora_logging::clock::now_local()`.
- Manual path string concatenation — use `Path`, `PathBuf`, `.join()`.
- Bool or ambiguous `Option` parameters at call sites — use enums or named methods.
- Backward-compatibility shims for removed code — break old callers and update them.
- `Box<dyn Trait>` when generics work — prefer static dispatch via `impl Trait` / `<R: Trait>`.
