use crate::node::{NewFile, NodeRepository, NodeRepositoryError};
use pretty_assertions::assert_eq;
use sonora_domain::{Node, NodeKind};
use sonora_logging::set_trace_logging;
use testcontainers::ImageExt;
use testcontainers::runners::AsyncRunner;
use testcontainers_modules::postgres::Postgres as PgContainer;

use crate::{
    DatabaseBootstrapper, DatabaseLocation, SystemTimestampSource, default_migration_catalog,
};

use super::PostgresNodeRepository;

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
    (container, DatabaseLocation::new(url))
}

async fn make_repo(location: &DatabaseLocation) -> PostgresNodeRepository<SystemTimestampSource> {
    let catalog = default_migration_catalog().expect("catalog build failed");
    let db = DatabaseBootstrapper::system()
        .bootstrap(location, &catalog)
        .await
        .expect("bootstrap failed");
    PostgresNodeRepository::new(db.into_pool(), SystemTimestampSource)
}

// ── create_dir ─────────────────────────────────────────────────────────────

/// Creates a root-level directory and verifies all fields are stored correctly.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn create_dir_at_root() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let node = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect("create_dir failed");

    assert_eq!(node.name, "Rock");
    assert_eq!(node.kind, NodeKind::Directory);
    assert_eq!(node.parent_id, None);
    assert_eq!(node.size, None);
    assert_eq!(node.mime_type, None);
    assert!(!node.is_deleted);
}

/// Creates a nested directory and verifies the parent link is set.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn create_dir_nested() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let parent = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect("create parent failed");
    let child = repo
        .create_dir(Some(parent.id), "Classic Rock".to_string())
        .await
        .expect("create child failed");

    assert_eq!(child.parent_id, Some(parent.id));
    assert_eq!(child.name, "Classic Rock");
}

/// Two directories with the same name under the same parent must be rejected.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn create_dir_duplicate_name_is_rejected() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    repo.create_dir(None, "Rock".to_string())
        .await
        .expect("first create failed");

    let err = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect_err("expected NameConflict");

    assert!(matches!(err, NodeRepositoryError::NameConflict { .. }));
}

// ── create_file ────────────────────────────────────────────────────────────

/// Creates a file with size and MIME type and verifies the stored snapshot.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn create_file_with_metadata() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let dir = repo
        .create_dir(None, "Jazz".to_string())
        .await
        .expect("create dir failed");

    let file = repo
        .create_file(NewFile {
            parent_id: Some(dir.id),
            name: "track01.flac".to_string(),
            size: Some(10_485_760),
            mime_type: Some("audio/flac".to_string()),
        })
        .await
        .expect("create_file failed");

    assert_eq!(file.name, "track01.flac");
    assert_eq!(file.kind, NodeKind::File);
    assert_eq!(file.parent_id, Some(dir.id));
    assert_eq!(file.size, Some(10_485_760));
    assert_eq!(file.mime_type, Some("audio/flac".to_string()));
    assert!(!file.is_deleted);
}

// ── get_node_by_id ─────────────────────────────────────────────────────────

/// Round-trips a node through the repository and compares the full struct.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_node_by_id_returns_stored_node() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let created = repo
        .create_dir(None, "Soul".to_string())
        .await
        .expect("create failed");

    let fetched = repo
        .get_node_by_id(created.id)
        .await
        .expect("get failed")
        .expect("node should exist");

    assert_eq!(created, fetched);
}

/// Returns `None` for an ID that was never inserted.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_node_by_id_returns_none_for_unknown_id() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let result = repo
        .get_node_by_id(sonora_domain::NodeId::generate())
        .await
        .expect("get failed");

    assert_eq!(result, None);
}

// ── get_node_by_path ───────────────────────────────────────────────────────

