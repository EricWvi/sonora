use axum::Router;
use axum::routing::{get, post, put};

use crate::app_state::AppState;
use crate::handlers::node::{
    create_dir, create_file, delete_node, get_node_by_id, get_node_by_path, get_path,
    list_children, list_root_children, move_node, node_exists,
};

/// Builds the application router with all VFS node routes registered.
pub fn build_router(state: AppState) -> Router {
    Router::new()
        .route("/api/nodes/dirs", post(create_dir))
        .route("/api/nodes/files", post(create_file))
        .route("/api/nodes/by-path", get(get_node_by_path))
        .route("/api/nodes/exists", get(node_exists))
        .route("/api/nodes/root/children", get(list_root_children))
        .route("/api/nodes/{id}", get(get_node_by_id).delete(delete_node))
        .route("/api/nodes/{id}/children", get(list_children))
        .route("/api/nodes/{id}/path", get(get_path))
        .route("/api/nodes/{id}/move", put(move_node))
        .with_state(state)
}
