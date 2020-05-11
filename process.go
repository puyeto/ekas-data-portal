package main

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
	"sync"
	"time"

	"github.com/ekas-data-portal/core"
	"github.com/ekas-data-portal/models"
)

const queueLimit = 50

// HandleRequest Handles incoming requests.
func HandleRequest(conn net.Conn) {

	var byteSize = 70
	byteData := make([]byte, 700)

	for {
		reqLen, err := conn.Read(byteData)
		if err != nil {
			if err != io.EOF {
				fmt.Println("End of file error:", err)
			}
			fmt.Println("Error reading:", err.Error(), reqLen)
			return
		}

		// return Response
		result := "Received byte size = " + strconv.Itoa(reqLen) + "\n"
		conn.Write([]byte(string(result)))

		if reqLen == 0 {
			return // connection already closed by client
		}

		if reqLen > 0 {
			byteRead := bytes.NewReader(byteData)

			for i := 0; i < (reqLen / byteSize); i++ {

				byteRead.Seek(int64((byteSize * i)), 0)

				mb := make([]byte, byteSize)
				n1, _ := byteRead.Read(mb)

				go processRequest(conn, mb, n1)
			}

		}
		opsRate.Mark(1)
	}
}

func readNextBytes(conn net.Conn, number int) (int, []byte) {
	bytes := make([]byte, number)

	reqLen, err := conn.Read(bytes)
	if err != nil {
		if err != io.EOF {
			core.Logger.Warnf("End of file error: %v", err)
		}
		core.Logger.Warnf("Error reading: %v %v", err.Error(), reqLen)
	}

	return reqLen, bytes
}

func processRequest(conn net.Conn, b []byte, byteLen int) {
	clientJobs := make(chan models.ClientJob)
	go generateResponses(clientJobs)

	var deviceData models.DeviceData

	if byteLen != 70 {
		core.Logger.Errorf("Invalid Byte Length: %v", byteLen)
		return
	}

	byteReader := bytes.NewReader(b)

	scode := processSeeked(byteReader, 4, 0)
	deviceData.SystemCode = string(scode)
	if deviceData.SystemCode != "MCPG" {
		return
	}

	did := processSeeked(byteReader, 4, 5)
	deviceData.DeviceID = binary.LittleEndian.Uint32(did)
	if deviceData.DeviceID == 0 {
		return
	}
	// Transmission Reason – 1 byte
	reason := processSeeked(byteReader, 1, 18)
	deviceData.TransmissionReason = int(reason[0])

	// Transmission Reason Specific data – 1 byte
	trsd := 0
	if deviceData.TransmissionReason == 255 {
		specific := processSeeked(byteReader, 1, 17)

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
	satellites := processSeeked(byteReader, 1, 43)
	deviceData.NoOfSatellitesUsed = int(satellites[0])

	// Longitude – 4 bytes
	long := processSeeked(byteReader, 4, 44)
	deviceData.Longitude = readInt32(long)

	//  Latitude – 4 bytes
	lat := processSeeked(byteReader, 4, 48)
	deviceData.Latitude = readInt32(lat)

	// Altitude
	alt := processSeeked(byteReader, 4, 52)
	deviceData.Altitude = readInt32(alt)

	// Ground speed – 4 bytes
	gspeed := processSeeked(byteReader, 4, 56)
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(gspeed)

	// Speed direction – 2 bytes
	speedd := processSeeked(byteReader, 2, 60)
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(speedd))

	sec := processSeeked(byteReader, 1, 62)
	deviceData.UTCTimeSeconds = int(sec[0])

	min := processSeeked(byteReader, 1, 63)
	deviceData.UTCTimeMinutes = int(min[0])

	hrs := processSeeked(byteReader, 1, 64)
	deviceData.UTCTimeHours = int(hrs[0])

	day := processSeeked(byteReader, 1, 65)
	deviceData.UTCTimeDay = int(day[0])

	mon := processSeeked(byteReader, 1, 66)
	deviceData.UTCTimeMonth = int(mon[0])

	yr := processSeeked(byteReader, 2, 67)
	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(yr))

	// if deviceData.DeviceID == 1212208985 && deviceData.GroundSpeed > 85 {
	// 	rand.Seed(time.Now().UnixNano())
	// 	min := 75
	// 	max := 83

	// 	deviceData.GroundSpeed = uint32(rand.Intn(max-min+1) + min)
	// }

	loc, _ := time.LoadLocation("Africa/Nairobi")
	later := time.Now().In(loc)
	before := time.Now().In(loc)
	oneHourLater := later.Add(time.Hour * 1).Unix()
	oneHourBefore := before.Add(time.Hour * -1).Unix()

	// if checkIdleState(deviceData) != "idle3" {
	if deviceData.DateTimeStamp < oneHourBefore || deviceData.DateTimeStamp > oneHourLater {
		now := time.Now().In(loc)
		deviceData.UTCTimeMinutes = now.Minute()
		deviceData.UTCTimeHours = now.Hour()
		deviceData.UTCTimeDay = now.Day()
		deviceData.UTCTimeMonth = int(now.Month())
		deviceData.UTCTimeYear = now.Year()
	}
	deviceData.DateTime = time.Date(deviceData.UTCTimeYear, time.Month(deviceData.UTCTimeMonth), deviceData.UTCTimeDay, deviceData.UTCTimeHours, deviceData.UTCTimeMinutes, deviceData.UTCTimeSeconds, 0, time.UTC)
	deviceData.DateTimeStamp = deviceData.DateTime.Unix()
	clientJobs <- models.ClientJob{deviceData, conn}
	// }

	// if deviceData.DeviceID == 1205205360 {
	// 	deviceData.GroundSpeed = 0
	// 	deviceData.DeviceID = 1161512252
	// 	clientJobs <- models.ClientJob{deviceData, conn}
	// }

	// send data to ntsa
	// go sendToNTSA(deviceData)

	// send to association
	// go sendToAssociation(deviceData)
	// conn.Close()

}

