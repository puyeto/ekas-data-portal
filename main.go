package main

import (
	"bytes"
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

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			fmt.Println("Added new connection!")
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	var byteSize = 70
	for {
		message := make([]byte, 700)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("Length: " + strconv.Itoa(length))
			byteRead := bytes.NewReader(message)

			for i := 0; i < (length / byteSize); i++ {

				byteRead.Seek(int64((byteSize * i)), 0)

				mb := make([]byte, byteSize)
				n1, _ := byteRead.Read(mb)

				go core.ProcessRequest(mb, n1)
			}
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		result := "Received - Portal\n"
		client.socket.Write([]byte(string(result)))
	}
}

func startServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", ":8083")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go manager.start()
	for {
		connection, _ := listener.Accept()
		if error != nil {
			fmt.Println(error)
		}
		client := &Client{socket: connection, data: make(chan []byte)}
		manager.register <- client
		go manager.receive(client)
		go manager.send(client)
	}
}

func main() {
	startServerMode()

}

func checkError(err error) {
	if err != nil {
		fmt.Println(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
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
