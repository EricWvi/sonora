"""
Audio metadata extraction module

This module provides functions to extract metadata from various audio formats
including MP3 and OGG files.
"""

from pathlib import Path
from mutagen.mp3 import MP3
from mutagen.oggvorbis import OggVorbis
from mutagen.id3 import ID3NoHeaderError
from dataclasses import dataclass
from typing import Optional, List


@dataclass
class AudioMetadata:
    """Audio file metadata container"""
    title: str
    artists: List[str]  # List of artist names
    album: str
    genre: str
    track_number: str
    duration_seconds: Optional[float]  # Duration as number
    date: str
    file_path: str
    file_size: int
    format: str

    @property
    def artist(self) -> str:
        """Get artists as semicolon-separated string for display"""
        return '; '.join(self.artists) if self.artists else 'Unknown'

    @property
    def duration(self) -> str:
        """Get formatted duration string"""
        if self.duration_seconds is None:
            return "Unknown"
        minutes = int(self.duration_seconds // 60)
        seconds = int(self.duration_seconds % 60)
        return f"{minutes:02d}:{seconds:02d}"

    def __str__(self):
        """Nice string representation for display"""
        return f"""Title:        {self.title}
Track Number: {self.track_number}
Album:        {self.album}
Artist:       {self.artist}
Genre:        {self.genre}
Duration:     {self.duration}
Date:         {self.date}"""


def format_duration(seconds):
    """Convert seconds to MM:SS format"""
    if seconds is None:
        return "Unknown"
    minutes = int(seconds // 60)
    seconds = int(seconds % 60)
    return f"{minutes:02d}:{seconds:02d}"


def extract_metadata(file_path) -> Optional[AudioMetadata]:
    """
    Extract metadata from an audio file (MP3 or OGG)

    Args:
        file_path: Path to the audio file (Path object or string)

    Returns:
        AudioMetadata: Metadata object or None if extraction fails
    """
    try:
        file_path = Path(file_path)
        file_ext = file_path.suffix.lower()

        if file_ext == '.mp3':
            audio = MP3(file_path)
            # Extract MP3 metadata using ID3 tags
            title = audio.get('TIT2', ['Unknown'])[0] if 'TIT2' in audio else 'Unknown'

            # Handle multiple artists as list
            artists = []
            if 'TPE1' in audio:
                tpe1_tag = audio['TPE1']
                # TPE1 is a mutagen tag object, convert to string and handle delimiters
                artist_str = str(tpe1_tag)
                if '\x00' in artist_str:
                    # Handle null-byte separated artists (common in ID3 tags)
                    artists = [a.strip() for a in artist_str.split('\x00') if a.strip()]
                elif ';' in artist_str:
                    artists = [a.strip() for a in artist_str.split(';') if a.strip()]
                elif '/' in artist_str:
                    artists = [a.strip() for a in artist_str.split('/') if a.strip()]
                elif ',' in artist_str:
                    artists = [a.strip() for a in artist_str.split(',') if a.strip()]
                else:
                    artists = [artist_str.strip()] if artist_str.strip() else []

            album = audio.get('TALB', ['Unknown'])[0] if 'TALB' in audio else 'Unknown'
            genre = audio.get('TCON', ['Unknown'])[0] if 'TCON' in audio else 'Unknown'
            date = audio.get('TDRC', ['Unknown'])[0] if 'TDRC' in audio else 'Unknown'

            # Track number (can be "1/12" format)
            track_num = 'Unknown'
            if 'TRCK' in audio:
                track_info = str(audio['TRCK'][0])
                track_num = track_info.split('/')[0] if '/' in track_info else track_info

        elif file_ext == '.ogg':
            audio = OggVorbis(file_path)
            # Extract OGG metadata using Vorbis comments
            title = audio.get('title', ['Unknown'])[0] if 'title' in audio else 'Unknown'

            # Handle multiple artists as list
            artists = []
            if 'artist' in audio:
                ogg_artists = audio['artist']
                if len(ogg_artists) > 1:
                    artists = [str(a).strip() for a in ogg_artists]
                else:
                    # Check if single artist field contains multiple artists separated by delimiters
                    artist_str = str(ogg_artists[0])
                    if '\x00' in artist_str:
                        # Handle null-byte separated artists
                        artists = [a.strip() for a in artist_str.split('\x00') if a.strip()]
                    elif ';' in artist_str:
                        artists = [a.strip() for a in artist_str.split(';') if a.strip()]
                    elif '/' in artist_str:
                        artists = [a.strip() for a in artist_str.split('/') if a.strip()]
                    elif ',' in artist_str:
                        artists = [a.strip() for a in artist_str.split(',') if a.strip()]
                    else:
                        artists = [artist_str.strip()] if artist_str.strip() else []

            album = audio.get('album', ['Unknown'])[0] if 'album' in audio else 'Unknown'
            genre = audio.get('genre', ['Unknown'])[0] if 'genre' in audio else 'Unknown'
            date = audio.get('date', ['Unknown'])[0] if 'date' in audio else 'Unknown'

            # Track number
            track_num = 'Unknown'
            if 'tracknumber' in audio:
                track_info = str(audio['tracknumber'][0])
                track_num = track_info.split('/')[0] if '/' in track_info else track_info
        else:
            return None

        return AudioMetadata(
            title=str(title),
            artists=artists,
            album=str(album),
            genre=str(genre),
            track_number=str(track_num),
            duration_seconds=audio.info.length,
            date=str(date),
            file_path=str(file_path),
            file_size=file_path.stat().st_size,
            format=file_ext[1:].upper()
        )

    except ID3NoHeaderError:
        return None
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return None


def get_audio_files(directory_path, recursive=False):
    """
    Get all audio files from a directory

    Args:
        directory_path: Path to directory to scan
        recursive: Whether to scan subdirectories

    Returns:
        list: List of Path objects for audio files
    """
    directory = Path(directory_path)
    if not directory.exists() or not directory.is_dir():
        return []

    patterns = ['*.mp3', '*.ogg']
    audio_files = []

    for pattern in patterns:
        if recursive:
            audio_files.extend(directory.rglob(pattern))
        else:
            audio_files.extend(directory.glob(pattern))

    return sorted(audio_files)


def is_supported_audio_file(file_path):
    """
    Check if a file is a supported audio format

    Args:
        file_path: Path to the file

    Returns:
        bool: True if supported, False otherwise
    """
    supported_extensions = {'.mp3', '.ogg'}
    return Path(file_path).suffix.lower() in supported_extensions
