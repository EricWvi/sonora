package migration

import (
	"fmt"
	"time"

	"github.com/EricWvi/sonora/log"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (m *Migration) TableName() string {
	return "migration"
}

// MigrationFunc represents a migration function
type MigrationFunc func(*gorm.DB) error

// MigrationStep contains both up and down migration functions
type MigrationStep struct {
	Version string
	Name    string
	Up      MigrationFunc
	Down    MigrationFunc
}

// Migrator handles database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []MigrationStep
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: []MigrationStep{},
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration MigrationStep) {
	m.migrations = append(m.migrations, migration)
}

// InitMigrationTable creates the migration tracking table
func (m *Migrator) InitMigrationTable() error {
	return m.db.AutoMigrate(&Migration{})
}

// GetAppliedMigrations returns a list of applied migration versions
func (m *Migrator) GetAppliedMigrations() (map[string]bool, error) {
	var migrations []Migration
	if err := m.db.Find(&migrations).Error; err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, migration := range migrations {
		applied[migration.Version] = true
	}

	return applied, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	return m.UpWithOptions(MigrationOptions{
		MaxMigrations: 0, // Unlimited
		DryRun:        false,
		Force:         false,
	})
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	log.Info(log.WorkerCtx, "Rolling back last migration...")

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Find the last applied migration
	var lastMigration *MigrationStep
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if applied[m.migrations[i].Version] {
			lastMigration = &m.migrations[i]
			break
		}
	}

	if lastMigration == nil {
		log.Info(log.WorkerCtx, "No migrations to roll back")
		return nil
	}

	log.Infof(log.WorkerCtx, "Rolling back migration %s: %s", lastMigration.Version, lastMigration.Name)

	// Start transaction for rollback
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Run the down migration
	if err := lastMigration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("rollback %s failed: %w", lastMigration.Version, err)
	}

	// Remove the migration record
	if err := tx.Where("version = ?", lastMigration.Version).Delete(&Migration{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %s: %w", lastMigration.Version, err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", lastMigration.Version, err)
	}

	log.Infof(log.WorkerCtx, "Migration %s rolled back successfully", lastMigration.Version)
	return nil
}

// Status shows the current migration status
func (m *Migrator) Status() error {
	status, err := m.GetMigrationStatus()
	if err != nil {
		return err
	}

	log.Info(log.WorkerCtx, "Migration Status:")
	log.Info(log.WorkerCtx, "================")

	pendingCount := 0
	appliedCount := 0

	for _, ms := range status {
		statusText := "PENDING"
		timeText := ""

		if ms.Applied {
			statusText = "APPLIED"
			appliedCount++
			if ms.AppliedAt != nil {
				timeText = fmt.Sprintf(" (applied: %s)", ms.AppliedAt.Format("2006-01-02 15:04:05"))
			}
		} else {
			pendingCount++
		}

		log.Infof(log.WorkerCtx, "%s | %s | %s%s", ms.Version, statusText, ms.Name, timeText)
	}

	log.Info(log.WorkerCtx, "================")
	log.Infof(log.WorkerCtx, "Total: %d | Applied: %d | Pending: %d", len(status), appliedCount, pendingCount)

	return nil
}

// MigrationOptions provides options for controlling migration execution
type MigrationOptions struct {
	MaxMigrations int  // Maximum number of migrations to run (0 = unlimited)
	DryRun        bool // If true, show what would be migrated without applying
	Force         bool // If true, ignore safety checks
}

// UpWithOptions runs migrations with specified options
func (m *Migrator) UpWithOptions(opts MigrationOptions) error {
	if err := m.InitMigrationTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Count pending migrations
	pendingMigrations := []MigrationStep{}
	for _, migration := range m.migrations {
		if !applied[migration.Version] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		log.Info(log.WorkerCtx, "No pending migrations to apply")
		return nil
	}

	log.Infof(log.WorkerCtx, "Found %d pending migrations", len(pendingMigrations))

	log.Info(log.WorkerCtx, "Starting database migration...")

	// Apply limit if specified
	migrationsToRun := pendingMigrations
	if opts.MaxMigrations > 0 && len(pendingMigrations) > opts.MaxMigrations {
		migrationsToRun = pendingMigrations[:opts.MaxMigrations]
		log.Infof(log.WorkerCtx, "Limiting to %d migrations as requested", opts.MaxMigrations)
	}

	if opts.DryRun {
		log.Info(log.WorkerCtx, "DRY RUN - Would apply the following migrations:")
		for i, migration := range migrationsToRun {
			log.Infof(log.WorkerCtx, "%d. %s: %s", i+1, migration.Version, migration.Name)
		}
		return nil
	}

	// Apply migrations
	for i, migration := range migrationsToRun {
		log.Infof(log.WorkerCtx, "Applying migration %d/%d: %s (%s)", i+1, len(migrationsToRun), migration.Version, migration.Name)

		// Start transaction for migration
		tx := m.db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		// Run the migration
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		// Record the migration as applied
		migrationRecord := Migration{
			Version:   migration.Version,
			Name:      migration.Name,
			AppliedAt: time.Now(),
		}

		if err := tx.Create(&migrationRecord).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		log.Infof(log.WorkerCtx, "Migration %s applied successfully", migration.Version)
	}

	if len(migrationsToRun) > 0 {
		log.Infof(log.WorkerCtx, "Applied %d migrations successfully", len(migrationsToRun))
	} else {
		log.Info(log.WorkerCtx, "No pending migrations to apply")
	}

	return nil
}

// GetMigrationStatus returns detailed status of all migrations
func (m *Migrator) GetMigrationStatus() ([]MigrationStatus, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get applied migration details
	var appliedMigrations []Migration
	if err := m.db.Find(&appliedMigrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migration details: %w", err)
	}

	appliedDetails := make(map[string]Migration)
	for _, migration := range appliedMigrations {
		appliedDetails[migration.Version] = migration
	}

	var status []MigrationStatus
	for _, migration := range m.migrations {
		ms := MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: applied[migration.Version],
		}

		if applied[migration.Version] {
			if details, exists := appliedDetails[migration.Version]; exists {
				ms.AppliedAt = &details.AppliedAt
			}
		}

		status = append(status, ms)
	}

	return status, nil
}

// MigrationStatus represents the status of a single migration
type MigrationStatus struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt *time.Time
}
