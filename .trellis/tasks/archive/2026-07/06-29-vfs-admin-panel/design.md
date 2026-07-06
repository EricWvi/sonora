# Design — VFS admin panel

## Architecture & boundaries

Two layers, mirroring the reference project (`example/apps/web/journal` +
`example/packages/journal`):

- **Host — `apps/web/client`** (`@sonora/web-client`): a thin Vite + React entry.
  Responsibilities only: bootstrap React, assemble the contracts client
  (`createContractsClient(createFetchTransport())`), and mount the provider stack
  + `<AppShell>`. Contains no business logic.
- **Feature package — `packages/admin`** (`@sonora/admin`): all VFS business
  logic. Exports `AppShell`. `peerDependencies`: react, react-dom. Regular deps:
  `@sonora/contracts` (workspace), `@tanstack/react-query`, `@heroui/react`,
  `framer-motion`, `lucide-react`.

Boundary rule: the host never imports heroui internals or admin modules beyond
`AppShell`; the feature package never touches Vite/env directly and receives the
transport only via the `client` prop / Context.

## Module layout (`packages/admin/src`)

```
src/
  index.ts                       # export { AppShell }
  app-shell.tsx                  # QueryClientProvider + HeroUIProvider + ClientProvider + NodeExplorer
  providers/
    client-context.tsx           # ClientProvider + useClient()
  lib/
    error-toast.ts               # ContractTransportError -> toast notification
  features/nodes/
    constants.ts                 # FILE_PLACEHOLDER_SIZE = 4 * 1024 * 1024
    node-explorer.tsx            # top-level: breadcrumb + toolbar + table
    components/
      node-table.tsx             # heroui Table listing
      node-breadcrumb.tsx        # path breadcrumb + navigation
      create-dir-dialog.tsx
      create-file-dialog.tsx
      rename-dialog.tsx
      delete-confirm-dialog.tsx
    hooks/
      use-node-children.ts       # query: listChildren / listRootChildren
      use-node-path.ts           # query: getNodePath
      use-node-mutations.ts      # mutations: createDir, createFile, moveNode, deleteNode
```

## Data flow & contracts

Client assembly (host `main.tsx`):

```ts
const client = createContractsClient(createFetchTransport());
```

`AppShell` wraps the tree in `<ClientProvider value={client}>`; hooks obtain it
via `const client = useClient()`.

