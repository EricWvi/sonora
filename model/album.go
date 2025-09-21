package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Album struct {
	gorm.Model
	AlbumField
}

type AlbumView struct {
	ID uint `json:"id"`
	AlbumField
}

const albumPageSize = 15

type AlbumField struct {
	Name  string `gorm:"size:255;not null" json:"name"`
	Cover string `gorm:"size:1024;not null" json:"cover"`
	Year  int16  `gorm:"not null;default:0" json:"year"`
}

const (
	Album_Table = "d_album"

	// Album column names
	Album_Name  = "name"
	Album_Cover = "cover"
	Album_Year  = "year"
)

func (a *Album) TableName() string {
	return Album_Table
}

func (a *Album) Get(db *gorm.DB, where map[string]any) error {
	rst := db.Where(where).Find(&a)
	if rst.Error != nil {
		return rst.Error
	}
	if rst.RowsAffected == 0 {
		return fmt.Errorf("can not find album")
	}
	return nil
}

func ListAllAlbums(db *gorm.DB, where map[string]any) ([]AlbumView, error) {
	albums := make([]AlbumView, 0)
	if err := db.Model(&Album{}).Where(where).
		Order("id DESC").
		Omit("created_at", "updated_at", "deleted_at").
		Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

func ListAlbums(db *gorm.DB, where WhereExpr, page uint) ([]AlbumView, bool, error) {
	if page < 1 {
		return nil, false, errors.New("page number must be greater than 0")
	}
	albums := make([]AlbumView, 0, albumPageSize+1)
	offset := (page - 1) * albumPageSize

	for i := range where {
		db = db.Where(where[i])
	}

	// Retrieve one extra to check if there are more entries
	if err := db.Model(&Album{}).
		Order("id DESC").
		Offset(int(offset)).
		Limit(albumPageSize+1).
		Omit("created_at", "updated_at", "deleted_at").
		Find(&albums).Error; err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(albums) > albumPageSize {
		hasMore = true
		albums = albums[:albumPageSize]
	}

	return albums, hasMore, nil
}

func (a *Album) Create(db *gorm.DB) error {
	return db.Create(a).Error
}

func (a *Album) Update(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Updates(a).Error
}

func (a *Album) Delete(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Delete(a).Error
}
