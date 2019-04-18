package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8082"
	CONN_TYPE = "tcp"
)

type DeviceData struct {
	SystemCode    string
	SystemMessage uint16
}

func main() {
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
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {

	SystemCode := readNextBytes(conn, 4)
	fmt.Printf("Parsed format: %s\n", SystemCode)
	if string(SystemCode) != "MCPG" {
		fmt.Println("data not valid")
	}

	SystemMessage := readNextBytes(conn, 1)
	fmt.Printf("System Message: %s\n", SystemMessage)
	fmt.Printf("System Message: %s\n", binary.LittleEndian.Uint16(SystemMessage))

	deviceID := binary.BigEndian.Uint32(readNextBytes(conn, 4))
	fmt.Println("System Device ID")
	fmt.Println(deviceID)

	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()
}

func readNextBytes(conn net.Conn, number int) []byte {
	bytes := make([]byte, number)

	reqLen, err := conn.Read(bytes)
	if err != nil {
		if err != io.EOF {
			fmt.Println("End of file error:", err)
		}
		fmt.Println("Error reading:", err.Error(), reqLen)
	}

	return bytes
}
