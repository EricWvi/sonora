package model

import (
	"gorm.io/gorm"
)

// ChangeLog represents a record in the d_change_log table
type ChangeLog struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Table     string `gorm:"column:table_name;size:50;not null" json:"tableName"`
	RecordID  uint   `gorm:"column:record_id;not null" json:"recordId"`
	Operation string `gorm:"column:operation;size:10;not null" json:"operation"`
	ChangedAt int64  `gorm:"column:changed_at;not null" json:"changedAt"`
}

// Operation types
const (
	OpInsert = "INSERT"
	OpUpdate = "UPDATE"
	OpDelete = "DELETE"
)

const (
	ChangeLog_Table = "d_change_log"

	// ChangeLog column names
	ChangeLog_TableName = "table_name"
	ChangeLog_RecordID  = "record_id"
	ChangeLog_Operation = "operation"
	ChangeLog_ChangedAt = "changed_at"
)

func (c *ChangeLog) TableName() string {
	return ChangeLog_Table
}

func ListChanges(db *gorm.DB, timestamp int64) ([]ChangeLog, error) {
	changes := make([]ChangeLog, 0)

	if err := db.Model(&ChangeLog{}).
		Where("changed_at > ?", timestamp).
		Order("changed_at ASC").
		Find(&changes).Error; err != nil {
		return nil, err
	}

	return changes, nil
}

// FullSyncResponse represents the response for full sync
type FullSyncResponse struct {
	Albums    []AlbumView  `json:"albums"`
	Lyrics    []LyricView  `json:"lyrics"`
	Singers   []SingerView `json:"singers"`
	Tracks    []TrackView  `json:"tracks"`
	Timestamp int64        `json:"timestamp"`
}

// ChangedEntries represents changes of a table in the incremental sync
type ChangedEntries struct {
	TableName string `json:"tableName"`
	Stale     []uint `json:"stale"`
	Deleted   []uint `json:"deleted"`
}
