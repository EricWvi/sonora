mod node;

use std::sync::Arc;

use axum::body::Body;
use axum::http::{Request, Response};
use sonora_db::{DatabaseBootstrapper, DatabaseLocation, default_migration_catalog};
use sonora_logging::set_trace_logging;
use testcontainers::ContainerAsync;
use testcontainers::ImageExt;
use testcontainers::runners::AsyncRunner;
use testcontainers_modules::postgres::Postgres as PgContainer;
use tower::ServiceExt;

use crate::app_state::AppState;
use crate::routes::build_router;
use crate::service::NodeApi;

/// Starts a Postgres 17 container, runs all migrations, and returns a fully wired `AppState`.
///
/// The returned `ContainerAsync` must be held by the caller for the lifetime of the test;
/// dropping it stops the container.
pub async fn bootstrap_test_state() -> (ContainerAsync<PgContainer>, AppState) {
    let _guard = set_trace_logging();

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
    let location = DatabaseLocation::new(url);
    let catalog = default_migration_catalog().expect("catalog build failed");

    let db = DatabaseBootstrapper::system()
        .bootstrap(&location, &catalog)
        .await
        .expect("database bootstrap failed");

    let pool = db.into_pool();
    let node_api = Arc::new(NodeApi::new(pool));
    let state = AppState { node_api };

    (container, state)
}

/// Sends a single request through the router built from `state` and returns the response.
pub async fn send(state: AppState, req: Request<Body>) -> Response<Body> {
    build_router(state)
        .oneshot(req)
        .await
        .expect("router oneshot failed")
}
