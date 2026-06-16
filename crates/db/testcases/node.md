# Node repository test cases

Repository under test: `PostgresNodeRepository` (`crates/db/src/node/repository.rs`)

---

## CD — create_dir

```
/// CD-01: root-level directory → all fields stored correctly.
async fn cd_01_root_stores_all_fields

/// CD-02: nested directory → parent link set correctly.
async fn cd_02_nested_sets_parent_link

/// CD-03: duplicate name under same parent → NameConflict.
async fn cd_03_duplicate_name_returns_conflict
```

---

## CF — create_file

```
/// CF-01: file with size and MIME type → all fields stored correctly.
async fn cf_01_with_metadata_stores_all_fields
```

---

## GI — get_node_by_id

```
/// GI-01: existing id → full node returned.
async fn gi_01_existing_id_returns_node

/// GI-02: unknown id → None.
async fn gi_02_unknown_id_returns_none
```

---

## GP — get_node_by_path

```
/// GP-01: three-level path → correct file node returned.
async fn gp_01_deep_path_returns_node

/// GP-02: path with a missing component → None.
async fn gp_02_missing_component_returns_none
```

---

## LC — list_children

```
/// LC-01: root contains a directory and a file → directories returned before files.
async fn lc_01_dirs_returned_before_files

/// LC-02: directory with no children → empty vec.
async fn lc_02_empty_dir_returns_empty_vec
```

---

## MN — move_node

```
/// MN-01: same parent, new name → node renamed in place.
async fn mn_01_same_parent_renames_node

/// MN-02: different parent → parent_id updated.
async fn mn_02_new_parent_updates_parent_id

/// MN-03: unknown node id → NotFound.
async fn mn_03_unknown_id_returns_not_found

/// MN-04: target name already taken at destination → NameConflict.
async fn mn_04_name_taken_returns_conflict
```

---

## DN — delete_node

```
/// DN-01: file deleted → no longer visible via get_node_by_id.
async fn dn_01_file_no_longer_visible_after_delete

/// DN-02: directory with nested children deleted → entire subtree soft-deleted.
async fn dn_02_directory_recursively_deletes_subtree

/// DN-03: unknown id → NotFound.
async fn dn_03_unknown_id_returns_not_found

/// DN-04: name reused after soft-delete → new node created with a different id.
async fn dn_04_deleted_name_can_be_reused
```

---

## EX — exists

```
/// EX-01: known path → true; unknown path → false.
async fn ex_01_known_path_true_unknown_path_false
```

---

## PT — get_path

```
/// PT-01: deeply nested node → full slash-separated path returned.
async fn pt_01_nested_node_returns_full_path

/// PT-02: root-level node → single name returned with no separator.
async fn pt_02_root_node_returns_single_name

/// PT-03: unknown id → NotFound.
async fn pt_03_unknown_id_returns_not_found
```
