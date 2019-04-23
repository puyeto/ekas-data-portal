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

// DeviceData ...
type DeviceData struct {
	SystemCode                     string `json:"system_code,omitempty"`
	SystemMessage                  uint16 `json:"system_message,omitempty"`
	DeviceID                       uint32 `json:"device_id,omitempty"`
	CommunicationControlField      uint32 `json:"communication_control_field,omitempty"`
	MessageNumerator               uint16 `json:"message_numerator,omitempty"`
	HardwareVersion                uint16 `json:"hardware_version,omitempty"`
	SoftwareVersion                uint16 `json:"software_version,omitempty"`
	ProtocolVersionIdentifier      uint16 `json:"protocol_version_identifier,omitempty"`
	Status                         uint16 `json:"status,omitempty"`
	ConfigurationFlags             uint16 `json:"configuration_flags,omitempty"`
	TransmissionReasonSpecificData uint16 `json:"transmission_reason_specificData,omitempty"`
	TransmissionReason             uint16 `json:"transmission_reason,omitempty"`
	ModeOfOperation                uint16 `json:"mode_of_operation,omitempty"`
	IOStatus                       uint16 `json:"io_status,omitempty"`
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
	NoOfSatellitesUsed             uint16
	Longitude                      uint16
	Latitude                       uint16
	Altitude                       uint16
	GroundSpeed                    uint16
	SpeedDirection                 uint16
	UTCTimeSeconds                 uint16
	UTCTimeMinutes                 uint16
	UTCTimeHours                   uint16
	UTCTimeDay                     uint16
	UTCTimeMonth                   uint16
	UTCTimeYear                    uint16
	ErrorDetectionCode             uint16
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

	deviceData.SystemMessage = binary.ByteOrder.Uint16(readNextBytes(conn, 1))
	deviceData.DeviceID = binary.LittleEndian.Uint32(readNextBytes(conn, 4))
	// deviceData.CommunicationControlField = binary.ByteOrder.Uint32(readNextBytes(conn, 2))
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
