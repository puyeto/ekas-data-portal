package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ekas-data-portal/models"
)

// HandleRequest Handles incoming requests.
func HandleRequest(conn net.Conn, clientJobs chan models.ClientJob) {

	var byteSize = 70
	totalBytes, res := readNextBytes(conn, 700)

	// return Response
	result := "Received byte size = " + strconv.Itoa(totalBytes) + "\n"
	conn.Write([]byte(string(result)))

	if totalBytes > 0 {
		byteRead := bytes.NewReader(res)

		for i := 0; i < (totalBytes / byteSize); i++ {

			byteRead.Seek(int64((byteSize * i)), 0)

			mb := make([]byte, byteSize)
			n1, _ := byteRead.Read(mb)

			go processRequest(conn, mb, n1, clientJobs)
		}

	}
}

func processRequest(conn net.Conn, b []byte, byteLen int, clientJobs chan models.ClientJob) {
	var deviceData models.DeviceData

	if byteLen != 70 {
		fmt.Println("Invalid Byte Length = ", byteLen)
		return
	}

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

	fmt.Println(deviceData.DeviceID, time.Now(), " data received")

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

	// if checkIdleState(deviceData) != "idle3" {
	clientJobs <- models.ClientJob{deviceData, conn}
	//}

	// send data to ntsa
	// go sendToNTSA(deviceData)

	// send to association
	// go sendToAssociation(deviceData)

}

// LogToRedis log data to redis
func LogToRedis(m models.DeviceData) {
	var device = strconv.FormatUint(uint64(m.DeviceID), 10)
	lastSeen(m, "lastseen:"+device)
	lastSeen(m, "lastseen")
	// if m.TransmissionReason != 255 && m.GroundSpeed != 0 {
	SetRedisLog(m, "data:"+device)
	// }
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
	var deviceid = strconv.FormatUint(uint64(m.DeviceID), 10)
	squery := "SELECT device_status FROM vehicle_configuration "
	squery += " where device_id=? AND status=? LIMIT ?"
	// Execute the query
	err = tx.QueryRow(squery, deviceid, 1, 1).Scan(&deviceStatus)
	if err != nil {
		return "err"
	}

	var query string
	if m.GroundSpeed > 0 {
		query = "UPDATE vehicle_configuration SET device_status='online' WHERE device_id=? AND status=?"
	} else if m.GroundSpeed == 0 && deviceStatus == "online" {
		query = "UPDATE vehicle_configuration SET device_status='idle1' WHERE device_id=? AND status=?"
	} else if deviceStatus == "idle1" {
		query = "UPDATE vehicle_configuration SET device_status='idle2' WHERE device_id=? AND status=?"
	} else {
		// device data will not be store but redis logs last seen
		SetRedisLog(m, "offline:"+deviceid)
		query = "UPDATE vehicle_configuration SET device_status='idle3' WHERE device_id=? AND status=?"
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return "err"
	}

	_, err = stmt.Exec(deviceid, 1)
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
	squery += " where device_id=? AND status=? LIMIT ?"
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
}

