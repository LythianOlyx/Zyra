package scaffold

// Database identifies the database engine selected during `zyra create`'s
// interactive prompt. Only Postgres, MySQL and SQLite are actually wired
// into Zyra's data layer (internal/data) today — MongoDB, Firebase and
// Supabase are listed (matching zyraStrategy/10-CLI-AND-PROJECT-TEMPLATES.md's
// documented prompt) so the CLI's UX matches the design doc, but selecting
// one of them currently degrades gracefully to DatabaseSkip: no driver/url
// is configured, and the generated project simply runs without a database
// until first-class support ships.
type Database string

const (
	DatabasePostgres Database = "postgres"
	DatabaseMySQL    Database = "mysql"
	DatabaseSQLite   Database = "sqlite"
	DatabaseMongoDB  Database = "mongodb"
	DatabaseFirebase Database = "firebase"
	DatabaseSupabase Database = "supabase"
	DatabaseSkip     Database = "skip"
)

// SupportedDatabases lists every database chosen presented by the
// interactive prompt, in display order.
func SupportedDatabases() []Database {
	return []Database{
		DatabasePostgres,
		DatabaseMySQL,
		DatabaseSQLite,
		DatabaseMongoDB,
		DatabaseFirebase,
		DatabaseSupabase,
		DatabaseSkip,
	}
}

// wired reports whether d has real internal/data driver support today.
func (d Database) wired() bool {
	switch d {
	case DatabasePostgres, DatabaseMySQL, DatabaseSQLite:
		return true
	default:
		return false
	}
}

// driverAndURL returns the zyra.config.json "database.driver"/"database.url"
// values for d. The second return value is empty for drivers whose real
// connection string must come from the DATABASE_URL environment variable
// (see .env.example in every template) rather than being hardcoded.
func (d Database) driverAndURL() (driver, url string) {
	if !d.wired() {
		return "", ""
	}
	if d == DatabaseSQLite {
		return "sqlite", "zyra.db"
	}
	return string(d), ""
}
