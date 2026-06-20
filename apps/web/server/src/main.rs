mod app_state;
mod bootstrap;
mod config;
mod error;
mod handlers;
mod routes;
mod service;

use sonora_logging::{LogOutput, LoggingConfig, init_logging, sonora_info};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenvy::dotenv().ok();

    let log_level = config::read_log_level()?;
    let _logging_guard = init_logging(LoggingConfig::new(log_level, LogOutput::Stdout))?;

    let db_config = config::DatabaseRuntimeConfig::from_env()?;

    let state = bootstrap::bootstrap(&db_config.connection_string(), db_config.timezone()).await?;

    let router = routes::build_router(state);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], 8080));
    let listener = tokio::net::TcpListener::bind(addr).await?;

    sonora_info!(address = %addr, "server listening");

    axum::serve(listener, router)
        .with_graceful_shutdown(async {
            let _ = tokio::signal::ctrl_c().await;
        })
        .await?;

    Ok(())
}
