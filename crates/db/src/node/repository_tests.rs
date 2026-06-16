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

// ── CD — create_dir ────────────────────────────────────────────────────────

/// CD-01: root-level directory → all fields stored correctly.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn cd_01_root_stores_all_fields() {
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

/// CD-02: nested directory → parent link set correctly.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn cd_02_nested_sets_parent_link() {
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

/// CD-03: duplicate name under same parent → NameConflict.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn cd_03_duplicate_name_returns_conflict() {
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

// ── CF — create_file ───────────────────────────────────────────────────────

/// CF-01: file with size and MIME type → all fields stored correctly.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn cf_01_with_metadata_stores_all_fields() {
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

// ── GI — get_node_by_id ───────────────────────────────────────────────────

/// GI-01: existing id → full node returned.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn gi_01_existing_id_returns_node() {
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

/// GI-02: unknown id → None.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn gi_02_unknown_id_returns_none() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let result = repo
        .get_node_by_id(sonora_domain::NodeId::generate())
        .await
        .expect("get failed");

    assert_eq!(result, None);
}

// ── GP — get_node_by_path ─────────────────────────────────────────────────

/// GP-01: three-level path → correct file node returned.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn gp_01_deep_path_returns_node() {
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

/// GP-02: path with a missing component → None.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn gp_02_missing_component_returns_none() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let result = repo
        .get_node_by_path("NonExistent")
        .await
        .expect("path lookup failed");

    assert_eq!(result, None);
}

// ── LC — list_children ────────────────────────────────────────────────────

/// LC-01: root contains a directory and a file → directories returned before files.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn lc_01_dirs_returned_before_files() {
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

/// LC-02: directory with no children → empty vec.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn lc_02_empty_dir_returns_empty_vec() {
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

// ── MN — move_node ────────────────────────────────────────────────────────

/// MN-01: same parent, new name → node renamed in place.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn mn_01_same_parent_renames_node() {
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

/// MN-02: different parent → parent_id updated.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn mn_02_new_parent_updates_parent_id() {
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

/// MN-03: unknown node id → NotFound.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn mn_03_unknown_id_returns_not_found() {
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

/// MN-04: target name already taken at destination → NameConflict.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn mn_04_name_taken_returns_conflict() {
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

// ── DN — delete_node ──────────────────────────────────────────────────────

/// DN-01: file deleted → no longer visible via get_node_by_id.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn dn_01_file_no_longer_visible_after_delete() {
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

/// DN-02: directory with nested children deleted → entire subtree soft-deleted.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn dn_02_directory_recursively_deletes_subtree() {
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

/// DN-03: unknown id → NotFound.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn dn_03_unknown_id_returns_not_found() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let err = repo
        .delete_node(sonora_domain::NodeId::generate())
        .await
        .expect_err("expected NotFound");

    assert!(matches!(err, NodeRepositoryError::NotFound(_)));
}

/// DN-04: name reused after soft-delete → new node created with a different id.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn dn_04_deleted_name_can_be_reused() {
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

// ── EX — exists ───────────────────────────────────────────────────────────

/// EX-01: known path → true; unknown path → false.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn ex_01_known_path_true_unknown_path_false() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    repo.create_dir(None, "Country".to_string())
        .await
        .expect("create failed");

    assert!(repo.exists("Country").await.expect("exists failed"));
    assert!(!repo.exists("Classical").await.expect("exists failed"));
}

// ── PT — get_path ─────────────────────────────────────────────────────────

/// PT-01: deeply nested node → full slash-separated path returned.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn pt_01_nested_node_returns_full_path() {
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

/// PT-02: root-level node → single name returned with no separator.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn pt_02_root_node_returns_single_name() {
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

/// PT-03: unknown id → NotFound.
#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn pt_03_unknown_id_returns_not_found() {
    let _guard = set_trace_logging();
    let (_container, location) = start_postgres().await;
    let repo = make_repo(&location).await;

    let err = repo
        .get_path(sonora_domain::NodeId::generate())
        .await
        .expect_err("expected NotFound");

    assert!(matches!(err, NodeRepositoryError::NotFound(_)));
}
