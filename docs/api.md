## API Doc

### Common Response Format

All API responses follow this format:

```json
{
    "requestId": "uuid",
    "code": 200,
    "message": { ... }
}
```

Error responses have a `message` field as a string.

### Health Check

#### Ping
Health check endpoint

GET /api/ping

```json
{"message": "pong"}
```

---

### Album

#### CreateAlbum
Create a new album

POST /api/album?Action=CreateAlbum

Request Body:
```json
{
    "name": "Album Name",
    "cover": "uuid-of-cover-media",
    "year": 2023
}
```

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "id": 1
    }
}
```

#### GetAlbum
Get album by ID

GET /api/album?Action=GetAlbum&id=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "album": {
            "id": 1,
            "name": "Album Name",
            "cover": "uuid",
            "year": 2023
        }
    }
}
```

#### UpdateAlbum
Update an existing album (partial model)

POST /api/album?Action=UpdateAlbum

Request Body:
```json
{
    "id": 1,
    "name": "Updated Name"
}
```

#### DeleteAlbum
Delete an album

POST /api/album?Action=DeleteAlbum

Request Body:
```json
{
    "id": 1
}
```

#### ListAlbums
List albums with pagination (15 per page)

GET /api/album?Action=ListAlbums&page=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "albums": [...],
        "hasMore": true
    }
}
```

#### ListAllAlbums
List all albums (no pagination)

GET /api/album?Action=ListAllAlbums

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "albums": [
            {
                "id": 3,
                "name": "Spectrum",
                "cover": "ae84e80c-4d75-480c-8950-0a9e932ef332",
                "year": 2023
            }
        ]
    }
}
```

#### SearchAlbum
Search albums by name (case-insensitive partial match)

GET /api/album?Action=SearchAlbum&query=rock

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "albums": [...]
    }
}
```

---

### Singer

#### CreateSinger
Create a new singer

POST /api/singer?Action=CreateSinger

Request Body:
```json
{
    "name": "Singer Name",
    "avatar": "uuid-of-avatar-media"
}
```

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "id": 1
    }
}
```

#### GetSinger
Get singer by ID

GET /api/singer?Action=GetSinger&id=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "singer": {
            "id": 1,
            "name": "Singer Name",
            "avatar": "uuid"
        }
    }
}
```

#### UpdateSinger
Update an existing singer (partial model)

POST /api/singer?Action=UpdateSinger

Request Body:
```json
{
    "id": 1,
    "name": "Updated Name"
}
```

#### DeleteSinger
Delete a singer

POST /api/singer?Action=DeleteSinger

Request Body:
```json
{
    "id": 1
}
```

#### ListSingers
List singers with pagination (15 per page)

GET /api/singer?Action=ListSingers&page=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "singers": [...],
        "hasMore": true
    }
}
```

#### ListAllSingers
List all singers (no pagination)

GET /api/singer?Action=ListAllSingers

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "singers": [...]
    }
}
```

#### SearchSinger
Search singers by name (case-insensitive partial match)

GET /api/singer?Action=SearchSinger&query=taylor

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "singers": [...]
    }
}
```

---

### Track

#### CreateTrack
Create a new track

POST /api/track?Action=CreateTrack

Request Body:
```json
{
    "name": "Track Name",
    "singer": "singer-id-or-name",
    "album": 0,
    "cover": "uuid",
    "url": "uuid",
    "lyric": 0,
    "duration": 180,
    "year": 2023,
    "trackNumber": 1,
    "genre": "Pop",
    "albumText": "Album Name"
}
```

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "id": 1
    }
}
```

#### GetTrack
Get track by ID

GET /api/track?Action=GetTrack&id=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "track": {
            "id": 1,
            "name": "Track Name",
            "singer": "Singer",
            "album": 0,
            "cover": "uuid",
            "url": "uuid",
            "lyric": 0,
            "duration": 180,
            "year": 2023,
            "trackNumber": 1,
            "genre": "Pop",
            "albumText": "Album Name"
        }
    }
}
```

#### UpdateTrack
Update an existing track (partial model)

POST /api/track?Action=UpdateTrack

Request Body:
```json
{
    "id": 1,
    "name": "Updated Name"
}
```

#### DeleteTrack
Delete a track

POST /api/track?Action=DeleteTrack

Request Body:
```json
{
    "id": 1
}
```

#### ListTracks
List tracks with pagination (15 per page)

GET /api/track?Action=ListTracks&page=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "tracks": [...],
        "hasMore": true
    }
}
```

#### SearchTrack
Search tracks by name (case-insensitive partial match)

GET /api/track?Action=SearchTrack&query=love

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "tracks": [...]
    }
}
```

#### ListAlbumTracks
List all tracks for a specific album

GET /api/track?Action=ListAlbumTracks&albumId=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "tracks": [...]
    }
}
```

#### ListSingles
List all singles (tracks where album=0)

GET /api/track?Action=ListSingles

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "tracks": [...]
    }
}
```

---

### Lyric

#### CreateLyric
Create a new lyric

POST /api/track?Action=CreateLyric

Request Body:
```json
{
    "content": "Lyric content here..."
}
```

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "id": 1
    }
}
```

#### GetLyric
Get lyric content by ID

GET /api/track?Action=GetLyric&id=1

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "lyric": "Lyric content here..."
    }
}
```

#### UpdateLyric
Update an existing lyric

POST /api/track?Action=UpdateLyric

Request Body:
```json
{
    "id": 1,
    "content": "Updated lyric content..."
}
```

---

### Media

#### Upload
Upload media files (multipart/form-data)

POST /api/upload

Form Data:
- `photos`: file(s) to upload

Response:
```json
{
    "photos": ["uuid-1", "uuid-2"]
}
```

#### Serve
Serve a media file by UUID link

GET /api/m/:link

Returns the file directly with appropriate headers.

#### DeleteMedia
Delete media files by UUID IDs

POST /api/media?Action=DeleteMedia

Request Body:
```json
{
    "ids": ["uuid-1", "uuid-2"]
}
```

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "ids": ["uuid-1", "uuid-2"]
    }
}
```

Only successfully deleted IDs are returned in the response.

---

### Sync

The sync endpoints support client-side caching with incremental updates. All changes to `d_album`, `d_lyric`, `d_singer`, and `d_track` tables are automatically tracked via database triggers.

#### GetFullSync
Get all data for initial cache population

GET /api/sync?Action=GetFullSync

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "data": {
            "albums": [...],
            "lyrics": [...],
            "singers": [...],
            "tracks": [...],
            "timestamp": 1768054202
        }
    }
}
```

#### GetUpdates
Get incremental updates since a specific timestamp. If last sync happened 28 days ago, use `GetFullSync` instead.

GET /api/sync?Action=GetUpdates&since=1768054202873

Response:
```json
{
    "requestId": "uuid",
    "code": 200,
    "message": {
        "entries": [
            {
                "tableName": "d_track",
                "stale": [1, 2, 3],
                "deleted": [4, 5, 6]
            }
        ],
        "timestamp": 1768120438034
    }
}
```

**Notes:**
- For each record, only the **latest operation** is returned if multiple changes occurred.
- `operation` values: `INSERT`, `UPDATE`, `DELETE`
- Use the returned `timestamp` as the `since` parameter for the next sync request
