# PRD: VFS Admin Panel — Web Client

## Goal

Build a developer-facing management UI for browsing and managing the VFS node tree:
navigate directories, create directories, create file nodes (metadata only), rename/move
nodes, delete nodes, and inspect node details.

## Package Architecture

Following the `example/` pattern — `apps` is entry-only, all logic lives in `packages`:

| Package             | Name                | Role                                               |
|---------------------|---------------------|----------------------------------------------------|
| `packages/contracts`| `@sonora/contracts` | Existing generated types + HTTP client SDK         |
| `packages/ui`       | `@sonora/ui`        | Generic system-level UI primitives (no domain)     |
| `packages/features` | `@sonora/features`  | VFS domain: node tree, list, detail, dialogs, hooks|
| `packages/admin`    | `@sonora/admin`     | Page-level shell: QueryClient, layout, Toaster     |
| `apps/web/client`   | `@sonora/web-client`| Vite entry: creates client, renders `<AppShell />` |

`pnpm-workspace.yaml` at project root wires all packages.  
Vite dev proxy in `apps/web/client`: `/api` → `http://localhost:8080`.

## Confirmed Facts (from codebase inspection)

**Backend API** (all implemented, `apps/web/server`, port **8080**):
- `POST /api/nodes/dirs` — create directory
- `POST /api/nodes/files` — create file node (name, size, mimeType; metadata only)
- `GET /api/nodes/root/children` — list root nodes
- `GET /api/nodes/{id}/children` — list children of a directory
- `GET /api/nodes/{id}` — get node by id
- `GET /api/nodes/by-path?path=...` — get node by path
- `GET /api/nodes/exists?path=...` — check path existence
- `GET /api/nodes/{id}/path` — get virtual path
- `PUT /api/nodes/{id}/move` — move/rename
- `DELETE /api/nodes/{id}` — soft-delete

**Node fields**: id, parentId, name, kind ("directory"|"file"), size (bigint|null),
mimeType (string|null), md5 (string|null), storageStatus ("pending_upload"|"available"),
createdAt (bigint ms), updatedAt (bigint ms)

**Tech stack**: React 19 + TypeScript, Vite 7, Tailwind CSS v4, Radix UI, TanStack Query v5,
Lucide React, Sonner, pnpm

## Requirements

1. **Layout**: Two-panel — left collapsible directory tree, right children list panel.
2. **Toolbar**: Toolbar above the right panel with "New Folder" and "New File" buttons.
   Row-level actions (Rename, Delete) appear in a context menu (right-click) and via a
   row action button (three-dot or similar).
3. **Create directory**: Modal — name field only; parent defaults to focused directory.
4. **Create file node**: Modal — name (required), size in bytes (optional), MIME type
   (optional, free-text). No actual file upload.
5. **File size limit**: 4 MB = `4 * 1024 * 1024` bytes, validated client-side before the request.
6. **Rename**: Dialog that calls `PUT /api/nodes/{id}/move` with same `newParentId`.
7. **Delete**: Confirmation dialog → `DELETE /api/nodes/{id}`.
8. **Node detail**: Right-side detail panel (or bottom info bar) showing id, full path,
   storageStatus, md5 (or "—"), createdAt, updatedAt as human-readable timestamps.
9. **Error toasts**: Sonner toast for API errors (409 conflict, 404 not found, etc.).
10. **Dev proxy**: `/api` proxied to `http://localhost:8080` in Vite config.

## Constraints

- No actual file upload.
- 4 MB limit hardcoded in frontend only.
- MIME type: free-text input, no validation or predefined list.
- No authentication.
- No pagination for MVP.
- No drag-and-drop move for MVP.
- `bigint` timestamps from contracts must be converted to `number` for `Date` formatting.

## Acceptance Criteria

- [ ] Clicking a directory in the left tree loads its children in the right panel.
- [ ] Creating a directory with a duplicate name shows a 409 conflict toast.
- [ ] Creating a file node with size > 4 MB is blocked client-side (error shown, no API call).
- [ ] Renaming a node updates the display without a full page reload.
- [ ] Deleting a node removes it from the UI after confirmation.
- [ ] Node detail panel shows storageStatus, md5 (or "—"), and formatted timestamps.
- [ ] `pnpm dev` in `apps/web/client` starts the app; `/api` proxies to port 8080.

## Out of Scope

- File upload / content management
- Drag-and-drop move
- Pagination / virtual scrolling
- Authentication
- Production build pipeline
- Search / filter / sort
