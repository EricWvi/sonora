`sonora` is a self-hosting music streaming service. It comprises two parts, admin management system and music player.

The backend is designed around models and actions on the models. Database tables are in `migration/migrations.go`. Models are defined in `model`.

The overall HTTP interface is through a `base` handler and using `Action` query and reflection to choose method on `base`. For each group of handlers, say `singer`, we will register it in `router.go`, using `GET` and `POST`. The corresponding handlers for actions of a specific model, say `singer`, are defined in `handler/<model>`, say `handler/singer`.

## Structure

config                              # Configuration management
├── config.go                       # Config loading and initialization
└── db.go                           # Database connection setup
handler                             # HTTP request handlers
├── album                           # Album-related handlers
│   ├── CreateAlbum.go              # Create new album
│   ├── DeleteAlbum.go              # Delete album by ID
│   ├── GetAlbum.go                 # Get single album by ID
│   ├── ListAlbums.go               # List albums with pagination
│   ├── ListAllAlbums.go            # List all albums without pagination
│   ├── SearchAlbum.go              # Search albums by name
│   ├── UpdateAlbum.go              # Update album by ID
│   └── base.go                     # Album handler base struct
├── handler.go                      # Common handler utilities
├── media                           # Media file management
│   ├── DeleteMedia.go              # Delete media file
│   ├── media.go                    # Media handler base
│   ├── serve.go                    # Serve media files
│   └── upload.go                   # Upload media files
├── ping.go                         # Health check endpoint
├── singer                          # Singer-related handlers
│   ├── CreateSinger.go             # Create new singer
│   ├── DeleteSinger.go             # Delete singer by ID
│   ├── GetSinger.go                # Get single singer by ID
│   ├── ListAllSingers.go           # List all singers without pagination
│   ├── ListSingers.go              # List singers with pagination
│   ├── SearchSinger.go             # Search singers by name
│   ├── UpdateSinger.go             # Update singer by ID
│   └── base.go                     # Singer handler base struct
└── track                           # Track-related handlers
    ├── CreateLyric.go              # Create lyrics for a track
    ├── CreateTrack.go              # Create new track
    ├── DeleteTrack.go              # Delete track by ID
    ├── GetLyric.go                 # Get lyrics for a track
    ├── GetTrack.go                 # Get single track by ID
    ├── ListAlbumTracks.go          # List tracks for an album
    ├── ListSingles.go              # List single tracks (no album)
    ├── ListTracks.go               # List tracks with pagination
    ├── SearchTrack.go              # Search tracks by name
    ├── UpdateLyric.go              # Update lyrics for a track
    ├── UpdateTrack.go              # Update track by ID
    └── base.go                     # Track handler base struct
log                                 # Logging utilities
└── logger.go                       # Logger setup and configuration
middleware                          # HTTP middleware
├── bodywriter.go                   # Response body writer for logging
├── idempotency.go                  # Idempotency key handling
├── jwt.go                          # JWT authentication
└── logging.go                      # Request/response logging
migration                           # Database migrations
├── helpers.go                      # Migration helper functions
├── migration.go                    # Migration runner
└── migrations.go                   # Database schema definitions
model                               # Data models
├── album.go                        # Album model and operations
├── lyric.go                        # Lyric model and operations
├── media.go                        # Media model and operations
├── model.go                        # Common model utilities
├── singer.go                       # Singer model and operations
├── track.go                        # Track model and operations
└── user.go                         # User model and operations
service                             # Business logic services
├── storage.go                      # File storage service
└── worker.go                       # Background worker service
tests                               # Test suites
├── api                             # API test helpers
│   ├── album.go                    # Album API test utilities
│   ├── media.go                    # Media API test utilities
│   └── track.go                    # Track API test utilities
├── integration                     # Integration tests
│   ├── album                       # Album integration tests
│   │   ├── DeleteAlbum_test.go     # Test album deletion
│   └── track                       # Track integration tests
│       └── DeleteTrack_test.go     # Test track deletion
└── tests.go                        # Test utilities and setup
config.yaml                         # Application configuration
main.go                             # Application entry point
router.go                           # Route definitions and setup

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
