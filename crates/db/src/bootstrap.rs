use sonora_logging::{sonora_error, sonora_info};
use sqlx::postgres::{PgConnectOptions, PgPoolOptions};
use sqlx::{Pool, Postgres};
use std::str::FromStr;

use crate::{
    DatabaseError, DatabaseLocation, MigrationCatalog, SystemTimestampSource, TimestampSource,
    migration,
};

/// Owns a PostgreSQL connection pool that has already been reconciled with the active migration target.
#[derive(Debug, Clone)]
pub struct Database {
    pool: Pool<Postgres>,
}

impl Database {
    /// Exposes the managed connection pool for query and repository work.
    pub fn pool(&self) -> &Pool<Postgres> {
        &self.pool
    }

    /// Transfers ownership of the managed pool to callers that need direct control.
    pub fn into_pool(self) -> Pool<Postgres> {
        self.pool
    }
}

/// Coordinates opening PostgreSQL connection pools and reconciling them with the migration catalog.
#[derive(Debug)]
pub struct DatabaseBootstrapper<T> {
    timestamp_source: T,
}

impl DatabaseBootstrapper<SystemTimestampSource> {
    /// Builds a bootstrapper that timestamps applied migrations from the system clock.
    pub fn system() -> Self {
        Self::new(SystemTimestampSource)
    }
}

impl<T> DatabaseBootstrapper<T>
where
    T: TimestampSource,
{
    /// Builds a bootstrapper around a caller-provided timestamp source for deterministic tests.
    pub fn new(timestamp_source: T) -> Self {
        Self { timestamp_source }
    }

    /// Opens a connection pool, reconciles it with the target migration prefix, and returns the ready database.
    pub async fn bootstrap(
        &self,
        location: &DatabaseLocation,
        catalog: &MigrationCatalog,
    ) -> Result<Database, DatabaseError> {
        self.bootstrap_with_timezone(location, catalog, None).await
    }

    /// Opens a connection pool, applies an optional session timezone, reconciles it with the
    /// target migration prefix, and returns the ready database.
    pub async fn bootstrap_with_timezone(
        &self,
        location: &DatabaseLocation,
        catalog: &MigrationCatalog,
        timezone: Option<&str>,
    ) -> Result<Database, DatabaseError> {
        sonora_info!(
            message = "opening database pool",
            operation = "database_open"
        );

        let pool = match open_pool(location, timezone).await {
            Ok(pool) => pool,
            Err(error) => {
                sonora_error!(
                    message = "failed to open database pool",
                    operation = "database_open",
                    error.kind = "database_open",
                    error.message = error.to_string()
                );
                return Err(DatabaseError::Sqlx(error));
            }
        };

        sonora_info!(
            message = "opened database pool",
            operation = "database_open"
        );

        if let Err(error) =
            migration::reconcile_database(&pool, catalog, &self.timestamp_source).await
        {
            sonora_error!(
                message = "database bootstrap failed",
                operation = "database_bootstrap",
                error.kind = "database_bootstrap",
                error.message = error.to_string()
            );
            return Err(error);
        }

        sonora_info!(
            message = "database bootstrap complete",
            operation = "database_bootstrap"
        );

        Ok(Database { pool })
    }
}

/// Opens a PostgreSQL pool and applies the configured session timezone on every new connection.
async fn open_pool(
    location: &DatabaseLocation,
    timezone: Option<&str>,
) -> Result<Pool<Postgres>, sqlx::Error> {
    let connect_options = PgConnectOptions::from_str(location.connection_string())?;
    let pool_options = if let Some(timezone) = timezone.map(str::to_owned) {
        PgPoolOptions::new().after_connect(move |connection, _meta| {
            let timezone = timezone.clone();
            Box::pin(async move {
                sqlx::query("SELECT pg_catalog.set_config('TimeZone', $1, false)")
                    .bind(timezone)
                    .execute(connection)
                    .await?;
                Ok(())
            })
        })
    } else {
        PgPoolOptions::new()
    };

    pool_options.connect_with(connect_options).await
}
