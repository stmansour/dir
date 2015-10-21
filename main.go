// This application takes raw data from Accord's "ALL BFF Databases" spreadsheet
// and creates a MySql database of the information.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)
import _ "github.com/go-sql-driver/mysql"

var titlemap map[string]int

func errcheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// FixEligibleForRehire changes the EligibleForRehire column from a string to a boolean
func FixEligibleForRehire(db *sql.DB) {
	//--------------------------------------------------------------------------
	// create a map from EligibleForRehire to boolean
	//--------------------------------------------------------------------------
	comap := make(map[int]string)
	var EligibleForRehire string
	var uid int
	rows, err := db.Query("select uid,EligibleForRehire from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &EligibleForRehire))
		comap[uid] = EligibleForRehire
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// get rid of the current EligibleForRehire column, add it back as a boolean
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column EligibleForRehire")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN EligibleForRehire SMALLINT NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now update EligibleForRehire for each uid
	//--------------------------------------------------------------------------
	EligibleForRehireUpdate, err := db.Prepare("Update people set EligibleForRehire=? where uid=?")
	var b int
	for uid, stat := range comap {
		stat = strings.ToLower(stat)
		b = 1
		if strings.Contains(stat, "no") {
			b = 0
		}
		_, err = EligibleForRehireUpdate.Exec(b, uid)
		errcheck(err)
	}
}

func fixCompensationType(db *sql.DB) {
	//--------------------------------------------------------------------------
	// create a map from CompensationType to int
	//--------------------------------------------------------------------------
	comap := make(map[int]string)
	var CompensationType string
	var uid int
	rows, err := db.Query("select uid,CompensationType from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &CompensationType))
		comap[uid] = CompensationType
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// get rid of the current CompensationType column, add it back as a boolean
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column CompensationType")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now update CompensationType for each uid
	//--------------------------------------------------------------------------
	CompensationTypeAdd, err := db.Prepare("INSERT INTO compensation (uid,type) VALUES(?,?)")
	errcheck(err)
	var b int
	for uid, stat := range comap {
		b = compensationTypeToInt(stat)
		_, err = CompensationTypeAdd.Exec(uid, b)
		errcheck(err)
	}
}

func fixHealthInsurance(db *sql.DB) {
	//--------------------------------------------------------------------------
	// create a map from HealthInsuranceAccepted to int
	//--------------------------------------------------------------------------
	comap := make(map[int]string)
	var accepted string
	var uid int
	rows, err := db.Query("select uid,HealthInsuranceAccepted from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &accepted))
		comap[uid] = accepted
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// get rid of the current HealthInsuranceAccepted column, add it back as a smallint
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column HealthInsuranceAccepted")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN AcceptedHealthInsurance SMALLINT NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now update HealthInsuranceAccepted for each uid
	//--------------------------------------------------------------------------
	acceptStmt, err := db.Prepare("Update people set AcceptedHealthInsurance=? where uid=?")
	var b int
	for uid, stat := range comap {
		b = acceptTypeToInt(stat)
		_, err = acceptStmt.Exec(b, uid)
		errcheck(err)
	}
}

func fixDentalInsurance(db *sql.DB) {
	//--------------------------------------------------------------------------
	// create a map from DentalInsuranceAccepted to int
	//--------------------------------------------------------------------------
	comap := make(map[int]string)
	var accepted string
	var uid int
	rows, err := db.Query("select uid,DentalInsuranceAccepted from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &accepted))
		comap[uid] = accepted
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// get rid of the current DentalInsuranceAccepted column, add it back as a smallint
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column DentalInsuranceAccepted")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN AcceptedDentalInsurance SMALLINT NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now update DentalInsuranceAccepted for each uid
	//--------------------------------------------------------------------------
	acceptStmt, err := db.Prepare("Update people set AcceptedDentalInsurance=? where uid=?")
	var b int
	for uid, stat := range comap {
		b = acceptTypeToInt(stat)
		_, err = acceptStmt.Exec(b, uid)
		errcheck(err)
	}
}

func fix401K(db *sql.DB) {
	//--------------------------------------------------------------------------
	// get rid of the current 401Kaccepted column, add it back as a smallint
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column Accepted401K")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN Accepted401K SMALLINT NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
}

func main() {
	db, err := sql.Open("mysql", "sman:@/accord?charset=utf8&parseTime=True")
	if nil != err {
		fmt.Printf("sql.Open: Error = %v\n", err)
	}
	defer db.Close()

	err = db.Ping()
	if nil != err {
		fmt.Printf("db.Ping: Error = %v\n", err)
	}

	CreateDeductionsTable(db)
	CreateJobTitlesTable(db)
	FixDate(db, "hire")
	FixDate(db, "termination")
	FixBirthday(db)
	FixManager(db)
	LoadDepartments(db)
	LoadCompanies(db)
	AddPeopleColumns(db)
	FixStatus(db)
	FixEligibleForRehire(db)
	fixCompensationType(db)
	fixHealthInsurance(db)
	fixDentalInsurance(db)
	fix401K(db)
}
