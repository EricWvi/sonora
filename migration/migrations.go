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
	}
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
			lyric TEXT DEFAULT '' NOT NULL,
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