/// Resolves a three-level virtual path to the correct file node.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_node_by_path_resolves_deep_path() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let rock = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect("create rock failed");
    let album = repo
        .create_dir(Some(rock.id), "Led Zeppelin IV".to_string())
        .await
        .expect("create album failed");
    let track = repo
        .create_file(NewFile {
            parent_id: Some(album.id),
            name: "Stairway to Heaven.flac".to_string(),
            size: Some(50_000_000),
            mime_type: Some("audio/flac".to_string()),
        })
        .await
        .expect("create track failed");

    let found = repo
        .get_node_by_path("Rock/Led Zeppelin IV/Stairway to Heaven.flac")
        .await
        .expect("path lookup failed")
        .expect("node should exist");

    assert_eq!(track, found);
}

/// Returns `None` when a path component is missing.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_node_by_path_returns_none_for_missing_component() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let result = repo
        .get_node_by_path("NonExistent")
        .await
        .expect("path lookup failed");

    assert_eq!(result, None);
}

//── list_children ──────────────────────────────────────────────────────────

/// Lists root-level children and verifies directories are returned before files.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn list_children_returns_directories_before_files() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let dir = repo
        .create_dir(None, "Blues".to_string())
        .await
        .expect("create dir failed");
    let _file = repo
        .create_file(NewFile {
            parent_id: None,
            name: "compilation.flac".to_string(),
            size: None,
            mime_type: None,
        })
        .await
        .expect("create file failed");

    let children = repo
        .list_children(None)
        .await
        .expect("list_children failed");

    assert_eq!(children.len(), 2);
    assert_eq!(children[0], dir);
    assert_eq!(children[1].name, "compilation.flac");
}

/// Returns an empty vec when the directory has no children.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn list_children_returns_empty_for_leaf_dir() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let dir = repo
        .create_dir(None, "Empty".to_string())
        .await
        .expect("create dir failed");

    let children = repo
        .list_children(Some(dir.id))
        .await
        .expect("list_children failed");

    assert_eq!(children, Vec::<Node>::new());
}

// ── move_node ──────────────────────────────────────────────────────────────

/// Renames a node in place (same parent, new name).
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn move_node_renames_in_place() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let node = repo
        .create_dir(None, "OldName".to_string())
        .await
        .expect("create failed");

    let renamed = repo
        .move_node(node.id, None, "NewName".to_string())
        .await
        .expect("move_node failed");

    assert_eq!(renamed.id, node.id);
    assert_eq!(renamed.name, "NewName");
    assert_eq!(renamed.parent_id, None);
}

/// Moves a node to a different parent directory.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn move_node_reparents_to_new_directory() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let src = repo
        .create_dir(None, "Source".to_string())
        .await
        .expect("create src failed");
    let dst = repo
        .create_dir(None, "Destination".to_string())
        .await
        .expect("create dst failed");
    let child = repo
        .create_dir(Some(src.id), "Child".to_string())
        .await
        .expect("create child failed");

    let moved = repo
        .move_node(child.id, Some(dst.id), "Child".to_string())
        .await
        .expect("move_node failed");

    assert_eq!(moved.parent_id, Some(dst.id));
    assert_eq!(moved.name, "Child");
}

/// Returns `NotFound` when the node being moved does not exist.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn move_node_returns_not_found_for_unknown_id() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let unknown = sonora_domain::NodeId::generate();
    let err = repo
        .move_node(unknown, None, "whatever".to_string())
        .await
        .expect_err("expected NotFound");

    assert!(matches!(err, NodeRepositoryError::NotFound(_)));
}

/// Moving a node to a destination where its name already exists returns NameConflict.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn move_node_returns_conflict_when_name_taken() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let _existing = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect("create existing failed");
    let other = repo
        .create_dir(None, "Jazz".to_string())
        .await
        .expect("create other failed");

    let err = repo
        .move_node(other.id, None, "Rock".to_string())
        .await
        .expect_err("expected NameConflict");

    assert!(matches!(err, NodeRepositoryError::NameConflict { .. }));
}

// ── delete_node ────────────────────────────────────────────────────────────

