`sonora` is a self-hosting music streaming service. It comprises two parts, admin management system and music player.

The backend is designed around models and actions on the models. Database tables are in `migration/migrations.go`. Models are defined in `model`.

The overall HTTP interface is through a `base` handler and using `Action` query and reflection to choose method on `base`. For each group of handlers, say `singer`, we will register it in `router.go`, using `GET` and `POST`. The corresponding handlers for actions of a specific model, say `singer`, are defined in `handler/<model>`, say `handler/singer`.

The current development is in admin part.

## Database Schema (from migration/migrations.go)

- **d_singer**: id, name, avatar, timestamps
- **d_album**: id, name, cover, timestamps
- **d_track**: id, name, singer, album, cover, url, lyric, duration, timestamps
- **d_user**: id, email, avatar, username, language, timestamps
- **d_media**: id, link, key, timestamps

## Implemented Models & Handlers

- **Singer**: Complete CRUD (Create, List, Update, Delete)
- **Album**: Complete CRUD
- **Track**: Complete CRUD (isolated, no relations)

## Router Endpoints

- `/singer` (GET/POST) → singer.DefaultHandler
- `/album` (GET/POST) → album.DefaultHandler
- `/track` (GET/POST) → track.DefaultHandler

## Implementation Pattern

Each model follows identical structure:

1. Model in `model/<name>.go` with:
   - TableName(), Get(), Create(), Update(), Delete()
   - ListAll<Name>() - returns all <Name>View records without pagination (includes id, omits timestamps)
   - List<Name>() - returns paginated <Name>View results with hasMore flag (includes id, omits timestamps)
   - <name>PageSize constant for pagination
   - <Name>View struct for API responses (includes id, excludes gorm.Model timestamps)
   - <Name>Field struct for data fields only (excludes id and timestamps)
2. Handlers in `handler/<name>/` with:
   - base.go, Create*.go, Get*.go, Update*.go, Delete*.go
   - List<Name>.go - paginated listing with Page parameter, returns []<Name>View
   - ListAll<Name>.go - non-paginated listing, returns []<Name>View
   - Get<Name>.go - single record retrieval by ID, returns <Name>View
3. Routes registered in router.go with import and GET/POST endpoints
4. The database columns use snake_case while the JSON API uses camelCase.

**Note**: All list operations return `<Model>View` structs that include the id field for frontend operations, while excluding database timestamps and internal fields from API responses.

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

#### UI Design
- **Compact styling**: Small fonts, reduced spacing, minimal padding
- **Consistent sizing**: All buttons use `text-xs px-2 py-1` for uniformity
- **Dark mode support**: Full light/dark theme switching
- **Visual feedback**: Loading states, error handling, pending actions

## CLI
Python CLI tool for audio file operations located in `cli/`.

### Structure
- **app.py** - Main CLI application with argparse interface
- **audio_metadata.py** - Audio metadata extraction module with AudioMetadata dataclass
- **config.py** - Configuration for database, paths, and application settings

### AudioMetadata Class
```python
@dataclass
class AudioMetadata:
    title: str
    artists: List[str]          # List of artist names
    album: str
    genre: str
    track_number: str
    duration_seconds: Optional[float]  # Duration as number
    date: str
    file_path: str
    file_size: int
    format: str

    @property
    def artist(self) -> str:    # Artists joined with "; " for display

    @property
    def duration(self) -> str:  # Formatted MM:SS duration
```

### Commands
- **show** `<file_path>` - Display metadata for a single audio file
- **get-art** `<file_path>` - Extract cover art from audio file and save to same directory

### Supported Formats
- **MP3** (ID3 tags) - TPE1, TIT2, TALB, TCON, TDRC, TRCK, APIC (cover art)
- **OGG** (Vorbis comments) - artist, title, album, genre, date, tracknumber, metadata_block_picture (cover art)

### Multiple Artists Support
Handles various artist delimiters and joins with `; `:
- **Null bytes** (`\x00`) - Primary format in ID3 tags
- **Semicolons** (`;`) - Preserved as-is
- **Slashes** (`/`) - Split and joined
- **Commas** (`,`) - Split and joined

### Cover Art Extraction
The `extract_cover_art()` utility function extracts embedded album art:
- **MP3**: Extracts APIC frames from ID3 tags
- **OGG**: Parses FLAC picture blocks from metadata_block_picture (base64-encoded)
- **Formats**: Supports JPEG, PNG, and WebP image formats
- **Output**: Saves to same directory as audio file with appropriate extension

### Usage
```bash
python app.py show music.mp3
python app.py show song.ogg
python app.py get-art music.mp3      # Saves as music.jpg
python app.py get-art song.ogg       # Saves as song.webp (or .jpg/.png)
```
