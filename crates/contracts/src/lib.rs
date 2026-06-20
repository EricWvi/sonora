mod frontend;
mod node;

pub use frontend::{FrontendEndpoint, FrontendHttpMethod, FrontendPathParam, frontend_endpoints};
pub use node::{
    CreateDirRequest, CreateDirResponse, CreateFileRequest, CreateFileResponse, DeleteNodeRequest,
    GetNodeByIdRequest, GetNodeByPathRequest, GetNodePathRequest, GetNodeResponse, GetPathResponse,
    ListChildrenRequest, ListChildrenResponse, ListRootChildrenRequest, MoveNodeRequest,
    MoveNodeResponse, MoveNodeWithIdRequest, NodeExistsRequest, NodeView,
};

use std::path::Path;
use ts_rs::{Config, ExportError, TS};

/// Exports every contract DTO family into the shared TypeScript package for frontend consumers.
pub fn export_typescript_bindings_to(
    output_directory: impl AsRef<Path>,
) -> Result<(), ExportError> {
    let config = Config::new().with_out_dir(output_directory.as_ref());

    NodeView::export(&config)?;
    CreateDirRequest::export(&config)?;
    CreateDirResponse::export(&config)?;
    CreateFileRequest::export(&config)?;
    CreateFileResponse::export(&config)?;
    GetNodeResponse::export(&config)?;
    ListChildrenResponse::export(&config)?;
    MoveNodeRequest::export(&config)?;
    MoveNodeResponse::export(&config)?;
    MoveNodeWithIdRequest::export(&config)?;
    GetNodeByIdRequest::export(&config)?;
    ListChildrenRequest::export(&config)?;
    GetNodePathRequest::export(&config)?;
    DeleteNodeRequest::export(&config)?;
    GetNodeByPathRequest::export(&config)?;
    NodeExistsRequest::export(&config)?;
    ListRootChildrenRequest::export(&config)?;
    GetPathResponse::export(&config)?;

    Ok(())
}
