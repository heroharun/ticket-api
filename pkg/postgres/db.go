package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // load postgres driver
	"log"
	"os"
)

func CreateConnection() *sql.DB {
	connStr := fmt.Sprintf(os.Getenv("DATABASE_URL"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
