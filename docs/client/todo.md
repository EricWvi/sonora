## TODO: Polish Styles for Phase 2

Status: ✅ completed

## Goal
Previously the main site was developed under wrong style guidelines. Read `docs/client/client.md` Main site's UI Design, and polish styles to meet the need of Phase 2.

### Completed
- [x] Updated Layout.tsx with modern gradients and dark mode support
- [x] Enhanced Sidebar.tsx with glassmorphism and smooth transitions
- [x] Improved App.tsx with vibrant gradient cards and animations
- [x] Added custom CSS animations (fade-in, slide-in)
- [x] Implemented proper dark mode theming (#1E1E1E background)
- [x] Applied micro-interactions and hover effects throughout


## TODO: Client Side Caching

Status: ✅ completed

## Goal
- Use `Dexie.js` to Manage cache for the main site. 
- Do not interfere with admin panel's queryClient. Define the main site's query client in `client/src/lib/localCache.ts`. The hooks in `client/src/hooks/player/` do not fetch data from server, instead it fetch data from `dbClient`. It is the `dbClient`'s responsibility to read data from indexed db and update cache during start.
- The main site can access album/singer/track/media in read-only way.
- When the main site starts up, it will check for local cache, use `/api/sync?Action=GetFullSync` if local cache does not exist or last sync happened 28 days ago, and use `/api/sync?Action=GetUpdates&since=<unixTimestamp>` otherwise.
- The main site can only do CRUD operations on playlist for now. The site cache all records and check for update on every start.
- create some indexes to support efficient operations:
    - look up for tracks of a specific album (on `d_track`'s `album`)
    - read `migration/migrations.go`, and build some necessary indexes
- Support efficient search for album's name, singer's name, track's name, track's singer.

### Plan
- [x] Create `client/src/lib/localCache.ts`
- [x] Create `client/src/hooks/player/use-albums.ts`
- [x] Create `client/src/hooks/player/use-singers.ts`
- [x] Create `client/src/hooks/player/use-tracks.ts`


## TODO: Music Player Layout Structure

Status: ✅ completed

## Goal
The first start will load full data from server, display a simple loading animation (We will refine it later).

Normal start will fetch updates not blocking the ui.

Focus on Phase 2 implementation starting with Layout.tsx, Sidebar.tsx, and basic song list view.

### Plan

- [x] Create `client/src/components/player/Layout.tsx`
- [x] Create `client/src/components/player/Sidebar.tsx`
