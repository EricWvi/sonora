package migration

import (
	"fmt"
	"time"

	"github.com/EricWvi/sonora/log"
	"gorm.io/gorm"
)

// SafeColumnAdd safely adds a column to a table
func SafeColumnAdd(db *gorm.DB, tableName, columnName, columnType string, defaultValue *string) error {
	// Check if column already exists
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = ? AND column_name = ? AND table_schema = 'public'
	`

	if err := db.Raw(query, tableName, columnName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if column exists: %w", err)
	}

	if count > 0 {
		log.Infof(log.WorkerCtx, "Column %s.%s already exists, skipping", tableName, columnName)
		return nil
	}

	// Add the column
	var sql string
	if defaultValue != nil {
		sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s DEFAULT %s", tableName, columnName, columnType, *defaultValue)
	} else {
		sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
	}

	log.Infof(log.WorkerCtx, "Adding column %s.%s", tableName, columnName)
	return db.Exec(sql).Error
}

// SafeColumnDrop safely drops a column from a table
func SafeColumnDrop(db *gorm.DB, tableName, columnName string) error {
	// Check if column exists
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = ? AND column_name = ? AND table_schema = 'public'
	`

	if err := db.Raw(query, tableName, columnName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if column exists: %w", err)
	}

	if count == 0 {
		log.Infof(log.WorkerCtx, "Column %s.%s doesn't exist, skipping drop", tableName, columnName)
		return nil
	}

	// Drop the column
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, columnName)
	log.Infof(log.WorkerCtx, "Dropping column %s.%s", tableName, columnName)
	return db.Exec(sql).Error
}

// SafeIndexAdd safely adds an index
func SafeIndexAdd(db *gorm.DB, tableName, indexName, columns string, unique bool) error {
	// Check if index already exists
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE tablename = ? AND indexname = ? AND schemaname = 'public'
	`

	if err := db.Raw(query, tableName, indexName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if count > 0 {
		log.Infof(log.WorkerCtx, "Index %s already exists, skipping", indexName)
		return nil
	}

	// Create the index
	var sql string
	if unique {
		sql = fmt.Sprintf("CREATE UNIQUE INDEX CONCURRENTLY %s ON %s (%s)", indexName, tableName, columns)
	} else {
		sql = fmt.Sprintf("CREATE INDEX CONCURRENTLY %s ON %s (%s)", indexName, tableName, columns)
	}

	log.Infof(log.WorkerCtx, "Creating index %s", indexName)
	return db.Exec(sql).Error
}

// SafeIndexDrop safely drops an index
func SafeIndexDrop(db *gorm.DB, indexName string) error {
	// Check if index exists
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE indexname = ? AND schemaname = 'public'
	`

	if err := db.Raw(query, indexName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if count == 0 {
		log.Infof(log.WorkerCtx, "Index %s doesn't exist, skipping drop", indexName)
		return nil
	}

	// Drop the index
	sql := fmt.Sprintf("DROP INDEX CONCURRENTLY IF EXISTS %s", indexName)
	log.Infof(log.WorkerCtx, "Dropping index %s", indexName)
	return db.Exec(sql).Error
}

// BackfillData helps with data backfill operations
func BackfillData(db *gorm.DB, query string, batchSize int, processor func(*gorm.DB, []map[string]any) error) error {
	offset := 0

	for {
		var results []map[string]any

		// Execute query with limit and offset
		limitQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, batchSize, offset)
		if err := db.Raw(limitQuery).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to fetch batch: %w", err)
		}

		// No more results
		if len(results) == 0 {
			break
		}

		// Process the batch
		if err := processor(db, results); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}

		log.Infof(log.WorkerCtx, "Processed %d records (offset: %d)", len(results), offset)
		offset += batchSize

		// Add a small delay to prevent overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
