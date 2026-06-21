# Implementation Plan: VFS Admin Panel

## Ordered Checklist

### Step 1 — Workspace & contracts package

- [ ] Create `pnpm-workspace.yaml` at project root:
      `packages: ["packages/*", "apps/**"]`
- [ ] Create root `package.json` (private workspace anchor, no scripts)
- [ ] Create `packages/contracts/package.json`:
      name `@sonora/contracts`, type `module`, exports `"."` → `"./src/index.ts"`

### Step 2 — `@sonora/ui` scaffold & primitives

- [ ] Create `packages/ui/package.json`:
      name `@sonora/ui`, peerDeps react ^19, deps: all Radix primitives used
      (`@radix-ui/react-dialog`, `react-alert-dialog`, `react-context-menu`,
      `react-dropdown-menu`), `clsx`, `tailwind-merge`, `tailwindcss`, `lucide-react`,
      `tw-animate-css`, `motion` (framer-motion v12, imported as `motion/react`)
- [ ] Create `packages/ui/tsconfig.json`
- [ ] Create `packages/ui/src/lib/cn.ts` — `clsx` + `tailwind-merge` helper
- [ ] Create `packages/ui/src/styles.css` — Tailwind v4 base + `tw-animate-css` import +
      CSS custom properties (neutral scale, accent color, radius, font-size tokens)
- [ ] Create `packages/ui/src/button.tsx` — filled / ghost / destructive variants, 32px height
- [ ] Create `packages/ui/src/icon-button.tsx` — 28×28px ghost rounded button
- [ ] Create `packages/ui/src/input.tsx` — flat style, 32px height, focus ring
- [ ] Create `packages/ui/src/label.tsx` — 11px uppercase tracking label
- [ ] Create `packages/ui/src/dialog.tsx` — Radix Dialog: Root/Trigger/Content/Header/Footer
- [ ] Create `packages/ui/src/alert-dialog.tsx` — Radix AlertDialog with destructive action
- [ ] Create `packages/ui/src/context-menu.tsx` — Radix ContextMenu thin wrapper
- [ ] Create `packages/ui/src/dropdown-menu.tsx` — Radix DropdownMenu thin wrapper
- [ ] Create `packages/ui/src/badge.tsx` — pill chip; storageStatus color variants
- [ ] Create `packages/ui/src/index.ts` — barrel export all components + cn

### Step 3 — `@sonora/features` scaffold & logic

- [ ] Create `packages/features/package.json`:
      name `@sonora/features`, deps: `@sonora/contracts workspace:*`,
      `@sonora/ui workspace:*`, `@tanstack/react-query`, `lucide-react`, `sonner`
      peerDeps: react ^19, react-dom ^19
- [ ] Create `packages/features/tsconfig.json`
- [ ] Create `packages/features/src/lib/constants.ts`:
      `export const MAX_FILE_SIZE = 4 * 1024 * 1024`
- [ ] Create `packages/features/src/lib/format.ts`:
      `formatTimestamp(ts: bigint): string`, `formatSize(bytes: bigint | null): string`
- [ ] Create `packages/features/src/hooks/use-nodes.ts` — all TanStack Query hooks:
  - `useRootChildren()` — query `["nodes","root"]` → `client.listRootChildren({})`
  - `useChildren(id)` — query `["nodes", id, "children"]` → `client.listChildren`
  - `useNodePath(id)` — query `["nodes", id, "path"]` → `client.getNodePath`
  - `useCreateDir()` — mutation, invalidates parent children on success
  - `useCreateFile()` — mutation, invalidates parent children on success
  - `useRenameNode()` — mutation (moveNode same parentId, new name), invalidates parent
  - `useDeleteNode()` — mutation, invalidates parent children on success
  - All mutations show `toast.error(...)` on failure via `onError`
- [ ] Create `packages/features/src/components/create-dir-dialog.tsx`
- [ ] Create `packages/features/src/components/create-file-dialog.tsx`
      — size field validates ≤ `MAX_FILE_SIZE` before submit; shows inline error
