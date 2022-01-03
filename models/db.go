package models

import (
	"database/sql"

	"github.com/GuiaBolso/darwin"
	"github.com/icco/gutil/logging"
	"go.uber.org/zap"

	// Needed for database connection
	_ "github.com/lib/pq"
)

var (
	db         *sql.DB
	log        = logging.Must(logging.NewLogger("cacophony"))
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table saved_urls",
			Script: `
      CREATE TABLE saved_urls (
        id serial primary key,
        link text,
        tweet_ids text[],
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     2,
			Description: "Add unique index to link",
			Script:      "CREATE UNIQUE INDEX link_idx ON saved_urls(link);",
		},
	}
)

// InitDB creates the database and migrates it to the correct version.
func InitDB(dataSourceName string) {
	// Connect to Database
	dbConn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Panicw("could not connect to DB", zap.Error(err))
	}

	if err := db.Ping(); err != nil {
		log.Panicw("could not ping DB", zap.Error(err))
	}

	log.Debug("connected to DB")
	db = dbConn

	// Migrate
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)

	if err := d.Migrate(); err != nil {
		log.Panicw("could not migrate database", zap.Error(err))
	}
	log.Info("database migration complete")
}
