package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ekas-data-portal/models"
)

// HandleRequest Handles incoming requests.
func HandleRequest(conn net.Conn, clientJobs chan models.ClientJob) {
	var byteSize = 70
	totalBytes, res := readNextBytes(conn, 1024)
	if totalBytes > 0 {
		byteRead := bytes.NewReader(res)
		for i := 0; i < (totalBytes / byteSize); i++ {
			if i > 0 {
				byteRead.Seek(int64((byteSize * i)), 0)
			}

			mb := make([]byte, byteSize)
			n1, _ := byteRead.Read(mb)

			processRequest(conn, mb, n1, clientJobs, i)
		}
	}
}

func processRequest(conn net.Conn, b []byte, byteLen int, clientJobs chan models.ClientJob, i int) {
	var deviceData models.DeviceData

	if byteLen != 70 {
		fmt.Println(i, " - Invalid Byte Length = ", byteLen)
		return
	}

	// fmt.Println(time.Now(), " data ", string(b))

	byteReader := bytes.NewReader(b)

	scode := make([]byte, 4)
	byteReader.Read(scode)
	deviceData.SystemCode = string(scode)
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid")
	}

	byteReader.Seek(5, 0)
	did := make([]byte, 4)
	byteReader.Read(did)
	deviceData.DeviceID = binary.LittleEndian.Uint32(did)

	// Transmission Reason – 1 byte
	byteReader.Seek(18, 0)
	reason := make([]byte, 1)
	byteReader.Read(reason)
	deviceData.TransmissionReason = int(reason[0])

	// Transmission Reason Specific data – 1 byte
	trsd := 0
	if deviceData.TransmissionReason == 255 {
		byteReader.Seek(17, 0)
		specific := make([]byte, 1)
		byteReader.Read(specific)

		var a = int8(specific[0])
		fmt.Printf("%08b\n", a)
		fmt.Printf("%08b\n", a&1<<1)
		fmt.Printf("%08b\n", a&1<<2)
		fmt.Printf("%08b\n", a&1<<3)

		trsd = int(a << 1)
	}
	deviceData.TransmissionReasonSpecificData = trsd

	// Number of satellites used (from GPS) – 1 byte
	byteReader.Seek(43, 0)
	satellites := make([]byte, 1)
	byteReader.Read(satellites)
	deviceData.NoOfSatellitesUsed = int(satellites[0])

	// Longitude – 4 bytes
	byteReader.Seek(44, 0)
	long := make([]byte, 4)
	byteReader.Read(long)
	deviceData.Longitude = readInt32(long)

	//  Latitude – 4 bytes
	byteReader.Seek(48, 0)
	lat := make([]byte, 4)
	byteReader.Read(lat)
	deviceData.Latitude = readInt32(lat)

	// Altitude
	byteReader.Seek(52, 0)
	alt := make([]byte, 4)
	byteReader.Read(alt)
	deviceData.Altitude = readInt32(alt)

	// Ground speed – 4 bytes
	byteReader.Seek(56, 0)
	gspeed := make([]byte, 4)
	byteReader.Read(gspeed)
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(gspeed)

	// Speed direction – 2 bytes
	byteReader.Seek(60, 0)
	speedd := make([]byte, 2)
	byteReader.Read(speedd)
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(speedd))

	// UTC time – 3 bytes (hours, minutes, seconds)
	byteReader.Seek(62, 0)
	sec := make([]byte, 1)
	byteReader.Read(sec)
	deviceData.UTCTimeSeconds = int(sec[0])

	byteReader.Seek(63, 0)
	min := make([]byte, 1)
	byteReader.Read(min)
	deviceData.UTCTimeMinutes = int(min[0])

	byteReader.Seek(64, 0)
	hrs := make([]byte, 1)
	byteReader.Read(hrs)
	deviceData.UTCTimeHours = int(hrs[0])

	// UTC date – 4 bytes (day, month, year)
	byteReader.Seek(65, 0)
	day := make([]byte, 1)
	byteReader.Read(day)
	deviceData.UTCTimeDay = int(day[0])

	byteReader.Seek(66, 0)
	mon := make([]byte, 1)
	byteReader.Read(mon)
	deviceData.UTCTimeMonth = int(mon[0])

	byteReader.Seek(67, 0)
	yr := make([]byte, 2)
	byteReader.Read(yr)
	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(yr))

	fmt.Println(deviceData)
	clientJobs <- models.ClientJob{deviceData, conn}
	conn.Close()

}

