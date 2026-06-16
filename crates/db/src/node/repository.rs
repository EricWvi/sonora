use sonora_domain::{Node, NodeId, NodeKind};

use crate::node::{NewFile, NodeRepository, NodeRepositoryError};
use sqlx::{Pool, Postgres, Row};

use crate::time::TimestampSource;

/// PostgreSQL-backed implementation of [`NodeRepository`] against the `nodes` table.
pub struct PostgresNodeRepository<T> {
    pool: Pool<Postgres>,
    timestamp_source: T,
}

impl<T: TimestampSource> PostgresNodeRepository<T> {
    /// Wraps an existing connection pool and timestamp source.
    pub fn new(pool: Pool<Postgres>, timestamp_source: T) -> Self {
        Self {
            pool,
            timestamp_source,
        }
    }
}

impl<T: TimestampSource + Send + Sync> NodeRepository for PostgresNodeRepository<T> {
    async fn create_dir(
        &self,
        parent_id: Option<NodeId>,
        name: String,
    ) -> Result<Node, NodeRepositoryError> {
        let id = NodeId::generate();
        let now = self.timestamp_source.current_timestamp_millis();

        let row = sqlx::query(
            r#"
            INSERT INTO nodes (id, parent_id, name, kind, size, mime_type,
                               created_at, updated_at, is_deleted)
            VALUES ($1, $2, $3, 'directory', NULL, NULL, $4, $4, FALSE)
            RETURNING id, parent_id, name, kind, size, mime_type,
                      created_at, updated_at, server_version, is_deleted
            "#,
        )
        .bind(id)
        .bind(parent_id)
        .bind(&name)
        .bind(now)
        .fetch_one(&self.pool)
        .await
        .map_err(|e| map_sqlx_error(e, &name))?;

        row_to_node(&row)
    }

    async fn create_file(&self, file: NewFile) -> Result<Node, NodeRepositoryError> {
        let id = NodeId::generate();
        let now = self.timestamp_source.current_timestamp_millis();

        let row = sqlx::query(
            r#"
            INSERT INTO nodes (id, parent_id, name, kind, size, mime_type,
                               created_at, updated_at, is_deleted)
            VALUES ($1, $2, $3, 'file', $4, $5, $6, $6, FALSE)
            RETURNING id, parent_id, name, kind, size, mime_type,
                      created_at, updated_at, server_version, is_deleted
            "#,
        )
        .bind(id)
        .bind(file.parent_id)
        .bind(&file.name)
        .bind(file.size)
        .bind(&file.mime_type)
        .bind(now)
        .fetch_one(&self.pool)
        .await
        .map_err(|e| map_sqlx_error(e, &file.name))?;

        row_to_node(&row)
    }

