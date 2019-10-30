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
	mysqlUsername = "doadmin"
	mysqlPassword = "vnz5itaj19jqco1n"
	mysqlIP       = "db-mysql-cluster-do-user-4666162-0.db.ondigitalocean.com"
	mysqlPort     = 25060
)

const (
	//Keeping a connection idle for a long time can cause problems
	//http://go-database-sql.org/connection-pool.html
	maxIdleConns = 0

	driverName = "mysql"
)

// DBconnect Initialise a database connection
func DBconnect(dbname string) *sql.DB {

	//Construct the host
	//Note: Values are set using a config file
	mysqlHost := mysqlUsername + ":" + mysqlPassword + "@tcp(" + mysqlIP + ":" + strconv.Itoa(mysqlPort) + ")/" + dbname + "?parseTime=true"

	db, err := sql.Open(driverName, mysqlHost)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(maxIdleConns)

	return db
}
