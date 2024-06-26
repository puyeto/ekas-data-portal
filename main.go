package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	"github.com/bamzi/jobrunner"
	"github.com/ekas-data-portal/core"
	"github.com/ekas-data-portal/cron/expireddevices"
	"github.com/ekas-data-portal/models"
	"github.com/rcrowley/go-metrics"
)

const (
	// CONNHOST connection host
	CONNHOST = "0.0.0.0"
	// CONNPORT connection port
	CONNPORT = 8083
	// CONNTYPE connection type
	CONNTYPE = "tcp"
)

var (
	startTime time.Time
	opsRate   = metrics.NewRegisteredMeter("portal-data", nil)
)

type heartbeatMessage struct {
	Status string `json:"status"`
	Build  string `json:"build"`
	Uptime string `json:"uptime"`
}

func init() {
	startTime = time.Now()

	core.InitLogger()

	//Open the database once when the system loads
	//Do not reopen unless required as Go manages this database from here on
	//Do NOT CLOSE the db as it is ment to be long lasting
	core.DBCONN = core.DBconnect("ekas_portal")
	core.DBCONDATA = core.DBconnect("ekas_portal_data")
	core.MongoDB = core.InitializeMongoDB("mongodb://root:safcom2012@144.76.140.105:27017/?authSource=admin", "ekas_portal")
	core.InitializeRedis()
}

func main() {
	time.Now().UnixNano()
	// setLimit()
	// go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	jobrunner.Schedule("@every 30m", expireddevices.Status{})
	jobrunner.In(2*time.Second, expireddevices.Status{})

	go runHeartbeatService(":7001")

	// ticker := time.NewTicker(10 * time.Minute)
	// go func() {
	// 	for range ticker.C {
	// 		checkLastSeen()
	// 	}
	// }()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(CONNPORT))
	checkError(err)
	// Listen for incoming connections.
	l, err := net.ListenTCP(CONNTYPE, tcpAddr)
	if err != nil {
		panic(err)
	}

	defer l.Close()
	rand.Seed(time.Now().Unix())

	// fmt.Println("Listening on " + CONNHOST + ":" + strconv.Itoa(CONNPORT))
	core.Logger.Infoln("Listening on " + CONNHOST + ":" + strconv.Itoa(CONNPORT))

	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				fmt.Printf("accept temp err: %v", ne)
				continue
			}

			fmt.Printf("accept err: %v", err)
			break
		}

		// log.Println("Client ", conn.RemoteAddr(), " connected")
		// Handle connections in a new goroutine.
		go HandleRequest(conn)

		// Create a goroutine that closes a session after 15 seconds
		// go func() {
		// 	<-time.After(time.Duration(10) * time.Second)
		// 	defer conn.Close()
		// }()

		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(c, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}

}

func checkError(err error) {
	if err != nil {
		core.Logger.Errorf("Fatal error: %v", err.Error())
		return
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
	core.Logger.Errorf("%v\n", message)
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
			if CallTime(value) > 1440 {
				value.Offline = true
				SaveData(value)
			}
		}
	}
}

func CallTime(m models.DeviceData) int {
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
