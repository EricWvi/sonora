/// Wraps a PostgreSQL connection string so callers don't pass raw strings through the bootstrap API.
#[derive(Clone, Debug)]
pub struct DatabaseLocation {
    connection_string: String,
}

impl DatabaseLocation {
    /// Builds a location from a libpq-compatible connection string or URL.
    pub fn new(connection_string: impl Into<String>) -> Self {
        Self {
            connection_string: connection_string.into(),
        }
    }

    /// Returns the connection string for use by the tokio-postgres connector.
    pub fn connection_string(&self) -> &str {
        &self.connection_string
    }
}
