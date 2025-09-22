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

### Search Endpoints
- `/album?Action=SearchAlbum&query=<search_term>` - Search albums by name
- `/singer?Action=SearchSinger&query=<search_term>` - Search singers by name
- `/track?Action=SearchTrack&query=<search_term>` - Search tracks by name

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
   - Search<Name>.go - search by name using ILIKE, returns []<Name>View
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
- **upload.py** - Album upload functionality with API integration

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
- **upload** `<subcommand>` - Upload songs
   - `album <folder>`: upload album, create cover and year extracted from a song inside

### Album Upload Workflow (`upload album <folder>`)
1. **Scan folder** for supported audio files (.mp3, .ogg)
2. **Extract metadata** from all tracks using existing metadata extraction
3. **Extract and save album cover** to `cli/output/<ab>/<cd>/<album_name>.<ext>` (UUID directory structure)
4. **Create album record** via API with cover path and year from track metadata
5. **Process each track**:
   - Find or create singer using SearchSinger API (searches existing, creates if not found)
   - Move audio file to `cli/output/<ab>/<cd>/<filename>` (UUID directory structure)
   - Create track record with proper singer, album, and file path relationships

### API Integration
- **Base URL**: `http://localhost:8765/api` (configurable via `API_BASE_URL` env var)
- **Request Format**: `POST /<endpoint>?Action=<ActionName>` with JSON body
- **Response Format**: All API responses return data in `message` field
- **Endpoints Used**:
  - `/singer?Action=SearchSinger` - Search existing singers by name
  - `/singer?Action=CreateSinger` - Create new singer
  - `/album?Action=CreateAlbum` - Create new album with cover and year
  - `/track?Action=CreateTrack` - Create new track with relationships

### File Organization
- **Output Directory**: `cli/output/` with UUID-based subdirectory structure
- **Directory Pattern**: `<ab>/<cd>/` (first 4 characters of UUID)
- **Cover Art**: Saved as `<album_name>.<ext>` in output directory
- **Audio Files**: Moved to output directory preserving original filename
- **Path Storage**: Relative paths stored in database (e.g., `ab/cd/filename.ext`)

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
python app.py upload album /path/to/album/folder
```

### Example Album Upload
```bash
python app.py upload album input/spectrum/
# Output:
# Processing album folder: spectrum
# Found 2 audio file(s)
# Album: Spectrum (2019)
# Album cover saved with media UUID: 93710dae-9ebd-487f-bea2-f9194cf2f1c0
# Creating album record...
# Album created with ID: 11
# Processing track 1/2: Another Life
#   Singer: Westlife (ID: 15)
#   Audio file moved to: output/33/b3/Westlife-Another Life.mp3
#   Lyrics created with ID: 6
#   Track created with ID: 14
# Processing track 2/2: Better Man
#   Singer: Westlife (ID: 15)
#   Audio file moved to: output/ca/22/Westlife-Better Man.mp3
#   Lyrics created with ID: 7
#   Track created with ID: 15
# Album upload completed successfully!
```

### Database Testing
You can verify the upload results using psql:

```bash
# Check media records
psql postgresql://onlyquant:123456@localhost:5432/sonora -c "SELECT link, key FROM d_media ORDER BY created_at DESC LIMIT 1;"

# Check album covers using UUIDs
psql postgresql://onlyquant:123456@localhost:5432/sonora -c "SELECT name, cover FROM d_album WHERE id = 11;"

# Check track covers and lyrics
psql postgresql://onlyquant:123456@localhost:5432/sonora -c "SELECT name, cover, lyric FROM d_track WHERE album = 11;"

# Check lyrics content
psql postgresql://onlyquant:123456@localhost:5432/sonora -c "SELECT id, LENGTH(content) as lyric_length FROM d_lyric WHERE id IN (6, 7);"
```
