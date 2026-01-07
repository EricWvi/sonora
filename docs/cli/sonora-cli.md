## CLI
Python CLI tool for audio file operations located in `cli/`.

### Structure

cli
├── app.py - Main CLI application with argparse interface
├── audio_metadata.py - Audio metadata extraction module with AudioMetadata dataclass
├── config.py - Configuration for database, paths, and application settings
└── upload.py - Album upload functionality with API integration

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
   - `single <folder>`: upload individual tracks without creating album record

### Album Upload Workflow (`upload album <folder>`)
1. **Scan folder** for supported audio files (.mp3, .ogg)
2. **Extract metadata** from all tracks using existing metadata extraction
3. **Extract and save album cover** to `cli/output/<ab>/<cd>/<album_name>.<ext>` (UUID directory structure)
4. **Create album record** via API with cover path and year from track metadata
5. **Process each track**:
   - Find or create singer using SearchSinger API (searches existing, creates if not found)
   - Move audio file to `cli/output/<ab>/<cd>/<filename>` (UUID directory structure)
   - Create track record with proper singer, album, and file path relationships

### Single Track Upload Workflow (`upload single <folder>`)
1. **Scan folder** for supported audio files (.mp3, .ogg)
2. **Process each track individually**:
   - Extract metadata from the track
   - Find or create singer(s) using SearchSinger API
   - Extract and save track's own cover art (if available)
   - Move audio file to `cli/output/<ab>/<cd>/<filename>` (UUID directory structure)
   - Create track record with `album=0` and `albumText=<album name from track metadata>`
   - No album record is created

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
python app.py upload single /path/to/tracks/folder
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
#   Audio file saved with media UUID: d746ff24-4f8b-4ef3-816d-dd03a3944c98
#   Lyrics created with ID: 6
#   Track created with ID: 14
# Processing track 2/2: Better Man
#   Singer: Westlife (ID: 15)
#   Audio file saved with media UUID: 38f14b5d-62b6-4656-83ca-5e836dc330bc
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
