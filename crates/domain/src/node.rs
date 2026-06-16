use crate::NodeId;

/// Distinguishes whether a VFS entry represents a container or a leaf file.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum NodeKind {
    Directory,
    File,
}

/// A single entry in the virtual file system, representing either a directory or a file.
///
/// `parent_id = None` means the node sits at the VFS root. There is no root node itself.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Node {
    pub id: NodeId,
    pub parent_id: Option<NodeId>,
    pub name: String,
    pub kind: NodeKind,
    /// `None` for directories; byte count for files.
    pub size: Option<i64>,
    pub mime_type: Option<String>,
    /// Unix timestamp in milliseconds.
    pub created_at: i64,
    /// Unix timestamp in milliseconds.
    pub updated_at: i64,
    /// Monotonically increasing version assigned on every write, used for sync.
    pub server_version: i64,
    pub is_deleted: bool,
}

impl Node {
    /// Constructs a complete node snapshot from persistence-layer values.
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        id: NodeId,
        parent_id: Option<NodeId>,
        name: impl Into<String>,
        kind: NodeKind,
        size: Option<i64>,
        mime_type: Option<String>,
        created_at: i64,
        updated_at: i64,
        server_version: i64,
        is_deleted: bool,
    ) -> Self {
        Self {
            id,
            parent_id,
            name: name.into(),
            kind,
            size,
            mime_type,
            created_at,
            updated_at,
            server_version,
            is_deleted,
        }
    }
}