func processSeeked(byteReader *bytes.Reader, bytesize, seek int64) []byte {
	byteReader.Seek(seek, 0)
	val := make([]byte, bytesize)
	byteReader.Read(val)
	return val
}

func generateResponses(clientJobs chan models.ClientJob) {
	for {
		// use a WaitGroup
		var wg sync.WaitGroup

		// Wait for the next job to come off the queue.
		clientJob := <-clientJobs

		// make a channel with a capacity of 100.
		jobChan := make(chan models.DeviceData, queueLimit)

		worker := func(jobChan <-chan models.DeviceData) {
			defer wg.Done()
			for job := range jobChan {
				// SaveAllData(job)
				LogToRedis(job)
				if err := core.LogToMongoDB(job); err != nil {
					core.Logger.Warnf("Mongo DB - logging error: %v", err)
				}
				if err := core.LoglastSeenMongoDB(job); err != nil {
					core.Logger.Warnf("Mongo DB - logging last seen error: %v", err)
				}
				if job.TransmissionReason == 255 || job.GroundSpeed > 84 || job.Offline == true {
					if err := core.LogCurrentViolationSeenMongoDB(job); err != nil {
						core.Logger.Warnf("Mongo DB - logging current violations error: %v", err)
					}
				}
			}
		}

		// increment the WaitGroup before starting the worker
		wg.Add(1)
		go worker(jobChan)

		// enqueue a job
		jobChan <- clientJob.DeviceData

		// to stop the worker, first close the job channel
		close(jobChan)

		// then wait using the WaitGroup
		WaitTimeout(&wg, 1*time.Second)
	}
}

// WaitTimeout does a Wait on a sync.WaitGroup object but with a specified
// timeout. Returns true if the wait completed without timing out, false
// otherwise.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
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
func checkIdleStateMysql(m models.DeviceData) string {
	err := core.DBCONN.Ping()
	if err != nil {
		fmt.Println(err)
	}

	tx, err := core.DBCONN.Begin()
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
	defer stmt.Close()

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
	err := core.DBCONN.Ping()
	if err != nil {
		fmt.Println(err)
	}

	tx, err := core.DBCONN.Begin()
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
			defer stmt.Close()
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
			defer stmt.Close()
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
	if core.DBCONDATA == nil {
		fmt.Println("db nil")
	}
	err := core.DBCONDATA.Ping()
	if err != nil {
		return err
	}

	tx, err := core.DBCONDATA.Begin()
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
	createquery += "`date_time_stamp` INT(11) NOT NULL DEFAULT 0, "
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
		fmt.Println(m.DeviceID, err)
	}

	defer stmt.Close()
	_, err = stmt.Exec(m.DeviceID, m.SystemCode, m.DateTime, m.GroundSpeed, m.SpeedDirection, m.Longitude, m.Latitude, m.Altitude, m.NoOfSatellitesUsed, m.HardwareVersion,
		m.SoftwareVersion, m.TransmissionReason, m.TransmissionReasonSpecificData, m.Failsafe, m.Disconnect, m.Offline, m.DateTimeStamp)

	if err != nil {
		fmt.Println(err)
	}

	tx.Commit()

	if m.TransmissionReason == 255 || m.GroundSpeed > 84 || m.Offline == true {
		// log data to redis
		currentViolations(m, "currentviolations:"+device)
		currentViolations(m, "currentviolations")
		SetRedisLogViolations(m, "violations")
		SetRedisLog(m, "violations:"+device)

		if m.Offline == true {
			SetRedisLog(m, "offline:"+device)
		}
	}

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
	_, err := core.SetValue(key, data)
	if err != nil {
		fmt.Println(err)
	}
}

func currentViolations(m models.DeviceData, key string) {
	// SET object
	_, err := core.SetValue(key, m)
	if err != nil {
		fmt.Println(err)
	}
}

// SetRedisLogViolations log to redis
func SetRedisLogViolations(m models.DeviceData, key string) {
	_, err := core.ZAdd(key, int64(m.DeviceID), m)
	if err != nil {
		fmt.Println(err)
	}
}

// SetRedisLog log to redis
func SetRedisLog(m models.DeviceData, key string) {
	_, err := core.ZAdd(key, m.DateTimeStamp, m)
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
