"""
Configuration settings for Sonora CLI Tool

This module contains configuration settings for database connections,
file paths, and other application settings.
"""

import os
from pathlib import Path

# Database configuration
DATABASE_CONFIG = {
    'host': os.getenv('DB_HOST', 'localhost'),
    'port': os.getenv('DB_PORT', 5432),
    'database': os.getenv('DB_NAME', 'sonora'),
    'username': os.getenv('DB_USER', 'postgres'),
    'password': os.getenv('DB_PASSWORD', ''),
}

# File paths
BASE_DIR = Path(__file__).parent
INPUT_DIR = BASE_DIR / 'input'
OUTPUT_DIR = BASE_DIR / 'output'

# Audio processing settings
SUPPORTED_AUDIO_FORMATS = {'.mp3', '.ogg', '.flac', '.m4a'}
DEFAULT_PAGE_SIZE = 20

# Application settings
APP_NAME = 'Sonora CLI'
APP_VERSION = '1.0.0'
DEBUG = os.getenv('DEBUG', 'False').lower() == 'true'

# Logging configuration
LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO')
LOG_FILE = BASE_DIR / 'sonora.log'

def get_database_url():
    """Get database connection URL"""
    config = DATABASE_CONFIG
    return f"postgresql://{config['username']}:{config['password']}@{config['host']}:{config['port']}/{config['database']}"

def ensure_directories():
    """Create necessary directories if they don't exist"""
    INPUT_DIR.mkdir(exist_ok=True)
    OUTPUT_DIR.mkdir(exist_ok=True)