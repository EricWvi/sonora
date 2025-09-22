#!/usr/bin/env python3
"""
Sonora CLI Tool

A command-line interface for common operations on audio files.

Usage:
    python app.py show <file_path>         - Show metadata for a single audio file
    python app.py get-art <file_path>      - Extract cover art from audio file
    python app.py upload album <folder>    - Upload album from folder
    python app.py upload single <folder>   - Upload individual tracks from folder
    python app.py help                     - Show this help message

Examples:
    python app.py show music.mp3
    python app.py show song.ogg
    python app.py get-art music.mp3
    python app.py upload album /path/to/album/folder
    python app.py upload single /path/to/tracks/folder
"""

import sys
import argparse
from pathlib import Path
from audio_metadata import extract_metadata, extract_cover_art, get_audio_files
from upload import upload_album, upload_single


def show_command(file_path):
    """Show metadata for a single audio file"""
    file_path = Path(file_path)

    if not file_path.exists():
        print(f"Error: File '{file_path}' does not exist")
        return False

    if not file_path.is_file():
        print(f"Error: '{file_path}' is not a file")
        return False

    # Check if it's a supported audio file
    supported_extensions = {'.mp3', '.ogg'}
    if file_path.suffix.lower() not in supported_extensions:
        print(f"Error: Unsupported file type '{file_path.suffix}'. Supported: {', '.join(supported_extensions)}")
        return False

    print(f"File: {file_path.name}")
    print("=" * 60)

    metadata = extract_metadata(file_path)

    if metadata:
        print(metadata)
        return True
    else:
        print("Could not extract metadata (no tags found)")
        return False


def get_art_command(file_path):
    """Extract and save cover art from an audio file"""
    file_path = Path(file_path)

    if not file_path.exists():
        print(f"Error: File '{file_path}' does not exist")
        return False

    if not file_path.is_file():
        print(f"Error: '{file_path}' is not a file")
        return False

    # Check if it's a supported audio file
    supported_extensions = {'.mp3', '.ogg'}
    if file_path.suffix.lower() not in supported_extensions:
        print(f"Error: Unsupported file type '{file_path.suffix}'. Supported: {', '.join(supported_extensions)}")
        return False

    print(f"Extracting cover art from: {file_path.name}")

    cover_path = extract_cover_art(file_path)

    if cover_path:
        print(f"Cover art saved to: {cover_path}")
        return True
    else:
        print("No cover art found in the audio file")
        return False


def upload_album_command(folder_path):
    """Upload an album from a folder"""
    folder_path = Path(folder_path)

    if not folder_path.exists():
        print(f"Error: Folder '{folder_path}' does not exist")
        return False

    if not folder_path.is_dir():
        print(f"Error: '{folder_path}' is not a directory")
        return False

    return upload_album(folder_path)


def upload_single_command(folder_path):
    """Upload individual tracks from a folder"""
    folder_path = Path(folder_path)

    if not folder_path.exists():
        print(f"Error: Folder '{folder_path}' does not exist")
        return False

    if not folder_path.is_dir():
        print(f"Error: '{folder_path}' is not a directory")
        return False

    return upload_single(folder_path)


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(
        description="Sonora CLI Tool - Audio file operations",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python app.py show music.mp3            Show metadata for music.mp3
  python app.py show song.ogg             Show metadata for song.ogg
  python app.py get-art music.mp3         Extract cover art from music.mp3
  python app.py upload album /path/folder Upload album from folder
  python app.py upload single /path/folder Upload individual tracks from folder
        """
    )

    subparsers = parser.add_subparsers(dest='command', help='Available commands')

    # Show command
    show_parser = subparsers.add_parser('show', help='Show metadata for an audio file')
    show_parser.add_argument('file', help='Path to the audio file')

    # Get-art command
    get_art_parser = subparsers.add_parser('get-art', help='Extract cover art from an audio file')
    get_art_parser.add_argument('file', help='Path to the audio file')

    # Upload command
    upload_parser = subparsers.add_parser('upload', help='Upload content to the server')
    upload_subparsers = upload_parser.add_subparsers(dest='upload_type', help='Upload types')

    # Upload album subcommand
    album_parser = upload_subparsers.add_parser('album', help='Upload an album from a folder')
    album_parser.add_argument('folder', help='Path to the album folder')

    # Upload single tracks subcommand
    single_parser = upload_subparsers.add_parser('single', help='Upload individual tracks from a folder')
    single_parser.add_argument('folder', help='Path to the folder containing tracks')

    # Parse arguments
    if len(sys.argv) == 1:
        parser.print_help()
        return

    args = parser.parse_args()

    if args.command == 'show':
        success = show_command(args.file)
        sys.exit(0 if success else 1)
    elif args.command == 'get-art':
        success = get_art_command(args.file)
        sys.exit(0 if success else 1)
    elif args.command == 'upload':
        if args.upload_type == 'album':
            success = upload_album_command(args.folder)
            sys.exit(0 if success else 1)
        elif args.upload_type == 'single':
            success = upload_single_command(args.folder)
            sys.exit(0 if success else 1)
        else:
            upload_parser.print_help()
    else:
        parser.print_help()


if __name__ == "__main__":
    main()