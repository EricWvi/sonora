use axum::body::Body;
use axum::http::{Request, Response, StatusCode};
use pretty_assertions::assert_eq;
use sonora_contracts::{
    CreateDirResponse, CreateFileResponse, GetNodeResponse, GetPathResponse, ListChildrenResponse,
    MoveNodeResponse, NodeView,
};

use super::{bootstrap_test_state, send};
use crate::app_state::AppState;

async fn parse_json<T: serde::de::DeserializeOwned>(resp: Response<Body>) -> T {
    let bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .expect("failed to collect response body");
    serde_json::from_slice(&bytes).expect("failed to deserialize response JSON")
}

/// Creates a directory node via HTTP and returns the saved view.
async fn create_dir(state: &AppState, parent_id: Option<&str>, name: &str) -> NodeView {
    let body = serde_json::json!({"parentId": parent_id, "name": name});
    let req = Request::builder()
        .method("POST")
        .uri("/api/nodes/dirs")
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::CREATED);
    let r: CreateDirResponse = parse_json(resp).await;
    r.node
}

/// Creates a file node via HTTP and returns the saved view.
async fn create_file(
    state: &AppState,
    parent_id: Option<&str>,
    name: &str,
    size: Option<i64>,
    mime_type: Option<&str>,
) -> NodeView {
    let body = serde_json::json!({"parentId": parent_id, "name": name, "size": size, "mimeType": mime_type});
    let req = Request::builder()
        .method("POST")
        .uri("/api/nodes/files")
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::CREATED);
    let r: CreateFileResponse = parse_json(resp).await;
    r.node
}

// ── CD — create_dir ────────────────────────────────────────────────────────

/// CD-01: root-level directory → 201, all fields stored correctly.
async fn cd_01_root_stores_all_fields(state: &AppState) {
    let node = create_dir(state, None, "Rock_CD01").await;

    assert_eq!(node.name, "Rock_CD01");
    assert_eq!(node.kind, "directory");
    assert_eq!(node.parent_id, None);
    assert_eq!(node.size, None);
    assert_eq!(node.mime_type, None);
}

/// CD-02: nested directory → 201, parentId set correctly.
async fn cd_02_nested_sets_parent_id(state: &AppState) {
    let parent = create_dir(state, None, "Rock_CD02").await;
    let child = create_dir(state, Some(&parent.id), "Classic Rock").await;

    assert_eq!(child.parent_id, Some(parent.id));
    assert_eq!(child.name, "Classic Rock");
}

/// CD-03: duplicate name under same parent → 409.
async fn cd_03_duplicate_name_returns_409(state: &AppState) {
    create_dir(state, None, "Rock_CD03").await;

    let body = serde_json::json!({"parentId": null, "name": "Rock_CD03"});
    let req = Request::builder()
        .method("POST")
        .uri("/api/nodes/dirs")
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::CONFLICT);
}

// ── CF — create_file ───────────────────────────────────────────────────────

/// CF-01: file with size and MIME type → 201, all fields stored correctly.
async fn cf_01_with_metadata_stores_all_fields(state: &AppState) {
    let dir = create_dir(state, None, "Jazz_CF01").await;
    let file = create_file(
        state,
        Some(&dir.id),
        "track01.flac",
        Some(10_485_760),
        Some("audio/flac"),
    )
    .await;

    assert_eq!(file.name, "track01.flac");
    assert_eq!(file.kind, "file");
    assert_eq!(file.parent_id, Some(dir.id));
    assert_eq!(file.size, Some(10_485_760));
    assert_eq!(file.mime_type, Some("audio/flac".to_string()));
}

// ── GI — get_node_by_id ───────────────────────────────────────────────────

/// GI-01: existing id → 200, full node returned.
async fn gi_01_existing_id_returns_node(state: &AppState) {
    let created = create_dir(state, None, "Soul_GI01").await;

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}", created.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: GetNodeResponse = parse_json(resp).await;
    assert_eq!(r.node, created);
}

