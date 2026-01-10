package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Lyric struct {
	gorm.Model
	LyricField
}

type LyricView struct {
	ID uint `json:"id"`
	LyricField
}

type LyricField struct {
	Content string `gorm:"type:text;not null" json:"content"`
}

const (
	Lyric_Table = "d_lyric"

	// Lyric column names
	Lyric_Content = "content"
)

func (l *Lyric) TableName() string {
	return Lyric_Table
}

func (l *Lyric) Get(db *gorm.DB, where map[string]any) error {
	rst := db.Where(where).Find(&l)
	if rst.Error != nil {
		return rst.Error
	}
	if rst.RowsAffected == 0 {
		return fmt.Errorf("can not find lyric")
	}
	return nil
}

func (l *Lyric) Create(db *gorm.DB) error {
	return db.Create(l).Error
}

func (l *Lyric) Update(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Updates(l).Error
}

func ListAllLyrics(db *gorm.DB, where WhereExpr) ([]LyricView, error) {
	lyrics := []LyricView{}
	for i := range where {
		db = db.Where(where[i])
	}
	rst := db.Table(Lyric_Table).Omit("created_at", "updated_at", "deleted_at").Find(&lyrics)
	if rst.Error != nil {
		return nil, rst.Error
	}
	return lyrics, nil
}
