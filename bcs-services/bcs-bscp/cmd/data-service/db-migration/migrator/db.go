package migrator

import (
	"database/sql"
	"fmt"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/sharding"
)

// NewDB new a db instance
func NewDB() (*sql.DB, error) {

	fmt.Println("Connecting to MySQL database...")

	dbConf := cc.DataService().Sharding.AdminDatabase
	db, err := sql.Open("mysql", sharding.URI(dbConf))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database, err: %s", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database, err: %s", err)
	}

	fmt.Println("Database connected!")

	return db, nil
}
