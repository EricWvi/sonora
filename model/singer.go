package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Singer struct {
	gorm.Model
	SingerField
}

type SingerField struct {
	Name   string `gorm:"size:255;not null" json:"name"`
	Avatar string `gorm:"size:1024;not null" json:"avatar"`
}

const (
	Singer_Table = "d_singer"
)

func (s *Singer) TableName() string {
	return Singer_Table
}

func (s *Singer) Get(db *gorm.DB, where map[string]any) error {
	rst := db.Where(where).Find(&s)
	if rst.Error != nil {
		return rst.Error
	}
	if rst.RowsAffected == 0 {
		return fmt.Errorf("can not find singer")
	}
	return nil
}

func ListSingers(db *gorm.DB, where map[string]any) ([]Singer, error) {
	singers := make([]Singer, 0)
	if err := db.Where(where).
		Order("created_at DESC").
		Find(&singers).Error; err != nil {
		return nil, err
	}
	return singers, nil
}

func (s *Singer) Create(db *gorm.DB) error {
	return db.Create(s).Error
}

func (s *Singer) Update(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Updates(s).Error
}

func (s *Singer) Delete(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Delete(s).Error
}
