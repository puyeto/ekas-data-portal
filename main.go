package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

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

	ticker := time.NewTicker(50000 * time.Millisecond)
	go func() {
		for range ticker.C {
			checkLastSeen()
		}
	}()

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

		// Do something thats keeps the CPU busy for a whole second.
		// for start := time.Now(); time.Now().Sub(start) < time.Second; {
		core.SaveData(clientJob.DeviceData)
		// }

		// Send back the response.
		clientJob.Conn.Write([]byte("Hello, " + string(clientJob.DeviceData.DeviceID)))
	}
}

func checkLastSeen() {
	keysList, err := core.ListKeys("lastseen:*")
	if err != nil {
		fmt.Println("Getting Keys Failed : " + err.Error())
	}

	for i := 0; i < len(keysList); i++ {
		fmt.Println("Getting " + keysList[i])
		value, err := core.GetLastSeenValue(keysList[i])
		if err != nil {
			return
		}
		if value.SystemCode == "MCPG" {
			if callTime(value) >= 5 {
				fmt.Println("device_id", value.DeviceID)
				value.Offline = true
				core.SaveData(value)
				var device = strconv.FormatUint(uint64(value.DeviceID), 10)
				core.SetRedisLog(value, "violations")
				core.SetRedisLog(value, "violations:"+device)
				core.SetRedisLog(value, "offline:"+device)
			}
		}
	}
}

func callTime(m models.DeviceData) int {
	nowd := time.Now()
	now := dateF(nowd.Year(), nowd.Month(), nowd.Day(), nowd.Hour(), nowd.Minute(), nowd.Second())
	pastDate := dateF(m.UTCTimeYear, time.Month(m.UTCTimeMonth), m.UTCTimeDay, m.UTCTimeHours, m.UTCTimeMinutes, m.UTCTimeSeconds)
	diff := now.Sub(pastDate)

	mins := int(diff.Minutes())
	fmt.Println("mins = ", mins)
	return mins
}

func dateF(year int, month time.Month, day int, hr, min, sec int) time.Time {
	return time.Date(year, month, day, hr, min, sec, 0, time.UTC)
}
