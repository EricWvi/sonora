`sonora` is a self-hosting music streaming service. It comprises two parts, admin management system and music player.

The backend is designed around models and actions on the models. Database tables are in `migration/migrations.go`. Models are defined in `model`.

The overall HTTP interface is through a `base` handler and using `Action` query and reflection to choose method on `base`. For each group of handlers, say `singer`, we will register it in `router.go`, using `GET` and `POST`. The corresponding handlers for actions of a specific model, say `singer`, are defined in `handler/<model>`, say `handler/singer`.

The current development is in admin part.

## Database Schema (from migration/migrations.go)

- **d_singer**: id, name, avatar, timestamps
- **d_album**: id, name, cover, year, timestamps
- **d_track**: id, name, singer, album, cover, url, lyric, duration, year, track_number, genre, album_text, timestamps
- **d_user**: id, email, avatar, username, language, timestamps
- **d_media**: id, link (UUID), key, timestamps
- **d_lyric**: id, content, timestamps

notes:
- **Database Triggers**: Automatically updates `album_text` or `cover` in tracks when album name or cover changes
- **Performance Indexes**:
  - `idx_media_link` - Index on d_media.link for UUID lookups
  - `idx_track_singer_trgm` - Trigram GIN index on d_track.singer for fuzzy search
  - `idx_track_album` - Index on d_track.album for joins/filtering
  - `idx_track_name_trgm`, `idx_singer_name_trgm`, `idx_album_name_trgm` - Trigram indexes for search
  - **Note**: Genre index was removed (v0.9.0) due to low cardinality - not efficient for filtering

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
