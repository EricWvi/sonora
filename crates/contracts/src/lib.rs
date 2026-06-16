mod frontend;
mod node;

pub use frontend::{FrontendEndpoint, FrontendHttpMethod, FrontendPathParam, frontend_endpoints};
pub use node::{
    CreateDirRequest, CreateDirResponse, CreateFileRequest, CreateFileResponse, GetNodeResponse,
    GetPathResponse, ListChildrenResponse, MoveNodeRequest, MoveNodeResponse, NodeView,
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
    GetPathResponse::export(&config)?;

    Ok(())
}
