package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
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

		var a = int(specific[0])
		// Failsafe
		failsafe := hasBit(a, 1)
		deviceData.Failsafe = failsafe
		// main power disconnected
		disconnect := hasBit(a, 2)
		deviceData.Disconnect = disconnect
		trsd = int(a)
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

func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
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

	squery := "SELECT vehicle_reg_no FROM vehicle_configuration "
	squery += " INNER JOIN vehicle_details ON (vehicle_details.vehicle_id = vehicle_configuration.vehicle_id) "
	squery += " where data->'$.governor_details.device_id'=? AND status=? LIMIT ?"
	// Execute the query
	err = tx.QueryRow(squery, strconv.FormatUint(uint64(m.DeviceID), 10), 1, 1).Scan(&m.Name)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(m.Name)

	// strDate := string(m.UTCTimeYear) + "-" + string(m.UTCTimeMonth) + "-" + string(m.UTCTimeDay) // + " " + string(m.UTCTimeHours) + ":" + string(m.UTCTimeMinutes) + ":" + string(m.UTCTimeSeconds)
	t := time.Date(m.UTCTimeYear, time.Month(m.UTCTimeMonth), m.UTCTimeDay, m.UTCTimeHours, m.UTCTimeMinutes, m.UTCTimeSeconds, 0, time.UTC)
	// t, err:= time.Parse(time.RFC3339, strDate)
	if err != nil {
		fmt.Println(err)
	}
	m.DateTime = t

	// perform a db.Query insert
	query := "INSERT INTO trip_data (device_id, system_code, data_date, speed, speed_direction, "
	query += " longitude, latitude, altitude, satellites, hardware_version, software_version, "
	query += " transmission_reason, transmission_reason_specific_data, failsafe, disconnect) "
	query += " VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	stmt, err := tx.Prepare(query)
	if err != nil {
		fmt.Println(err)
	}

	defer stmt.Close()
	_, err = stmt.Exec(m.DeviceID, m.SystemCode, t, m.GroundSpeed, m.SpeedDirection, m.Longitude, m.Latitude, m.Altitude, m.NoOfSatellitesUsed, m.HardwareVersion,
		m.SoftwareVersion, m.TransmissionReason, m.TransmissionReasonSpecificData, m.Failsafe, m.Disconnect)

	if err != nil {
		fmt.Println(err)
	}
	tx.Commit()
	// log data to redis
	lastSeen(m)
	setRedisLog(t, m)
	if m.TransmissionReason == 255 || m.GroundSpeed > 80 {
		currentViolations(m)
	}
}

type lastSeenStruct struct {
	DateTime   time.Time
	DeviceData models.DeviceData
}

func lastSeen(m models.DeviceData) {
	const lastSeenPrefix string = "lastseen:"
	var device = strconv.FormatUint(uint64(m.DeviceID), 10)
	var data = lastSeenStruct{
		DateTime:   m.DateTime,
		DeviceData: m,
	}
	// SET object
	_, err := SetValue(lastSeenPrefix+device, data)
	if err != nil {
		fmt.Println(err)
	}
}

func currentViolations(m models.DeviceData) {
	const cvPrefix string = "currentviolations:"
	var device = strconv.FormatUint(uint64(m.DeviceID), 10)
	// SET object
	_, err := SetValue(cvPrefix+device, m)
	if err != nil {
		fmt.Println(err)
	}
}

func setRedisLog(t time.Time, m models.DeviceData) {
	const dataPrefix string = "data:"
	var device = strconv.FormatUint(uint64(m.DeviceID), 10)

	// SET object
	_, err := SAdd(dataPrefix+device, m)
	if err != nil {
		fmt.Println(err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
