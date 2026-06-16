use std::collections::BTreeMap;

use crate::DatabaseError;

use super::{schema_v0001, schema_v0002, schema_v0003, schema_v0004};

/// Captures one versioned migration and the SQL needed to move schema state up or down.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Migration {
    version: &'static str,
    up_statements: &'static [&'static str],
    down_statements: &'static [&'static str],
}

impl Migration {
    /// Builds a migration from its ordered version and paired SQL statement lists.
    pub fn new(
        version: &'static str,
        up_statements: &'static [&'static str],
        down_statements: &'static [&'static str],
    ) -> Self {
        Self {
            version,
            up_statements,
            down_statements,
        }
    }

    /// Returns the stable version identifier used for persistence and logging.
    pub fn version(&self) -> &'static str {
        self.version
    }

    /// Returns the forward SQL statements that install this migration.
    pub fn up_statements(&self) -> &'static [&'static str] {
        self.up_statements
    }

    /// Returns the rollback SQL statements that remove this migration's schema changes.
    pub fn down_statements(&self) -> &'static [&'static str] {
        self.down_statements
    }
}

/// Holds every migration definition plus the active target prefix the bootstrapper should enforce.
#[derive(Clone, Debug)]
pub struct MigrationCatalog {
    migrations: Vec<Migration>,
    target_versions: Vec<&'static str>,
    migrations_by_version: BTreeMap<&'static str, usize>,
}

impl MigrationCatalog {
    /// Builds a catalog whose active target includes every supplied migration.
    pub fn new(migrations: Vec<Migration>) -> Result<Self, DatabaseError> {
        let target_versions = migrations.iter().map(Migration::version).collect();

        Self::with_target_versions(migrations, target_versions)
    }

    /// Builds a catalog with an explicit active target prefix for controlled rollback scenarios.
    pub fn with_target_versions(
        migrations: Vec<Migration>,
        target_versions: Vec<&'static str>,
    ) -> Result<Self, DatabaseError> {
        validate_migration_versions(&migrations)?;
        validate_target_prefix(&migrations, &target_versions)?;

        let migrations_by_version = migrations
            .iter()
            .enumerate()
            .map(|(index, migration)| (migration.version(), index))
            .collect();

        Ok(Self {
            migrations,
            target_versions,
            migrations_by_version,
        })
    }

    /// Returns the active target versions the database should match after reconciliation.
    pub fn target_versions(&self) -> &[&'static str] {
        &self.target_versions
    }

    /// Finds a migration definition by version so reconciliation can execute it.
    pub fn migration(&self, version: &str) -> Option<&Migration> {
        self.migrations_by_version
            .get(version)
            .map(|index| &self.migrations[*index])
    }
}

/// Builds the default migration catalog shipped by the crate.
pub fn default_migration_catalog() -> Result<MigrationCatalog, DatabaseError> {
    MigrationCatalog::new(vec![
        schema_v0001::migration(),
        schema_v0002::migration(),
        schema_v0003::migration(),
        schema_v0004::migration(),
    ])
}

/// Validates that migration versions are unique.
/// Version strings are semantic labels (e.g. "v2.10.0") and do not necessarily sort
/// lexicographically in release order, so only uniqueness is enforced here. The application
/// order is determined by insertion order in the Vec.
fn validate_migration_versions(migrations: &[Migration]) -> Result<(), DatabaseError> {
    let mut seen_versions = BTreeMap::new();

    for migration in migrations {
        if seen_versions.insert(migration.version(), ()).is_some() {
            return Err(DatabaseError::DuplicateMigrationVersion(
                migration.version().to_string(),
            ));
        }
    }

    Ok(())
}

/// Validates that the active target is a prefix of the full catalog so rollback remains deterministic.
fn validate_target_prefix(
    migrations: &[Migration],
    target_versions: &[&'static str],
) -> Result<(), DatabaseError> {
    for (position, target_version) in target_versions.iter().enumerate() {
        let expected = migrations
            .get(position)
            .map(Migration::version)
            .unwrap_or_default();

        if target_version != &expected {
            return Err(DatabaseError::InvalidTargetVersionPrefix {
                position,
                expected: expected.to_string(),
                found: (*target_version).to_string(),
            });
        }
    }

    Ok(())
}
