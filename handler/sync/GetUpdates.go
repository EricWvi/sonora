package sync

import (
	"time"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetUpdates(c *gin.Context, req *GetUpdatesRequest) *GetUpdatesResponse {
	db := config.ContextDB(c)

	if req.Since == 0 {
		handler.Errorf(c, "since parameter is required")
		return nil
	}

	// Fetch all change logs since the given unix timestamp
	changeLogs, err := model.ListChanges(db, req.Since)
	if err != nil {
		handler.Errorf(c, "failed to fetch change logs: %s", err.Error())
		return nil
	}

	// Build a map to track the latest operation for each record
	finalChanges := make(map[string]map[uint]int)
	for i, log := range changeLogs {
		if _, exists := finalChanges[log.Table]; !exists {
			finalChanges[log.Table] = make(map[uint]int)
		}
		finalChanges[log.Table][log.RecordID] = i
	}

	// Collect unique changes (only the latest operation for each record)
	changedEntries := make([]model.ChangedEntries, 0)
	for tableName, changes := range finalChanges {
		staleEntries := make([]uint, 0)
		deletedEntries := make([]uint, 0)
		for recordID, idx := range changes {
			if changeLogs[idx].Operation == model.OpDelete {
				deletedEntries = append(deletedEntries, recordID)
			} else {
				staleEntries = append(staleEntries, recordID)
			}
		}
		changedEntries = append(changedEntries, model.ChangedEntries{
			TableName: tableName,
			Stale:     staleEntries,
			Deleted:   deletedEntries,
		})
	}

	return &GetUpdatesResponse{
		Entries:   changedEntries,
		Timestamp: time.Now().UnixMilli(),
	}
}

type GetUpdatesRequest struct {
	Since int64 `form:"since" binding:"required"`
}

type GetUpdatesResponse struct {
	Entries   []model.ChangedEntries `json:"entries"`
	Timestamp int64                  `json:"timestamp"`
}
