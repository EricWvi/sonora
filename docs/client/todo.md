## TODO: Finish Phase 2

Status: ✅ completed

## Goal
Complete remaining Phase 2 tasks: singer views and search functionality to enable full library browsing.

### Completed

#### Fixed Navigation
- [x] Fixed sidebar navigation
  - Added `onNavigate` callback from App → Layout → Sidebar
  - Added `currentView` prop to highlight active items
  - Navigation properly updates view state

#### Singer Views
- [x] Create `client/src/components/player/SingerList.tsx`
  - Grid view of all singers (2-6 columns responsive)
  - Avatar display with gradient fallback
  - Show singer name and "Artist" label
  - Click navigates to singer detail
  - Hover effects matching album grid style

- [x] Create `client/src/components/player/SingerDetail.tsx`
  - Hero section with large avatar (256px)
  - Singer name with gradient text
  - "Singles" section: all singles displayed
  - "Albums" section: grid of first 10 albums
  - Play All and Shuffle buttons
  - Click album to navigate to album detail

#### Search Functionality
- [x] Create `client/src/components/player/SearchBar.tsx`
  - Debounced search input (300ms)
  - Search icon with glassmorphism styling
  - Clear button (X icon) when text entered
  - Keyboard shortcut: "/" to focus
  - Shows keyboard hint when empty

- [x] Create `client/src/components/player/SearchResults.tsx`
  - Grouped results: Artists, Albums, Songs sections
  - Shows top 5 results per category
  - Click to navigate to detail views
  - Shows "+N more" count for overflow
  - Empty state when no results
  - Search prompt when query is empty

- [x] Update `App.tsx` for search integration
  - Added searchQuery state
  - Shows SearchResults view when typing
  - Clears search on sidebar navigation
  - Search bar visible on all views

#### Integration
- [x] Updated sidebar navigation for singers
- [x] All views follow design system (gradients, glassmorphism, transitions)
- [x] Tested navigation flows (sidebar, search, clicks)
- [x] Mobile responsive (grid adapts, sidebar slides)

## TODO: Create song list view and album grid view

Status: ✅ completed

## Goal
Build core navigation views for browsing music library with Apple Music-inspired aesthetics.

### Completed
- [x] Created `client/src/components/player/SongList.tsx`
  - Table layout with columns: #, Title, Artist, Album, Duration
  - Play on click (integration ready)
  - Hover effects with smooth transitions
  - Responsive: Stack columns on mobile
  
- [x] Created `client/src/components/player/AlbumGrid.tsx`
  - Responsive grid (2-6 columns based on screen size)
  - Album cover images with aspect-ratio preservation
  - Show: cover, album name, year
  - Hover: Scale effect + shadow + overlay
  - Click navigates to album detail view
  
- [x] Created `client/src/components/player/AlbumDetail.tsx`
  - Hero section with large album cover
  - Display album metadata (name, year)
  - Tracklist with play buttons
  - "Play All" and "Shuffle" actions
  
- [x] Updated `App.tsx` routing/navigation
  - Added navigation between stats, song list, albums, and album detail
  - Implemented view switching logic with state management
  - Converted stat cards to clickable buttons
  - Added quick action buttons on home view
  
- [x] Applied design standards:
  - Vibrant gradients and glassmorphism
  - Smooth 200-300ms transitions
  - Dark mode: #1E1E1E background
  - Micro-interactions on hover
  - Loading states handled

- [x] Bug fixes:
  - Fixed album ID 0 query error by adding `enabled: id > 0` to useAlbum hook

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