// SaveAllData save all records to second db
func SaveAllData(m models.DeviceData) error {
	err := DBCONDATA.Ping()
	if err != nil {
		return err
	}

	tx, err := DBCONDATA.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer tx.Rollback()

	// create data table
	var device = strconv.FormatUint(uint64(m.DeviceID), 10)
	tablename := "data_" + device
	createquery := "CREATE TABLE IF NOT EXISTS " + tablename + " ( "
	createquery += "`trip_id` int(10) unsigned NOT NULL AUTO_INCREMENT, "
	createquery += "`device_id` BIGINT(20) UNSIGNED NOT NULL, "
	createquery += "`system_code` varchar(10) NOT NULL DEFAULT '0', "
	createquery += "`data_date` datetime DEFAULT NULL, "
	createquery += "`speed` decimal(10,2) NOT NULL DEFAULT 0.00, "
	createquery += "`speed_direction` varchar(45) NOT NULL DEFAULT '0', "
	createquery += "`longitude` int(11) NOT NULL DEFAULT 0, "
	createquery += "`latitude` int(11) NOT NULL DEFAULT 0, "
	createquery += "`altitude` int(11) NOT NULL DEFAULT 0, "
	createquery += "`satellites` int(3) NOT NULL DEFAULT 1, "
	createquery += "`hardware_version` varchar(45) DEFAULT NULL, "
	createquery += "`software_version` varchar(45) DEFAULT NULL, "
	createquery += "`transmission_reason` int(3) DEFAULT 0, "
	createquery += "`transmission_reason_specific_data` varchar(200) DEFAULT NULL, "
	createquery += "`failsafe` tinyint(1) NOT NULL DEFAULT 0, "
	createquery += "`disconnect` tinyint(1) NOT NULL DEFAULT 0, "
	createquery += "`offline` tinyint(1) NOT NULL DEFAULT 0, "
	createquery += "`created_on` timestamp NULL DEFAULT current_timestamp(), "
	createquery += "`date_time_stamp` INT(10) NOT NULL DEFAULT 0, "
	createquery += "PRIMARY KEY (`trip_id`,`device_id`) "
	createquery += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	stmt, err := tx.Prepare(createquery)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer stmt.Close() // danger!

	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}

	// perform a db.Query insert
	query := "INSERT INTO " + tablename + " (device_id, system_code, data_date, speed, speed_direction, "
	query += " longitude, latitude, altitude, satellites, hardware_version, software_version, "
	query += " transmission_reason, transmission_reason_specific_data, failsafe, disconnect, offline, date_time_stamp) "
	query += " VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	stmt, err = tx.Prepare(query)
	if err != nil {
		fmt.Println(err)
	}

	defer stmt.Close()
	_, err = stmt.Exec(m.DeviceID, m.SystemCode, m.DateTime, m.GroundSpeed, m.SpeedDirection, m.Longitude, m.Latitude, m.Altitude, m.NoOfSatellitesUsed, m.HardwareVersion,
		m.SoftwareVersion, m.TransmissionReason, m.TransmissionReasonSpecificData, m.Failsafe, m.Disconnect, m.Offline, m.DateTimeStamp)

	if err != nil {
		fmt.Println(err)
	}

	tx.Commit()

	return nil
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

func sendToAssociation(deviceData models.DeviceData) {
	if deviceData.SystemCode == "MCPG" && deviceData.DeviceID == 12751145 {
		t := deviceData.DateTime
		url := "http://134.209.85.190:8888/api/raw/data"
		powerWireStatus := "off"
		if deviceData.Disconnect {
			powerWireStatus = "on"
		}
		speedSignalStatus := "off"
		if deviceData.Failsafe {
			speedSignalStatus = "on"
		}
		requestBody, err := json.Marshal(map[string]string{
			"companyId":          "ekasfk2017",
			"dateonly":           t.Format("2006-01-02"),
			"deviceNumber":       strconv.Itoa(int(deviceData.DeviceID)),
			"latitude":           strconv.Itoa(int(deviceData.Latitude)),
			"longitude":          strconv.Itoa(int(deviceData.Longitude)),
			"powerWireStatus":    powerWireStatus,
			"registrationNumber": "KDH201-5009832",
			"speed":              strconv.Itoa(int(deviceData.GroundSpeed)),
			"speedSignalStatus":  speedSignalStatus,
			"timeonly":           t.Format("15:04:05"),
		})
		if err != nil {
			fmt.Println(err)
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		fmt.Println(string(body))
	}
}

func sendToNTSA(deviceData models.DeviceData) {
	if deviceData.DeviceID == 100000071 || deviceData.DeviceID == 1000061 {
		t := deviceData.DateTime
		disconnect := "0"
		failsafe := "0"
		if deviceData.Disconnect {
			disconnect = "1"
		}
		if deviceData.Failsafe {
			failsafe = "1"
		}

		lat := FloatToString(float64(deviceData.Latitude) / 10000000)
		long := FloatToString(float64(deviceData.Longitude) / 10000000)
		latdirection := "N"
		if deviceData.Latitude < 0 {
			latdirection = "S"
		}
		longdirection := "E"
		if deviceData.Longitude < 0 {
			longdirection = "W"
		}

		datastr := t.Format("02/01/2006") + "," + t.Format("15:04:05") + "," + strconv.Itoa(int(deviceData.DeviceID)) + ",ekasfk2017,"
		datastr += "KCF 861X," + strconv.Itoa(int(deviceData.GroundSpeed)) + "," + lat + "," + latdirection + ","
		datastr += long + "," + longdirection + "," + disconnect + "," + failsafe

		url := "http://api.speedlimiter.co.ke/ekas"
		payload := strings.NewReader(datastr)

		req, _ := http.NewRequest("POST", url, payload)
		req.Header.Add("content-type", "text/plain")
		req.Header.Add("cache-control", "no-cache")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)

		fmt.Println("association 2", string(body))
	}
}

// FloatToString ...
func FloatToString(inputnum float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(inputnum, 'f', 6, 64)
}
