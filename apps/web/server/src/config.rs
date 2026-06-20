use sonora_logging::LogLevel;

use crate::error::WebBootstrapError;

const DB_HOST_VAR: &str = "SONORA_DB_HOST";
const DB_PORT_VAR: &str = "SONORA_DB_PORT";
const DB_NAME_VAR: &str = "SONORA_DB_NAME";
const DB_USERNAME_VAR: &str = "SONORA_DB_USERNAME";
const DB_PASSWORD_VAR: &str = "SONORA_DB_PASSWORD";
const DB_SSL_MODE_VAR: &str = "SONORA_DB_SSL_MODE";
const DB_TIMEZONE_VAR: &str = "SONORA_DB_TIMEZONE";

const DEFAULT_DB_PORT: &str = "5432";
const DEFAULT_DB_SSL_MODE: &str = "disable";
const DEFAULT_DB_TIMEZONE: &str = "UTC";

const LOG_LEVEL_VAR: &str = "SONORA_LOG_LEVEL";
const DEFAULT_LOG_LEVEL: &str = "info";

/// Runtime configuration for the PostgreSQL database connection.
#[derive(Debug)]
pub struct DatabaseRuntimeConfig {
    host: String,
    port: u16,
    name: String,
    username: String,
    password: String,
    ssl_mode: String,
    timezone: String,
}

impl DatabaseRuntimeConfig {
    /// Reads configuration from the process environment.
    pub fn from_env() -> Result<Self, WebBootstrapError> {
        Self::from_reader(|key| std::env::var(key).ok())
    }

    /// Reads configuration from a caller-supplied variable reader, enabling test isolation.
    pub fn from_reader(
        mut read_var: impl FnMut(&str) -> Option<String>,
    ) -> Result<Self, WebBootstrapError> {
        let host = required_non_empty(
            &mut read_var,
            DB_HOST_VAR,
            WebBootstrapError::DatabaseHostEmpty,
        )?;
        let name = required_non_empty(
            &mut read_var,
            DB_NAME_VAR,
            WebBootstrapError::DatabaseNameEmpty,
        )?;
        let username = required_non_empty(
            &mut read_var,
            DB_USERNAME_VAR,
            WebBootstrapError::DatabaseUsernameEmpty,
        )?;
        let password = required_non_empty(
            &mut read_var,
            DB_PASSWORD_VAR,
            WebBootstrapError::DatabasePasswordEmpty,
        )?;

        let port_raw = read_var(DB_PORT_VAR).unwrap_or_else(|| DEFAULT_DB_PORT.to_string());
        let port =
            port_raw
                .parse::<u16>()
                .map_err(|source| WebBootstrapError::DatabasePortInvalid {
                    value: port_raw.clone(),
                    source,
                })?;

        let ssl_mode = read_var(DB_SSL_MODE_VAR).unwrap_or_else(|| DEFAULT_DB_SSL_MODE.to_string());
        let timezone = read_var(DB_TIMEZONE_VAR).unwrap_or_else(|| DEFAULT_DB_TIMEZONE.to_string());

        Ok(Self {
            host,
            port,
            name,
            username,
            password,
            ssl_mode,
            timezone,
        })
    }

    /// Builds a libpq-compatible PostgreSQL connection URL from the configured fields.
    pub fn connection_string(&self) -> String {
        format!(
            "postgres://{}:{}@{}:{}/{}?sslmode={}",
            self.username, self.password, self.host, self.port, self.name, self.ssl_mode,
        )
    }

    /// Returns the configured PostgreSQL session timezone.
    pub fn timezone(&self) -> &str {
        &self.timezone
    }
}

/// Reads the log level from the process environment.
pub fn read_log_level() -> Result<LogLevel, WebBootstrapError> {
    read_log_level_from(|key| std::env::var(key).ok())
}

fn read_log_level_from(
    mut read_var: impl FnMut(&str) -> Option<String>,
) -> Result<LogLevel, WebBootstrapError> {
    let raw = read_var(LOG_LEVEL_VAR).unwrap_or_else(|| DEFAULT_LOG_LEVEL.to_string());
    match raw.to_ascii_lowercase().as_str() {
        "debug" => Ok(LogLevel::Debug),
        "info" => Ok(LogLevel::Info),
        "warn" => Ok(LogLevel::Warn),
        "error" => Ok(LogLevel::Error),
        _ => Err(WebBootstrapError::LogLevelInvalid { value: raw }),
    }
}

