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

## Coding Guidelines

1. use `model.WhereMap` instead of `gin.H`, and name variables of `model.WhereMap` and `model.WhereExpr` simply as `m`.

## Testing Considerations

When writing tests for handlers, be aware of the following requirements:

### 1. Init Stage
- Tests must call `config.Init()` before using database or storage services
- Tests should use `config.TestRouter()` to get router.

### 2. Action Middleware for Tests
- **Issue**: The handler dispatch system requires the `Action` parameter to be set in the gin context via `c.Set("Action", action)`
- **Production**: The `middleware.Logging()` middleware extracts Action from query parameters and sets it
- **Test Solution**: Add a simple middleware to test routers:
```go
router.Use(func(c *gin.Context) {
    action := c.Request.URL.Query().Get("Action")
    if action != "" {
        c.Set("Action", action)
    }
    c.Next()
})
```

### 3. Test Structure for Operations
- always create new records to test update/delete

### 4. Database Access in Tests
- Use `config.DB()` to get the database instance in tests
- The database is shared across test runs - ensure proper cleanup or use unique test data
- Tests use soft deletes (GORM), so deleted records have `deleted_at` timestamp set
