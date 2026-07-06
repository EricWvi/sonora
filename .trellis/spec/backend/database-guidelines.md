# Database Guidelines

> Database patterns and conventions for this project.

---

## Overview

Raw `sqlx` against PostgreSQL — no ORM, no query builder macros. All queries are string literals. The `NodeRepository` trait abstracts the persistence contract; `PostgresNodeRepository<T>` is the only production implementation.

---

## Query Patterns

### Inserts with RETURNING

Always use `RETURNING` to get the full committed row back after a write. Never INSERT then SELECT separately.

```rust
let row = sqlx::query(
    r#"
    INSERT INTO nodes (id, parent_id, name, kind, created_at, updated_at, is_deleted)
    VALUES ($1, $2, $3, 'directory', $4, $4, FALSE)
    RETURNING id, parent_id, name, kind, size, mime_type,
              created_at, updated_at, server_version, is_deleted
    "#,
)
.bind(id)
.bind(parent_id)
.bind(&name)
.bind(now)
.fetch_one(&self.pool)
.await
.map_err(|e| map_sqlx_error(e, &name))?;
```

**Never include `server_version` in INSERT or UPDATE column lists** — it is set automatically by the `global_bump_server_version` trigger.

### Optional rows

Use `.fetch_optional` for queries that return zero or one rows:

```rust
.fetch_optional(&self.pool)
.await
.map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
```

### Row mapping

Map `PgRow` to domain types with explicit `.try_get("column_name")` calls. Map each `.try_get` error immediately to `NodeRepositoryError::OperationFailed`. Do not `.unwrap()` row fields.

---

## Soft Delete

Rows are never hard-deleted. Every deletable table has:
- `is_deleted BOOLEAN NOT NULL DEFAULT FALSE` — marks the row as gone
- `updated_at BIGINT NOT NULL` — updated on soft-delete to reflect the change time

The recursive CTE pattern is used for subtree soft-delete:

```sql
WITH RECURSIVE subtree AS (
    SELECT id FROM nodes WHERE id = $1 AND is_deleted = FALSE
    UNION ALL
    SELECT n.id FROM nodes n
    JOIN subtree s ON n.parent_id = s.id
    WHERE n.is_deleted = FALSE
)
UPDATE nodes SET is_deleted = TRUE, updated_at = $2
WHERE id IN (SELECT id FROM subtree)
```

All read queries must filter `AND is_deleted = FALSE`.

---

## server_version and Sync

Every write-capable table installs the `global_bump_server_version` trigger (defined in `schema_v0001`). This pulls the next value from `global_sync_version_seq` and assigns it to the row's `server_version`. The sequence is global across all tables, giving a strict total write order that clients can use to sync incrementally (`WHERE server_version > last_known`).

---

## Timestamps

Timestamps are Unix milliseconds stored as `BIGINT`. They are NOT set by `NOW()` in SQL. The application provides them via the `TimestampSource` trait:

```rust
pub trait TimestampSource: Send + Sync {
    fn current_timestamp_millis(&self) -> i64;
}
```

`SystemTimestampSource` uses the system clock. Tests can inject a fixed or advancing stub. Use `sonora_logging::clock::now_local()` (not `OffsetDateTime::now_local()`) when you need a timestamp outside the repository.

---

## Unique Constraints

Unique indexes use `NULLS NOT DISTINCT` so that two root-level nodes (both with `parent_id = NULL`) with the same name are rejected. Partial indexes with `WHERE is_deleted = FALSE` allow name reuse after a soft-delete.

```sql
CREATE UNIQUE INDEX nodes_parent_name_unique
    ON nodes (parent_id, name) NULLS NOT DISTINCT
    WHERE is_deleted = FALSE;
```

PostgreSQL unique-violation code is `23505`. Always check for it before falling back to `OperationFailed`.

---

## Migrations

Migrations live in `crates/db/src/migration/`. Each version is its own module (`schema_vNNNN.rs`) that returns a `Migration` struct:

```rust
pub fn migration() -> Migration {
    Migration::new("v0.2.0", UP_STATEMENTS, DOWN_STATEMENTS)
}

static UP_STATEMENTS: &[&str] = &[r#"CREATE TABLE ..."#, r#"CREATE INDEX ..."#];
static DOWN_STATEMENTS: &[&str] = &[r#"DROP TABLE IF EXISTS ... CASCADE;"#];
```

Register new migrations in `catalog.rs` via `default_migration_catalog()`. The runner applies missing migrations in version order and records them in a `migrations` tracking table.

---

## Naming Conventions

| Object | Convention | Example |
|--------|-----------|---------|
| Table | `snake_case` plural | `nodes` |
| Column | `snake_case` | `parent_id`, `is_deleted` |
| Index | `idx_{table}_{column}` | `idx_nodes_server_version` |
| Unique index | `{table}_{columns}_unique` | `nodes_parent_name_unique` |
| Trigger | `trg_{table}_{purpose}` | `trg_nodes_version` |

---

## Common Mistakes

- **Including `server_version` in INSERT/UPDATE** — the trigger overwrites it anyway, but the column list mismatch can cause confusing errors.
- **Hard-coding path separators** — use `Path`/`PathBuf`/`.join()` for filesystem paths. For virtual path parsing, split on both `/` and `\`.
- **Using `NOW()` in SQL for timestamps** — timestamps come from the app via `TimestampSource`, not the DB clock.
- **Forgetting `is_deleted = FALSE` in WHERE clauses** — soft-deleted rows are invisible by convention, but the filter must be explicit.
- **Adding a mapped domain field without updating every SELECT/RETURNING** — when a column is added to a struct that a row→struct mapper reads (e.g. `row_to_node` → `Node`, which does `row.try_get("md5")` / `try_get("storage_status")`), EVERY query whose result flows through that mapper must list the new column in its `SELECT`/`RETURNING` — not only the write paths. `list_children` once omitted `md5`/`storage_status` after they were added to `Node`: it compiled, type-checked, and passed empty-result tests, then threw `no column found for name: md5` at runtime the instant any row existed. When you add a mapped field, grep for all `SELECT`/`RETURNING` lists that feed the mapper and update them in the same change.