Server state via `@tanstack/react-query` (type-safe — hooks call the typed
`ContractsClient` directly, unlike the reference project's URL-based queryFn):

- Query keys: `["nodes", "children", parentId]`, `["nodes", "path", id]`.
- `useNodeChildren(parentId)`:
  `parentId == null ? client.listRootChildren({}) : client.listChildren({ id: parentId })`.
- `useNodePath(id)`: `client.getNodePath({ id })` (path string for the breadcrumb title).
- Mutations (`useNodeMutations`):
  - `createDir({ parentId, name })` -> `client.createDir(...)`.
  - `createFile({ parentId, name, mimeType })` ->
    `client.createFile({ parentId, name, size: FILE_PLACEHOLDER_SIZE, mimeType })`.
  - `rename({ id, parentId, newName })` ->
    `client.moveNode({ id, newParentId: parentId, newName })`.
  - `remove({ id })` -> `client.deleteNode({ id })`.
- Cache invalidation: on mutation success,
  `queryClient.invalidateQueries({ queryKey: ["nodes"] })`.
- Errors: each mutation's `onError` routes `ContractTransportError` to a toast.

## Navigation state

No client-side router — the URL stays fixed at `/`. The current directory lives
in component state inside `NodeExplorer` as a directory stack
`dirStack: string[]` (ids from root to the current folder; `[]` means root).

- Open folder: `setDirStack([...stack, id])`.
- Click a breadcrumb segment: truncate the stack to that index.
- Go up: pop the last id.

`getNodePath({ id })` is used only to render the human-readable path title in
the breadcrumb; navigation itself is driven by the stack, so path strings are
never parsed back to ids. Refreshing the page returns to root, which is
acceptable for an admin tool.

Fallback if router integration proves costly: plain `useState` for
`currentDirId` (refresh returns to root). To be decided during implementation,
router is the default.

## HeroUI v3 + Tailwind v4 setup

- Dependencies: `@heroui/react@^3`, `framer-motion` (heroui runtime dep),
  `tailwindcss@^4`, `@tailwindcss/vite`, `@vitejs/plugin-react-swc`.
- Setup path (preferred): run the official HeroUI CLI inside `apps/web/client`:
  - `npx heroui@latest init` -> generates the `hero.ts` theme + the Tailwind v4
    CSS plugin wiring and the component-folder convention.
  - `npx heroui@latest add table button modal input ...` -> adds each used
    component.
  The CLI is preferred because several v3 doc pages are still 404; the CLI is the
  most reliable v3 setup route.
  - Manual fallback: write the CSS by hand per v3 docs
    (`@import "tailwindcss";` + heroui `@plugin`/`hero.ts`).
- Provider: `main.tsx` wraps the app in `<HeroUIProvider>`.
- pnpm: add to repo-root `.npmrc`:
  ```
  public-hoist-pattern[]=*@heroui*
  public-hoist-pattern[]=*framer-motion*
  ```
  Tailwind v4 auto-scans `node_modules`; hoisting ensures heroui's class
  definitions are reachable.

## Workspace package resolution & Vite config

Workspace packages are resolved via **conditional `exports`** — the same pattern
`@sonora/contracts` already uses — not via Vite aliases.

`packages/admin/package.json`:
```json
{
  "exports": {
    ".": {
      "development": "./src/index.ts",
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    }
  }
}
```
- **Dev**: Vite activates the `development` condition by default, so
  `@sonora/admin` resolves straight to its source — full HMR, no background
  watch/build process.
- **Build**: the `import` condition resolves to `dist/`, so `packages/admin`
  must be built (`tsup`) before `@sonora/web-client` is built (mirrors
  `@sonora/contracts`).

`apps/web/client/tsconfig.json`: set `customConditions: ["development"]` so TS
type resolution in the IDE follows the same source path; no `paths` aliases
needed.

`apps/web/client/vite.config.ts`:
- `plugins: [react(), tailwindcss()]`.
- `server.fs.allow: [".."]` so Vite can read workspace sibling sources.
- `server.watch.ignored: ["!**/node_modules/@sonora/**"]` so the symlinked
  workspace packages stay watched.
- `server.proxy`: `/api` -> `http://localhost:8080` (backend).
- `server.port`: `5173`.
- `build.outDir`: `dist`.
- No `resolve.alias` for workspace packages (only the local `@` -> `./src`
  convenience alias, optional).

## Key UX flows

1. **Enter**: `AppShell` defaults to root (`?dir` absent). `useNodeChildren(null)`
   -> `listRootChildren`. Table renders.
2. **Open folder**: click a directory row -> push id onto `dirStack` ->
   `useNodeChildren(id)` -> `listChildren`.
3. **Breadcrumb**: directory stack drives segments; click a segment to jump to
   that ancestor. `getNodePath` shows the path string as the page title.
4. **New directory**: toolbar button -> name dialog -> `createDir` -> refresh.
5. **New file**: toolbar button -> name + optional mimeType dialog ->
   `createFile({ size: 4 MB, mimeType })` -> refresh. mimeType is not validated.
6. **Rename**: row action -> newName dialog ->
   `moveNode({ newParentId: current, newName })` -> refresh.
7. **Delete**: row action -> confirm -> `deleteNode` -> refresh.

## Error handling

- Transport throws `ContractTransportError` (`code`, `status`, `message`).
- Mutation `onError` -> heroui toast showing `code` + `message`.
- List query failure -> table area shows an error state with a retry button.

## Constants & validation rules

- `FILE_PLACEHOLDER_SIZE = 4 * 1024 * 1024` — the `size` sent on every
  `createFile`. Placeholder only; real size arrives later via upload.
- `mimeType`: optional, free-text, **not** format-validated on the client (sent
  as-is, may be `null`).
- `name`: required, non-empty (client trims and rejects empty before submit).

## Tradeoffs & risks

- **HeroUI v3 is new (released 2026-03; some docs still 404).** Mitigation: use
  the official CLI for setup; validate `init` runs cleanly in the monorepo
  sub-directory before building anything else. Manual CSS is the fallback.
- **pnpm hoist**: a missing hoist pattern makes heroui styles silently absent.
  Mitigation: `.npmrc` hoist + verify styles render after Phase B.
- **Hardcoded 4 MB size** is a placeholder semantic; the backend accepts any
  `size`, so no contract change is needed.
- **No client router**: the URL is fixed at `/` (admin tool, not shareable
  links); refreshing returns to root. This shrinks the dependency surface.
- **Conditional exports**: dev reads `packages/admin` source (HMR, no watch
  process); production reads its `dist`, so `packages/admin` is built with
  `tsup` before `@sonora/web-client` is built — same shape as
  `@sonora/contracts`. Build ordering = build admin first.

## Compatibility & rollback

- Pure additive: `apps/web/client` and `packages/admin` start empty; no backend
  changes, no edits to generated `@sonora/contracts` files.
- Rollback: delete `apps/web/client/src`, `packages/admin/src`, and revert
  `.npmrc`. No data migration, no impact on `task run:web-backend` / db tasks.
