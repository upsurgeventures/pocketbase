package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// Adds _collections.rowid for catalogs that already applied 1640988000_init.go
// before the BIGSERIAL column was added to the CREATE TABLE.
// Fresh databases get the column from init; IF NOT EXISTS makes this a no-op there.
//
// Must run before 1763020353 (that migration calls FindAllCollections, which
// ORDER BY rowid).
func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		_, err := txApp.DB().NewQuery(`
			ALTER TABLE {{_collections}} ADD COLUMN IF NOT EXISTS [[rowid]] BIGSERIAL
		`).Execute()
		if err != nil {
			return fmt.Errorf("add _collections.rowid: %w", err)
		}
		return nil
	}, nil)
}
