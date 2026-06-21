# Design: VFS Admin Panel

## Directory Layout

```
sonora/
├── pnpm-workspace.yaml                    ← NEW
├── packages/
│   ├── contracts/
│   │   ├── src/                           ← existing generated files (unchanged)
│   │   └── package.json                   ← NEW: "@sonora/contracts"
│   │
│   ├── ui/                                ← NEW: generic system-level UI primitives
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── src/
│   │       ├── index.ts
│   │       ├── styles.css                 — Tailwind base + design tokens
│   │       ├── button.tsx
│   │       ├── icon-button.tsx
│   │       ├── input.tsx
│   │       ├── label.tsx
│   │       ├── dialog.tsx                 — Radix Dialog wrapper
│   │       ├── alert-dialog.tsx           — Radix AlertDialog wrapper
│   │       ├── context-menu.tsx           — Radix ContextMenu wrapper
│   │       ├── dropdown-menu.tsx          — Radix DropdownMenu wrapper
│   │       ├── badge.tsx                  — status / tag chip
│   │       └── lib/
│   │           └── cn.ts                  — clsx + tailwind-merge helper
│   │
│   ├── features/                          ← NEW: VFS domain components + hooks
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── src/
│   │       ├── index.ts
│   │       ├── hooks/
│   │       │   └── use-nodes.ts           — all TanStack Query hooks
│   │       ├── lib/
│   │       │   ├── constants.ts           — MAX_FILE_SIZE = 4 * 1024 * 1024
│   │       │   └── format.ts             — formatTimestamp, formatSize
│   │       └── components/
│   │           ├── node-tree.tsx          — left panel: collapsible dir tree
│   │           ├── node-list.tsx          — right panel: children table + toolbar
│   │           ├── node-detail.tsx        — detail area: id, path, meta fields
│   │           ├── create-dir-dialog.tsx
│   │           ├── create-file-dialog.tsx
│   │           ├── rename-dialog.tsx
│   │           └── delete-confirm-dialog.tsx
│   │
│   └── admin/                             ← NEW: page-level admin shell
│       ├── package.json
│       ├── tsconfig.json
│       └── src/
│           ├── index.ts
│           ├── query-client.ts
│           ├── styles.css
│           └── app-shell.tsx              — QueryClientProvider + layout + ContractsClient prop
│
└── apps/
    └── web/
        └── client/                        ← NEW: entry point only
            ├── package.json
            ├── tsconfig.json
            ├── vite.config.ts
            ├── index.html
            └── src/
                └── main.tsx
```

## Package Graph

```
@sonora/contracts   (generated, no deps)
      ↑
@sonora/ui          (Radix UI, Tailwind — no domain knowledge)
      ↑
@sonora/features    (contracts + ui + tanstack-query — VFS domain)
      ↑
@sonora/admin       (features + query-client — page composition)
      ↑
@sonora/web-client  (admin + react/react-dom — Vite entry)
```

## UI Design Principles (`@sonora/ui`)

Target aesthetic: **Android/iOS system-level** — minimal, tactile, purposeful.

- **Color**: neutral palette; accent only for primary actions; no decorative color.
- **Typography**: system-sans stack (`-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`);
  tight hierarchy (13px body, 11px caption, 15px heading).
- **Spacing**: 4px base unit; generous internal padding (12–16px) on interactive elements.
- **Elevation**: no heavy box-shadows; use background-color steps (`bg-neutral-50/100/200`)
  to distinguish layers.
- **Borders**: 1px `border-neutral-200` only as dividers; no borders on buttons.
- **Motion**: fast micro-animations (100–150ms ease-out); no bouncy or slow transitions.
- **Interactive states**: clear hover (bg shift), active (slight scale or bg darken), focus
  (2px ring in accent color, no default browser outline).
- **Density**: compact — admin tools should show maximum information without scrolling.

### Component-level notes

