use sonora_db::DatabaseError;
use sonora_logging::LoggingInitError;
use thiserror::Error;

/// Errors that can occur during server bootstrap.
#[derive(Debug, Error)]
pub enum WebBootstrapError {
    #[error("SONORA_DB_HOST is required and must not be empty")]
    DatabaseHostEmpty,

    #[error("SONORA_DB_NAME is required and must not be empty")]
    DatabaseNameEmpty,

    #[error("SONORA_DB_USERNAME is required and must not be empty")]
    DatabaseUsernameEmpty,

    #[error("SONORA_DB_PASSWORD is required and must not be empty")]
    DatabasePasswordEmpty,

    #[error("invalid SONORA_DB_PORT value {value:?}: expected a valid port number")]
    DatabasePortInvalid {
        value: String,
        #[source]
        source: std::num::ParseIntError,
    },

    #[error(
        "invalid SONORA_LOG_LEVEL value {value:?}: expected 'debug', 'info', 'warn', or 'error'"
    )]
    LogLevelInvalid { value: String },

    #[error("logging initialization failed: {0}")]
    LoggingInit(#[from] LoggingInitError),

    #[error("database bootstrap failed: {0}")]
    Database(#[from] DatabaseError),
}