/// Soft-deletes a single file and verifies it no longer appears in queries.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn delete_node_removes_file() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let file = repo
        .create_file(NewFile {
            parent_id: None,
            name: "orphan.flac".to_string(),
            size: None,
            mime_type: None,
        })
        .await
        .expect("create failed");

    repo.delete_node(file.id).await.expect("delete_node failed");

    let result = repo.get_node_by_id(file.id).await.expect("get failed");

    assert_eq!(result, None);
}

/// Deleting a non-empty directory recursively soft-deletes the entire subtree.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn delete_node_recursively_deletes_subtree() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let root_dir = repo
        .create_dir(None, "Metal".to_string())
        .await
        .expect("create root_dir failed");
    let sub_dir = repo
        .create_dir(Some(root_dir.id), "Black Sabbath".to_string())
        .await
        .expect("create sub_dir failed");
    let track = repo
        .create_file(NewFile {
            parent_id: Some(sub_dir.id),
            name: "Paranoid.flac".to_string(),
            size: None,
            mime_type: None,
        })
        .await
        .expect("create track failed");

    repo.delete_node(root_dir.id)
        .await
        .expect("delete_node failed");

    assert_eq!(repo.get_node_by_id(root_dir.id).await.unwrap(), None);
    assert_eq!(repo.get_node_by_id(sub_dir.id).await.unwrap(), None);
    assert_eq!(repo.get_node_by_id(track.id).await.unwrap(), None);
}

/// Returns `NotFound` when deleting an ID that does not exist.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn delete_node_returns_not_found_for_unknown_id() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let err = repo
        .delete_node(sonora_domain::NodeId::generate())
        .await
        .expect_err("expected NotFound");

    assert!(matches!(err, NodeRepositoryError::NotFound(_)));
}

/// After soft-deleting a node, the same name can be reused under the same parent.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn deleted_name_can_be_reused() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let original = repo
        .create_dir(None, "Funk".to_string())
        .await
        .expect("create failed");

    repo.delete_node(original.id).await.expect("delete failed");

    let reused = repo
        .create_dir(None, "Funk".to_string())
        .await
        .expect("reuse after delete should succeed");

    assert_ne!(original.id, reused.id);
    assert_eq!(reused.name, "Funk");
}

// ── exists ─────────────────────────────────────────────────────────────────

/// Returns `true` for a path that exists and `false` for one that does not.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn exists_returns_true_for_known_path_and_false_for_missing() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    repo.create_dir(None, "Country".to_string())
        .await
        .expect("create failed");

    assert!(repo.exists("Country").await.expect("exists failed"));
    assert!(!repo.exists("Classical").await.expect("exists failed"));
}

// ── get_path ───────────────────────────────────────────────────────────────

/// Resolves the full virtual path of a deeply nested node.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_path_resolves_nested_node() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let rock = repo
        .create_dir(None, "Rock".to_string())
        .await
        .expect("create rock failed");
    let album = repo
        .create_dir(Some(rock.id), "Led Zeppelin IV".to_string())
        .await
        .expect("create album failed");
    let track = repo
        .create_file(NewFile {
            parent_id: Some(album.id),
            name: "Stairway to Heaven.flac".to_string(),
            size: None,
            mime_type: None,
        })
        .await
        .expect("create track failed");

    let path = repo.get_path(track.id).await.expect("get_path failed");

    assert_eq!(path, "Rock/Led Zeppelin IV/Stairway to Heaven.flac");
}

/// Returns the correct path for a root-level node.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_path_of_root_level_node() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let node = repo
        .create_dir(None, "Folk".to_string())
        .await
        .expect("create failed");

    let path = repo.get_path(node.id).await.expect("get_path failed");

    assert_eq!(path, "Folk");
}

/// Returns `NotFound` for a node ID that does not exist.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn get_path_returns_not_found_for_unknown_id() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let err = repo
        .get_path(sonora_domain::NodeId::generate())
        .await
        .expect_err("expected NotFound");

    assert!(matches!(err, NodeRepositoryError::NotFound(_)));
}