- [ ] Create `packages/features/src/components/rename-dialog.tsx`
- [ ] Create `packages/features/src/components/delete-confirm-dialog.tsx`
- [ ] Create `packages/features/src/components/node-detail.tsx`
- [ ] Create `packages/features/src/components/node-list.tsx`
      — toolbar (New Folder, New File), table rows, three-dot + context menu
- [ ] Create `packages/features/src/components/node-tree.tsx`
      — recursive expandable tree, root + per-dir children queries
- [ ] Create `packages/features/src/index.ts` — export all public components

### Step 4 — `@sonora/admin` page shell

- [ ] Create `packages/admin/package.json`:
      name `@sonora/admin`, deps: `@sonora/features workspace:*`, `@tanstack/react-query`,
      `sonner`, peerDeps: react ^19, react-dom ^19
- [ ] Create `packages/admin/tsconfig.json`
- [ ] Create `packages/admin/src/query-client.ts` — `new QueryClient()`
- [ ] Create `packages/admin/src/styles.css` — import `@sonora/ui` styles
- [ ] Create `packages/admin/src/app-shell.tsx`:
      Props: `{ client: ContractsClient }`. Provides: `QueryClientProvider`,
      `ClientContext`, `<Toaster />`, layout with `NodeTree` + `NodeList` + `NodeDetail`
- [ ] Create `packages/admin/src/index.ts` — `export { AppShell }`

### Step 5 — `apps/web/client` entry

- [ ] Create `apps/web/client/package.json`:
      name `@sonora/web-client`, deps: `@sonora/admin workspace:*`, `react ^19`, `react-dom ^19`
      devDeps: vite, `@vitejs/plugin-react-swc`, `@tailwindcss/vite`, typescript, @types/react,
      @types/react-dom
      scripts: `"dev": "vite"`, `"build": "tsc -b && vite build"`
- [ ] Create `apps/web/client/tsconfig.json`
- [ ] Create `apps/web/client/vite.config.ts`:
      - `@tailwindcss/vite` plugin + `@vitejs/plugin-react-swc` plugin
      - proxy `/api` → `http://localhost:8080`
      - resolve aliases for workspace packages (source-direct)
- [ ] Create `apps/web/client/index.html`
- [ ] Create `apps/web/client/src/main.tsx`:
      create `ContractsClient` via `createFetchTransport()`, render `<AppShell client={client} />`

### Step 6 — Install & smoke test

- [ ] Run `pnpm install` from project root
- [ ] Verify no TS errors: `cd packages/features && npx tsc --noEmit`
- [ ] Start dev: `pnpm --filter @sonora/web-client dev`
- [ ] Browser smoke-test: root list loads, dir navigation, create dir, create file,
      rename, delete, detail panel

## Validation Commands

```bash
pnpm install
pnpm --filter @sonora/features exec tsc --noEmit
pnpm --filter @sonora/admin exec tsc --noEmit
pnpm --filter @sonora/web-client dev
```

## Risky Files / Rollback Points

- `pnpm-workspace.yaml` at root — if this conflicts with Cargo tooling (unlikely), fall back
  to placing the workspace yaml inside `apps/web/` instead and use vite aliases.
- `packages/contracts/package.json` — adding this should be additive only; generated source
  files remain untouched.

## Notes

- No build step (tsup) for any package — Vite resolves source directly in dev mode.
- Tailwind v4: only `@tailwindcss/vite` plugin needed; CSS `@import "tailwindcss"` + `@import "tw-animate-css"` in styles.css.
- `motion/react`: use `<AnimatePresence>` for dialog mount/unmount, `<motion.div>` for list item enter, tree expand, and detail panel slide-in.
- Context for hooks: pass `ContractsClient` via React context from `AppShell`; hooks call
  `useContext(ClientContext)` internally so callers don't thread the client through props.
- `bigint` → `number` coercion happens inside `format.ts` helpers, not scattered in components.
