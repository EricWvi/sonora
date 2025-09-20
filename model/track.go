package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Track struct {
	gorm.Model
	TrackField
}

type TrackField struct {
	Name     string `gorm:"size:255;not null" json:"name"`
	Singer   string `gorm:"size:255;not null" json:"singer"`
	Album    int    `gorm:"not null;default:0" json:"album"`
	Cover    string `gorm:"size:255;not null" json:"cover"`
	URL      string `gorm:"size:255;not null" json:"url"`
	Lyric    string `gorm:"type:text;not null" json:"lyric"`
	Duration int    `gorm:"not null;default:0" json:"duration"`
}

const (
	Track_Table = "d_track"
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

func ListTracks(db *gorm.DB, where map[string]any) ([]Track, error) {
	tracks := make([]Track, 0)
	if err := db.Where(where).
		Order("created_at DESC").
		Find(&tracks).Error; err != nil {
		return nil, err
	}
	return tracks, nil
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
