## TODO: test tauri build

now add tauri setup, i want to test authelia oidc
1. if `isTauri` and no `sonoraAuthToken` in local storage, open `https://auth.onlyquant.top/api/oidc/authorization?client_id={$clientId}&redirect_uri={$redirectUri}&response_type=code&scope=openid%20profile%20email`
2. 

## TODO: Finish Phase 3

Status: 🚧 in progress

### Goal
Implement audio playback system with mini player and full player views for complete music playback experience.

### Tasks

#### Audio Engine
- [x] Install howler.js and types
  - `npm install howler`
  - `npm install -D @types/howler`

- [ ] Create `client/src/lib/AudioContext.tsx`
  - Global audio state provider
  - Current track state
  - Playback state (playing, paused, loading)
  - Queue state and management
  - No volume and mute state (User do not need to control volume, user just need to concern about system's volume)
  - Progress tracking

- [ ] Create `client/src/hooks/player/use-audio-player.ts`
  - Core audio playback hook using howler.js
  - Play, pause, seek
  - Track queue management (add, remove, reorder)
  - Auto-advance to next track
  - Shuffle and repeat logic (off/all/one)
  - Media Session API for system controls
  - Error handling and recovery
  - Preload next track for seamless playback

#### Playback Components
- [ ] Create `client/src/components/player/ProgressBar.tsx`
  - Seekable progress bar
  - Click to seek to position
  - Drag to scrub through track
  - Show buffered range
  - Display current time / total duration
  - Smooth transitions and hover effects

- [ ] Create `client/src/components/player/PlaybackControls.tsx`
  - Previous track button
  - Play/Pause toggle button
  - Next track button
  - Shuffle toggle button
  - Repeat mode button (off/all/one)
  - Disabled states when appropriate
  - Gradient hover effects

#### Mini Player
- [ ] Create `client/src/components/player/MiniPlayer.tsx`
  - Fixed position at bottom: `fixed bottom-0 left-0 right-0 h-20`
  - Backdrop blur: `backdrop-blur-lg bg-white/80 dark:bg-black/80`
  - **Left section**: Cover thumbnail (48x48), track title, artist name
  - **Center section**: PlaybackControls, ProgressBar, Time display
  - **Right section**: Queue button, Expand button
  - Smooth slide-up animation on track start
  - Click expand to open full player

#### Full Player
- [ ] Create `client/src/components/player/FullPlayer.tsx`
  - Modal/overlay covering entire screen
  - Large album art (70% of screen height)
  - Track metadata (title, artist, album)
  - Full playback controls
  - Large seekbar with timestamp
  - Queue/Lyrics toggle tabs
  - Close/minimize button
  - Animated background with album art colors
  - Swipe down to minimize (mobile gesture)

#### Queue Management
- [ ] Create `client/src/components/player/Queue.tsx`
  - Display current queue
  - Highlight currently playing track
  - Drag to reorder tracks
  - Remove track from queue
  - Clear queue button
  - Save queue as playlist (future)
  - Scrollable list

#### Integration
- [ ] Update `client/src/App.tsx`
  - Wrap app with AudioContext provider
  - Pass audio controls to all views
  
- [ ] Update `client/src/components/player/Layout.tsx`
  - Add MiniPlayer component at bottom
  - Reserve h-20 space for mini player
  - Add padding-bottom to main content

- [ ] Update `client/src/components/player/SongList.tsx`
  - Click track to play (use audio context)
  - Show playing indicator on current track
  - Play icon on hover

- [ ] Update `client/src/components/player/AlbumDetail.tsx`
  - Play All button functionality
  - Shuffle button functionality
  - Click track to play

- [ ] Update `client/src/components/player/SingerDetail.tsx`
  - Play All button functionality
  - Shuffle button functionality

#### Polish & Testing
- [ ] Add keyboard shortcuts
  - `Space` - Play/Pause
  - `←/→` - Seek backward/forward 5s
  - `Shift + ←/→` - Previous/Next track
  
- [ ] Error handling
  - Network errors during playback
  - Missing/invalid audio files
  - Show error toast messages
  
- [ ] Loading states
  - Buffering indicator
  - Skeleton loaders
  
- [ ] Smooth transitions
  - Fade in/out between tracks
  - Animated player entrance/exit
  - Progress bar animations

- [ ] Mobile responsive
  - Touch gestures for seek
  - Swipe controls
  - Adapted layout for small screens

- [ ] Test complete playback flow
  - Play from song list
  - Play from album detail
  - Queue management
  - Shuffle and repeat modes
  - Seek functionality

## TODO: Finish Phase 2

Status: ✅ completed

### Goal
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

### Goal
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

### Goal
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
