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
	}
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
