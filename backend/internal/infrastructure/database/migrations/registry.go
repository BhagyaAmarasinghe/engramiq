package migrations

// GetAllMigrations returns all available migrations in order
func GetAllMigrations() []Migration {
	return []Migration{
		CreatePopulateSiteDataMigration(),
		// Add future migrations here in chronological order
	}
}