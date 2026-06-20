use crate::NodeId;

/// Distinguishes whether a VFS entry represents a container or a leaf file.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum NodeKind {
    Directory,
    File,
}

/// Whether the server holds the physical bytes for a file node.
///
/// - `PendingUpload` (0): no file stored yet; a client must upload before others can download.
/// - `Available` (1): bytes are on the server and any client may download.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum StorageStatus {
    PendingUpload,
    Available,
}

impl StorageStatus {
    /// Converts the integer discriminant stored in the database to a `StorageStatus`.
    pub fn from_db(value: i32) -> Result<Self, i32> {
        match value {
            0 => Ok(Self::PendingUpload),
            1 => Ok(Self::Available),
            other => Err(other),
        }
    }

    /// Returns the integer discriminant used in the database column.
    pub fn as_db(self) -> i32 {
        match self {
            Self::PendingUpload => 0,
            Self::Available => 1,
        }
    }
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
    /// MD5 hex digest of the latest file content; `None` until the first upload completes.
    pub md5: Option<String>,
    /// Whether the server holds the physical bytes for this node.
    pub storage_status: StorageStatus,
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
        md5: Option<String>,
        storage_status: StorageStatus,
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
            md5,
            storage_status,
            created_at,
            updated_at,
            server_version,
            is_deleted,
        }
    }
}