func handleRequest2(conn net.Conn, clientJobs chan models.ClientJob) {
	var deviceData models.DeviceData

	_, scode := readNextBytes(conn, 4)
	deviceData.SystemCode = string(scode)
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid")
	}

	_, sm := readNextBytes(conn, 1)
	deviceData.SystemMessage = int(sm[0])

	_, did := readNextBytes(conn, 4)
	deviceData.DeviceID = binary.LittleEndian.Uint32(did)

	readNextBytes(conn, 2)

	_, mn := readNextBytes(conn, 1)
	deviceData.MessageNumerator = int(mn[0])

	// HardwareVersion
	_, hv := readNextBytes(conn, 1)
	deviceData.HardwareVersion = int(hv[0])

	// SoftwareVersion
	_, sv := readNextBytes(conn, 1)
	deviceData.SoftwareVersion = int(sv[0])

	// ProtocolVersionIdentifier
	_, pvi := readNextBytes(conn, 1)
	deviceData.ProtocolVersionIdentifier = int(pvi[0])

	// Status
	_, Status := readNextBytes(conn, 1)
	deviceData.Status = int(Status[0])

	// ConfigurationFlags
	_, cf := readNextBytes(conn, 1)
	deviceData.ConfigurationFlags = int(cf[0])

	// TransmissionReasonSpecificData
	_, trsd := readNextBytes(conn, 1)
	deviceData.TransmissionReasonSpecificData = int(trsd[0])

	// TransmissionReason
	_, tr := readNextBytes(conn, 1)
	deviceData.TransmissionReason = int(tr[0])

	// ModeOfOperation
	_, moo := readNextBytes(conn, 1)
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
	_, satellites := readNextBytes(conn, 1)
	deviceData.NoOfSatellitesUsed = int(satellites[0])

	// Longitude – 4 bytes
	_, long := readNextBytes(conn, 4)
	deviceData.Longitude = readInt32(long)
	//  Latitude – 4 bytes
	_, lat := readNextBytes(conn, 4)
	deviceData.Latitude = readInt32(lat)
	// Altitude
	_, alt := readNextBytes(conn, 4)
	deviceData.Altitude = readInt32(alt)
	// Ground speed – 4 bytes
	_, gspeed := readNextBytes(conn, 4)
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(gspeed)

	// Speed direction – 2 bytes
	_, speedd := readNextBytes(conn, 2)
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(speedd))

	// UTC time – 3 bytes (hours, minutes, seconds)
	_, sec := readNextBytes(conn, 1)
	deviceData.UTCTimeSeconds = int(sec[0])
	_, min := readNextBytes(conn, 1)
	deviceData.UTCTimeMinutes = int(min[0])
	_, hrs := readNextBytes(conn, 1)
	deviceData.UTCTimeHours = int(hrs[0])

	// UTC date – 4 bytes (day, month, year)
	_, day := readNextBytes(conn, 1)
	deviceData.UTCTimeDay = int(day[0])

	_, mon := readNextBytes(conn, 1)
	deviceData.UTCTimeMonth = int(mon[0])

	_, yr := readNextBytes(conn, 2)
	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(yr))

	// Send a response back to person contacting us.
	// conn.Write([]byte("Message received."))

	// Close the connection when you're done with it.
	conn.Close()

	// Save data
	// if deviceData.DeviceID > 0 || deviceData.SystemCode == "MCPG" {
	//	saveData(deviceData)
	// }

	clientJobs <- models.ClientJob{deviceData, conn}
}

func readNextBytes(conn net.Conn, number int) (int, []byte) {
	bytes := make([]byte, number)

	reqLen, err := conn.Read(bytes)
	if err != nil {
		if err != io.EOF {
			fmt.Println("End of file error:", err)
		}
		fmt.Println("Error reading:", err.Error(), reqLen)
	}

	return reqLen, bytes
}

func readInt32(data []byte) (ret int32) {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	return ret
}

//SaveData Save data to db
func SaveData(m models.DeviceData) {

	fmt.Println(m)
	err := DBCONN.Ping()
	if err != nil {
		fmt.Println(err)
	}

	tx, err := DBCONN.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer tx.Rollback()

	// strDate := string(m.UTCTimeYear) + "-" + string(m.UTCTimeMonth) + "-" + string(m.UTCTimeDay) // + " " + string(m.UTCTimeHours) + ":" + string(m.UTCTimeMinutes) + ":" + string(m.UTCTimeSeconds)
	t := time.Date(m.UTCTimeYear, time.Month(m.UTCTimeMonth), m.UTCTimeDay, m.UTCTimeHours, m.UTCTimeMinutes, m.UTCTimeSeconds, 0, time.UTC)
	// t, err:= time.Parse(time.RFC3339, strDate)
	if err != nil {
		fmt.Println(err)
	}

	// perform a db.Query insert
	query := "INSERT INTO trip_data (device_id, system_code, data_date, speed, speed_direction, "
	query += " longitude, latitude, altitude, satellites, hardware_version, software_version, "
	query += " transmission_reason, transmission_reason_specific_data) "
	query += " VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"

	stmt, err := tx.Prepare(query)
	if err != nil {
		panic(err.Error())
	}

	defer stmt.Close()
	_, err = stmt.Exec(m.DeviceID, m.SystemCode, t, m.GroundSpeed, m.SpeedDirection, m.Longitude, m.Latitude, m.Altitude, m.NoOfSatellitesUsed, m.HardwareVersion, m.SoftwareVersion, m.TransmissionReason, m.TransmissionReasonSpecificData)

	if err != nil {
		fmt.Println(err)
	}
	tx.Commit()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
