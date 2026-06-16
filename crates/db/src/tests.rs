use pretty_assertions::assert_eq;

use crate::{DatabaseError, Migration, MigrationCatalog, default_migration_catalog};

/// Verifies the default catalog builds without errors and contains the expected migration versions.
#[test]
fn default_catalog_contains_expected_migrations() {
    let catalog = default_migration_catalog().unwrap();

    assert_eq!(
        catalog
            .target_versions()
            .iter()
            .copied()
            .collect::<Vec<_>>(),
        vec!["v0.1.0", "v0.2.0"]
    );
}

/// Verifies catalog construction rejects duplicate version identifiers.
#[test]
fn rejects_duplicate_migration_versions() {
    let error = MigrationCatalog::new(vec![
        Migration::new("v0.1.0", &[], &[]),
        Migration::new("v0.1.0", &[], &[]),
    ])
    .unwrap_err();

    assert_eq!(
        matches!(error, DatabaseError::DuplicateMigrationVersion(ref v) if v == "v0.1.0"),
        true
    );
}

/// Verifies that target versions must form a prefix of the full catalog list.
#[test]
fn rejects_target_versions_that_diverge_from_catalog_prefix() {
    let migrations = vec![
        Migration::new("v0.1.0", &[], &[]),
        Migration::new("v0.2.0", &[], &[]),
        Migration::new("v0.3.0", &[], &[]),
    ];

    let error =
        MigrationCatalog::with_target_versions(migrations, vec!["v0.1.0", "v0.3.0"]).unwrap_err();

    assert_eq!(
        matches!(
            error,
            DatabaseError::InvalidTargetVersionPrefix {
                position: 1,
                ref expected,
                ref found,
            } if expected == "v0.2.0" && found == "v0.3.0"
        ),
        true
    );
}

/// Verifies that a migration can be looked up by version string after catalog construction.
#[test]
fn catalog_lookup_returns_migration_for_known_version() {
    let migration = Migration::new(
        "v0.1.0",
        &["CREATE TABLE foo (id INT);"],
        &["DROP TABLE foo;"],
    );
    let catalog = MigrationCatalog::new(vec![migration.clone()]).unwrap();

    assert_eq!(catalog.migration("v0.1.0"), Some(&migration));
    assert_eq!(catalog.migration("v9.9.9"), None);
}
