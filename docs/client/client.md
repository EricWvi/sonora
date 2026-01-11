## Client
Client side code are in `client`. `index.html` is for main site music player and `admin.html` is for admin management.

### Structure

src                                 # Source code directory
├── AdminBoard.tsx                  # Main admin dashboard component
├── App.tsx                         # Main music player app component
├── admin.css                       # Admin-specific styles
├── admin.tsx                       # Admin app entry point
├── components                      # Reusable UI components
│   ├── admin                       # Admin panel components
│   │   ├── album.tsx               # Album CRUD management UI
│   │   ├── singer.tsx              # Singer CRUD management UI
│   │   ├── track.tsx               # Track CRUD management UI
│   │   └── MiniPlayer.tsx          # Mini player at bottom-right
│   └── player                      # Main site components
│       ├── Layout.tsx              # Main layout with sidebar and content area
│       └── Sidebar.tsx             # Responsive navigation sidebar
├── hooks                           # Custom React hooks
│   ├── admin                       # Admin hooks (server-fetched data)
│   │   ├── use-albums.ts           # Album data fetching and mutations
│   │   ├── use-miniplayer.ts       # Audio player state management for admin panel
│   │   ├── use-singers.ts          # Singer data fetching and mutations
│   │   └── use-tracks.ts           # Track data fetching and mutations
│   └── player                      # Player hooks (cached data)
│       ├── use-albums.ts           # Album hooks using local cache
│       ├── use-singers.ts          # Singer hooks using local cache
│       └── use-tracks.ts           # Track hooks using local cache
├── index.css                       # Global styles and Tailwind directives
├── lib                             # Utility libraries
│   ├── fileUpload.ts               # File upload utilities
│   ├── localCache.ts               # Dexie database and sync manager
│   ├── queryClient.ts              # TanStack Query configuration
│   └── utils.ts                    # Common utility functions (cn, etc.)
├── main.tsx                        # Player app entry point
└── vite-env.d.ts                   # Vite TypeScript declarations
admin.html                          # Admin panel HTML entry
eslint.config.js                    # ESLint configuration
index.html                          # Music player HTML entry
vite.config.ts                      # Vite build configuration


### Architecture
- **React + TypeScript** with Vite build system
- **TanStack Query** for API state management and caching
- **Dexie.js** for IndexedDB-based client-side caching (main site)
- **Tailwind CSS** for styling with dark mode support
- **Dual data strategy**:
  - Admin hooks: Direct server queries with cache invalidation
  - Player hooks: Local IndexedDB cache with sync manager

### Client-Side Caching (Main Site)
The main site uses Dexie.js for offline-capable data storage:
- **Database**: SonoraDB with tables for albums, singers, tracks, lyrics, and sync metadata
- **Sync Strategy**:
  - Full sync on first load or if last sync was 28+ days ago
  - Incremental sync for regular updates (fetches only changed records)
  - Sync happens on app initialization, non-blocking for incremental updates
- **Indexes**: Optimized for album lookup, genre filtering, and name-based search
- **Query Client**: Separate from admin panel's queryClient, uses dbClient for data access

### API Hooks Pattern
Each model follows identical hook structure:
- `use<Model>()` - fetch all records (uses ListAll<Model> endpoint)
- `use<Model>(id)` - fetch single record (uses Get<Model> endpoint)
- `useCreate<Model>()` - create mutation
- `useUpdate<Model>()` - update mutation
- `useDelete<Model>()` - delete mutation
- `list<Model>(page, conditions)` - paginated fetch function (uses List<Model> endpoint)

### Main Site
#### Features

#### UI Design
- **Modern, immersive layout**: Edge-to-edge album art, glassmorphism overlays, and smooth transitions for a premium feel.
- **Dynamic theming**: Auto light/dark mode, main Body: pure white (light mode) or #1E1E1E (dark mode). Use vibrant accent colors based on album art, and animated backgrounds in the full size player.
- **Responsive grid**: Adaptive album/song grid with hover effects, optimized for desktop and mobile.
- **Sticky player**: Persistent bottom music bar with queue, and mini-player popout.
- **Rich search**: Instant, fuzzy search and category tabs (songs, albums, artists).
- **Detail pages**: Full-bleed album/artist pages with hero images, tracklists, and related content.
- **Microinteractions**: Animated play/pause, loading skeletons, and subtle feedback for all actions.


### Admin
#### Admin Features
- **CRUD Operations**: Create, Read, Update, Delete for all models
- **Search**: Client-side filtering by model name (triggered on Enter key)
- **Pagination**: Client-side pagination (10 items per page)
- **Compact Layout**: Optimized for PC screens to fit 10 entries without scrolling
- **Edit in Place**: Click Edit to populate form with existing data
- **Responsive Forms**: Grid layouts that adapt to screen size
- **Delete Confirmation**: Named confirmation dialogs (e.g., "Delete singer 'John Doe'?")
- **Track Form Fields**: Includes name, singer, year, genre, album_text, cover image, audio file, and lyrics

#### UI Design
- **Compact styling**: Small fonts, reduced spacing, minimal padding
- **Consistent sizing**: All buttons use `text-xs px-2 py-1` for uniformity
- **Dark mode support**: Full light/dark theme switching
- **Visual feedback**: Loading states, error handling, pending actions
