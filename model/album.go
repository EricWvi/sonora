package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Album struct {
	gorm.Model
	AlbumField
}

type AlbumField struct {
	Name  string `gorm:"size:255;not null" json:"name"`
	Cover string `gorm:"size:1024;not null" json:"cover"`
}

const (
	Album_Table = "d_album"
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

func ListAlbums(db *gorm.DB, where map[string]any) ([]Album, error) {
	albums := make([]Album, 0)
	if err := db.Where(where).
		Order("created_at DESC").
		Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
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