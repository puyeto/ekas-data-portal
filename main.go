package main

import (
	"fmt"
	"net"
	"os"

	// "time"

	"github.com/ekas-data-portal/core"
	"github.com/ekas-data-portal/models"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8083"
	CONN_TYPE = "tcp"
)

func init() {
	//Open the database once when the system loads
	//Do not reopen unless required as Go manages this database from here on
	//Do NOT CLOSE the db as it is ment to be long lasting
	core.DBCONN = core.DBconnect()
	err := core.InitializeRedis()
	if err != nil {
		// panic(err)
		fmt.Println(err)
	}
}

func main() {

	clientJobs := make(chan models.ClientJob)
	go generateResponses(clientJobs)

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go core.HandleRequest(conn, clientJobs)
	}
}

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", message)
}

func generateResponses(clientJobs chan models.ClientJob) {
	for {
		// Wait for the next job to come off the queue.
		clientJob := <-clientJobs

		// Do something thats keeps the CPU buys for a whole second.
		// for start := time.Now(); time.Now().Sub(start) < time.Second; {
		core.SaveData(clientJob.DeviceData)
		// }

		// Send back the response.
		clientJob.Conn.Write([]byte("Hello, " + string(clientJob.DeviceData.DeviceID)))
	}
}
