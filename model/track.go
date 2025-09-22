package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Track struct {
	gorm.Model
	TrackField
}

const trackPageSize = 15

type TrackView struct {
	ID uint `json:"id"`
	TrackField
}

type TrackField struct {
	Name        string `gorm:"size:255;not null" json:"name"`
	Singer      string `gorm:"size:255;not null" json:"singer"`
	Album       int    `gorm:"not null;default:0" json:"album"`
	Cover       string `gorm:"size:255;not null" json:"cover"`
	URL         string `gorm:"size:255;not null" json:"url"`
	Lyric       int    `gorm:"not null;default:0" json:"lyric"`
	Duration    int    `gorm:"not null;default:0" json:"duration"`
	Year        int16  `gorm:"not null;default:0" json:"year"`
	TrackNumber int16  `gorm:"not null;default:0" json:"trackNumber"`
	Genre       string `gorm:"size:50;not null" json:"genre"`
	AlbumText   string `gorm:"size:255;not null" json:"albumText"`
}

const (
	Track_Table = "d_track"

	// Track column names
	Track_Name        = "name"
	Track_Singer      = "singer"
	Track_Album       = "album"
	Track_Cover       = "cover"
	Track_URL         = "url"
	Track_Lyric       = "lyric"
	Track_Duration    = "duration"
	Track_Year        = "year"
	Track_TrackNumber = "track_number"
	Track_Genre       = "genre"
	Track_AlbumText   = "album_text"
)

func (t *Track) TableName() string {
	return Track_Table
}

func (t *Track) Get(db *gorm.DB, where map[string]any) error {
	rst := db.Where(where).Find(&t)
	if rst.Error != nil {
		return rst.Error
	}
	if rst.RowsAffected == 0 {
		return fmt.Errorf("can not find track")
	}
	return nil
}

func ListAllTracks(db *gorm.DB, where WhereExpr) ([]TrackView, error) {
	tracks := make([]TrackView, 0)
	for i := range where {
		db = db.Where(where[i])
	}
	if err := db.Model(&Track{}).
		Order("id DESC").
		Omit("created_at", "updated_at", "deleted_at").
		Find(&tracks).Error; err != nil {
		return nil, err
	}
	return tracks, nil
}

func ListTracks(db *gorm.DB, where WhereExpr, page uint) ([]TrackView, bool, error) {
	if page < 1 {
		return nil, false, errors.New("page number must be greater than 0")
	}
	tracks := make([]TrackView, 0, trackPageSize+1)
	offset := (page - 1) * trackPageSize

	for i := range where {
		db = db.Where(where[i])
	}

	// Retrieve one extra to check if there are more entries
	if err := db.Model(&Track{}).
		Order("id DESC").
		Offset(int(offset)).
		Limit(trackPageSize+1).
		Omit("created_at", "updated_at", "deleted_at").
		Find(&tracks).Error; err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(tracks) > trackPageSize {
		hasMore = true
		tracks = tracks[:trackPageSize]
	}

	return tracks, hasMore, nil
}

func (t *Track) Create(db *gorm.DB) error {
	return db.Create(t).Error
}

func (t *Track) Update(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Updates(t).Error
}

func (t *Track) Delete(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Delete(t).Error
}
