"""
Upload functionality for Sonora CLI

This module provides functions for uploading albums, including API client
functions and file organization.
"""

import requests
import json
import uuid
import shutil
from pathlib import Path
from audio_metadata import extract_metadata, extract_cover_art, get_audio_files
from config import API_BASE_URL, API_ENDPOINTS, OUTPUT_DIR, ensure_directories


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
                track_number: int = 0) -> int:
    """Create a new track"""
    response = make_api_request(API_ENDPOINTS['track'], 'CreateTrack', {
        'name': name,
        'singer': singer,
        'album': album,
        'cover': cover,
        'url': url,
        'lyric': 0,
        'duration': duration,
        'year': year,
        'trackNumber': track_number
    })
    if response and 'message' in response and 'id' in response['message']:
        return response['message']['id']

    raise Exception(f"Failed to create track: {name}")


def extract_cover_art_to_output(file_path: Path, output_dir: Path, album_name: str) -> str:
    """Extract cover art from audio file and save to output directory with UUID structure"""
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

        # Return relative path
        return f"{first_two}/{second_two}/{cover_filename}"
    except Exception as e:
        print(f"Failed to extract cover art: {e}")
        return ""


def move_audio_file_to_output(file_path: Path, output_dir: Path) -> str:
    """Move audio file to output directory with UUID structure"""
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

    # Return relative path
    return f"{first_two}/{second_two}/{file_path.name}"


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
        cover_path = extract_cover_art_to_output(Path(first_track.file_path), OUTPUT_DIR, album_name)
        if cover_path:
            print(f"Album cover saved to: output/{cover_path}")

        # Create album record
        print("Creating album record...")
        album_id = create_album(album_name, cover_path, album_year)
        print(f"Album created with ID: {album_id}")

        # Process each track
        for i, metadata in enumerate(tracks_metadata, 1):
            print(f"Processing track {i}/{len(tracks_metadata)}: {metadata.title}")

            # Get primary artist (first one)
            primary_artist = metadata.artists[0] if metadata.artists else "Unknown Artist"

            # Find or create singer
            singer_id = find_or_create_singer(primary_artist)
            print(f"  Singer: {primary_artist} (ID: {singer_id})")

            # Move audio file to output directory
            audio_path = move_audio_file_to_output(Path(metadata.file_path), OUTPUT_DIR)
            print(f"  Audio file moved to: output/{audio_path}")

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

            # Create track record
            track_id = create_track(
                name=metadata.title,
                singer=primary_artist,  # API expects singer name, not ID
                album=album_id,
                cover=cover_path,  # Use album cover for track
                url=audio_path,
                duration=duration,
                year=album_year,
                track_number=track_number
            )
            print(f"  Track created with ID: {track_id}")

        print(f"\nAlbum upload completed successfully!")
        print(f"Album: {album_name} (ID: {album_id})")
        print(f"Tracks: {len(tracks_metadata)} tracks processed")
        return True

    except Exception as e:
        print(f"Error during album upload: {e}")
        return False