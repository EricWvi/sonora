## Client
Client side code are in `client`. `index.html` is for music player and `admin.html` is for admin management.

### Architecture
- **React + TypeScript** with Vite build system
- **TanStack Query** for API state management and caching
- **Tailwind CSS** for styling with dark mode support
- **API hooks** in `client/src/hooks/` - one file per model (`use-singers.ts`, `use-albums.ts`, `use-tracks.ts`)

### API Hooks Pattern
Each model follows identical hook structure:
- `use<Model>()` - fetch all records (uses ListAll<Model> endpoint)
- `use<Model>(id)` - fetch single record (uses Get<Model> endpoint)
- `useCreate<Model>()` - create mutation
- `useUpdate<Model>()` - update mutation
- `useDelete<Model>()` - delete mutation
- `list<Model>(page, conditions)` - paginated fetch function (uses List<Model> endpoint)

### Admin
#### Admin Components
Located in `client/src/components/admin/`:
- **singer.tsx** - Singer management with CRUD operations
- **album.tsx** - Album management with CRUD operations
- **track.tsx** - Track management with CRUD operations

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
