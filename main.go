package main

import "bufio"
import "fmt"
import "log"
import "net"
import "strings" // only needed below for sample processing

func main() {
	fmt.Println("Launching server...")
	fmt.Println("Listen on port")
	ln, err := net.Listen("tcp", "127.0.0.1:8082")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("Accept connection on port")
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Entering loop")
	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// output message received
		fmt.Print("Message Received:", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
}
