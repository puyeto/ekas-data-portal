package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ekas-data-portal/core"
	"github.com/ekas-data-portal/models"
	"github.com/pkg/profile"
)

const (
	CONNHOST = "0.0.0.0"
	CONNPORT = 8083
	CONNTYPE = "tcp"
)

var startTime time.Time

type heartbeatMessage struct {
	Status string `json:"status"`
	Build  string `json:"build"`
	Uptime string `json:"uptime"`
}

func init() {
	startTime = time.Now()

	//Open the database once when the system loads
	//Do not reopen unless required as Go manages this database from here on
	//Do NOT CLOSE the db as it is ment to be long lasting
	core.DBCONN = core.DBconnect("ekas_portal")
	core.DBCONDATA = core.DBconnect("ekas_portal_data")
	err := core.InitializeRedis()
	if err != nil {
		// panic(err)
		fmt.Println(err)
	}
}

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	time.Now().UnixNano()

	// ticker := time.NewTicker(5 * time.Minute)
	// go func() {
	// 	for range ticker.C {
	// 		checkLastSeen()
	// 	}
	// }()

	// Listen for incoming connections.
	l, err := net.ListenTCP(CONNTYPE, &net.TCPAddr{Port: CONNPORT})
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + CONNHOST + ":" + strconv.Itoa(CONNPORT))

	go runHeartbeatService(":7001")

	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}

		// Handle connections in a new goroutine.
		go core.HandleRequest(conn)
	}

}

func handler(rw http.ResponseWriter, r *http.Request) {
	enableCors(&rw)
	s := rand.New(rand.NewSource(99))
	hash := s.Int()

	uptime := time.Since(startTime).String()
	err := json.NewEncoder(rw).Encode(heartbeatMessage{"OK", strconv.Itoa(hash), uptime})
	if err != nil {
		log.Fatalf("Failed to write heartbeat message. Reason: %s", err.Error())
	}
}

func runHeartbeatService(address string) {
	http.HandleFunc("/heartbeat", handler)
	log.Println(http.ListenAndServe(address, nil))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", message)
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
			if callTime(value) > 1440 {
				value.Offline = true
				core.SaveData(value)
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
