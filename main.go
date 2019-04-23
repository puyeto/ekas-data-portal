package main

import (
	"bytes"
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

// DeviceData ...
type DeviceData struct {
	SystemCode                     string `json:"system_code,omitempty"`                 // 4 bytes
	SystemMessage                  int    `json:"system_message,omitempty"`              // 1 byte
	DeviceID                       uint32 `json:"device_id,omitempty"`                   // 4 bytes
	CommunicationControlField      uint32 `json:"communication_control_field,omitempty"` // 2 bytes
	MessageNumerator               int    `json:"message_numerator,omitempty"`           // 1 byte
	HardwareVersion                int    `json:"hardware_version,omitempty"`            // 1 byte
	SoftwareVersion                int    `json:"software_version,omitempty"`            // 1 byte
	ProtocolVersionIdentifier      uint16 `json:"protocol_version_identifier,omitempty"` // 1 byte
	Status                         uint16 `json:"status,omitempty"`                      // 1 byte
	ConfigurationFlags             uint16 `json:"configuration_flags,omitempty"`         // 2 bytes
	TransmissionReasonSpecificData uint16 `json:"transmission_reason_specificData,omitempty"`
	TransmissionReason             uint16 `json:"transmission_reason,omitempty"` // 1 byte
	ModeOfOperation                uint16 `json:"mode_of_operation,omitempty"`   // 1 byte
	IOStatus                       uint16 `json:"io_status,omitempty"`           // 5 bytes
	AnalogInput1Value              uint16 `json:"analog_Input_1_value,omitempty"`
	AnalogInput1Value1             uint16 `json:"analog_Input_1_value_1,omitempty"`
	AnalogInput2Value              uint16 `json:"analog_Input_2_value,omitempty"`
	AnalogInput2Value2             uint16 `json:"analog_Input_2_value_2,omitempty"`
	MileageCounter                 uint16 `json:"mileage_counter,omitempty"`
	DriverID                       uint16 `json:"driver_id,omitempty"`
	LastGPSFix                     uint16 `json:"last_gps_fix,omitempty"`
	LocationStatus                 uint16 `json:"location_status,omitempty"`
	Mode1                          uint16 `json:"mode_1,omitempty"`
	Mode2                          uint16 `json:"mode_2,omitempty"`
	NoOfSatellitesUsed             uint16 `json:"no_of_satellites_used,omitempty"`
	Longitude                      uint16 `json:"longitude,omitempty"`
	Latitude                       uint16 `json:"latitude,omitempty"`
	Altitude                       uint16 `json:"altitude,omitempty"`
	GroundSpeed                    uint16 `json:"ground_speed,omitempty"`
	SpeedDirection                 uint16 `json:"speed_direction,omitempty"`
	UTCTimeSeconds                 uint16 `json:"utc_time_seconds,omitempty"`
	UTCTimeMinutes                 uint16 `json:"utc_time_minutes,omitempty"`
	UTCTimeHours                   uint16 `json:"utc_time_hours,omitempty"`
	UTCTimeDay                     uint16 `json:"utc_time_day,omitempty"`
	UTCTimeMonth                   uint16 `json:"utc_time_month,omitempty"`
	UTCTimeYear                    uint16 `json:"utc_time_year,omitempty"`
	ErrorDetectionCode             uint16 `json:"error_detection_code,omitempty"`
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
	var deviceData DeviceData

	deviceData.SystemCode = string(readNextBytes(conn, 4))
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid")
	}

	// deviceData.SystemMessage = int(readNextBytes(conn, 1))
	buf := bytes.NewBuffer(readNextBytes(conn, 1)) // b is []byte
	sm, err := binary.ReadVarint(buf)
	if err != nil {
		fmt.Println("Error reading SystemMessage:", err.Error())
	}
	deviceData.SystemMessage = int(sm)
	deviceData.DeviceID = binary.LittleEndian.Uint32(readNextBytes(conn, 4))

	countries := readNextBytes(conn, 2)
	space := []byte{'\'}
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
