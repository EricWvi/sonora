"""
Upload functionality for Sonora CLI

This module provides functions for uploading albums, including API client
functions and file organization.
"""

import requests
import json
import uuid
import shutil
import psycopg2
from pathlib import Path
from audio_metadata import extract_metadata, extract_cover_art, get_audio_files
from config import API_BASE_URL, API_ENDPOINTS, OUTPUT_DIR, ensure_directories, get_database_url


def create_media_record(file_path: str) -> str:
    """Create a media record in the database and return the UUID"""
    try:
        # Generate UUID for the media record
        media_uuid = str(uuid.uuid4())

        # Connect to database
        conn = psycopg2.connect(get_database_url())
        cur = conn.cursor()

        # Insert into d_media table
        cur.execute(
            "INSERT INTO d_media (link, key, created_at, updated_at) VALUES (%s, %s, NOW(), NOW())",
            (media_uuid, file_path)
        )

        conn.commit()
        cur.close()
        conn.close()

        return media_uuid

    except Exception as e:
        print(f"Failed to create media record: {e}")
        raise Exception(f"Failed to create media record for {file_path}")


def make_api_request(endpoint: str, action: str, data: dict = None) -> dict:
    """Make a POST request to the API with the given action and data"""
    url = f"{API_BASE_URL}{endpoint}?Action={action}"

    payload = data if data else {}

    try:
        response = requests.post(url, json=payload)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"API request failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}")
        return {}


def find_or_create_singer(name: str) -> int:
    """Find an existing singer or create a new one"""
    # Search for existing singer
    response = make_api_request(API_ENDPOINTS['singer'], 'SearchSinger', {'query': name})
    if response and 'message' in response and 'singers' in response['message']:
        for singer in response['message']['singers']:
            if singer.get('name', '').lower() == name.lower():
                return singer.get('id')

    # If not found, create new singer
    response = make_api_request(API_ENDPOINTS['singer'], 'CreateSinger', {
        'name': name,
        'avatar': ''
    })
    if response and 'message' in response and 'id' in response['message']:
        return response['message']['id']

    raise Exception(f"Failed to create singer: {name}")


def create_album(name: str, cover: str = "", year: int = 0) -> int:
    """Create a new album"""
    response = make_api_request(API_ENDPOINTS['album'], 'CreateAlbum', {
        'name': name,
        'cover': cover,
        'year': year
    })
    if response and 'message' in response and 'id' in response['message']:
        return response['message']['id']

    raise Exception(f"Failed to create album: {name}")


def create_track(name: str, singer: str, album: int, cover: str = "",
                url: str = "", duration: int = 0, year: int = 0,
                track_number: int = 0, lyric_id: int = 0, genre: str = "",
                album_text: str = "") -> int:
    """Create a new track"""
    response = make_api_request(API_ENDPOINTS['track'], 'CreateTrack', {
        'name': name,
        'singer': singer,
        'album': album,
        'cover': cover,
        'url': url,
        'lyric': lyric_id,
        'duration': duration,
        'year': year,
        'trackNumber': track_number,
        'genre': genre,
        'albumText': album_text
    })
    if response and 'message' in response and 'id' in response['message']:
        return response['message']['id']

    raise Exception(f"Failed to create track: {name}")


def create_lyric(content: str) -> int:
    """Create a new lyric record"""
    if not content.strip():
        return 0  # Return 0 for empty lyrics (default value)

    response = make_api_request(API_ENDPOINTS['lyric'], 'CreateLyric', {
        'content': content.strip()
    })
    if response and 'message' in response and 'id' in response['message']:
        return response['message']['id']

    raise Exception(f"Failed to create lyric")


def extract_cover_art_to_output(file_path: Path, output_dir: Path, album_name: str) -> str:
    """Extract cover art from audio file, save to output directory, create media record, and return UUID"""
    try:
        # Generate UUID for cover art
        cover_uuid = str(uuid.uuid4())
        first_two = cover_uuid[:2]
        second_two = cover_uuid[2:4]

        # Create directory structure
        cover_dir = output_dir / first_two / second_two
        cover_dir.mkdir(parents=True, exist_ok=True)

        # Extract cover art using existing function first
        temp_cover = extract_cover_art(file_path)
        if not temp_cover:
            return ""

        # Determine file extension from temporary cover
        temp_cover_path = Path(temp_cover)
        extension = temp_cover_path.suffix

        # Move to proper output location with album name
        cover_filename = f"{album_name}{extension}"
        final_cover_path = cover_dir / cover_filename
        shutil.move(temp_cover, final_cover_path)

        # Create media record in database
        relative_path = f"{first_two}/{second_two}/{cover_filename}"
        media_uuid = create_media_record(relative_path)

        # Return media UUID (not file path)
        return media_uuid
    except Exception as e:
        print(f"Failed to extract cover art: {e}")
        return ""