    async fn get_node_by_id(&self, id: NodeId) -> Result<Option<Node>, NodeRepositoryError> {
        let row = sqlx::query(
            r#"
            SELECT id, parent_id, name, kind, size, mime_type,
                   created_at, updated_at, server_version, is_deleted
            FROM nodes
            WHERE id = $1 AND is_deleted = FALSE
            "#,
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

        row.map(|r| row_to_node(&r)).transpose()
    }

    async fn get_node_by_path(&self, path: &str) -> Result<Option<Node>, NodeRepositoryError> {
        let components = parse_path(path)?;

        let mut current_parent: Option<NodeId> = None;
        let mut current_node: Option<Node> = None;

        for component in &components {
            let row = if let Some(parent_id) = current_parent {
                sqlx::query(
                    r#"
                    SELECT id, parent_id, name, kind, size, mime_type,
                           created_at, updated_at, server_version, is_deleted
                    FROM nodes
                    WHERE parent_id = $1 AND name = $2 AND is_deleted = FALSE
                    "#,
                )
                .bind(parent_id)
                .bind(component)
                .fetch_optional(&self.pool)
                .await
            } else {
                sqlx::query(
                    r#"
                    SELECT id, parent_id, name, kind, size, mime_type,
                           created_at, updated_at, server_version, is_deleted
                    FROM nodes
                    WHERE parent_id IS NULL AND name = $1 AND is_deleted = FALSE
                    "#,
                )
                .bind(component)
                .fetch_optional(&self.pool)
                .await
            }
            .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

            match row {
                None => return Ok(None),
                Some(row) => {
                    let node = row_to_node(&row)?;
                    current_parent = Some(node.id);
                    current_node = Some(node);
                }
            }
        }

        Ok(current_node)
    }

    async fn list_children(
        &self,
        parent_id: Option<NodeId>,
    ) -> Result<Vec<Node>, NodeRepositoryError> {
        let rows = if let Some(pid) = parent_id {
            sqlx::query(
                r#"
                SELECT id, parent_id, name, kind, size, mime_type,
                       created_at, updated_at, server_version, is_deleted
                FROM nodes
                WHERE parent_id = $1 AND is_deleted = FALSE
                ORDER BY kind ASC, name ASC
                "#,
            )
            .bind(pid)
            .fetch_all(&self.pool)
            .await
        } else {
            sqlx::query(
                r#"
                SELECT id, parent_id, name, kind, size, mime_type,
                       created_at, updated_at, server_version, is_deleted
                FROM nodes
                WHERE parent_id IS NULL AND is_deleted = FALSE
                ORDER BY kind ASC, name ASC
                "#,
            )
            .fetch_all(&self.pool)
            .await
        }
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

        rows.iter().map(row_to_node).collect()
    }

    async fn move_node(
        &self,
        node_id: NodeId,
        new_parent_id: Option<NodeId>,
        new_name: String,
    ) -> Result<Node, NodeRepositoryError> {
        let now = self.timestamp_source.current_timestamp_millis();

        let row = sqlx::query(
            r#"
            UPDATE nodes
            SET parent_id = $2, name = $3, updated_at = $4
            WHERE id = $1 AND is_deleted = FALSE
            RETURNING id, parent_id, name, kind, size, mime_type,
                      created_at, updated_at, server_version, is_deleted
            "#,
        )
        .bind(node_id)
        .bind(new_parent_id)
        .bind(&new_name)
        .bind(now)
        .fetch_optional(&self.pool)
        .await
        .map_err(|e| map_sqlx_error(e, &new_name))?;

        row.map(|r| row_to_node(&r))
            .transpose()?
            .ok_or(NodeRepositoryError::NotFound(node_id))
    }

    async fn delete_node(&self, node_id: NodeId) -> Result<(), NodeRepositoryError> {
        let now = self.timestamp_source.current_timestamp_millis();

        // Soft-delete the node and its entire subtree using a recursive CTE.
        let result = sqlx::query(
            r#"
            WITH RECURSIVE subtree AS (
                SELECT id FROM nodes WHERE id = $1 AND is_deleted = FALSE
                UNION ALL
                SELECT n.id FROM nodes n
                JOIN subtree s ON n.parent_id = s.id
                WHERE n.is_deleted = FALSE
            )
            UPDATE nodes SET is_deleted = TRUE, updated_at = $2
            WHERE id IN (SELECT id FROM subtree)
            "#,
        )
        .bind(node_id)
        .bind(now)
        .execute(&self.pool)
        .await
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

        if result.rows_affected() == 0 {
            return Err(NodeRepositoryError::NotFound(node_id));
        }

        Ok(())
    }

    async fn exists(&self, path: &str) -> Result<bool, NodeRepositoryError> {
        self.get_node_by_path(path).await.map(|n| n.is_some())
    }

    async fn get_path(&self, node_id: NodeId) -> Result<String, NodeRepositoryError> {
        // Walk ancestors from the node upward; depth=0 is the node itself, increasing toward root.
        // Ordering by depth DESC puts the root-closest ancestor first, giving path components
        // in top-down order ready for joining.
        let names: Vec<String> = sqlx::query_scalar(
            r#"
            WITH RECURSIVE path_cte AS (
                SELECT id, parent_id, name, 0 AS depth
                FROM nodes
                WHERE id = $1 AND is_deleted = FALSE
                UNION ALL
                SELECT n.id, n.parent_id, n.name, p.depth + 1
                FROM nodes n
                JOIN path_cte p ON n.id = p.parent_id
                WHERE n.is_deleted = FALSE
            )
            SELECT name FROM path_cte ORDER BY depth DESC
            "#,
        )
        .bind(node_id)
        .fetch_all(&self.pool)
        .await
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

        if names.is_empty() {
            return Err(NodeRepositoryError::NotFound(node_id));
        }

        Ok(names.join("/"))
    }
}

/// Splits a virtual path into its component names.
///
/// Both `/` and `\` are accepted as separators. Leading, trailing, and repeated
/// separators are ignored. Returns `Err(InvalidPath)` when the path produces no
/// components.
fn parse_path(path: &str) -> Result<Vec<&str>, NodeRepositoryError> {
    let components: Vec<&str> = path.split(['/', '\\']).filter(|s| !s.is_empty()).collect();

    if components.is_empty() {
        return Err(NodeRepositoryError::InvalidPath(
            "path must not be empty".to_string(),
        ));
    }

    Ok(components)
}

/// Converts a unique-constraint violation from sqlx into a user-facing NameConflict error;
/// all other database errors become OperationFailed.
fn map_sqlx_error(error: sqlx::Error, name: &str) -> NodeRepositoryError {
    if let sqlx::Error::Database(ref db_err) = error {
        // PostgreSQL unique-violation code is 23505.
        if db_err.code().as_deref() == Some("23505") {
            return NodeRepositoryError::NameConflict {
                name: name.to_string(),
            };
        }
    }
    NodeRepositoryError::OperationFailed(error.to_string())
}

/// Constructs a `Node` from a PostgreSQL result row.
fn row_to_node(row: &sqlx::postgres::PgRow) -> Result<Node, NodeRepositoryError> {
    let id: NodeId = row
        .try_get("id")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let parent_id: Option<NodeId> = row
        .try_get("parent_id")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let name: String = row
        .try_get("name")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let kind_str: String = row
        .try_get("kind")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let size: Option<i64> = row
        .try_get("size")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let mime_type: Option<String> = row
        .try_get("mime_type")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let created_at: i64 = row
        .try_get("created_at")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let updated_at: i64 = row
        .try_get("updated_at")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let server_version: i64 = row
        .try_get("server_version")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;
    let is_deleted: bool = row
        .try_get("is_deleted")
        .map_err(|e| NodeRepositoryError::OperationFailed(e.to_string()))?;

    let kind = match kind_str.as_str() {
        "directory" => NodeKind::Directory,
        "file" => NodeKind::File,
        other => {
            return Err(NodeRepositoryError::OperationFailed(format!(
                "unknown node kind in database: `{other}`"
            )));
        }
    };

    Ok(Node::new(
        id,
        parent_id,
        name,
        kind,
        size,
        mime_type,
        created_at,
        updated_at,
        server_version,
        is_deleted,
    ))
}
