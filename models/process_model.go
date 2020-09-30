package models

import (
	"net"
	"time"
)

// ClientJob ...
type ClientJob struct {
	DeviceData DeviceData
	Conn       net.Conn
}

// DeviceData ...
type DeviceData struct {
	SystemCode                     string    `json:"system_code,omitempty"`                      // 4 bytes
	SystemMessage                  int       `json:"system_message,omitempty"`                   // 1 byte
	DeviceID                       uint32    `json:"device_id,omitempty"`                        // 4 bytes
	CommunicationControlField      uint32    `json:"communication_control_field,omitempty"`      // 2 bytes
	MessageNumerator               int       `json:"message_numerator,omitempty"`                // 1 byte
	HardwareVersion                int       `json:"hardware_version,omitempty"`                 // 1 byte
	SoftwareVersion                int       `json:"software_version,omitempty"`                 // 1 byte
	ProtocolVersionIdentifier      int       `json:"protocol_version_identifier,omitempty"`      // 1 byte
	Status                         int       `json:"status,omitempty"`                           // 1 byte
	ConfigurationFlags             int       `json:"configuration_flags,omitempty"`              // 2 bytes
	TransmissionReasonSpecificData int       `json:"transmission_reason_specificData,omitempty"` // 1 byte
	Failsafe                       bool      `json:"failsafe"`
	Disconnect                     bool      `json:"disconnect"`
	Offline                        bool      `json:"offline"`
	TransmissionReason             int       `json:"transmission_reason,omitempty"` // 1 byte
	ModeOfOperation                int       `json:"mode_of_operation,omitempty"`   // 1 byte
	IOStatus                       uint16    `json:"io_status,omitempty"`           // 5 bytes
	AnalogInput1Value              uint16    `json:"analog_Input_1_value,omitempty"`
	AnalogInput1Value1             uint16    `json:"analog_Input_1_value_1,omitempty"`
	AnalogInput2Value              uint16    `json:"analog_Input_2_value,omitempty"`
	AnalogInput2Value2             uint16    `json:"analog_Input_2_value_2,omitempty"`
	MileageCounter                 uint16    `json:"mileage_counter,omitempty"` // 3 bytes
	DriverID                       uint16    `json:"driver_id,omitempty"`       // 6 bytes
	LastGPSFix                     uint16    `json:"last_gps_fix,omitempty"`
	LocationStatus                 uint16    `json:"location_status,omitempty"`
	Mode1                          uint16    `json:"mode_1,omitempty"`
	Mode2                          uint16    `json:"mode_2,omitempty"`
	NoOfSatellitesUsed             int       `json:"no_of_satellites_used,omitempty"` // 1 byte
	Longitude                      int32     `json:"longitude,omitempty"`             // 4 byte
	Latitude                       int32     `json:"latitude,omitempty"`              // 4 byte
	Altitude                       int32     `json:"altitude,omitempty"`              // 4 byte
	GroundSpeed                    uint32    `json:"ground_speed,omitempty"`          // 4 byte
	SpeedDirection                 int       `json:"speed_direction,omitempty"`       // 2 byte
	UTCTimeSeconds                 int       `json:"utc_time_seconds,omitempty"`      // 1 byte
	UTCTimeMinutes                 int       `json:"utc_time_minutes,omitempty"`      // 1 byte
	UTCTimeHours                   int       `json:"utc_time_hours,omitempty"`        // 1 byte
	UTCTimeDay                     int       `json:"utc_time_day,omitempty"`          // 1 byte
	UTCTimeMonth                   int       `json:"utc_time_month,omitempty"`        // 1 byte
	UTCTimeYear                    int       `json:"utc_time_year,omitempty"`         // 2 byte
	ErrorDetectionCode             uint16    `json:"error_detection_code,omitempty"`
	DateTime                       time.Time `json:"date_time,omitempty"`
	Name                           string    `json:"name,omitempty"`
	DateTimeStamp                  int64     `json:"date_time_stamp,omitempty"`
	Checksum                       int       `json:"checksum,omitempty"`
}

// LastSeenStruct ...
type LastSeenStruct struct {
	DateTime   time.Time
	DeviceData DeviceData
}

// AlertsDeviceData ...
type AlertsDeviceData struct {
	DeviceID           uint32    `json:"device_id"`
	DateTime           time.Time `json:"date_time"`
	Failsafe           bool      `json:"failsafe"`
	Disconnect         bool      `json:"disconnect"`
	TransmissionReason int       `json:"transmission_reason,omitempty"`
	Speed              uint32    `json:"speed,omitempty"`
}
