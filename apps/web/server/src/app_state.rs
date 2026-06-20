use std::sync::Arc;

use crate::service::NodeApi;

/// Shared state injected into every HTTP handler via axum's `State` extractor.
#[derive(Clone)]
pub struct AppState {
    pub node_api: Arc<NodeApi>,
}
