package migration

import (
	"gorm.io/gorm"
)

// GetAllMigrations returns all defined migrations in order
func GetAllMigrations() []MigrationStep {
	return []MigrationStep{
		{
			Version: "v0.1.0",
			Name:    "Create user+media+migration",
			Up:      InitTables,
			Down:    DropInitTables,
		},
		{
			Version: "v0.2.0",
			Name:    "Create singer+album+track",
			Up:      CreateSingerAlbumTrack,
			Down:    DropSingerAlbumTrack,
		},
		{
			Version: "v0.4.0",
			Name:    "Create lyric table",
			Up:      CreateLyricTable,
			Down:    DropLyricTable,
		},
		{
			Version: "v0.5.0",
			Name:    "Create search index for track, singer, album name",
			Up:      CreateSearchIndex,
			Down:    DropSearchIndex,
		},
		{
			Version: "v0.6.0",
			Name:    "Add genre and album_text columns to d_track",
			Up:      AddTrackGenreAlbumText,
			Down:    DropTrackGenreAlbumText,
		},
		{
			Version: "v0.7.0",
			Name:    "Add multiple indexes for performance optimization",
			Up:      CreatePerformanceIndexes,
			Down:    DropPerformanceIndexes,
		},
		{
			Version: "v0.8.0",
			Name:    "Add trigger to update track covers when album cover changes",
			Up:      CreateAlbumCoverUpdateTrigger,
			Down:    DropAlbumCoverUpdateTrigger,
		},
	}
}

// ------------------- v0.8.0 -------------------
func CreateAlbumCoverUpdateTrigger(db *gorm.DB) error {
	return db.Exec(`
		-- Create trigger function to update track covers when album cover changes
		CREATE OR REPLACE FUNCTION update_track_cover()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Update cover for all tracks with this album
			UPDATE public.d_track
			SET cover = NEW.cover
			WHERE album = NEW.id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		-- Create trigger
		CREATE TRIGGER trigger_update_track_cover
		AFTER UPDATE OF cover ON public.d_album
		FOR EACH ROW
		WHEN (OLD.cover IS DISTINCT FROM NEW.cover)
		EXECUTE FUNCTION update_track_cover();
	`).Error
}

func DropAlbumCoverUpdateTrigger(db *gorm.DB) error {
	return db.Exec(`
		-- Drop trigger and function
		DROP TRIGGER IF EXISTS trigger_update_track_cover ON public.d_album;
		DROP FUNCTION IF EXISTS update_track_cover();
	`).Error
}

// ------------------- v0.7.0 -------------------
func CreatePerformanceIndexes(db *gorm.DB) error {
	return db.Exec(`
		-- Index on d_media link column
		CREATE INDEX idx_media_link ON public.d_media (link);

		-- Trigram index on d_track singer column for text search
		CREATE INDEX idx_track_singer_trgm ON public.d_track USING gin (singer gin_trgm_ops);

		-- Index on d_track album column for joins and filtering
		CREATE INDEX idx_track_album ON public.d_track (album);

		-- Index on d_track genre column for filtering
		CREATE INDEX idx_track_genre ON public.d_track (genre);
	`).Error
}

func DropPerformanceIndexes(db *gorm.DB) error {
	return db.Exec(`
		DROP INDEX IF EXISTS idx_media_link;
		DROP INDEX IF EXISTS idx_track_singer_trgm;
		DROP INDEX IF EXISTS idx_track_album;
		DROP INDEX IF EXISTS idx_track_genre;
	`).Error
}