/// GI-02: unknown id → 404.
async fn gi_02_unknown_id_returns_404(state: &AppState) {
    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/00000000-0000-0000-0000-000000000000")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

// ── GP — get_node_by_path ─────────────────────────────────────────────────

/// GP-01: three-level path → 200, correct file node returned.
async fn gp_01_deep_path_returns_node(state: &AppState) {
    let rock = create_dir(state, None, "Rock_GP01").await;
    let album = create_dir(state, Some(&rock.id), "Led Zeppelin IV").await;
    let track = create_file(
        state,
        Some(&album.id),
        "Stairway to Heaven.flac",
        Some(50_000_000),
        Some("audio/flac"),
    )
    .await;

    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/by-path?path=Rock_GP01/Led%20Zeppelin%20IV/Stairway%20to%20Heaven.flac")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: GetNodeResponse = parse_json(resp).await;
    assert_eq!(r.node, track);
}

/// GP-02: path with a missing component → 404.
async fn gp_02_missing_component_returns_404(state: &AppState) {
    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/by-path?path=NonExistent_GP02")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

// ── LC — list_children ────────────────────────────────────────────────────

/// LC-01: directory contains a subdirectory and a file → directories returned before files.
async fn lc_01_dirs_returned_before_files(state: &AppState) {
    let parent = create_dir(state, None, "Blues_LC01").await;
    let inner_dir = create_dir(state, Some(&parent.id), "Subdir_LC01").await;
    let _file = create_file(state, Some(&parent.id), "compilation.flac", None, None).await;

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}/children", parent.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: ListChildrenResponse = parse_json(resp).await;
    assert_eq!(r.nodes.len(), 2);
    assert_eq!(r.nodes[0], inner_dir);
    assert_eq!(r.nodes[1].name, "compilation.flac");
}

/// LC-02: directory with no children → empty list.
async fn lc_02_empty_dir_returns_empty_list(state: &AppState) {
    let dir = create_dir(state, None, "Empty_LC02").await;

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}/children", dir.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: ListChildrenResponse = parse_json(resp).await;
    assert_eq!(r.nodes, vec![]);
}

// ── MN — move_node ────────────────────────────────────────────────────────

/// MN-01: same parent, new name → node renamed in place.
async fn mn_01_same_parent_renames_node(state: &AppState) {
    let node = create_dir(state, None, "OldName_MN01").await;

    let body = serde_json::json!({"newParentId": null, "newName": "NewName_MN01"});
    let req = Request::builder()
        .method("PUT")
        .uri(format!("/api/nodes/{}/move", node.id))
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: MoveNodeResponse = parse_json(resp).await;
    assert_eq!(r.node.id, node.id);
    assert_eq!(r.node.name, "NewName_MN01");
    assert_eq!(r.node.parent_id, None);
}

/// MN-02: different parent → parentId updated.
async fn mn_02_new_parent_updates_parent_id(state: &AppState) {
    let src = create_dir(state, None, "Source_MN02").await;
    let dst = create_dir(state, None, "Destination_MN02").await;
    let child = create_dir(state, Some(&src.id), "Child_MN02").await;

    let body = serde_json::json!({"newParentId": dst.id, "newName": "Child_MN02"});
    let req = Request::builder()
        .method("PUT")
        .uri(format!("/api/nodes/{}/move", child.id))
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: MoveNodeResponse = parse_json(resp).await;
    assert_eq!(r.node.parent_id, Some(dst.id));
    assert_eq!(r.node.name, "Child_MN02");
}

/// MN-03: unknown node id → 404.
async fn mn_03_unknown_id_returns_404(state: &AppState) {
    let body = serde_json::json!({"newParentId": null, "newName": "whatever"});
    let req = Request::builder()
        .method("PUT")
        .uri("/api/nodes/00000000-0000-0000-0000-000000000000/move")
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

/// MN-04: target name already taken at destination → 409.
async fn mn_04_name_taken_returns_409(state: &AppState) {
    let _existing = create_dir(state, None, "Rock_MN04").await;
    let other = create_dir(state, None, "Jazz_MN04").await;

    let body = serde_json::json!({"newParentId": null, "newName": "Rock_MN04"});
    let req = Request::builder()
        .method("PUT")
        .uri(format!("/api/nodes/{}/move", other.id))
        .header("content-type", "application/json")
        .body(Body::from(body.to_string()))
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::CONFLICT);
}

// ── DN — delete_node ──────────────────────────────────────────────────────

