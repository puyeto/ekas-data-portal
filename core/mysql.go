package core

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	// ...
	_ "github.com/go-sql-driver/mysql"
)

var (
	// DBCONN ...
	DBCONN *sql.DB
	// DBCONDATA for data store.
	DBCONDATA     *sql.DB
	mysqlUsername = "remote"
	mysqlPassword = "2030-Ekas12"
	mysqlIP       = "178.62.75.148"
	mysqlPort     = 3306
)

// DBconnect Initialise a database connection
func DBconnect(dbname string) *sql.DB {

	//Construct the host
	//Note: Values are set using a config file
	mysqlHost := mysqlUsername + ":" + mysqlPassword + "@tcp(" + mysqlIP + ":" + strconv.Itoa(mysqlPort) + ")/" + dbname + "?parseTime=true&net_write_timeout=6000"

	db, err := sql.Open("mysql", mysqlHost)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 10)

	return db
}
