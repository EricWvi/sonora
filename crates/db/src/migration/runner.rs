use sonora_logging::{sonora_error, sonora_info};
use sqlx::{Pool, Postgres, Row};

use crate::{DatabaseError, MigrationCatalog, MigrationDirection, TimestampSource};

use super::{AppliedMigration, Migration};

/// Applied migrations are ordered by the SERIAL sequence column rather than version string because
/// semantic versions like "v2.10.0" do not sort lexicographically after "v2.8.0".
const CREATE_MIGRATIONS_TABLE_SQL: &str = r#"
CREATE TABLE IF NOT EXISTS migrations (
    sequence    SERIAL NOT NULL,
    version     TEXT   PRIMARY KEY,
    executed_at BIGINT NOT NULL
);
"#;

/// Reconciles a PostgreSQL pool with the catalog's target prefix by applying or rolling back migrations.
pub async fn reconcile_database<T>(
    pool: &Pool<Postgres>,
    catalog: &MigrationCatalog,
    timestamp_source: &T,
) -> Result<(), DatabaseError>
where
    T: TimestampSource,
{
    ensure_migrations_table(pool).await?;

    let applied_migrations = load_applied_migrations(pool).await?;
    let target_versions = catalog.target_versions();
    let shared_prefix_length = applied_migrations.len().min(target_versions.len());
    let pending_up_count = target_versions
        .len()
        .saturating_sub(applied_migrations.len());
    let pending_down_count = applied_migrations
        .len()
        .saturating_sub(target_versions.len());

    sonora_info!(
        message = "evaluated migration reconciliation",
        operation = "migration_reconciliation",
        applied_migration_count = applied_migrations.len(),
        target_migration_count = target_versions.len(),
        pending_up_count,
        pending_down_count
    );

    for (position, (applied, expected)) in applied_migrations
        .iter()
        .zip(target_versions.iter())
        .take(shared_prefix_length)
        .enumerate()
    {
        if applied.version != *expected {
            sonora_error!(
                message = "migration history diverged",
                operation = "migration_reconciliation",
                migration_position = position,
                error.kind = "diverged_migration_history",
                error.message = format!(
                    "expected migration version {expected}, found {}",
                    applied.version
                )
            );
            return Err(DatabaseError::DivergedMigrationHistory {
                position,
                expected: (*expected).to_string(),
                found: applied.version.clone(),
            });
        }
    }

    if applied_migrations.len() > target_versions.len() {
        sonora_info!(
            message = "rolling back trailing migrations",
            operation = "migration_reconciliation",
            rollback_count = pending_down_count
        );

        for applied_migration in applied_migrations.iter().skip(target_versions.len()).rev() {
            let migration = catalog
                .migration(&applied_migration.version)
                .ok_or_else(|| {
                    sonora_error!(
                        message = "encountered unknown applied migration version",
                        operation = "migration_reconciliation",
                        migration_version = applied_migration.version.clone(),
                        error.kind = "unknown_applied_migration_version",
                        error.message = format!(
                            "database contains unknown applied migration version {}",
                            applied_migration.version
                        )
                    );
                    DatabaseError::UnknownAppliedMigrationVersion {
                        version: applied_migration.version.clone(),
                    }
                })?;

            execute_migration(
                pool,
                migration,
                MigrationDirection::Down,
                /*executed_at=*/ 0,
            )
            .await?;
        }
    }

    if target_versions.len() > applied_migrations.len() {
        sonora_info!(
            message = "applying pending migrations",
            operation = "migration_reconciliation",
            apply_count = pending_up_count
        );

        for target_version in target_versions.iter().skip(applied_migrations.len()) {
            let migration = catalog.migration(target_version).ok_or_else(|| {
                sonora_error!(
                    message = "target migration version is missing from the catalog",
                    operation = "migration_reconciliation",
                    migration_version = (*target_version).to_string(),
                    error.kind = "unknown_applied_migration_version",
                    error.message = format!(
                        "target migration version {target_version} is missing from the catalog"
                    )
                );
                DatabaseError::UnknownAppliedMigrationVersion {
                    version: (*target_version).to_string(),
                }
            })?;

            let executed_at = timestamp_source.current_timestamp_millis();
            execute_migration(pool, migration, MigrationDirection::Up, executed_at).await?;
        }
    }

    if pending_up_count == 0 && pending_down_count == 0 {
        sonora_info!(
            message = "database schema already matches the target migration prefix",
            operation = "migration_reconciliation"
        );
    }

    Ok(())
}

