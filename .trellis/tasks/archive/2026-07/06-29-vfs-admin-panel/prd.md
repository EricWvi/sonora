# VFS admin panel (apps/web/client)

## Goal

Build a web admin panel that lets an operator browse and manage the existing VFS
node model like a lightweight web disk: navigate the directory tree, create
folders and file placeholders, rename/move, and soft-delete nodes. The panel is
the primary UI surface for the VFS while file *content* delivery (upload) is
still being built.

## User Value

- Operators can visually browse the VFS tree (root + nested directories) without
  hitting raw HTTP endpoints.
- Create folders and file metadata placeholders; rename, move, and delete nodes
  through idiomatic UI instead of curl.
- See node metadata at a glance (size, mime type, md5, storage status).

## Confirmed Facts (from codebase)

- Backend `apps/web/server` (axum) listens on `0.0.0.0:8080` and exposes a full
  VFS node API: createDir, createFile, getNodeByPath, nodeExists,
  listRootChildren, getNodeById, listChildren, getNodePath, moveNode, deleteNode.
- `@sonora/contracts` SDK is generated and ready: `createContractsClient(
  createFetchTransport())` from `@sonora/contracts` and `@sonora/contracts/fetch`.
  `createFetchTransport({ baseUrl })` defaults to relative paths (empty baseUrl),
  so dev can ride the Vite `/api` proxy.
- `apps/web/client` is an empty placeholder (only empty `src/`, no `package.json`).
- `packages/admin`, `packages/ui`, `packages/features` are empty placeholder
  packages (no source files yet).
- `Taskfile.yml` already has `run:admin` (`apps/web/client` -> `pnpm dev`) and
  `run:web-backend`.
- Reference project `example/apps/web/journal` uses: React 19, Vite 7,
  `@vitejs/plugin-react-swc`, Tailwind CSS v4 (`@tailwindcss/vite`), TS 5.8.
  Pattern: thin `main.tsx` entry (`createContractsClient(createFetchTransport())`
  + `<AppShell client={client} />`) backed by a separate feature package.
- `NodeView` fields: id, parentId, name, kind (`directory`|`file`), size,
  mimeType, md5, storageStatus (`pending_upload`|`available`), createdAt,
  updatedAt. `CreateFileRequest` takes parentId, name, size, mimeType.
- There is **no** file-upload backend endpoint; upload is out of scope this round.
- Reference UI package uses shadcn/Radix; user explicitly requested **heroui**, so
  the UI library choice intentionally diverges from the reference project.

## Requirements

- Scaffold a Vite + React + TypeScript app at `apps/web/client` runnable via
  `pnpm dev` (consumed by `task run:admin`), with Tailwind v4 and heroui wired up.
- Vite dev server proxies `/api` -> `http://localhost:8080` (backend).
- Integrate `@sonora/contracts` SDK via `createFetchTransport()` + a single client.
- Folder/file browsing: list root children, open folders to list their children,
  show breadcrumb path navigation (via getNodePath), navigate up.
- Node table/list view showing name, kind, size (files), mime type, storage
  status, updated time; click folder to descend.
- Create directory (createDir) with a name dialog under the current folder.
- Create file placeholder (createFile) with a name + lenient mime type; the
  request `size` is hardcoded to **4 MB** in the frontend (placeholder, since
  content arrives via upload later). Mime type is not strictly validated.
- Rename / move node (moveNode) — at minimum inline rename; move-to-folder if
  feasible in scope.
- Soft-delete node (deleteNode) with a confirmation.
- Surface transport errors (`ContractTransportError`) via heroui toasts/notifications.
- Loading and empty states for the listing.

## Constraints

- React + Vite + Tailwind v4 + heroui. No shadcn/Radix.
- No file upload UI or upload endpoint work this round.
- Do not hand-edit generated `@sonora/contracts` files.
- File content language: English in all written file content (code, comments,
  docs). Conversation in Chinese with the user.

## Acceptance Criteria

- [ ] `task run:admin` (or `pnpm --filter ... dev`) starts the Vite dev server
      and the app loads in the browser.
- [ ] With `task run:web-backend` + DB running, the root listing renders real
      nodes from `listRootChildren`.
- [ ] Operator can navigate into folders and back via breadcrumbs; current path
      reflects `getNodePath`.
- [ ] Operator can create a directory under the current folder; it appears in the
      list after refresh.
- [ ] Operator can create a file placeholder; the request sends `size` = 4 MB and
      a (leniently accepted) mime type; node appears with storageStatus
      `pending_upload`.
- [ ] Operator can rename a node (moveNode) and see the updated name.
- [ ] Operator can delete a node after confirmation; it appears gone from the list.
- [ ] Transport errors surface as user-visible notifications, not console-only.
- [ ] `pnpm -r exec tsc --noEmit` (or equivalent) passes with no type errors.

## Decisions

- **Monorepo layout (Q1, confirmed):** follow the reference "thin entry + feature
  package" pattern. `apps/web/client` is a thin Vite entry (`main.tsx` creates the
  contracts client via `createFetchTransport()` and renders `<AppShell>`, plus
  `index.html`, `vite.config.ts`, `tsconfig.json`, `package.json`, heroui
  Provider). Business logic (AppShell, node pages, dialogs, hooks) lives in the
  pre-reserved `packages/admin` package, consumed as `@sonora/admin`.
- **Move scope (Q2, confirmed):** inline rename only. `moveNode` is used with the
  new name and the existing `parentId` (unchanged). Move-to-folder is out of
  scope for this round.
- **HeroUI version (Q3, confirmed):** HeroUI **v3** (released 2026-03), which is
  built natively on Tailwind CSS v4 and React Aria and is the version HeroUI
  recommends for new projects. v2 is heading toward deprecation, so starting on
  v3 avoids an immediate migration debt.

## Strong defaults adopted (no separate question needed)

- Create-file form: `name` (required text), `mimeType` (optional text input, not
  strictly validated), and `size` hardcoded to a 4 MB frontend constant
  (`4 * 1024 * 1024`) — content arrives via upload later.
- Listing rendered as a heroui Table (admin semantics, dense metadata: name,
  kind, size, mimeType, storageStatus, updatedAt); folder rows descend on click.
- Breadcrumb navigation built from `getNodePath` results; navigate up by clicking
  a breadcrumb segment.
- Server state via `@tanstack/react-query`; the `ContractsClient` is provided
  through React Context inside `AppShell` and consumed by hooks.

## Out of Scope

- File content upload (any transport of bytes).
- Backend changes (no new endpoints, no Rust work).
- Authentication / authorization / multi-user.
- Search, sorting, pagination, drag-and-drop move.
- File preview / download.
- Android / Mac clients.

## Open Questions (blocking planning)

- None. All product/scope/version decisions resolved (Q1 layout, Q2 move scope,
  Q3 HeroUI v3). Remaining choices (HeroUI CLI setup, react-router, tanstack-query)
  are technical design decisions captured in `design.md`.
