package models

import (
	"database/sql"

	"github.com/GuiaBolso/darwin"
	sd "github.com/icco/logrus-stackdriver-formatter"

	// Needed for database connection
	_ "github.com/lib/pq"
)

var (
	db         *sql.DB
	log        = sd.InitLogging()
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
	var err error

	// Connect to Database
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.WithError(err).Panic("could not connect to DB")
	}

	if err = db.Ping(); err != nil {
		log.WithError(err).Panic("could not ping DB")
	}

	log.Debug("connected to DB")

	// Migrate
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()
	if err != nil {
		log.WithError(err).Panic("could not migrate database")
	}
	log.Info("database migration complete")
}