/// Ensures the bookkeeping table exists before reconciliation starts reading or mutating migration state.
async fn ensure_migrations_table(pool: &Pool<Postgres>) -> Result<(), DatabaseError> {
    sqlx::raw_sql(CREATE_MIGRATIONS_TABLE_SQL)
        .execute(pool)
        .await?;
    Ok(())
}

/// Loads applied migration rows ordered by insertion sequence so prefix comparison is deterministic
/// regardless of whether version strings sort lexicographically in release order.
async fn load_applied_migrations(
    pool: &Pool<Postgres>,
) -> Result<Vec<AppliedMigration>, DatabaseError> {
    let rows = sqlx::query("SELECT version, executed_at FROM migrations ORDER BY sequence ASC")
        .fetch_all(pool)
        .await?;

    Ok(rows
        .iter()
        .map(|row| {
            AppliedMigration::new(
                row.get::<String, _>("version"),
                row.get::<i64, _>("executed_at"),
            )
        })
        .collect())
}

/// Executes one migration step inside a single transaction so the DDL and the bookkeeping record
/// are committed together — leaving no window where schema and migration history can diverge.
async fn execute_migration(
    pool: &Pool<Postgres>,
    migration: &Migration,
    direction: MigrationDirection,
    executed_at: i64,
) -> Result<(), DatabaseError> {
    let version = migration.version();
    sonora_info!(
        message = "executing migration step",
        operation = "migration_execute",
        migration_version = version,
        direction = direction.to_string()
    );

    let statements = match direction {
        MigrationDirection::Up => migration.up_statements(),
        MigrationDirection::Down => migration.down_statements(),
    };

    let mut tx = pool.begin().await?;

    for statement in statements {
        sqlx::raw_sql(sqlx::AssertSqlSafe(*statement))
            .execute(&mut *tx)
            .await
            .map_err(|source| {
                sonora_error!(
                    message = "migration step failed",
                    operation = "migration_execute",
                    migration_version = version,
                    direction = direction.to_string(),
                    error.kind = "migration_step_failed",
                    error.message = source.to_string()
                );
                DatabaseError::MigrationStepFailed {
                    version: version.to_string(),
                    direction,
                    source,
                }
            })?;
    }

    let bookkeeping_result = match direction {
        MigrationDirection::Up => {
            sqlx::query("INSERT INTO migrations (version, executed_at) VALUES ($1, $2)")
                .bind(version)
                .bind(executed_at)
                .execute(&mut *tx)
                .await
        }
        MigrationDirection::Down => {
            sqlx::query("DELETE FROM migrations WHERE version = $1")
                .bind(version)
                .execute(&mut *tx)
                .await
        }
    };

    bookkeeping_result.map_err(|source| {
        sonora_error!(
            message = "migration bookkeeping failed",
            operation = "migration_execute",
            migration_version = version,
            direction = direction.to_string(),
            error.kind = "migration_bookkeeping_failed",
            error.message = source.to_string()
        );
        DatabaseError::MigrationStepFailed {
            version: version.to_string(),
            direction,
            source,
        }
    })?;

    tx.commit().await?;

    sonora_info!(
        message = "executed migration step",
        operation = "migration_execute",
        migration_version = version,
        direction = direction.to_string()
    );

    Ok(())
}
