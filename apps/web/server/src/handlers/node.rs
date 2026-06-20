use axum::{
    Json,
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use serde::Deserialize;
use sonora_application::NodeError;
use sonora_contracts::{CreateDirRequest, CreateFileRequest, MoveNodeRequest};

use crate::app_state::AppState;

/// Maps a [`NodeError`] to an HTTP response with an appropriate status code.
fn node_error_response(error: NodeError) -> Response {
    match error {
        NodeError::NotFound { id } => {
            (StatusCode::NOT_FOUND, format!("node not found: {id}")).into_response()
        }
        NodeError::NameConflict { name } => {
            (StatusCode::CONFLICT, format!("name already exists: {name}")).into_response()
        }
        NodeError::InvalidPath { path } => {
            (StatusCode::BAD_REQUEST, format!("invalid path: {path}")).into_response()
        }
        NodeError::InvalidId { id } => {
            (StatusCode::BAD_REQUEST, format!("invalid node id: {id}")).into_response()
        }
        NodeError::Repository { message } => {
            (StatusCode::INTERNAL_SERVER_ERROR, message).into_response()
        }
    }
}

#[derive(Deserialize)]
pub struct PathQuery {
    path: String,
}

/// `POST /api/nodes/dirs` — creates a new directory.
pub async fn create_dir(
    State(state): State<AppState>,
    Json(body): Json<CreateDirRequest>,
) -> Response {
    match state.node_api.create_dir(body).await {
        Ok(resp) => (StatusCode::CREATED, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `POST /api/nodes/files` — creates a new file node.
pub async fn create_file(
    State(state): State<AppState>,
    Json(body): Json<CreateFileRequest>,
) -> Response {
    match state.node_api.create_file(body).await {
        Ok(resp) => (StatusCode::CREATED, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/by-path?path=...` — resolves a virtual path to a node.
pub async fn get_node_by_path(
    State(state): State<AppState>,
    Query(q): Query<PathQuery>,
) -> Response {
    match state.node_api.get_by_path(&q.path).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/exists?path=...` — checks whether a node exists at the given path.
pub async fn node_exists(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    match state.node_api.exists(&q.path).await {
        Ok(exists) => (StatusCode::OK, Json(exists)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/root/children` — lists root-level nodes.
pub async fn list_root_children(State(state): State<AppState>) -> Response {
    match state.node_api.list_children(None).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/{id}` — fetches a single node by id.
pub async fn get_node_by_id(State(state): State<AppState>, Path(id): Path<String>) -> Response {
    match state.node_api.get_by_id(&id).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/{id}/children` — lists the direct children of a directory.
pub async fn list_children(State(state): State<AppState>, Path(id): Path<String>) -> Response {
    match state.node_api.list_children(Some(&id)).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `GET /api/nodes/{id}/path` — resolves the full virtual path of a node.
pub async fn get_path(State(state): State<AppState>, Path(id): Path<String>) -> Response {
    match state.node_api.get_path(&id).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `PUT /api/nodes/{id}/move` — moves or renames a node.
pub async fn move_node(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(body): Json<MoveNodeRequest>,
) -> Response {
    match state.node_api.move_node(&id, body).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(e) => node_error_response(e),
    }
}

/// `DELETE /api/nodes/{id}` — soft-deletes a node and its entire subtree.
pub async fn delete_node(State(state): State<AppState>, Path(id): Path<String>) -> Response {
    match state.node_api.delete_node(&id).await {
        Ok(()) => StatusCode::NO_CONTENT.into_response(),
        Err(e) => node_error_response(e),
    }
}
