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
import imghdr
import base64
import struct


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
    lyric: str = ""  # Embedded lyrics text

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
        result = f"""Title:        {self.title}
Track Number: {self.track_number}
Album:        {self.album}
Artist:       {self.artist}
Genre:        {self.genre}
Duration:     {self.duration}
Date:         {self.date}"""

        if self.lyric.strip():
            result += f"\nLyrics:\n{self.lyric}"

        return result


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

            # Extract lyrics (USLT tag for unsynchronized lyrics or TXXX:lyrics)
            lyrics = ""
            # First check for standard USLT tags
            for key in audio.keys():
                if key.startswith('USLT'):
                    uslt = audio[key]
                    lyrics = str(uslt.text)
                    break

            # If no USLT found, check for TXXX tags that might contain lyrics
            if not lyrics:
                for key in audio.keys():
                    if key.startswith('TXXX:'):
                        txxx = audio[key]
                        desc = getattr(txxx, 'desc', '').lower()
                        # Check for various lyrics-related descriptions
                        if any(term in desc for term in ['lyrics', 'lyric', 'uslt', 'unsynced']):
                            lyrics = str(txxx.text[0]) if hasattr(txxx, 'text') and txxx.text else str(txxx)
                            break

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

            # Extract lyrics (LYRICS tag for OGG)
            lyrics = ""
            if 'lyrics' in audio:
                lyrics = str(audio['lyrics'][0])
            elif 'unsyncedlyrics' in audio:
                lyrics = str(audio['unsyncedlyrics'][0])
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
            format=file_ext[1:].upper(),
            lyric=lyrics.strip()
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


def extract_cover_art(file_path, output_path=None):
    """
    Extract cover art from an audio file and save it to disk

    Args:
        file_path: Path to the audio file (Path object or string)
        output_path: Optional output path. If None, saves to same directory as audio file

    Returns:
        str: Path to saved cover art file, or None if no cover art found
    """
    try:
        file_path = Path(file_path)
        file_ext = file_path.suffix.lower()

        cover_data = None
        image_type = None

        if file_ext == '.mp3':
            audio = MP3(file_path)
            # Look for attached pictures (APIC frames)
            for key in audio.keys():
                if key.startswith('APIC'):
                    apic = audio[key]
                    cover_data = apic.data
                    # Determine image type from MIME type or data
                    if apic.mime == 'image/jpeg':
                        image_type = 'jpeg'
                    elif apic.mime == 'image/png':
                        image_type = 'png'
                    else:
                        # Try to detect from data
                        image_type = imghdr.what(None, h=cover_data)
                    break

        elif file_ext == '.ogg':
            audio = OggVorbis(file_path)
            # OGG files have embedded images in METADATA_BLOCK_PICTURE
            if 'metadata_block_picture' in audio:
                metadata_blocks = audio['metadata_block_picture']
                if metadata_blocks:
                    # Decode the base64-encoded FLAC picture block
                    try:
                        block_data = base64.b64decode(metadata_blocks[0])

                        # Parse FLAC picture block format
                        # Skip picture type (4 bytes), MIME type length (4 bytes)
                        mime_length = struct.unpack('>I', block_data[4:8])[0]
                        mime_type = block_data[8:8+mime_length].decode('ascii')

                        # Skip description length and description
                        desc_start = 8 + mime_length
                        desc_length = struct.unpack('>I', block_data[desc_start:desc_start+4])[0]

                        # Skip width, height, depth, colors (16 bytes total)
                        # Skip data length (4 bytes)
                        data_start = desc_start + 4 + desc_length + 16 + 4

                        # Extract the actual image data
                        cover_data = block_data[data_start:]

                        # Determine image type from MIME type
                        if 'jpeg' in mime_type.lower():
                            image_type = 'jpeg'
                        elif 'png' in mime_type.lower():
                            image_type = 'png'
                        elif 'webp' in mime_type.lower():
                            image_type = 'webp'
                        else:
                            # Try to detect from data
                            image_type = imghdr.what(None, h=cover_data)

                    except Exception as e:
                        print(f"Error parsing OGG picture block: {e}")
                        pass

        if cover_data and image_type:
            # Determine output file path
            if output_path is None:
                base_name = file_path.stem
                parent_dir = file_path.parent
                if image_type == 'jpeg':
                    output_path = parent_dir / f"{base_name}.jpg"
                elif image_type == 'png':
                    output_path = parent_dir / f"{base_name}.png"
                elif image_type == 'webp':
                    output_path = parent_dir / f"{base_name}.webp"
                else:
                    output_path = parent_dir / f"{base_name}.{image_type}"
            else:
                output_path = Path(output_path)

            # Write cover art to file
            with open(output_path, 'wb') as f:
                f.write(cover_data)

            return str(output_path)

    except Exception as e:
        print(f"Error extracting cover art from {file_path}: {e}")

    return None


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
