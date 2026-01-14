# Roadmap

## Vision
Build a beautiful, performant music player that rivals Apple Music in aesthetics and user experience. Support both web and mobile (Tauri Android) with full offline capabilities.

## Current Focus

**Next Immediate Steps:**
1. ✅ Create comprehensive Layout.tsx with Sidebar + MainContent + Player slots
2. ✅ Build Sidebar with navigation items for user-facing player
3. ✅ Implement client-side caching with Dexie.js
4. ✅ Create song list view components
5. ✅ Build album grid view
6. ✅ Create album detail view
7. Add mini player component for playback control
8. Implement playlist/queue management
9. Add keyboard shortcuts (Space = play/pause, arrow keys = seek)
10. Create full player view (expand from mini player)

**Success Criteria:**
- ✅ Data syncs from server to IndexedDB
- ✅ Layout with sidebar and main content area works
- ✅ Stats dashboard shows library overview
- ✅ User can browse list of songs
- ✅ User can browse album grid
- ✅ User can view album details with tracklist
- User can click a song to play (playback integration needed)
- Mini player shows currently playing track
- Play/pause/next/previous controls work
- User do not need to control volume (user just need to concern about system's volume)
- User can create and manage playlists
- Beautiful, responsive UI that works on mobile and desktop

## Design System

### Component Library
Build reusable components following Apple Music aesthetics:

**Buttons:**
- Primary: Vibrant color with shadow
- Secondary: Outlined, transparent background
- Icon: Circular, subtle hover effect
- Sizes: sm, md, lg

**Cards:**
- Album card: Square, rounded corners
- Track row: Hover highlight
- Singer card: Circular avatar

**Typography Scale:**
- Display: 48px / 3rem
- Heading 1: 32px / 2rem
- Heading 2: 24px / 1.5rem
- Heading 3: 20px / 1.25rem
- Body: 16px / 1rem
- Small: 14px / 0.875rem
- Tiny: 12px / 0.75rem

**Spacing:**
- Use Tailwind spacing scale (4px increments)
- Consistent padding: p-4, p-6, p-8
- Generous margins between sections

**Colors:**
- Extract dominant colors from album covers
- Use for backgrounds and accents
- Ensure WCAG AA contrast compliance

---

## Phase 1: Core Architecture & Foundation ✅

### 1.1 Project Setup
- [x] React + TypeScript with Vite
- [x] TanStack Query for API state management
- [x] Tailwind CSS with dark mode
- [x] Admin panel for content management

### 1.2 Data Layer
- [x] API hooks architecture (`use-singers.ts`, `use-albums.ts`, `use-tracks.ts`)
- [x] CRUD operations for all models
- [x] Query caching and invalidation

---

## Phase 2: Music Player UI ✅

### 2.1 Layout Structure ✅
**Components Created:**
- ✅ `Layout.tsx` - App shell with modern gradients, glassmorphism, and loading states
  - Vibrant gradient loading screen with animations
  - Pure white/dark mode (#1E1E1E) backgrounds
  - Error boundary with modern styling
  - Responsive container with generous spacing
  
- ✅ `Sidebar.tsx` - Glassmorphism navigation sidebar
  - Backdrop blur and translucent backgrounds
  - Gradient brand logo and tagline
  - Gradient active states with shadow effects
  - Mobile menu with smooth animations
  - Hover scale effects on navigation items
  
- ✅ `App.tsx` - Modern stats dashboard
  - Vibrant gradient stat cards (blue-cyan, purple-pink, pink-rose)
  - Hover animations with scale and shadow transitions
  - Large gradient heading typography
  - Staggered fade-in animations
  
- ✅ `index.css` - Custom animations and base styles
  - Fade-in and slide-in animations
  - Smooth scrolling behavior
  - Dark mode base styles
  - Font feature settings for better typography
  
- ✅ `localCache.ts` - Client-side caching with Dexie.js
  - Full sync and incremental sync support
  - Database schema with indexes for performance
  - Query functions for albums, singers, tracks, lyrics
  - Search capabilities across all entities

**Design Implementation:**
- ✅ Modern, immersive layout with edge-to-edge design
- ✅ Dynamic theming: Pure white (light) / #1E1E1E (dark)
- ✅ Glassmorphism: Backdrop blur and translucent overlays
- ✅ Vibrant gradients: Purple-pink-blue color scheme
- ✅ Smooth transitions: 200-300ms for all interactions
- ✅ Micro-interactions: Hover scales, animated icons
- ✅ Rich typography: Gradient headings, bold weights

**Technical Details:**
- ✅ Sidebar: `w-72 backdrop-blur-xl bg-white/95` on desktop, slide-in on mobile
- ✅ Main content: `max-w-7xl mx-auto` with generous padding
- ✅ IndexedDB caching with automatic sync on startup
- ✅ Custom CSS animations with stagger delays
- Reserve `h-20` at bottom for mini player (to be implemented)

### 2.2 Navigation Views ✅
**Song List View:**
- ✅ `SongList.tsx` - Display all tracks
  - Table layout with columns: #, Title, Artist, Album, Duration
  - Play on click (handler ready for integration)
  - Responsive grid that stacks columns on mobile
  - Hover effects with gradient backgrounds
  
**Album Grid View:**
- ✅ `AlbumGrid.tsx` - Grid of album covers
  - Responsive grid (2-6 columns based on screen size)
  - Cover image with fallback gradient
  - Hover effects: scale, shadow, overlay
  - Click navigates to album detail
  
**Album Detail View:**
- ✅ `AlbumDetail.tsx` - Full album view
  - Hero section with large cover (256px)
  - Album metadata display
  - Tracklist with track numbers and durations
  - "Play All" and "Shuffle" action buttons
  - Back navigation to album grid

**Navigation:**
- ✅ View state management in App.tsx
- ✅ Clickable stat cards for quick navigation
- ✅ Quick action buttons on home view
- ✅ Back buttons for all views
  
**Singer View:**
- ✅ `SingerList.tsx` - Grid of all singers
  - Responsive 2-6 column grid
  - Avatar display with fallback gradient icon
  - Click navigates to singer detail
  
- ✅ `SingerDetail.tsx` - Singer profile page
  - Hero section with large avatar
  - Singles collection (tracks without albums)
  - Albums grid showing first 10 albums
  - Play All and Shuffle buttons

**Playlist View:**
- [ ] `PlaylistList.tsx` - User playlists (future feature)

### 2.3 Search Functionality ✅
- ✅ `SearchBar.tsx` - Global search component
  - Debounced input (300ms)
  - Search across tracks, albums, singers
  - Keyboard shortcut: "/" to focus
  - Clear button
  
- ✅ `SearchResults.tsx` - Display search results
  - Grouped by type (Artists, Albums, Songs)
  - Shows top 5 results per category
  - Click to navigate to detail views
  - Empty state messaging

---

## Phase 3: Audio Player

### 3.1 Mini Player
- [ ] `MiniPlayer.tsx` - Bottom floating player
  - **Left section:** Cover thumbnail (48x48), track title, artist name
  - **Center section:** Play/Pause, Previous, Next, Progress bar, Time display
  - **Right section:** Volume control, Queue button, Expand button
  - Fixed position: `fixed bottom-0 left-0 right-0 h-20`
  - Backdrop blur effect: `backdrop-blur-lg bg-white/80 dark:bg-black/80`
  
### 3.2 Full Player
- [ ] `FullPlayer.tsx` - Expanded player view
  - Modal/overlay that covers entire screen
  - Large album art (70% of screen height)
  - Track metadata (title, artist, album)
  - Full playback controls
  - Seekbar with timestamp
  - Volume slider
  - Queue/lyrics toggle
  - Close/minimize button
  - Gesture support: swipe down to minimize (mobile)

### 3.3 Playback Controls
- [ ] `PlaybackControls.tsx` - Reusable control component
  - Previous track
  - Play/Pause toggle
  - Next track
  - Shuffle toggle
  - Repeat mode (off/all/one)
  
- [ ] `ProgressBar.tsx` - Seekable progress bar
  - Click to seek
  - Drag to scrub
  - Show buffered range
  - Display current time / total duration

### 3.4 Audio Engine
- [ ] `useAudioPlayer.ts` - Core audio playback hook
  - HTML5 Audio API integration
  - Play, pause, seek, volume control
  - Track queue management
  - Auto-advance to next track
  - Shuffle and repeat logic
  - Media session API for system controls
  - Error handling and recovery
  
- [ ] `AudioContext.tsx` - Global audio state provider
  - Current track
  - Playback state (playing, paused, loading)
  - Queue state
  - Volume and mute state
  - Progress tracking

---

## Phase 4: Advanced Features

### 4.1 Queue Management
- [ ] `Queue.tsx` - View and manage playback queue
  - Current track highlight
  - Drag to reorder
  - Remove from queue
  - Clear queue
  - Save queue as playlist
  
### 4.2 Lyrics Display
- [ ] `Lyrics.tsx` - Scrolling lyrics view
  - Fetch from GetLyric API
  - Time-synced highlighting (if LRC format)
  - Auto-scroll to current line
  - Click line to seek
  
### 4.3 Context Menus
- [ ] `ContextMenu.tsx` - Right-click menu component
  - Actions based on item type (track/album/singer)
  - Add to queue (Next/Last)
  - Add to playlist
  - Go to album/artist
  - Share link
  - Delete (admin mode)

### 4.4 Keyboard Shortcuts
- [ ] Implement global keyboard shortcuts
  - `Space` - Play/Pause
  - `←/→` - Previous/Next track
  - `↑/↓` - Volume up/down
  - `M` - Mute toggle
  - `S` - Shuffle toggle
  - `R` - Repeat cycle
  - `/` - Focus search
  - `F` - Toggle fullscreen player
  - `Esc` - Close modals/fullscreen

---

## Phase 5: Caching & Offline Support

### 5.1 Data Caching (Web)
- [ ] Configure TanStack Query cache
  - `gcTime`: 1 hour for track/album/singer lists
  - `staleTime`: 30 minutes
  - Persist cache to localStorage/IndexedDB
  
- [ ] `useCacheManager.ts` - Cache management hook
  - Prefetch albums when viewing singer
  - Prefetch tracks when viewing album
  - Background refresh strategy

### 5.2 Media Caching (Web)
- [ ] `useMediaCache.ts` - Cache audio files and covers
  - IndexedDB for storing binary data
  - Service Worker for offline playback
  - Cache size limits and eviction policy (LRU)
  - Download progress tracking
  
- [ ] `DownloadManager.tsx` - UI for managing downloads
  - Download album/track for offline
  - View cached items
  - Manage storage usage
  - Clear cache options

### 5.3 Offline Mode Indicator
- [ ] `OfflineIndicator.tsx` - Show online/offline status
  - Badge in UI when offline
  - Filter to show only cached content when offline
  - Queue sync when back online

### 5.4 Service Worker
- [ ] Create service worker for PWA
  - Cache app shell (HTML, CSS, JS)
  - Cache media files on-demand
  - Offline fallback page
  - Background sync for favorites/playlists

---

## Phase 6: Mobile (Tauri Android)

### 6.1 Tauri Setup
- [ ] Initialize Tauri project
  - Android build configuration
  - Permissions (storage, network, wake lock)
  - App icons and splash screen
  
### 6.2 Mobile-Specific Features
- [ ] Touch gesture support
  - Swipe to navigate
  - Pull to refresh
  - Long press for context menu
  
- [ ] Native integrations
  - Android media session
  - Lock screen controls
  - Notification controls
  - Background playback
  - Audio focus handling
  
- [ ] Mobile file storage
  - Save to app's private storage
  - Manage storage permissions
  - Download manager integration

### 6.3 Responsive Optimizations
- [ ] Mobile-first components
  - Bottom sheet modals
  - Mobile navigation (tabs instead of sidebar)
  - Touch-friendly button sizes (min 44x44px)
  - Optimized grid layouts for small screens

---

## Phase 7: UI/UX Polish

### 7.1 Animations & Transitions
- [ ] Smooth page transitions
  - Fade in/out
  - Slide animations
  - Shared element transitions (album cover)
  
- [ ] Micro-interactions
  - Button hover effects
  - Play button pulse
  - Loading skeletons
  - Progress bar animations

### 7.2 Apple Music-Inspired Design
**Color Palette:**
- Use album cover colors for dynamic theming
- Vibrant gradient backgrounds in player
- Glassmorphism effects (backdrop blur)

**Typography:**
- SF Pro Display for headings (fallback to system fonts)
- SF Pro Text for body (fallback to system fonts)
- Font weights: 300 (Light), 400 (Regular), 600 (Semibold), 700 (Bold)

**Visual Hierarchy:**
- Large, bold album titles
- Subtle, smaller artist names
- Generous whitespace
- Card-based layouts with subtle shadows

**Components to Match:**
- [ ] `AlbumCard.tsx` - Apple Music style album cards
  - Rounded corners (`rounded-lg`)
  - Shadow on hover
  - Smooth scale animation
  
- [ ] `NowPlayingBar.tsx` - Bottom bar with blur effect
- [ ] `HeroSection.tsx` - Large header for album/playlist views

### 7.3 Dark Mode Excellence
- [ ] Perfect dark mode support
  - OLED-friendly pure black option
  - Reduce brightness of images at night
  - Smooth theme transitions
  - System preference detection
  - User preference persistence

### 7.4 Accessibility
- [ ] Keyboard navigation
- [ ] ARIA labels
- [ ] Focus indicators
- [ ] Screen reader support
- [ ] High contrast mode

---

## Phase 8: Performance Optimization

### 8.1 Virtualization
- [ ] Implement virtual scrolling for long lists
  - `react-window` or `@tanstack/react-virtual`
  - Render only visible items
  - 1000+ tracks should scroll smoothly

### 8.2 Image Optimization
- [ ] Lazy loading images
  - Intersection Observer
  - Blur placeholder while loading
  - Progressive image loading
  
- [ ] Responsive images
  - Multiple sizes for album covers (thumbnail, medium, large)
  - WebP format support
  - Fallback to JPEG

### 8.3 Code Splitting
- [ ] Route-based code splitting
  - Lazy load views: Albums, Singers, Playlists
  - Separate chunks for admin vs player
  
- [ ] Component lazy loading
  - Load FullPlayer only when needed
  - Load Lyrics on demand

### 8.4 Bundle Optimization
- [ ] Tree shaking
- [ ] Minification
- [ ] Compression (Gzip/Brotli)
- [ ] Analyze bundle size
- [ ] Remove unused dependencies

---

## Phase 9: Testing & Quality

### 9.1 Testing Strategy
- [ ] Unit tests for hooks (`vitest`)
- [ ] Component tests (`@testing-library/react`)
- [ ] E2E tests (`playwright`)
- [ ] Visual regression tests

### 9.2 Error Handling
- [ ] Error boundaries
- [ ] Retry logic for failed API calls
- [ ] User-friendly error messages
- [ ] Fallback UI for missing media

### 9.3 Monitoring
- [ ] Performance monitoring
- [ ] Error tracking
- [ ] Usage analytics (privacy-respecting)

---

## Phase 10: Future Enhancements

### 10.1 Playlists
- [ ] Create/edit/delete playlists
- [ ] Add tracks to playlists
- [ ] Collaborative playlists
- [ ] Smart playlists (auto-generated based on criteria)

### 10.2 Social Features
- [ ] User accounts
- [ ] Share tracks/albums/playlists
- [ ] Follow other users
- [ ] Activity feed

### 10.3 Advanced Audio
- [ ] Equalizer
- [ ] Crossfade between tracks
- [ ] Gapless playback
- [ ] ReplayGain normalization
- [ ] Spatial audio

### 10.4 Discovery
- [ ] Recently played
- [ ] Most played
- [ ] Recommended tracks (based on listening history)
- [ ] Genre exploration
- [ ] Similar artists

### 10.5 Sync Across Devices
- [ ] Sync queue across devices
- [ ] Continue playback on another device
- [ ] Sync favorites and playlists

---

## Technical Debt & Refactoring

### Ongoing Improvements
- [ ] Replace placeholder App.tsx with real player layout
- [ ] Consistent error handling patterns
- [ ] API hook error types
- [ ] Storybook for component documentation
- [ ] Comprehensive TypeScript types for all API responses
- [ ] i18n setup for internationalization
- [ ] Theme customization system


