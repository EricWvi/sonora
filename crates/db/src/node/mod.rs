use std::future::Future;

use sonora_domain::{Node, NodeId};
use thiserror::Error;

mod repository;

#[cfg(test)]
mod repository_tests;

pub use repository::PostgresNodeRepository;

/// Metadata supplied when inserting a new file node.
pub struct NewFile {
    pub parent_id: Option<NodeId>,
    pub name: String,
    pub size: Option<i64>,
    pub mime_type: Option<String>,
}

/// Errors that a [`NodeRepository`] implementation may return.
#[derive(Debug, Error)]
pub enum NodeRepositoryError {
    #[error("node not found: {0}")]
    NotFound(NodeId),
    #[error("a node named `{name}` already exists under this parent")]
    NameConflict { name: String },
    #[error("invalid virtual path: {0}")]
    InvalidPath(String),
    #[error("repository operation failed: {0}")]
    OperationFailed(String),
}

/// Persistence contract for virtual file system node operations.
///
/// Implementations are expected to operate against the `nodes` table and must exclude
/// soft-deleted rows (`is_deleted = true`) from all read operations unless otherwise noted.
/// Root-level nodes have `parent_id = None`; there is no explicit root node in the database.
pub trait NodeRepository: Send + Sync {
    /// Creates a directory node under the given parent (or at the VFS root when `parent_id` is `None`).
    fn create_dir(
        &self,
        parent_id: Option<NodeId>,
        name: String,
    ) -> impl Future<Output = Result<Node, NodeRepositoryError>> + Send;

    /// Creates a file node with optional size and MIME type metadata.
    fn create_file(
        &self,
        file: NewFile,
    ) -> impl Future<Output = Result<Node, NodeRepositoryError>> + Send;

    /// Fetches a single node by its primary key.
    fn get_node_by_id(
        &self,
        id: NodeId,
    ) -> impl Future<Output = Result<Option<Node>, NodeRepositoryError>> + Send;

    /// Resolves a virtual path (e.g. `Rock/Album`) to a node.
    ///
    /// Returns `None` when any component of the path does not exist or is soft-deleted.
    /// Returns `Err(InvalidPath)` when the path is empty.
    fn get_node_by_path(
        &self,
        path: &str,
    ) -> impl Future<Output = Result<Option<Node>, NodeRepositoryError>> + Send;

    /// Lists the direct children of a directory. Pass `None` to list root-level nodes.
    fn list_children(
        &self,
        parent_id: Option<NodeId>,
    ) -> impl Future<Output = Result<Vec<Node>, NodeRepositoryError>> + Send;

    /// Re-parents and/or renames a node.
    ///
    /// Passing the existing `parent_id` with a new name is a rename.
    /// Passing a new `parent_id` with the same name is a move.
    fn move_node(
        &self,
        node_id: NodeId,
        new_parent_id: Option<NodeId>,
        new_name: String,
    ) -> impl Future<Output = Result<Node, NodeRepositoryError>> + Send;

    /// Soft-deletes a node. For directories, recursively marks the entire subtree as deleted.
    fn delete_node(
        &self,
        node_id: NodeId,
    ) -> impl Future<Output = Result<(), NodeRepositoryError>> + Send;

    /// Returns `true` when a non-deleted node exists at the given virtual path.
    fn exists(&self, path: &str) -> impl Future<Output = Result<bool, NodeRepositoryError>> + Send;

    /// Resolves the full virtual path of a node by walking its ancestor chain (e.g. `Rock/Album`).
    fn get_path(
        &self,
        node_id: NodeId,
    ) -> impl Future<Output = Result<String, NodeRepositoryError>> + Send;
}
