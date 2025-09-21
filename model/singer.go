package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Singer struct {
	gorm.Model
	SingerField
}

const singerPageSize = 15

type SingerView struct {
	ID uint `json:"id"`
	SingerField
}

type SingerField struct {
	Name   string `gorm:"size:255;not null" json:"name"`
	Avatar string `gorm:"size:1024;not null" json:"avatar"`
}

const (
	Singer_Table = "d_singer"

	// Singer column names
	Singer_Name   = "name"
	Singer_Avatar = "avatar"
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

func ListAllSingers(db *gorm.DB, where map[string]any) ([]SingerView, error) {
	singers := make([]SingerView, 0)
	if err := db.Model(&Singer{}).Where(where).
		Order("id DESC").
		Omit("created_at", "updated_at", "deleted_at").
		Find(&singers).Error; err != nil {
		return nil, err
	}
	return singers, nil
}

func ListSingers(db *gorm.DB, where WhereExpr, page uint) ([]SingerView, bool, error) {
	if page < 1 {
		return nil, false, errors.New("page number must be greater than 0")
	}
	singers := make([]SingerView, 0, singerPageSize+1)
	offset := (page - 1) * singerPageSize

	for i := range where {
		db = db.Where(where[i])
	}

	// Retrieve one extra to check if there are more entries
	if err := db.Model(&Singer{}).
		Order("id DESC").
		Offset(int(offset)).
		Limit(singerPageSize+1).
		Omit("created_at", "updated_at", "deleted_at").
		Find(&singers).Error; err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(singers) > singerPageSize {
		hasMore = true
		singers = singers[:singerPageSize]
	}

	return singers, hasMore, nil
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
