use sonora_application::{
    CreateDirHandler, CreateFileHandler, DeleteNodeHandler, ExistsHandler, GetNodeByIdHandler,
    GetNodeByPathHandler, GetPathHandler, ListChildrenHandler, MoveNodeHandler, NodeError,
};
use sonora_contracts::{
    CreateDirRequest, CreateDirResponse, CreateFileRequest, CreateFileResponse, GetNodeResponse,
    GetPathResponse, ListChildrenResponse, MoveNodeRequest, MoveNodeResponse,
};
use sonora_db::{PostgresNodeRepository, SystemTimestampSource};
use sqlx::{Pool, Postgres};

type Repo = PostgresNodeRepository<SystemTimestampSource>;

/// Groups all VFS node application handlers behind a single pool-bound façade.
pub struct NodeApi {
    create_dir: CreateDirHandler<Repo>,
    create_file: CreateFileHandler<Repo>,
    get_by_id: GetNodeByIdHandler<Repo>,
    get_by_path: GetNodeByPathHandler<Repo>,
    list_children: ListChildrenHandler<Repo>,
    move_node: MoveNodeHandler<Repo>,
    delete_node: DeleteNodeHandler<Repo>,
    exists: ExistsHandler<Repo>,
    get_path: GetPathHandler<Repo>,
}

impl NodeApi {
    /// Builds the node API from the shared connection pool.
    pub fn new(pool: Pool<Postgres>) -> Self {
        Self {
            create_dir: CreateDirHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            create_file: CreateFileHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            get_by_id: GetNodeByIdHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            get_by_path: GetNodeByPathHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            list_children: ListChildrenHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            move_node: MoveNodeHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            delete_node: DeleteNodeHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            exists: ExistsHandler::new(Repo::new(pool.clone(), SystemTimestampSource)),
            get_path: GetPathHandler::new(Repo::new(pool, SystemTimestampSource)),
        }
    }

    /// Creates a new directory node.
    pub async fn create_dir(
        &self,
        request: CreateDirRequest,
    ) -> Result<CreateDirResponse, NodeError> {
        self.create_dir.handle(request).await
    }

    /// Creates a new file node.
    pub async fn create_file(
        &self,
        request: CreateFileRequest,
    ) -> Result<CreateFileResponse, NodeError> {
        self.create_file.handle(request).await
    }

    /// Fetches a single node by its id.
    pub async fn get_by_id(&self, id: &str) -> Result<GetNodeResponse, NodeError> {
        self.get_by_id.handle(id).await
    }

    /// Resolves a virtual path to a node.
    pub async fn get_by_path(&self, path: &str) -> Result<GetNodeResponse, NodeError> {
        self.get_by_path.handle(path).await
    }

    /// Lists the direct children of a directory, or root-level nodes when `parent_id` is `None`.
    pub async fn list_children(
        &self,
        parent_id: Option<&str>,
    ) -> Result<ListChildrenResponse, NodeError> {
        self.list_children.handle(parent_id).await
    }

    /// Moves or renames a node.
    pub async fn move_node(
        &self,
        node_id: &str,
        request: MoveNodeRequest,
    ) -> Result<MoveNodeResponse, NodeError> {
        self.move_node.handle(node_id, request).await
    }

    /// Soft-deletes a node and its entire subtree.
    pub async fn delete_node(&self, node_id: &str) -> Result<(), NodeError> {
        self.delete_node.handle(node_id).await
    }

    /// Checks whether a node exists at the given virtual path.
    pub async fn exists(&self, path: &str) -> Result<bool, NodeError> {
        self.exists.handle(path).await
    }

    /// Resolves the full virtual path of a node.
    pub async fn get_path(&self, node_id: &str) -> Result<GetPathResponse, NodeError> {
        self.get_path.handle(node_id).await
    }
}
