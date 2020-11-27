package expireddevices

import (
	"fmt"
	"time"

	"github.com/ekas-data-portal/core"
)

// Status ...
type Status struct{}

type DeviceDetails struct {
	DeviceID     int32     `json:"device_id"`
	VehicleRegNo string    `json:"vehicle_reg_no"`
	ExpiryDate   time.Time `json:"expiry_date"`
}

// Run Status.Run() will get triggered automatically.
func (s Status) Run() {

	tx, err := core.DBCONN.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tx.Rollback()

	// Execute the query
	query := "SELECT vc.device_id, vd.vehicle_reg_no, "
	query += "DATE_ADD(DATE_ADD(COALESCE(vr.renewal_date, vd.created_on), INTERVAL -1 DAY), INTERVAL 1 YEAR) AS expiry_date FROM vehicle_configuration AS vc "
	query += "LEFT JOIN vehicle_details AS vd ON (vd.vehicle_string_id = vc.vehicle_string_id) "
	query += "LEFT JOIN vehicle_renewals AS vr ON (vr.vehicle_id = vd.vehicle_id) "
	query += "WHERE vc.STATUS=1 AND DATE_ADD(DATE_ADD(COALESCE(vr.renewal_date, vd.created_on), INTERVAL -1 DAY), INTERVAL 1 YEAR) < NOW() "
	results, err := tx.Query(query)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	core.ExpiredDeviceIDs = []int32{}
	for results.Next() {
		var details DeviceDetails
		// for each row, scan the result into our tag composite object
		err = results.Scan(&details.DeviceID, &details.VehicleRegNo, &details.ExpiryDate)
		if err != nil {
			continue
		}

		core.ExpiredDeviceIDs = append(core.ExpiredDeviceIDs, details.DeviceID)
	}

}
