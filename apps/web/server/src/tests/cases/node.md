# Node handler test cases

Handler under test: `apps/web/server/src/handlers/node.rs`

---

## CD — create_dir (`POST /api/nodes/dirs`)

```
/// CD-01: root-level directory → 201, all fields stored correctly.
async fn cd_01_root_stores_all_fields

/// CD-02: nested directory → 201, parentId set correctly.
async fn cd_02_nested_sets_parent_id

/// CD-03: duplicate name under same parent → 409.
async fn cd_03_duplicate_name_returns_409
```

---

## CF — create_file (`POST /api/nodes/files`)

```
/// CF-01: file with size and MIME type → 201, all fields stored correctly.
async fn cf_01_with_metadata_stores_all_fields
```

---

## GI — get_node_by_id (`GET /api/nodes/{id}`)

```
/// GI-01: existing id → 200, full node returned.
async fn gi_01_existing_id_returns_node

/// GI-02: unknown id → 404.
async fn gi_02_unknown_id_returns_404
```

---

## GP — get_node_by_path (`GET /api/nodes/by-path?path=...`)

```
/// GP-01: three-level path → 200, correct file node returned.
async fn gp_01_deep_path_returns_node

/// GP-02: path with a missing component → 404.
async fn gp_02_missing_component_returns_404
```

---

## LC — list_children (`GET /api/nodes/{id}/children`)

```
/// LC-01: directory contains a subdirectory and a file → directories returned before files.
async fn lc_01_dirs_returned_before_files

/// LC-02: directory with no children → empty list.
async fn lc_02_empty_dir_returns_empty_list
```

---

## MN — move_node (`PUT /api/nodes/{id}/move`)

```
/// MN-01: same parent, new name → node renamed in place.
async fn mn_01_same_parent_renames_node

/// MN-02: different parent → parentId updated.
async fn mn_02_new_parent_updates_parent_id

/// MN-03: unknown node id → 404.
async fn mn_03_unknown_id_returns_404

/// MN-04: target name already taken at destination → 409.
async fn mn_04_name_taken_returns_409
```

---

## DN — delete_node (`DELETE /api/nodes/{id}`)

```
/// DN-01: deleted file → 204, no longer accessible via get_node_by_id.
async fn dn_01_file_no_longer_visible_after_delete

/// DN-02: delete directory → subtree soft-deleted, all descendants return 404.
async fn dn_02_directory_recursively_deletes_subtree

/// DN-03: unknown id → 404.
async fn dn_03_unknown_id_returns_404

/// DN-04: name reused after soft-delete → 201, new node with a different id.
async fn dn_04_deleted_name_can_be_reused
```

---

## EX — exists (`GET /api/nodes/exists?path=...`)

```
/// EX-01: known path → true; unknown path → false.
async fn ex_01_known_path_true_unknown_path_false
```

---

## PT — get_path (`GET /api/nodes/{id}/path`)

```
/// PT-01: deeply nested node → 200, full slash-separated path returned.
async fn pt_01_nested_node_returns_full_path

/// PT-02: root-level node → 200, single name with no separator.
async fn pt_02_root_node_returns_single_name

/// PT-03: unknown id → 404.
async fn pt_03_unknown_id_returns_404
```
