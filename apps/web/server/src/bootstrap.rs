use std::sync::Arc;

use sonora_db::{
    Database, DatabaseBootstrapper, DatabaseLocation, SystemTimestampSource,
    default_migration_catalog,
};

use crate::app_state::AppState;
use crate::error::WebBootstrapError;
use crate::service::NodeApi;

/// Bootstraps the database and assembles the ready `AppState`.
pub async fn bootstrap(database_url: &str, timezone: &str) -> Result<AppState, WebBootstrapError> {
    let db = bootstrap_database(database_url, timezone).await?;
    let pool = db.into_pool();
    let node_api = Arc::new(NodeApi::new(pool));
    Ok(AppState { node_api })
}

async fn bootstrap_database(url: &str, timezone: &str) -> Result<Database, WebBootstrapError> {
    let catalog = default_migration_catalog().map_err(WebBootstrapError::Database)?;
    let location = DatabaseLocation::new(url);

    DatabaseBootstrapper::<SystemTimestampSource>::system()
        .bootstrap_with_timezone(&location, &catalog, Some(timezone))
        .await
        .map_err(WebBootstrapError::Database)
}