/// DN-01: deleted file → 204, no longer accessible via get_node_by_id.
async fn dn_01_file_no_longer_visible_after_delete(state: &AppState) {
    let file = create_file(state, None, "orphan_DN01.flac", None, None).await;

    let req = Request::builder()
        .method("DELETE")
        .uri(format!("/api/nodes/{}", file.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NO_CONTENT);

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}", file.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

/// DN-02: delete directory → subtree soft-deleted, all descendants return 404.
async fn dn_02_directory_recursively_deletes_subtree(state: &AppState) {
    let root_dir = create_dir(state, None, "Metal_DN02").await;
    let sub_dir = create_dir(state, Some(&root_dir.id), "Black Sabbath").await;
    let track = create_file(state, Some(&sub_dir.id), "Paranoid.flac", None, None).await;

    let req = Request::builder()
        .method("DELETE")
        .uri(format!("/api/nodes/{}", root_dir.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NO_CONTENT);

    for id in [&root_dir.id, &sub_dir.id, &track.id] {
        let req = Request::builder()
            .method("GET")
            .uri(format!("/api/nodes/{id}"))
            .body(Body::empty())
            .expect("request build failed");
        let resp = send(state.clone(), req).await;
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }
}

/// DN-03: unknown id → 404.
async fn dn_03_unknown_id_returns_404(state: &AppState) {
    let req = Request::builder()
        .method("DELETE")
        .uri("/api/nodes/00000000-0000-0000-0000-000000000000")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

/// DN-04: name reused after soft-delete → 201, new node with a different id.
async fn dn_04_deleted_name_can_be_reused(state: &AppState) {
    let original = create_dir(state, None, "Funk_DN04").await;

    let req = Request::builder()
        .method("DELETE")
        .uri(format!("/api/nodes/{}", original.id))
        .body(Body::empty())
        .expect("request build failed");
    send(state.clone(), req).await;

    let reused = create_dir(state, None, "Funk_DN04").await;
    assert_ne!(original.id, reused.id);
    assert_eq!(reused.name, "Funk_DN04");
}

// ── EX — exists ───────────────────────────────────────────────────────────

/// EX-01: known path → true; unknown path → false.
async fn ex_01_known_path_true_unknown_path_false(state: &AppState) {
    create_dir(state, None, "Country_EX01").await;

    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/exists?path=Country_EX01")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let exists: bool = parse_json(resp).await;
    assert!(exists);

    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/exists?path=Classical_EX01")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let exists: bool = parse_json(resp).await;
    assert!(!exists);
}

// ── PT — get_path ─────────────────────────────────────────────────────────

/// PT-01: deeply nested node → 200, full slash-separated path returned.
async fn pt_01_nested_node_returns_full_path(state: &AppState) {
    let rock = create_dir(state, None, "Rock_PT01").await;
    let album = create_dir(state, Some(&rock.id), "Led Zeppelin IV").await;
    let track = create_file(
        state,
        Some(&album.id),
        "Stairway to Heaven.flac",
        None,
        None,
    )
    .await;

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}/path", track.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: GetPathResponse = parse_json(resp).await;
    assert_eq!(r.path, "Rock_PT01/Led Zeppelin IV/Stairway to Heaven.flac");
}

/// PT-02: root-level node → 200, single name with no separator.
async fn pt_02_root_node_returns_single_name(state: &AppState) {
    let node = create_dir(state, None, "Folk_PT02").await;

    let req = Request::builder()
        .method("GET")
        .uri(format!("/api/nodes/{}/path", node.id))
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    let r: GetPathResponse = parse_json(resp).await;
    assert_eq!(r.path, "Folk_PT02");
}

/// PT-03: unknown id → 404.
async fn pt_03_unknown_id_returns_404(state: &AppState) {
    let req = Request::builder()
        .method("GET")
        .uri("/api/nodes/00000000-0000-0000-0000-000000000000/path")
        .body(Body::empty())
        .expect("request build failed");
    let resp = send(state.clone(), req).await;
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}

// ── test runner ────────────────────────────────────────────────────────────

#[tokio::test]
#[ignore = "requires RUN_TESTCONTAINERS=1"]
async fn node_handler_tests() {
    let (_container, state) = bootstrap_test_state().await;

    cd_01_root_stores_all_fields(&state).await;
    cd_02_nested_sets_parent_id(&state).await;
    cd_03_duplicate_name_returns_409(&state).await;
    cf_01_with_metadata_stores_all_fields(&state).await;
    gi_01_existing_id_returns_node(&state).await;
    gi_02_unknown_id_returns_404(&state).await;
    gp_01_deep_path_returns_node(&state).await;
    gp_02_missing_component_returns_404(&state).await;
    lc_01_dirs_returned_before_files(&state).await;
    lc_02_empty_dir_returns_empty_list(&state).await;
    mn_01_same_parent_renames_node(&state).await;
    mn_02_new_parent_updates_parent_id(&state).await;
    mn_03_unknown_id_returns_404(&state).await;
    mn_04_name_taken_returns_409(&state).await;
    dn_01_file_no_longer_visible_after_delete(&state).await;
    dn_02_directory_recursively_deletes_subtree(&state).await;
    dn_03_unknown_id_returns_404(&state).await;
    dn_04_deleted_name_can_be_reused(&state).await;
    ex_01_known_path_true_unknown_path_false(&state).await;
    pt_01_nested_node_returns_full_path(&state).await;
    pt_02_root_node_returns_single_name(&state).await;
    pt_03_unknown_id_returns_404(&state).await;
}
