use serde::{Deserialize, Serialize};
use ts_rs::TS;

/// Public representation of a VFS node shared across all node responses.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct NodeView {
    pub id: String,
    pub parent_id: Option<String>,
    pub name: String,
    /// `"directory"` or `"file"`.
    pub kind: String,
    pub size: Option<i64>,
    pub mime_type: Option<String>,
    /// MD5 hex digest of the latest file content; `null` until the first upload completes.
    pub md5: Option<String>,
    /// `"pending_upload"` or `"available"`.
    pub storage_status: String,
    pub created_at: i64,
    pub updated_at: i64,
}

/// Fields required to create a new directory node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct CreateDirRequest {
    pub parent_id: Option<String>,
    pub name: String,
}

/// Returns the newly created directory node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct CreateDirResponse {
    pub node: NodeView,
}

/// Fields required to create a new file node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct CreateFileRequest {
    pub parent_id: Option<String>,
    pub name: String,
    pub size: Option<i64>,
    pub mime_type: Option<String>,
}

/// Returns the newly created file node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct CreateFileResponse {
    pub node: NodeView,
}

/// Returns a single node looked up by id or path.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct GetNodeResponse {
    pub node: NodeView,
}

/// Returns the direct children of a directory.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct ListChildrenResponse {
    pub nodes: Vec<NodeView>,
}

/// Fields required to move or rename a node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct MoveNodeRequest {
    pub new_parent_id: Option<String>,
    pub new_name: String,
}

/// Returns the node after a successful move or rename.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct MoveNodeResponse {
    pub node: NodeView,
}

/// Returns the virtual path of a node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct GetPathResponse {
    pub path: String,
}

/// Frontend SDK request for move/rename: combines the path-level node id with the body fields.
///
/// Defined separately from [`MoveNodeRequest`] so the SDK caller passes a single flat object
/// instead of composing the id and body separately.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct MoveNodeWithIdRequest {
    pub id: String,
    pub new_parent_id: Option<String>,
    pub new_name: String,
}

/// Frontend SDK request for fetching a single node by id.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct GetNodeByIdRequest {
    pub id: String,
}

/// Frontend SDK request for listing the direct children of a directory.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct ListChildrenRequest {
    pub id: String,
}

/// Frontend SDK request for resolving the full virtual path of a node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct GetNodePathRequest {
    pub id: String,
}

/// Frontend SDK request for soft-deleting a node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct DeleteNodeRequest {
    pub id: String,
}

/// Frontend SDK request for resolving a virtual path to a node.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct GetNodeByPathRequest {
    pub path: String,
}

/// Frontend SDK request for checking whether a node exists at a virtual path.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct NodeExistsRequest {
    pub path: String,
}

/// Frontend SDK request for listing root-level nodes (no parameters required).
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, TS)]
#[serde(rename_all = "camelCase")]
#[ts(export, export_to = "node.ts")]
pub struct ListRootChildrenRequest {}
