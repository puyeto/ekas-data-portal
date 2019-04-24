package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/ekas-data-portal/models"
)

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	var deviceData models.DeviceData

	deviceData.SystemCode = string(readNextBytes(conn, 4))
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid")
	}

	// deviceData.SystemMessage = int(readNextBytes(conn, 1))
	sm, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading SystemMessage:", err.Error())
	}
	deviceData.SystemMessage = int(sm)
	deviceData.DeviceID = binary.LittleEndian.Uint32(readNextBytes(conn, 4))

	countries := readNextBytes(conn, 2)
	space := []byte{'\\'}
	splitExample := bytes.Split(countries, space)
	fmt.Printf("\nSplit split %q on a single space:", countries)
	for index, element := range splitExample {
		fmt.Printf("\n%d => %q", index, element)
	}

	deviceData.CommunicationControlField = 1

	mn, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading MessageNumerator:", err.Error())
	}
	deviceData.MessageNumerator = int(mn)

	// HardwareVersion
	hv, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading HardwareVersion:", err.Error())
	}
	deviceData.HardwareVersion = int(hv)

	// SoftwareVersion
	sv, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading SoftwareVersion:", err.Error())
	}
	deviceData.SoftwareVersion = int(sv)

	// ProtocolVersionIdentifier
	pvi, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading ProtocolVersionIdentifier:", err.Error())
	}
	deviceData.ProtocolVersionIdentifier = int(pvi)

	// Status
	Status, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading Status:", err.Error())
	}
	deviceData.Status = int(Status)

	// ConfigurationFlags
	cf, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading ConfigurationFlags:", err.Error())
	}
	deviceData.ConfigurationFlags = int(cf)

	// TransmissionReasonSpecificData
	trsd, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading TransmissionReasonSpecificData:", err.Error())
	}
	deviceData.TransmissionReasonSpecificData = int(trsd)

	// TransmissionReason
	tr, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading TransmissionReason:", err.Error())
	}
	deviceData.TransmissionReason = int(tr)

	// ModeOfOperation
	moo, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading ModeOfOperation:", err.Error())
	}
	deviceData.ModeOfOperation = int(moo)

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
	satellites, err := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	if err != nil {
		fmt.Println("Error reading Number of satellites :", err.Error())
	}
	deviceData.NoOfSatellitesUsed = int(satellites)

	// Longitude – 4 bytes
	deviceData.Longitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	//  Latitude – 4 bytes
	deviceData.Latitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	// Altitude
	deviceData.Altitude = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	// Ground speed – 4 bytes
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(readNextBytes(conn, 4))

	// Speed direction – 2 bytes
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(readNextBytes(conn, 2)))

	// UTC time – 3 bytes (hours, minutes, seconds)
	sec, _ := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	deviceData.UTCTimeSeconds = int(sec)
	min, _ := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	deviceData.UTCTimeMinutes = int(min)
	hrs, _ := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	deviceData.UTCTimeHours = int(hrs)

	// UTC date – 4 bytes (day, month, year)
	day, _ := binary.ReadVarint(bytes.NewBuffer(readNextBytes(conn, 1)))
	deviceData.UTCTimeDay = int(day)

	buffer := bytes.NewBuffer(readNextBytes(conn, 1))
	month, _ := binary.ReadVarint(buffer)
	deviceData.UTCTimeMonth = int(month)

	var m uint32
	err = binary.Read(buffer, binary.BigEndian, &m)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}

	fmt.Println(m)

	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(readNextBytes(conn, 2)))

	fmt.Println(deviceData)

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