def move_audio_file_to_output(file_path: Path, output_dir: Path) -> str:
    """Move audio file to output directory with UUID structure and create media record"""
    # Generate UUID for audio file
    file_uuid = str(uuid.uuid4())
    first_two = file_uuid[:2]
    second_two = file_uuid[2:4]

    # Create directory structure
    audio_dir = output_dir / first_two / second_two
    audio_dir.mkdir(parents=True, exist_ok=True)

    # Move file to output location
    final_audio_path = audio_dir / file_path.name
    shutil.copy2(file_path, final_audio_path)

    # Create media record in database
    relative_path = f"{first_two}/{second_two}/{file_path.name}"
    media_uuid = create_media_record(relative_path)

    # Return media UUID (not file path)
    return media_uuid


def upload_album(folder_path: Path) -> bool:
    """Upload an album from a folder"""
    try:
        # Ensure output directory exists
        ensure_directories()

        print(f"Processing album folder: {folder_path.name}")

        # Get all audio files from the folder
        audio_files = get_audio_files(folder_path)

        if not audio_files:
            print("No supported audio files found in the folder")
            return False

        print(f"Found {len(audio_files)} audio file(s)")

        # Extract metadata from all tracks
        tracks_metadata = []
        for audio_file in audio_files:
            metadata = extract_metadata(audio_file)
            if metadata:
                tracks_metadata.append(metadata)

        if not tracks_metadata:
            print("No metadata could be extracted from audio files")
            return False

        # Get album information from first track
        first_track = tracks_metadata[0]
        album_name = first_track.album
        album_year = 0
        try:
            if first_track.date and first_track.date.isdigit():
                album_year = int(first_track.date)
        except (ValueError, AttributeError):
            pass

        print(f"Album: {album_name} ({album_year if album_year else 'Unknown year'})")

        # Extract and save album cover art
        cover_uuid = extract_cover_art_to_output(Path(first_track.file_path), OUTPUT_DIR, album_name)
        if cover_uuid:
            print(f"Album cover saved with media UUID: {cover_uuid}")

        # Create album record
        print("Creating album record...")
        album_id = create_album(album_name, cover_uuid, album_year)
        print(f"Album created with ID: {album_id}")

        # Process each track
        for i, metadata in enumerate(tracks_metadata, 1):
            print(f"Processing track {i}/{len(tracks_metadata)}: {metadata.title}")

            # Find or create all artists
            artists_list = metadata.artists if metadata.artists else ["Unknown Artist"]
            artist_ids = []

            for artist_name in artists_list:
                artist_id = find_or_create_singer(artist_name)
                artist_ids.append(artist_id)
                print(f"  Singer: {artist_name} (ID: {artist_id})")

            # Concatenate all artists with "; "
            all_artists = "; ".join(artists_list)

            # Move audio file to output directory and create media record
            audio_uuid = move_audio_file_to_output(Path(metadata.file_path), OUTPUT_DIR)
            print(f"  Audio file moved with media UUID: {audio_uuid}")

            # Parse track number
            track_number = 0
            try:
                if metadata.track_number and metadata.track_number.isdigit():
                    track_number = int(metadata.track_number)
            except (ValueError, AttributeError):
                pass

            # Parse duration
            duration = 0
            if metadata.duration_seconds:
                duration = int(metadata.duration_seconds)

            # Create lyric record if lyrics exist
            lyric_id = 0
            if metadata.lyric.strip():
                lyric_id = create_lyric(metadata.lyric)
                print(f"  Lyrics created with ID: {lyric_id}")

            # Create track record
            track_id = create_track(
                name=metadata.title,
                singer=all_artists,  # All artists concatenated with "; "
                album=album_id,
                cover=cover_uuid,  # Use album cover UUID for track
                url=audio_uuid,  # Use media UUID for audio file
                duration=duration,
                year=album_year,
                track_number=track_number,
                lyric_id=lyric_id,
                genre=metadata.genre,  # Use genre from metadata
                album_text=album_name  # Use album name as album_text
            )
            print(f"  Track created with ID: {track_id}")

        print(f"\nAlbum upload completed successfully!")
        print(f"Album: {album_name} (ID: {album_id})")
        print(f"Tracks: {len(tracks_metadata)} tracks processed")
        return True

    except Exception as e:
        print(f"Error during album upload: {e}")
        return False