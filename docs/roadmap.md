# Roadmap

## Vision

Sonora is a multi-device music synchronization service that supports song file management, metadata management, and playlist management.

---

## Phase 1 — Database-Backed Virtual File System

### Goal

Implement a virtual file system (VFS) at the database layer that mirrors the on-disk layout of a user's music library. All file and directory entities are persisted in the database, enabling efficient traversal, querying, and manipulation without repeated disk I/O.

### Core Concepts

- **Library root (`sonora-library`)**: The single root node of the VFS. Everything lives under this node. It corresponds to a configured directory on disk.
- **Node**: A unified representation of either a file or a directory entry in the database.
- **Path**: A virtual path rooted at `sonora-library` (e.g., `sonora-library/Rock/Album/track01.flac`). Paths are stored as structured parent–child relationships, not raw strings.

### Data Model

```
Node
├── id              UUID, primary key
├── parent_id       UUID | NULL (NULL means direct child of library root)
├── name            VARCHAR(1024) NOT NULL
├── kind            ENUM { Directory, File }
├── size            INT8 | NULL (bytes, NULL for directories)
├── mime_type       TEXT | NULL
├── created_at      BIGINT NOT NULL (unix timestamp in millisec)
├── updated_at      BIGINT NOT NULL
├── server_version  BIGINT NOT NULL
└── is_deleted      BOOLEAN DEFAULT FALSE
```

### Supported Operations

| Operation        | Description                                              |
|------------------|----------------------------------------------------------|
| `create_dir`     | Create a directory node under a given parent             |
| `create_file`    | Insert a file node with metadata under a given parent    |
| `get_node`       | Fetch a single node by ID or by virtual path             |
| `list_children`  | List direct children of a directory node                 |
| `move_node`      | Re-parent a node (rename or move to another directory)   |
| `delete_node`    | Remove a node; recursively delete subtree for directories|
| `exists`         | Check whether a node at a given path exists              |
| `get_path`       | Resolve the full virtual path of a node by walking ancestors |

### Out of Scope (Phase 1)

- Actual disk I/O or file content storage
- Metadata extraction (ID3 tags, etc.)
- Synchronization across devices
- Playlist management

### Steps

#### 1.1 Crate Setup
- [ ] Create `sonora-db` crate with workspace integration
- [ ] Add database dependency (`sqlx`) and configure connection pool (see @example/Cargo.toml)
- [ ] Define `Node`, `NodeKind`, and error types (see @example/crates/domain)

#### 1.2 Database Migration
- [ ] Write migration to create `nodes` table with all columns
- [ ] Add unique constraint on `(parent_id, name)` to enforce sibling uniqueness

#### 1.3 Repository Layer
- [ ] Implement `create_dir` and `create_file`
- [ ] Implement `get_node` by ID and by virtual path
- [ ] Implement `list_children`
- [ ] Implement `move_node` (re-parent + rename)
- [ ] Implement `delete_node` with recursive subtree deletion
- [ ] Implement `exists` and `get_path`

#### 1.4 Tests
- [ ] Unit tests for path resolution logic
- [ ] Integration tests (testcontainers) for all repository operations
- [ ] Edge cases: delete non-empty directory, move to existing name, path of root node

