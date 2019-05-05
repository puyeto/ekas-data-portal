package core

import (
	"database/sql"
	"strconv"
	"log"

	// ...
	_ "github.com/go-sql-driver/mysql"
)

var (
	// DBCONN ...	
	DBCONN *sql.DB
	mysqlUsername = "root"
	mysqlPassword = ""
	mysqlIP = "" 
	mysqlDB = "ekas_portal"
	mysqlPort  = 3306
)

const (
	//Keeping a connection idle for a long time can cause problems
	//http://go-database-sql.org/connection-pool.html
	maxIdleConns = 0

	driverName = "mysql"
)

// DBconnect Initialise a database connection
func DBconnect() *sql.DB {

	//Construct the host
	//Note: Values are set using a config file
	mysqlHost := mysqlUsername + ":" + mysqlPassword + "@tcp(" + mysqlIP + ":" + strconv.Itoa(mysqlPort) + ")/" + mysqlDB + "?parseTime=true"

	db, err := sql.Open(driverName, mysqlHost)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(maxIdleConns)

	return db
}
