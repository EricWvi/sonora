use sonora_contracts::{
    CreateDirRequest, CreateDirResponse, CreateFileRequest, CreateFileResponse, GetNodeResponse,
    GetPathResponse, ListChildrenResponse, MoveNodeRequest, MoveNodeResponse, NodeView,
};
use sonora_db::{NewFile, NodeRepository, NodeRepositoryError};
use sonora_domain::{Node, NodeId, NodeKind, StorageStatus};
use thiserror::Error;

/// Application-level errors for VFS node operations.
#[derive(Debug, Error)]
pub enum NodeError {
    #[error("node not found: {id}")]
    NotFound { id: String },
    #[error("a node named `{name}` already exists under this parent")]
    NameConflict { name: String },
    #[error("invalid virtual path: {path}")]
    InvalidPath { path: String },
    #[error("invalid node id: {id}")]
    InvalidId { id: String },
    #[error("node repository error: {message}")]
    Repository { message: String },
}

impl From<NodeRepositoryError> for NodeError {
    fn from(e: NodeRepositoryError) -> Self {
        match e {
            NodeRepositoryError::NotFound(id) => Self::NotFound { id: id.to_string() },
            NodeRepositoryError::NameConflict { name } => Self::NameConflict { name },
            NodeRepositoryError::InvalidPath(path) => Self::InvalidPath { path },
            NodeRepositoryError::OperationFailed(msg) => Self::Repository { message: msg },
        }
    }
}

fn parse_node_id(id: &str) -> Result<NodeId, NodeError> {
    NodeId::parse(id).map_err(|_| NodeError::InvalidId { id: id.to_string() })
}

fn map_node(n: Node) -> NodeView {
    NodeView {
        id: n.id.to_string(),
        parent_id: n.parent_id.map(|id| id.to_string()),
        name: n.name,
        kind: match n.kind {
            NodeKind::Directory => "directory".to_string(),
            NodeKind::File => "file".to_string(),
        },
        size: n.size,
        mime_type: n.mime_type,
        md5: n.md5,
        storage_status: match n.storage_status {
            StorageStatus::PendingUpload => "pending_upload".to_string(),
            StorageStatus::Available => "available".to_string(),
        },
        created_at: n.created_at,
        updated_at: n.updated_at,
    }
}

/// Creates a new directory node.
pub struct CreateDirHandler<R> {
    repository: R,
}

impl<R> CreateDirHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> CreateDirHandler<R> {
    pub async fn handle(&self, request: CreateDirRequest) -> Result<CreateDirResponse, NodeError> {
        let parent_id = request
            .parent_id
            .as_deref()
            .map(parse_node_id)
            .transpose()?;
        let node = self.repository.create_dir(parent_id, request.name).await?;
        Ok(CreateDirResponse {
            node: map_node(node),
        })
    }
}

/// Creates a new file node.
pub struct CreateFileHandler<R> {
    repository: R,
}

impl<R> CreateFileHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> CreateFileHandler<R> {
    pub async fn handle(
        &self,
        request: CreateFileRequest,
    ) -> Result<CreateFileResponse, NodeError> {
        let parent_id = request
            .parent_id
            .as_deref()
            .map(parse_node_id)
            .transpose()?;
        let node = self
            .repository
            .create_file(NewFile {
                parent_id,
                name: request.name,
                size: request.size,
                mime_type: request.mime_type,
            })
            .await?;
        Ok(CreateFileResponse {
            node: map_node(node),
        })
    }
}

/// Fetches a single node by its id.
pub struct GetNodeByIdHandler<R> {
    repository: R,
}

impl<R> GetNodeByIdHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> GetNodeByIdHandler<R> {
    /// Returns the node view, or `NotFound` when no live node exists for the given id.
    pub async fn handle(&self, id: &str) -> Result<GetNodeResponse, NodeError> {
        let node_id = parse_node_id(id)?;
        let node = self
            .repository
            .get_node_by_id(node_id)
            .await?
            .ok_or_else(|| NodeError::NotFound { id: id.to_string() })?;
        Ok(GetNodeResponse {
            node: map_node(node),
        })
    }
}

/// Resolves a virtual path to a node.
pub struct GetNodeByPathHandler<R> {
    repository: R,
}

impl<R> GetNodeByPathHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> GetNodeByPathHandler<R> {
    /// Returns the node view, or `NotFound` when no live node exists at the given path.
    pub async fn handle(&self, path: &str) -> Result<GetNodeResponse, NodeError> {
        let node = self
            .repository
            .get_node_by_path(path)
            .await?
            .ok_or_else(|| NodeError::NotFound {
                id: path.to_string(),
            })?;
        Ok(GetNodeResponse {
            node: map_node(node),
        })
    }
}

/// Lists the direct children of a directory.
pub struct ListChildrenHandler<R> {
    repository: R,
}

impl<R> ListChildrenHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> ListChildrenHandler<R> {
    /// Pass `None` to list root-level nodes.
    pub async fn handle(&self, parent_id: Option<&str>) -> Result<ListChildrenResponse, NodeError> {
        let parent_id = parent_id.map(parse_node_id).transpose()?;
        let nodes = self.repository.list_children(parent_id).await?;
        Ok(ListChildrenResponse {
            nodes: nodes.into_iter().map(map_node).collect(),
        })
    }
}

/// Moves or renames a node.
pub struct MoveNodeHandler<R> {
    repository: R,
}

impl<R> MoveNodeHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> MoveNodeHandler<R> {
    pub async fn handle(
        &self,
        node_id: &str,
        request: MoveNodeRequest,
    ) -> Result<MoveNodeResponse, NodeError> {
        let node_id = parse_node_id(node_id)?;
        let new_parent_id = request
            .new_parent_id
            .as_deref()
            .map(parse_node_id)
            .transpose()?;
        let node = self
            .repository
            .move_node(node_id, new_parent_id, request.new_name)
            .await?;
        Ok(MoveNodeResponse {
            node: map_node(node),
        })
    }
}

/// Soft-deletes a node and its entire subtree.
pub struct DeleteNodeHandler<R> {
    repository: R,
}

impl<R> DeleteNodeHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> DeleteNodeHandler<R> {
    pub async fn handle(&self, node_id: &str) -> Result<(), NodeError> {
        let node_id = parse_node_id(node_id)?;
        self.repository.delete_node(node_id).await?;
        Ok(())
    }
}

/// Checks whether a node exists at the given virtual path.
pub struct ExistsHandler<R> {
    repository: R,
}

impl<R> ExistsHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> ExistsHandler<R> {
    pub async fn handle(&self, path: &str) -> Result<bool, NodeError> {
        Ok(self.repository.exists(path).await?)
    }
}

/// Resolves the full virtual path of a node.
pub struct GetPathHandler<R> {
    repository: R,
}

impl<R> GetPathHandler<R> {
    pub fn new(repository: R) -> Self {
        Self { repository }
    }
}

impl<R: NodeRepository> GetPathHandler<R> {
    pub async fn handle(&self, node_id: &str) -> Result<GetPathResponse, NodeError> {
        let node_id = parse_node_id(node_id)?;
        let path = self.repository.get_path(node_id).await?;
        Ok(GetPathResponse { path })
    }
}
