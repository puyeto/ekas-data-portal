package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8082"
	CONN_TYPE = "tcp"
)

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

	// Make a buffer to hold incoming data.
	// buf := make([]byte, 70)
	// // Read the incoming connection into the buffer.
	// reqLen, err := conn.Read(buf)
	// if err != nil {
	// 	if err != io.EOF {
	// 		fmt.Println("End of file error:", err)
	// 	}
	// 	fmt.Println("Error reading:", err.Error(), reqLen)
	// }

	// // Print to output
	// fmt.Println("\r\nRECVD: "+string(buf), reqLen)

	SystemCode := readNextBytes(conn, 4)
	fmt.Printf("Parsed format: %s\n", SystemCode)
	if string(SystemCode) != "MCPG" {
		fmt.Println("Provided replay file is not in correct format. Are you sure this is a SC2 replay file?")
	}

	mType := readNextBytes(conn, 1)
	fmt.Printf("Message Type: %s\n", mType)

	// var header interface{}
	// data := readNextBytes(conn, 4) //  3 * uint32 (4) + 5 * byte (1) + 22 * byte (1) = 43

	// buffer := bytes.NewBuffer(data)
	// err := binary.Read(buffer, binary.LittleEndian, &header)
	// if err != nil {
	// 	log.Fatal("binary.Read failed", err)
	// }

	// fmt.Printf("Parsed data:\n%+v\n", header)

	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()
}

func readNextBytes(conn net.Conn, number int) []byte {
	bytes := make([]byte, number)

	_, err := conn.Read(bytes)
	if err != nil {
		fmt.Println(err)
	}

	return bytes
}
