package model

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Media struct {
	gorm.Model
	Link uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null"`
	Key  string    `gorm:"type:varchar(1024);not null;unique"`
}

const (
	Media_Table = "d_media"
	Media_Id    = "id"
	Media_Link  = "link"
)

func (m *Media) TableName() string {
	return Media_Table
}

func (m *Media) Get(db *gorm.DB, where map[string]any) error {
	rst := db.Where(where).Find(&m)
	if rst.Error != nil {
		return rst.Error
	}
	if rst.RowsAffected == 0 {
		return fmt.Errorf("can not find media")
	}
	return nil
}

func (m *Media) Create(db *gorm.DB) error {
	return db.Create(m).Error
}

func (m *Media) Update(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Updates(m).Error
}

func (m *Media) Delete(db *gorm.DB, where map[string]any) error {
	return db.Where(where).Delete(m).Error
}
