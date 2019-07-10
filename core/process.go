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
		return
	}

	byteReader.Seek(5, 0)
	did := make([]byte, 4)
	byteReader.Read(did)
	deviceData.DeviceID = binary.LittleEndian.Uint32(did)
	if deviceData.DeviceID == 0 {
		return
	}

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

	deviceData.DateTime = time.Date(deviceData.UTCTimeYear, time.Month(deviceData.UTCTimeMonth), deviceData.UTCTimeDay, deviceData.UTCTimeHours, deviceData.UTCTimeMinutes, deviceData.UTCTimeSeconds, 0, time.UTC)
	deviceData.DateTimeStamp = deviceData.DateTime.Unix()
	fmt.Println(deviceData)
	if checkIdleState(deviceData) != "idle3" {
		clientJobs <- models.ClientJob{deviceData, conn}
	}

	conn.Close()

}

// check if Device is in idle state
func checkIdleState(m models.DeviceData) string {
	err := DBCONN.Ping()
	if err != nil {
		fmt.Println(err)
	}

	tx, err := DBCONN.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer tx.Rollback()

	var deviceStatus string
	squery := "SELECT device_status FROM vehicle_configuration "
	squery += " where device_id=? AND status=? LIMIT ?"
	// Execute the query
	err = tx.QueryRow(squery, strconv.FormatUint(uint64(m.DeviceID), 10), 1, 1).Scan(&deviceStatus)
	if err != nil {
		return "err"
	}

	var query string
	if m.GroundSpeed > 0 {
		query = "UPDATE vehicle_configuration SET device_status='online' WHERE device_id=? AND status=?"
	} else if m.GroundSpeed == 0 || deviceStatus == "online" {
		query = "UPDATE vehicle_configuration SET device_status='idle1' WHERE device_id=? AND status=?"
	} else if deviceStatus == "idle1" {
		query = "UPDATE vehicle_configuration SET device_status='idle2' WHERE device_id=? AND status=?"
	} else {
		query = "UPDATE vehicle_configuration SET device_status='idle3' WHERE device_id=? AND status=?"
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return "err"
	}

	_, err = stmt.Exec(strconv.FormatUint(uint64(m.DeviceID), 10), 1)
	if err != nil {
		return "err"
	}

	tx.Commit()

	return deviceStatus
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

// SaveData Save data to db
func SaveData(m models.DeviceData) {
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

	var device = strconv.FormatUint(uint64(m.DeviceID), 10)

	if m.TransmissionReason == 255 || m.GroundSpeed > 84 || m.Offline == true {
		// perform a db.Query insert
		query := "INSERT INTO trip_data (device_id, system_code, data_date, speed, speed_direction, "
		query += " longitude, latitude, altitude, satellites, hardware_version, software_version, "
		query += " transmission_reason, transmission_reason_specific_data, failsafe, disconnect, offline) "
		query += " VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

		stmt, err := tx.Prepare(query)
		if err != nil {
			fmt.Println(err)
		}

		defer stmt.Close()
		res, err := stmt.Exec(m.DeviceID, m.SystemCode, m.DateTime, m.GroundSpeed, m.SpeedDirection, m.Longitude, m.Latitude, m.Altitude, m.NoOfSatellitesUsed, m.HardwareVersion,
			m.SoftwareVersion, m.TransmissionReason, m.TransmissionReasonSpecificData, m.Failsafe, m.Disconnect, m.Offline)

		if err != nil {
			fmt.Println(err)
		}

		// update / save current violations
		//
		lid, err := res.LastInsertId()
		// check if a a vehicle id exists in the Current violation table
		var boo int8
		query = "SELECT EXISTS(SELECT 1 FROM current_violations WHERE device_id=? LIMIT 1)"
		tx.QueryRow(query, m.DeviceID).Scan(&boo)

		if boo == 1 {
			var q string
			if m.Offline == true {
				q = "UPDATE current_violations SET offline_trip_data=?, offline_trip_speed=? WHERE device_id=?"
			} else if m.GroundSpeed > 84 {
				q = "UPDATE current_violations SET overspeed_trip_data=?, overspeed_speed=? WHERE device_id=?"
			} else if m.Disconnect == true {
				q = "UPDATE current_violations SET disconnect_trip_data=?, disconnect_trip_speed=? WHERE device_id=?"
			} else if m.Failsafe == true {
				q = "UPDATE current_violations SET failsafe_trip_data=?, failsafe_trip_speed=? WHERE device_id=?"
			}
			stmt, _ := tx.Prepare(q)
			stmt.Exec(lid, m.GroundSpeed, m.DeviceID)
		} else {
			var q string
			if m.Offline {
				q = "INSERT INTO current_violations (device_id, name, offline_trip_data, offline_trip_speed) VALUES (?,?,?,?)"
			} else if m.GroundSpeed > 84 {
				q = "INSERT INTO current_violations (device_id, name, overspeed_trip_data, overspeed_speed) VALUES (?,?,?,?)"
			} else if m.Disconnect {
				q = "INSERT INTO current_violations (device_id, name, disconnect_trip_data, disconnect_trip_speed) VALUES (?,?,?,?)"
			} else if m.Failsafe {
				q = "INSERT INTO current_violations (device_id, name, failsafe_trip_data, failsafe_trip_speed) VALUES (?,?,?,?)"
			}
			stmt, _ := tx.Prepare(q)
			stmt.Exec(m.DeviceID, m.Name, lid, m.GroundSpeed)
		}

		// log data to redis
		currentViolations(m, "currentviolations:"+device)
		currentViolations(m, "currentviolations")
		SetRedisLog(m, "violations")
		SetRedisLog(m, "violations:"+device)
		SetRedisLog(m, "offline:"+device)
	}

	tx.Commit()

	lastSeen(m, "lastseen:"+device)
	lastSeen(m, "lastseen")
	// if m.TransmissionReason != 255 && m.GroundSpeed != 0 {
	SetRedisLog(m, "data:"+device)
	// }
}

type lastSeenStruct struct {
	DateTime   time.Time
	DeviceData models.DeviceData
}

func lastSeen(m models.DeviceData, key string) {
	var data = lastSeenStruct{
		DateTime:   m.DateTime,
		DeviceData: m,
	}
	// SET object
	_, err := SetValue(key, data)
	if err != nil {
		fmt.Println(err)
	}
}

func currentViolations(m models.DeviceData, key string) {
	// SET object
	_, err := SetValue(key, m)
	if err != nil {
		fmt.Println(err)
	}
}

// SetRedisLog log to redis
func SetRedisLog(m models.DeviceData, key string) {

	// SET object
	_, err := ZAdd(key, m.DateTimeStamp, m)
	if err != nil {
		fmt.Println(err)
	}
}
