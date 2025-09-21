#!/usr/bin/env python3
"""
Sonora CLI Tool

A command-line interface for common operations on audio files.

Usage:
    python app.py show <file_path>    - Show metadata for a single audio file
    python app.py help              - Show this help message

Examples:
    python app.py show music.mp3
    python app.py show song.ogg
"""

import sys
import argparse
from pathlib import Path
from audio_metadata import extract_metadata


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


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(
        description="Sonora CLI Tool - Audio file operations",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python app.py show music.mp3     Show metadata for music.mp3
  python app.py show song.ogg      Show metadata for song.ogg
        """
    )

    subparsers = parser.add_subparsers(dest='command', help='Available commands')

    # Show command
    show_parser = subparsers.add_parser('show', help='Show metadata for an audio file')
    show_parser.add_argument('file', help='Path to the audio file')

    # Parse arguments
    if len(sys.argv) == 1:
        parser.print_help()
        return

    args = parser.parse_args()

    if args.command == 'show':
        success = show_command(args.file)
        sys.exit(0 if success else 1)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()