| Component       | Style notes                                                              |
|-----------------|--------------------------------------------------------------------------|
| `Button`        | Filled (primary) / ghost / destructive variants; 32px height compact     |
| `IconButton`    | 28px × 28px, rounded, ghost by default; used for toolbar and row actions |
| `Input`         | No border-box, underline or flat style; 32px height                      |
| `Dialog`        | Centered overlay; 480px max-width; header + body + footer sections       |
| `AlertDialog`   | Same as Dialog; destructive action button uses red variant                |
| `ContextMenu`   | 200px min-width; 32px item height; 8px padding                           |
| `Badge`         | 5px radius pill; `pending_upload` → amber, `available` → green           |

## Data Flow

```
main.tsx
  └─ createContractsClient(createFetchTransport())
       └─ <AppShell client={client} />              [@sonora/admin]
            └─ QueryClientProvider
                 └─ layout: NodeTree | NodeList | NodeDetail   [@sonora/features]
```

### State topology (`AppShell`)

```ts
selectedDirId: string | null   // which directory the right panel shows
selectedNodeId: string | null  // which node is shown in NodeDetail
```

Both lifted to `AppShell`; passed down as props to avoid unnecessary context.

### TanStack Query Keys

```ts
["nodes", "root"]             → listRootChildren
["nodes", id, "children"]     → listChildren(id)
["nodes", id, "path"]         → getNodePath(id)
```

On create/rename/delete success → invalidate `["nodes", parentId, "children"]`
(and `["nodes", "root"]` when parentId is null).

## Component Responsibilities

### `NodeTree` (left panel, `@sonora/features`)
- Fetches root children; each directory row is expandable → fetches its own children.
- `expanded: Set<string>` local state; controlled expand via click.
- Clicking any node calls `onSelectDir(id)` / `onSelectNode(id)`.
- Files shown in tree only if they are at root; directories always shown.

### `NodeList` (right panel, `@sonora/features`)
- Fetches children of `selectedDirId` (or root if null).
- Toolbar: `New Folder` + `New File` `IconButton`s.
- Table columns: icon, name, kind badge, size, mimeType, storageStatus badge, updatedAt.
- Row: clicking dir → `onSelectDir(id)`; clicking file → `onSelectNode(id)`.
- Row three-dot `IconButton` → `DropdownMenu` → Rename / Delete.
- Row right-click → `ContextMenu` → same actions.

### `NodeDetail` (bottom strip or right rail, `@sonora/features`)
- Shown when `selectedNodeId` is set.
- Fetches path via `useNodePath(selectedNodeId)`.
- Displays: name, kind, id (monospace truncated), path, storageStatus `Badge`,
  md5 (or "—"), size formatted, createdAt, updatedAt.

### Dialogs (`@sonora/features` — use `@sonora/ui` primitives)
- All dialogs are uncontrolled-open-state; parent passes `open` + `onOpenChange`.
- `CreateDirDialog`: name field → `useCreateDir()`.
- `CreateFileDialog`: name (required), size (validates ≤ `MAX_FILE_SIZE`), mimeType (free-text).
- `RenameDialog`: pre-filled name → `useRenameNode()`.
- `DeleteConfirmDialog`: destructive `AlertDialog` → `useDeleteNode()`.

## bigint Coercion

Contracts emit `bigint` for `size`, `createdAt`, `updatedAt`. Coerce at the React boundary:

```ts
const sizeNum = node.size !== null ? Number(node.size) : null;
const createdAt = new Date(Number(node.createdAt));
```

Safe: timestamps < 2^53, sizes ≤ 4 MB.

## Error Handling

`ContractTransportError` has `status` and `message`. Map at the mutation's `onError`:

```ts
const label =
  err.status === 409 ? "A node with that name already exists"
  : err.status === 404 ? "Node not found"
  : err.message;
toast.error(label);
```

## Vite Config (apps/web/client)

```ts
proxy: { "/api": { target: "http://localhost:8080", changeOrigin: true } }
resolve.alias: {
  "@sonora/contracts": "<root>/../../packages/contracts/src/index.ts",
  "@sonora/admin":     "<root>/../../packages/admin/src/index.ts",
}
```

(Workspace `workspace:*` deps + Vite alias gives zero-build-step dev experience.)

## Workspace Wiring

`pnpm-workspace.yaml`:
```yaml
packages:
  - "packages/*"
  - "apps/**"
```

All packages use `"exports": { ".": "./src/index.ts" }` — source-direct, no pre-build for dev.
