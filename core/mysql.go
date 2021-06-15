package core

import (
	"database/sql"
	"log"
	"strconv"

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
	mysqlIP       = "138.197.205.177"
	mysqlPort     = 3306
)

// DBconnect Initialise a database connection
func DBconnect(dbname string) *sql.DB {

	//Construct the host
	//Note: Values are set using a config file
	mysqlHost := mysqlUsername + ":" + mysqlPassword + "@tcp(" + mysqlIP + ":" + strconv.Itoa(mysqlPort) + ")/" + dbname + "?parseTime=true"

	db, err := sql.Open("mysql", mysqlHost)
	if err != nil {
		log.Fatal(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(100)

	return db
}
