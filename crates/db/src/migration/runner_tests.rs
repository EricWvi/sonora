use sonora_logging::set_trace_logging;
use pretty_assertions::assert_eq;
use sqlx::{Pool, Postgres};
use testcontainers::ImageExt;
use testcontainers::runners::AsyncRunner;
use testcontainers_modules::postgres::Postgres as PgContainer;

use crate::{
    DatabaseBootstrapper, DatabaseError, DatabaseLocation, Migration, MigrationCatalog,
    SystemTimestampSource, default_migration_catalog,
};

/// Starts a fresh postgres container, enables the extensions required by our migrations,
/// and returns a handle that keeps the container alive for the duration of the test.
async fn start_postgres() -> (
    testcontainers::ContainerAsync<PgContainer>,
    DatabaseLocation,
) {
    let container = PgContainer::default()
        .with_tag("17-alpine")
        .start()
        .await
        .expect("postgres container failed to start");
    let port = container
        .get_host_port_ipv4(5432)
        .await
        .expect("postgres port not available");
    let url = format!("postgres://postgres:postgres@127.0.0.1:{port}/postgres");

    // pgcrypto provides gen_random_uuid() used by schema_v0001;
    // pg_trgm provides gin trigram operators used by schema_v0002.
    let pool = Pool::<Postgres>::connect(&url)
        .await
        .expect("failed to connect for extension setup");
    sqlx::raw_sql(
        "CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS pg_trgm;",
    )
    .execute(&pool)
    .await
    .expect("failed to enable extensions");
    pool.close().await;

    (container, DatabaseLocation::new(url))
}

fn bootstrapper() -> DatabaseBootstrapper<SystemTimestampSource> {
    DatabaseBootstrapper::system()
}

/// Queries the migrations table and returns version strings in application order.
async fn applied_versions(location: &DatabaseLocation) -> Vec<String> {
    let pool = Pool::<Postgres>::connect(location.connection_string())
        .await
        .expect("failed to connect to postgres");
    let versions: Vec<String> =
        sqlx::query_scalar("SELECT version FROM migrations ORDER BY sequence ASC")
            .fetch_all(&pool)
            .await
            .expect("failed to query migrations table");
    pool.close().await;
    versions
}

/// Verifies that bootstrapping an empty database applies every migration in catalog order.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn fresh_database_applies_all_migrations() {
    let _guard = set_trace_logging();

    let (_container, location) = start_postgres().await;
    let catalog = default_migration_catalog().expect("catalog build failed");

    bootstrapper()
        .bootstrap(&location, &catalog)
        .await
        .expect("bootstrap failed");

    assert_eq!(
        applied_versions(&location).await,
        vec!["v0.1.0", "v2.7.0", "v2.8.0", "v2.10.0"]
    );
}

/// Verifies that running bootstrap twice against an already-migrated database is a no-op.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn reconcile_is_idempotent() {
    let _guard = set_trace_logging();

    let (_container, location) = start_postgres().await;
    let catalog = default_migration_catalog().expect("catalog build failed");
    let bs = bootstrapper();

    bs.bootstrap(&location, &catalog)
        .await
        .expect("first bootstrap failed");
    bs.bootstrap(&location, &catalog)
        .await
        .expect("second bootstrap failed");

    assert_eq!(
        applied_versions(&location).await,
        vec!["v0.1.0", "v2.7.0", "v2.8.0", "v2.10.0"]
    );
}

/// Verifies that narrowing the target prefix causes trailing migrations to be rolled back.
/// All migration definitions must remain in the catalog so the runner can find their
/// down-statements even after the target prefix shrinks.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn rollback_removes_trailing_migrations() {
    let _guard = set_trace_logging();

    let (_container, location) = start_postgres().await;
    let full_catalog = default_migration_catalog().expect("catalog build failed");
    let bs = bootstrapper();

    bs.bootstrap(&location, &full_catalog)
        .await
        .expect("full bootstrap failed");

    let all_migrations: Vec<Migration> = full_catalog
        .target_versions()
        .iter()
        .map(|v| {
            full_catalog
                .migration(v)
                .expect("migration missing from catalog")
                .clone()
        })
        .collect();
    let partial_catalog =
        MigrationCatalog::with_target_versions(all_migrations, vec!["v0.1.0", "v2.7.0"])
            .expect("partial catalog build failed");

    bs.bootstrap(&location, &partial_catalog)
        .await
        .expect("rollback bootstrap failed");

    assert_eq!(applied_versions(&location).await, vec!["v0.1.0", "v2.7.0"]);
}

/// Verifies that reconciling against a database whose applied history diverges from the
/// catalog returns DivergedMigrationHistory at the correct position.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn detects_diverged_migration_history() {
    let _guard = set_trace_logging();

    let (_container, location) = start_postgres().await;
    let bs = bootstrapper();

    // Seed the database with a fictitious migration that will collide with v0.1.0
    let diverged_catalog = MigrationCatalog::new(vec![Migration::new("v99.0.0", &[], &[])])
        .expect("diverged catalog build failed");
    bs.bootstrap(&location, &diverged_catalog)
        .await
        .expect("diverged bootstrap failed");

    // Reconciling against the real catalog must detect the divergence at position 0
    let real_catalog = default_migration_catalog().expect("real catalog build failed");
    let error = bs
        .bootstrap(&location, &real_catalog)
        .await
        .expect_err("expected divergence error but bootstrap succeeded");

    assert_eq!(
        matches!(
            error,
            DatabaseError::DivergedMigrationHistory {
                position: 0,
                ref expected,
                ref found,
            } if expected == "v0.1.0" && found == "v99.0.0"
        ),
        true
    );
}
