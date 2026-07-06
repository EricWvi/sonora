# Implement — VFS admin panel

Execution plan for the design in `design.md`. Validate at each gate before
continuing.

## Package names

- Host app (`apps/web/client`): `@sonora/web-client`.
- Feature package (`packages/admin`): `@sonora/admin`.

## Ordered checklist

### Phase A — workspace & dependencies
1. Add repo-root `.npmrc` heroui/framer-motion hoist patterns (see design).
2. Create `apps/web/client/package.json` (`@sonora/web-client`): deps
   `react`, `react-dom`, `@heroui/react`, `framer-motion`, `@tanstack/react-query`,
   `lucide-react`, `@sonora/contracts` (workspace),
   `@sonora/admin` (workspace); devDeps `vite`, `@vitejs/plugin-react-swc`,
   `tailwindcss`, `@tailwindcss/vite`, `typescript`, `@types/react`,
   `@types/react-dom`; scripts `dev`/`build`/`typecheck`.
3. Create `packages/admin/package.json` (`@sonora/admin`): deps `@sonora/contracts`
   (workspace), `@tanstack/react-query`, `@heroui/react`, `framer-motion`,
   `lucide-react`; peerDeps `react`, `react-dom`; devDeps `tsup`, `typescript`;
   scripts `build: tsup`, `dev: tsup --watch`; **conditional `exports`**
   (`development` -> `./src/index.ts`, `types` -> `./dist/index.d.ts`,
   `import` -> `./dist/index.js`) — same pattern as `@sonora/contracts`.
4. `pnpm install` (from repo root).

### Phase B — host scaffold (`apps/web/client`)
5. Create `index.html`, `tsconfig.json` (`customConditions: ["development"]`,
   no `paths`), and `vite.config.ts` (proxy `/api` -> `localhost:8080`,
   `server.fs.allow: [".."]`, watch symlinked `@sonora/*` packages; **no**
   workspace `resolve.alias` — rely on conditional `exports`).
6. Run HeroUI CLI: `npx heroui@latest init` inside `apps/web/client`, then
   `npx heroui@latest add table button modal input ...` for the components used.
   **Gate:** confirm `hero.ts` + Tailwind v4 CSS wiring exist.
7. Write `src/index.css` (tailwind import + heroui plugin, as CLI emits).
8. Write `src/main.tsx` (assemble client, wrap in providers + `<AppShell>`).
   **Gate (smoke):** `pnpm dev` renders a heroui-styled placeholder before any
   feature work — styles must be visibly applied.

### Phase C — feature package (`packages/admin`)
9. `src/providers/client-context.tsx` (`ClientProvider`, `useClient`).
10. `src/lib/error-toast.ts` (`ContractTransportError` -> toast).
11. `src/features/nodes/constants.ts` (`FILE_PLACEHOLDER_SIZE`).
12. `src/features/nodes/hooks/use-node-children.ts`, `use-node-path.ts`,
    `use-node-mutations.ts`.
13. `src/features/nodes/components/`: `node-table.tsx`, `node-breadcrumb.tsx`,
    `create-dir-dialog.tsx`, `create-file-dialog.tsx`, `rename-dialog.tsx`,
    `delete-confirm-dialog.tsx`.
14. `src/features/nodes/node-explorer.tsx` (compose breadcrumb + toolbar + table).
15. `src/app-shell.tsx` (provider stack + `<NodeExplorer>`; no router — URL fixed
    at `/`, navigation via `dirStack` state in `NodeExplorer`).
16. `src/index.ts` exporting `AppShell`.
    **Gate:** `pnpm --filter @sonora/admin exec tsc --noEmit` passes.

### Phase D — wire & validate
17. `pnpm --filter @sonora/admin build` (tsup) produces `dist/` (needed for the
    client production build via the `import` export condition).
18. `pnpm --filter @sonora/web-client exec tsc --noEmit` passes.
19. `pnpm --filter @sonora/web-client build` (vite build) passes.
20. Manual end-to-end: `task db:up`, `task run:web-backend`, `task run:admin`.
    Verify each acceptance criterion in `prd.md` (list root, navigate, create
    dir, create file with size=4MB + lenient mimeType, rename, delete, error
    toast on failure).

## Validation commands

- `pnpm install`
- `pnpm --filter @sonora/admin exec tsc --noEmit`
- `pnpm --filter @sonora/admin build` (tsup -> dist)
- `pnpm --filter @sonora/web-client exec tsc --noEmit`
- `pnpm --filter @sonora/web-client build`
- `task run:admin` (manual browser check)

No Rust changes, so `task lint` / `task test` are unaffected and need not block
this task.

## Risky files & rollback points

- **`.npmrc`** — wrong hoist => heroui styles silently missing. Verify after
  Phase B smoke gate.
- **`vite.config.ts` aliases** — wrong path => import resolution failure.
- **`hero.ts` / `index.css`** — wrong Tailwind v4 plugin wiring => unstyled app.
- Rollback points: Phase B gate (styles render), Phase C gate (types compile),
  Phase D (full flow). Each phase is independently verifiable.

## Follow-up checks before `task.py start`

- Confirm `pnpm-workspace.yaml` includes `apps/**` (it does) and that
  `@sonora/web-client` / `@sonora/admin` resolve as workspace packages after
  `pnpm install`.
- Confirm `@sonora/contracts/fetch` source is reachable via the dev alias
  (`packages/contracts/src/fetch.ts`).
- Confirm the HeroUI CLI emits a working Tailwind v4 setup in a workspace
  sub-directory; if not, switch to the manual CSS fallback (design.md).
