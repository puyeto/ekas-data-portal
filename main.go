package main

import (
	"fmt"
	"net"
	"os"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/ekas-data-portal/models"
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
	var deviceData models.DeviceData

	deviceData.SystemCode = string(readNextBytes(conn, 4))
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid")
	}

	sm := readNextBytes(conn, 1)
	deviceData.SystemMessage = int(sm[0])
	deviceData.DeviceID = binary.LittleEndian.Uint32(readNextBytes(conn, 4))

	readNextBytes(conn, 2)	

	mn := readNextBytes(conn, 1)
	deviceData.MessageNumerator = int(mn[0])

	// HardwareVersion
	hv := readNextBytes(conn, 1)
	deviceData.HardwareVersion = int(hv[0])

	// SoftwareVersion
	sv := readNextBytes(conn, 1)
	deviceData.SoftwareVersion = int(sv[0])

	// ProtocolVersionIdentifier
	pvi := readNextBytes(conn, 1)
	deviceData.ProtocolVersionIdentifier = int(pvi[0])

	// Status
	Status := readNextBytes(conn, 1)
	deviceData.Status = int(Status[0])

	// ConfigurationFlags
	cf := readNextBytes(conn, 1)
	deviceData.ConfigurationFlags = int(cf[0])

	// TransmissionReasonSpecificData
	trsd := readNextBytes(conn, 1)
	deviceData.TransmissionReasonSpecificData = int(trsd[0])

	// TransmissionReason
	tr := readNextBytes(conn, 1)
	deviceData.TransmissionReason = int(tr[0])

	// ModeOfOperation
	moo := readNextBytes(conn, 1)
	deviceData.ModeOfOperation = int(moo[0])

	// IOStatus
	readNextBytes(conn, 5)

	// Analog Inputs
	readNextBytes(conn, 4)

	// Mileage counter
	readNextBytes(conn, 3)

	// Driver ID
	readNextBytes(conn, 6)

	// Last GPS Fix
	readNextBytes(conn, 2)

	// Location status (from unit)
	readNextBytes(conn, 1)

	// Mode 1 (from GPS)
	readNextBytes(conn, 1)

	// Mode 2 (from GPS)
	readNextBytes(conn, 1)

	// Number of satellites used (from GPS) – 1 byte
	satellites := readNextBytes(conn, 1)
	deviceData.NoOfSatellitesUsed = int(satellites[0])

	// Longitude – 4 bytes
	// deviceData.Longitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	deviceData.Longitude = readInt32(readNextBytes(conn, 4))
	//  Latitude – 4 bytes
	// deviceData.Latitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	deviceData.Latitude = readInt32(readNextBytes(conn, 4))
	// Altitude
	// deviceData.Altitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	deviceData.Altitude = readInt32(readNextBytes(conn, 4))
	// Ground speed – 4 bytes
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(readNextBytes(conn, 4))

	// Speed direction – 2 bytes
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(readNextBytes(conn, 2)))

	// UTC time – 3 bytes (hours, minutes, seconds)
	sec := readNextBytes(conn, 1)
	deviceData.UTCTimeSeconds = int(sec[0])
	min := readNextBytes(conn, 1)
	deviceData.UTCTimeMinutes = int(min[0])
	hrs := readNextBytes(conn, 1)
	deviceData.UTCTimeHours = int(hrs[0])

	// UTC date – 4 bytes (day, month, year)
	day := readNextBytes(conn, 1)
	deviceData.UTCTimeDay = int(day[0])

	mon := readNextBytes(conn, 1)
	deviceData.UTCTimeMonth = int(mon[0])

	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(readNextBytes(conn, 2)))

	fmt.Println(deviceData)

	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))

	// Close the connection when you're done with it.
	//conn.Close()
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

func readInt32(data []byte) (ret int32) {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Print("ret => ", ret)
    return ret
}