// ------------------- v0.6.0 -------------------
func AddTrackGenreAlbumText(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE public.d_track
		ADD COLUMN genre VARCHAR(50) DEFAULT '' NOT NULL,
		ADD COLUMN album_text VARCHAR(255) DEFAULT '' NOT NULL;

		-- Create trigger function to update album_text when album name changes
		CREATE OR REPLACE FUNCTION update_track_album_text()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Update album_text for all tracks with this album
			UPDATE public.d_track
			SET album_text = NEW.name
			WHERE album = NEW.id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		-- Create trigger
		CREATE TRIGGER trigger_update_track_album_text
		AFTER UPDATE OF name ON public.d_album
		FOR EACH ROW
		WHEN (OLD.name IS DISTINCT FROM NEW.name)
		EXECUTE FUNCTION update_track_album_text();
	`).Error
}

func DropTrackGenreAlbumText(db *gorm.DB) error {
	return db.Exec(`
		-- Drop trigger and function
		DROP TRIGGER IF EXISTS trigger_update_track_album_text ON public.d_album;
		DROP FUNCTION IF EXISTS update_track_album_text();

		-- Drop columns
		ALTER TABLE public.d_track
		DROP COLUMN IF EXISTS genre,
		DROP COLUMN IF EXISTS album_text;
	`).Error
}

// ------------------- v0.5.0 -------------------
func CreateSearchIndex(db *gorm.DB) error {
	return db.Exec(`
	    CREATE EXTENSION IF NOT EXISTS pg_trgm;

		CREATE INDEX idx_track_name_trgm
		ON public.d_track
		USING gin (name gin_trgm_ops);

		CREATE INDEX idx_singer_name_trgm
		ON public.d_singer
		USING gin (name gin_trgm_ops);

		CREATE INDEX idx_album_name_trgm
		ON public.d_album
		USING gin (name gin_trgm_ops);
	`).Error
}

func DropSearchIndex(db *gorm.DB) error {
	return db.Exec(`
		DROP INDEX IF EXISTS idx_track_name_trgm;
		DROP INDEX IF EXISTS idx_singer_name_trgm;
		DROP INDEX IF EXISTS idx_album_name_trgm;
	`).Error
}

// ------------------- v0.4.0 -------------------
func CreateLyricTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE public.d_lyric (
			id SERIAL PRIMARY KEY,
			content TEXT DEFAULT '' NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);
	`).Error
}

func DropLyricTable(db *gorm.DB) error {
	return db.Exec(`
		DROP TABLE IF EXISTS public.d_lyric CASCADE;
	`).Error
}

// ------------------- v0.2.0 -------------------
func CreateSingerAlbumTrack(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE public.d_singer (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			avatar VARCHAR(1024) DEFAULT '' NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);

		CREATE TABLE public.d_album (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			cover VARCHAR(1024) DEFAULT '' NOT NULL,
			year SMALLINT DEFAULT 0 NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);

		CREATE TABLE public.d_track (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			singer VARCHAR(255) DEFAULT '' NOT NULL,
			album INTEGER DEFAULT 0 NOT NULL,
			cover VARCHAR(255) DEFAULT '' NOT NULL,
			url VARCHAR(255) DEFAULT '' NOT NULL,
			year SMALLINT DEFAULT 0 NOT NULL,
			track_number SMALLINT DEFAULT 0 NOT NULL,
			lyric INTEGER DEFAULT 0 NOT NULL,
			duration INT DEFAULT 0 NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);
	`).Error
}

func DropSingerAlbumTrack(db *gorm.DB) error {
	return db.Exec(`
		DROP TABLE IF EXISTS public.d_track CASCADE;
		DROP TABLE IF EXISTS public.d_album CASCADE;
		DROP TABLE IF EXISTS public.d_singer CASCADE;
	`).Error
}

// ------------------- v0.1.0 -------------------
func InitTables(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE public.d_user (
			id SERIAL PRIMARY KEY,
			email VARCHAR(100) NOT NULL UNIQUE,
			avatar varchar(1024) DEFAULT '' NOT NULL,
			username varchar(255) DEFAULT '' NOT NULL,
			"language" varchar(10) DEFAULT 'zh-CN' NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);

		CREATE TABLE public.d_media (
			id SERIAL PRIMARY KEY,
			link uuid DEFAULT gen_random_uuid(),
			key VARCHAR(1024) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
		);
	`).Error
}

func DropInitTables(db *gorm.DB) error {
	return db.Exec(`
		DROP TABLE IF EXISTS public.d_user CASCADE;
		DROP TABLE IF EXISTS public.d_media CASCADE;
	`).Error
}
