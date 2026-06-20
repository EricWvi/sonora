use super::Migration;

/// v0.2.0 — creates the `nodes` table that backs the virtual file system.
///
/// `parent_id IS NULL` means the node sits at the VFS root. The UNIQUE index uses
/// `NULLS NOT DISTINCT` so that two root-level nodes with the same name are rejected even
/// though both have `parent_id = NULL`. The partial `WHERE is_deleted = FALSE` clause allows
/// the same name to be reused after a node has been soft-deleted.
///
/// `server_version` is set automatically by the `global_bump_server_version` trigger
/// (installed in v0.1.0); callers must not supply it in INSERT or UPDATE statements.
pub fn migration() -> Migration {
    Migration::new("v0.2.0", UP_STATEMENTS, DOWN_STATEMENTS)
}

static UP_STATEMENTS: &[&str] = &[
    r#"
CREATE TABLE nodes (
    id              UUID         NOT NULL PRIMARY KEY,
    parent_id       UUID         REFERENCES nodes(id),
    name            VARCHAR(1024) NOT NULL,
    kind            TEXT         NOT NULL CHECK (kind IN ('directory', 'file')),
    size            BIGINT,
    mime_type       TEXT,
    md5             VARCHAR(32),
    storage_status  INT          NOT NULL DEFAULT 0,
    created_at      BIGINT       NOT NULL,
    updated_at      BIGINT       NOT NULL,
    server_version  BIGINT       NOT NULL,
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE
);
"#,
    r#"
CREATE UNIQUE INDEX nodes_parent_name_unique
    ON nodes (parent_id, name) NULLS NOT DISTINCT
    WHERE is_deleted = FALSE;
"#,
    r#"
CREATE INDEX idx_nodes_server_version ON nodes USING btree (server_version);
"#,
    r#"
CREATE TRIGGER trg_nodes_version
BEFORE INSERT OR UPDATE ON nodes
FOR EACH ROW EXECUTE FUNCTION global_bump_server_version();
"#,
];

static DOWN_STATEMENTS: &[&str] = &[r#"
DROP TABLE IF EXISTS nodes CASCADE;
"#];
