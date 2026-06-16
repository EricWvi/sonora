use super::Migration;

/// v0.1.0 — creates the global server-version sequence and the trigger function that all
/// tables share for monotonically increasing `server_version` values.
///
/// Every write (INSERT or UPDATE) to a tracked table fires `global_bump_server_version`, which
/// pulls the next value from `global_sync_version_seq`. This guarantees a strict total order
/// across all table writes, enabling clients to sync by querying rows whose
/// `server_version > last_known_version`.
pub fn migration() -> Migration {
    Migration::new("v0.1.0", UP_STATEMENTS, DOWN_STATEMENTS)
}

static UP_STATEMENTS: &[&str] = &[r#"
CREATE SEQUENCE global_sync_version_seq;

CREATE OR REPLACE FUNCTION global_bump_server_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.server_version = nextval('global_sync_version_seq');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
"#];

static DOWN_STATEMENTS: &[&str] = &[r#"
DROP FUNCTION IF EXISTS global_bump_server_version;
DROP SEQUENCE IF EXISTS global_sync_version_seq;
"#];