fn required_non_empty(
    read_var: &mut impl FnMut(&str) -> Option<String>,
    key: &str,
    err: WebBootstrapError,
) -> Result<String, WebBootstrapError> {
    match read_var(key) {
        Some(v) if !v.trim().is_empty() => Ok(v),
        _ => Err(err),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use pretty_assertions::assert_eq;

    fn db_vars(key: &str) -> Option<String> {
        match key {
            DB_HOST_VAR => Some("localhost".to_string()),
            DB_NAME_VAR => Some("sonora".to_string()),
            DB_USERNAME_VAR => Some("user".to_string()),
            DB_PASSWORD_VAR => Some("pass".to_string()),
            _ => None,
        }
    }

    /// Verifies that all required fields must be present.
    #[test]
    fn rejects_missing_host() {
        let err = DatabaseRuntimeConfig::from_reader(|_| None).unwrap_err();
        assert!(matches!(err, WebBootstrapError::DatabaseHostEmpty));
    }

    /// Verifies that blank host is rejected.
    #[test]
    fn rejects_blank_host() {
        let err = DatabaseRuntimeConfig::from_reader(|key| match key {
            DB_HOST_VAR => Some("   ".to_string()),
            _ => None,
        })
        .unwrap_err();
        assert!(matches!(err, WebBootstrapError::DatabaseHostEmpty));
    }

    /// Verifies ssl_mode and timezone defaults and the connection string format.
    #[test]
    fn applies_defaults_and_builds_connection_string() {
        let config = DatabaseRuntimeConfig::from_reader(db_vars).unwrap();
        assert_eq!(
            config.connection_string(),
            "postgres://user:pass@localhost:5432/sonora?sslmode=disable"
        );
        assert_eq!(config.timezone(), "UTC");
    }

    /// Verifies custom ssl_mode and timezone override the defaults.
    #[test]
    fn uses_custom_ssl_mode_and_timezone() {
        let config = DatabaseRuntimeConfig::from_reader(|key| match key {
            DB_SSL_MODE_VAR => Some("require".to_string()),
            DB_TIMEZONE_VAR => Some("Asia/Shanghai".to_string()),
            _ => db_vars(key),
        })
        .unwrap();
        assert_eq!(
            config.connection_string(),
            "postgres://user:pass@localhost:5432/sonora?sslmode=require"
        );
        assert_eq!(config.timezone(), "Asia/Shanghai");
    }

    /// Verifies invalid port value is rejected.
    #[test]
    fn rejects_invalid_port() {
        let err = DatabaseRuntimeConfig::from_reader(|key| match key {
            DB_PORT_VAR => Some("not-a-port".to_string()),
            _ => db_vars(key),
        })
        .unwrap_err();
        assert!(matches!(
            err,
            WebBootstrapError::DatabasePortInvalid { value, .. } if value == "not-a-port"
        ));
    }

    /// Verifies valid log levels are accepted.
    #[test]
    fn parses_log_levels() {
        for (input, expected) in [
            ("debug", LogLevel::Debug),
            ("INFO", LogLevel::Info),
            ("Warn", LogLevel::Warn),
            ("ERROR", LogLevel::Error),
        ] {
            let level = read_log_level_from(|_| Some(input.to_string())).unwrap();
            assert_eq!(level, expected);
        }
    }

    /// Verifies default log level when env var is absent.
    #[test]
    fn defaults_log_level_to_info() {
        let level = read_log_level_from(|_| None).unwrap();
        assert_eq!(level, LogLevel::Info);
    }

    /// Verifies invalid log level is rejected.
    #[test]
    fn rejects_invalid_log_level() {
        let err = read_log_level_from(|_| Some("verbose".to_string())).unwrap_err();
        assert!(matches!(
            err,
            WebBootstrapError::LogLevelInvalid { value } if value == "verbose"
        ));
    }
}
