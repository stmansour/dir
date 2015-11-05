package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type aJobTitle struct {
	JobCode  int
	Title    string
	DeptCode int
}

func createDeductionsList(db *sql.DB) {
	Insrt, err := db.Prepare("INSERT INTO DeductionList (dcode,name) VALUES(?,?)")
	errcheck(err)

	for i := 0; i < DDEND; i++ {
		_, err := Insrt.Exec(i, deductionToString(i))
		errcheck(err)
	}
}

// CreateDeductionsTable takes the string value of "deductions" for every person
// in the people table, parses it, and creates an entry in the deductions table
// for each deduction
func CreateDeductionsTable(db *sql.DB) {
	// OK, we're connected to the database. On with the work. First thing to do is
	// to put all the deductions into a separate multivalued Deductions table organized
	// by employee id
	rows, err := db.Query("select uid, deductions from people")
	errcheck(err)
	defer rows.Close()

	InsrtDeduct, err := db.Prepare("INSERT INTO deductions (uid,deduction) VALUES(?,?)")
	errcheck(err)
	var (
		uid        int
		deductions string
	)

	for rows.Next() {
		errcheck(rows.Scan(&uid, &deductions))
		if len(deductions) > 0 {
			da := strings.Split(deductions, ",")
			for i := 0; i < len(da); i++ {
				d := deductionStringToInt(strings.Trim(da[i], " \n\r"))
				_, err := InsrtDeduct.Exec(uid, d)
				errcheck(err)
			}
		}
	}
	errcheck(rows.Err())
	RemovePeopleDeductCol, err := db.Prepare("alter table people drop column deductions")
	errcheck(err)
	_, err = RemovePeopleDeductCol.Exec()
	errcheck(err)

	createDeductionsList(db)
}

// CreateJobTitlesTable not only creates the JobTitles table, it makes a pass through
// the people table, replaces the  Title field with the appropriate deptcode field.
func CreateJobTitlesTable(db *sql.DB) {
	//--------------------------------------------------------------------------
	// Populate the JobTitles table
	//--------------------------------------------------------------------------
	InsertJT, err := db.Prepare("INSERT INTO JobTitles (title,department) VALUES(?,?)")
	errcheck(err)
	jobtitles := "sql/jobtitles.csv"
	f, err := os.Open(jobtitles)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		jta := strings.Split(line, ",")
		_, err := InsertJT.Exec(strings.Trim(jta[0], " \n\r"), strings.Trim(jta[1], " \n\r"))
		errcheck(err)
	}

	//--------------------------------------------------------------------------
	// create a map from jobtitle to jobcode
	//--------------------------------------------------------------------------
	titlemap = make(map[string]int)
	var title string
	var jobcode int
	var uid int
	rows, err := db.Query("select jobcode, title from jobtitles")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&jobcode, &title))
		titlemap[title] = jobcode
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// add a column for the jobtitle jobcode
	//--------------------------------------------------------------------------
	jobcodeAlter, err := db.Prepare("ALTER TABLE people add column jobcode MEDIUMINT")
	_, err = jobcodeAlter.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// statement to actually make the update
	//--------------------------------------------------------------------------
	jobcodeUpdate, err := db.Prepare("Update people set people.jobcode=? where people.uid=?")
	errcheck(err)

	//--------------------------------------------------------------------------
	// Spin through each entry in people, insert the jobcode associated with the
	// value currently in the Title field...
	//--------------------------------------------------------------------------
	rows, err = db.Query("select uid, title from people")
	errcheck(err)
	defer rows.Close()
	var jc int
	var notitle int
	for rows.Next() {
		errcheck(rows.Scan(&uid, &title))
		if len(title) == 0 {
			jc = 0
			notitle++
		} else {
			jc = titlemap[title]
		}
		_, err = jobcodeUpdate.Exec(jc, uid)
		errcheck(err)
	}
	errcheck(rows.Err())
	fmt.Printf("User entries with no title: %d\n", notitle)
	//--------------------------------------------------------------------------
	// now remove the old title column
	//--------------------------------------------------------------------------
	RemoveTitleCol, err := db.Prepare("alter table people drop column title")
	errcheck(err)
	_, err = RemoveTitleCol.Exec()
	errcheck(err)
}

func getJobInfo(db *sql.DB, jobcode int, job *aJobTitle) {
	job.DeptCode = 0
	job.Title = ""
	rows, err := db.Query("select jobcode,title,DeptCode from jobtitles where jobcode=?", jobcode)
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&job.JobCode, &job.Title, &job.DeptCode))
	}
	if "" == job.Title {
		fmt.Printf("Cound not find job title for jobcode = %d\n", jobcode)
	}
}

func getJobInfoByTitle(db *sql.DB, title string, job *aJobTitle) {
	m := 0
	job.JobCode = 0
	job.DeptCode = 0
	job.Title = ""
	rows, err := db.Query("select jobcode,title,DeptCode from jobtitles where title=?", title)
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&job.JobCode, &job.Title, &job.DeptCode))
		m++
	}
	if 0 == m {
		fmt.Printf("Cound not find job title for title = %s\n", title)
	}
}

// UpdateJobTitles reads the updated jobtitles that stacie sent me on Oct 30, 2015
// and updates them into the database. The data is in the format jobtitle,fullName
func UpdateJobTitles(db *sql.DB) {
	jobtitles := "sql/UpdatedEmployeeTitles.csv"
	f, err := os.Open(jobtitles)
	errcheck(err)
	defer f.Close()

	var uid, jobcode int
	var lastname, firstname string
	var found bool
	var job aJobTitle

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Printf("line:  %s\n", line)
		jta := strings.Split(line, ",")
		// fmt.Printf("jta[0] = %s    jta[1] = %s\n", jta[0], jta[1])
		// which is in the format firstname " " lastname
		nm := strings.Split(jta[1], " ")
		// fmt.Printf("%s %s\n", nm[0], nm[1])
		rows, err := db.Query("select uid,lastname,firstname,jobcode from people where firstname=? and lastname=?", nm[0], nm[1])
		errcheck(err)
		defer rows.Close()
		found = false
		for rows.Next() {
			errcheck(rows.Scan(&uid, &lastname, &firstname, &jobcode))
			found = true
			// fmt.Printf("Found %s %s, jobcode = %d\n", firstname, lastname, jobcode)
			if 0 == jobcode {
				// try to find a match on title
				// fmt.Printf("%s %s has jobcode 0.  Searching for match on job title %s\n", firstname, lastname, jta[0])
				getJobInfoByTitle(db, jta[0], &job)
				if job.JobCode > 0 {
					// fmt.Printf("Found match, job code is %d, updating record for %s %s\n", job.JobCode, firstname, lastname)
					update, err := db.Prepare("update people set jobcode=? where people.uid=?")
					_, err = update.Exec(job.JobCode, uid)
					if nil != err {
						fmt.Printf("err updating people record: %v\n", err)
					}
				} else {
					fmt.Printf("unable to find a job with title: %s\n", jta[0])
				}
			} else {
				getJobInfo(db, jobcode, &job)
				if job.Title != jta[0] {
					// fmt.Printf("%d - %s %s : current title: %s  -  updated title: %s\n", uid, firstname, lastname, job.Title, jta[0])
					getJobInfoByTitle(db, jta[0], &job)
					update, err := db.Prepare("update people set jobcode=? where people.uid=?")
					_, err = update.Exec(job.JobCode, uid)
					if nil != err {
						fmt.Printf("err updating people record: %v\n", err)
					}
				}
			}
		}
		if !found {
			fmt.Printf("Could not find %s %s\n", nm[0], nm[1])
		}
		errcheck(rows.Err())
	}
}

// FixDate deals with changing dates from the format
func FixDate(db *sql.DB, column string) {
	// First, read in everyone's date...
	thedateda := make(map[int]string, 0)

	rows, err := db.Query(fmt.Sprintf("select uid,%s from people", column))
	errcheck(err)
	defer rows.Close()

	var (
		uid     int
		thedate string
		year    int
		month   int
		day     int
	)
	for rows.Next() {
		errcheck(rows.Scan(&uid, &thedate))
		if len(thedate) > 0 {
			da := strings.Split(thedate, "/")
			month, _ = strconv.Atoi(da[0])
			day, _ = strconv.Atoi(da[1])
			year, _ = strconv.Atoi(da[2])
			if year > 50 {
				year += 1900
			} else {
				year += 2000
			}
		} else {
			year = 0
			month = 0
			day = 0
		}

		thedateda[uid] = fmt.Sprintf("%04d/%02d/%02d", year, month, day)
	}
	errcheck(rows.Err())

	//--------------------------------------------
	// Now remove the Hire column completely..
	//--------------------------------------------
	RemoveCol, err := db.Prepare(fmt.Sprintf("alter table people drop column %s", column))
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

	//--------------------------------------------
	// Add a Hire column back in as a Date...
	//--------------------------------------------
	AddCol, err := db.Prepare(fmt.Sprintf("ALTER TABLE people add column %s date", column))
	errcheck(err)
	_, err = AddCol.Exec()
	errcheck(err)

	//--------------------------------------------
	// Now add the date that we saved earlier for each UID
	//--------------------------------------------
	DateUpdate, err := db.Prepare(fmt.Sprintf("Update people set people.%s=? where people.uid=?", column))
	errcheck(err)
	for uid, dv := range thedateda {
		_, err = DateUpdate.Exec(dv, uid)
		errcheck(err)
	}
}

func fixReviewDates(db *sql.DB) {
	//--------------------------------------------
	// Now remove the Hire column completely..
	//--------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column LastReview")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

	RemoveCol, err = db.Prepare("alter table people drop column NextReview")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

	//--------------------------------------------
	// Add a Hire column back in as a Date...
	//--------------------------------------------
	AddCol, err := db.Prepare("ALTER TABLE people add column LastReview date")
	errcheck(err)
	_, err = AddCol.Exec()
	errcheck(err)

	AddCol, err = db.Prepare("ALTER TABLE people add column NextReview date")
	errcheck(err)
	_, err = AddCol.Exec()
	errcheck(err)

	//--------------------------------------------
	// Now add the date that we saved earlier for each UID
	//--------------------------------------------
	dt := fmt.Sprintf("%04d/%02d/%02d", 0, 0, 0)
	DateUpdate, err := db.Prepare(fmt.Sprintf("Update people set LastReview=?,NextReview=?"))
	errcheck(err)
	_, err = DateUpdate.Exec(dt, dt)
	errcheck(err)
}

// FixBirthday changes the "birthdate" string from it's present form (multiple formats)
// to two fields - birthMonth, birthDOM
func FixBirthday(db *sql.DB) {
	var bd struct {
		uid   int
		s     string
		month int
		day   int
	}
	var Months = []string{
		"January", "February", "March", "April",
		"May", "June", "July", "August",
		"September", "October", "November", "December",
	}

	//--------------------------------------------
	// Add a Hire column back in as a Date...
	//--------------------------------------------
	AddCol, err := db.Prepare(fmt.Sprintf("ALTER TABLE people add column birthMonth TINYINT after birthdate"))
	errcheck(err)
	_, err = AddCol.Exec()
	errcheck(err)

	AddCol, err = db.Prepare(fmt.Sprintf("ALTER TABLE people add column birthDoM TINYINT after birthMonth"))
	errcheck(err)
	_, err = AddCol.Exec()
	errcheck(err)

	Update, err := db.Prepare("Update people set birthMonth=?,birthDoM=? where uid=?")
	errcheck(err)

	rows, err := db.Query(fmt.Sprintf("select uid,birthdate from people"))
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&bd.uid, &bd.s))
		bd.day = 0
		bd.month = 0
		if bd.s != "" {
			// it will either have a slash (/) for dates like 12/25
			// or a minus (-) for dates like Aug-23
			if strings.Contains(bd.s, "/") {
				da := strings.Split(bd.s, "/")
				bd.month, _ = strconv.Atoi(da[0])
				bd.day, _ = strconv.Atoi(da[1])
			} else {
				da := strings.Split(bd.s, "-")
				da = strings.Split(bd.s, "-")
				bd.day, _ = strconv.Atoi(da[0])
				for i := 0; i < len(Months); i++ {
					//fmt.Printf("da[1] = %s,  Months[i][0:3]", ...)
					if da[1] == Months[i][0:3] {
						bd.month = 1 + i
						break
					}
				}
			}
			//fmt.Printf("%s maps to -> mon = %d, day = %d\n", bd.s, bd.month, bd.day)
		}
		_, err = Update.Exec(bd.month, bd.day, bd.uid)
		errcheck(err)
	}
	errcheck(rows.Err())

	//=================================
	// now remove the string column...
	//=================================
	RemoveCol, err := db.Prepare("alter table people drop column Birthdate")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
}

func findUID(db *sql.DB, first string, last string) int {
	rows, err := db.Query(fmt.Sprintf("select uid from people where firstname=\"%s\" and lastname=\"%s\"", first, last))
	errcheck(err)
	defer rows.Close()
	var uid int
	for rows.Next() {
		errcheck(rows.Scan(&uid))
	}
	errcheck(rows.Err())
	return uid
}

// FixManager takes the existing string ReportsTo field and tries to find the associated
// person uid. If successful it creates the reference to the manager's UID in the manager column.
func FixManager(db *sql.DB) {
	pplAlter, err := db.Prepare("ALTER TABLE people ADD COLUMN mgruid MEDIUMINT NOT NULL")
	_, err = pplAlter.Exec()
	errcheck(err)

	rows, err := db.Query("select uid, reportsto from people")
	errcheck(err)
	defer rows.Close()

	UpdateMgrUID, err := db.Prepare("Update people set people.mgruid=? where people.uid=?")
	errcheck(err)

	var uid int
	var mgruid int
	var reportsto string

	for rows.Next() {
		errcheck(rows.Scan(&uid, &reportsto))
		if len(reportsto) > 0 {
			da := strings.Split(reportsto, ",")
			if len(da) == 2 {
				first := strings.Trim(da[1], " \n\r")
				last := strings.Trim(da[0], " \n\r")
				mgruid = findUID(db, first, last)
				if 0 == mgruid {
					fmt.Printf("Could not find uid for: %s %s", first, last)
				}
			}
			// fmt.Printf("uid:%d  mgr:%d\n", uid, mgruid)
			_, err := UpdateMgrUID.Exec(mgruid, uid)
			errcheck(err)
		}
	}
	errcheck(rows.Err())
	RemoveCol, err := db.Prepare("alter table people drop column reportsto")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
}

// LoadDepartments populates the Departments table with its initial content
func LoadDepartments(db *sql.DB) {
	//--------------------------------------------------------------------------
	// Populate the JobTitles table
	//--------------------------------------------------------------------------
	InsertJT, err := db.Prepare("INSERT INTO Departments (name) VALUES(?)")
	errcheck(err)
	filename := "sql/departments.csv"
	f, err := os.Open(filename)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		da := strings.Split(line, ",")
		_, err := InsertJT.Exec(strings.Trim(da[0], " \n\r"))
		errcheck(err)
	}

	//--------------------------------------------------------------------------
	// create a map from DeptName to deptcode
	//--------------------------------------------------------------------------
	deptmap := make(map[string]int)
	var name string
	var deptcode int
	rows, err := db.Query("select deptcode,name from departments")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&deptcode, &name))
		deptmap[name] = deptcode
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// Put the deptcode into the JobTitles table
	//--------------------------------------------------------------------------
	var deptname string
	deptcodeUpdate, err := db.Prepare("Update jobtitles set deptcode=? where jobcode=?")
	errcheck(err)
	rows, err = db.Query("select jobcode,department from jobtitles")
	errcheck(err)
	defer rows.Close()
	var jc int
	for rows.Next() {
		errcheck(rows.Scan(&jc, &deptname))
		deptcode = deptmap[deptname]
		_, err = deptcodeUpdate.Exec(deptcode, jc)
		errcheck(err)
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// Now we can remove the text column of department from jobtitles
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table jobtitles drop column department")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// create a map from jobcode to deptcode
	//--------------------------------------------------------------------------
	j2d := make(map[int]int, 0)
	rows, err = db.Query("select jobcode,deptcode from jobtitles")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&jc, &deptcode))
		j2d[jc] = deptcode
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// add a column for the deptcode to people... based on the department
	// listed for the jobtitle
	//--------------------------------------------------------------------------
	deptcodeUpdate, err = db.Prepare("ALTER TABLE people add column deptcode MEDIUMINT")
	_, err = deptcodeUpdate.Exec()
	errcheck(err)
	var uid int
	deptcodeUpdate, err = db.Prepare("Update people set deptcode=? where uid=?")
	errcheck(err)
	rows, err = db.Query("select uid,jobcode from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &jc))
		_, err = deptcodeUpdate.Exec(j2d[jc], uid)
		errcheck(err)
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// now remove the department column from people
	//--------------------------------------------------------------------------
	RemoveCol, err = db.Prepare("alter table people drop column department")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)

}
func strToInt(s string) int {
	if len(s) == 0 {
		return 0
	}
	s = strings.Trim(s, " \n\r")
	n, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error converting %s to a number: %v\n", s, err)
		return 0
	}
	return n
}

// LoadCompanies reads companies.csv and initializes the companies table
func LoadCompanies(db *sql.DB) {
	//--------------------------------------------------------------------------
	// Populate the companies table
	//--------------------------------------------------------------------------
	InsertJT, err := db.Prepare("INSERT INTO companies (LegalName,CommonName,address,address2,city,state,postalcode,country,phone,fax,email,designation,Active,EmploysPersonnel) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	errcheck(err)
	filename := "sql/companies.csv"
	f, err := os.Open(filename)
	errcheck(err)

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sanity check, display to standard output
	for _, da := range rawCSVdata {
		for i := 0; i < len(da); i++ {
			da[i] = strings.Trim(da[i], " \n\r")
		}
		fmt.Printf("da[0]=\"%s\", da[1]=\"%s\", da[3]=\"%s\", da[4]=\"%s\"\n", da[0], da[1], da[2], da[3])
		_, err := InsertJT.Exec(da[0], da[1], da[2], da[3], da[4], da[5], da[6], da[7], da[8], da[9], da[10], da[11], strToInt(da[12]), strToInt(da[13]))
		errcheck(err)
	}
	//--------------------------------------------------------------------------
	// create a map from Employer to cocode
	//--------------------------------------------------------------------------
	comap := make(map[string]int)
	var name string
	var cocode int
	rows, err := db.Query("select cocode,CommonName from companies")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&cocode, &name))
		comap[name] = cocode
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// add a column for the cocode to people... based on the employer
	//--------------------------------------------------------------------------
	colUpdate, err := db.Prepare("ALTER TABLE people add column cocode MEDIUMINT")
	_, err = colUpdate.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now fill in the value for each employee
	//--------------------------------------------------------------------------
	var uid int
	var employer string
	deptcodeUpdate, err := db.Prepare("Update people set cocode=? where uid=?")
	errcheck(err)
	rows, err = db.Query("select uid,Employer from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &employer))
		employer = strings.Trim(employer, " \n\r")
		_, err = deptcodeUpdate.Exec(comap[employer], uid)
		errcheck(err)
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// Now we can remove the Employer column from people
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column employer")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
}

// AddPeopleColumns adds several remaining columns to db
func AddPeopleColumns(db *sql.DB) {
	//--------------------------------------------------------------------------
	// Country of employment
	//--------------------------------------------------------------------------
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN StateOfEmployment VARCHAR(25) NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
	colAdd, err = db.Prepare("ALTER TABLE people ADD COLUMN CountryOfEmployment VARCHAR(25) NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
	colAdd, err = db.Prepare("ALTER TABLE people ADD COLUMN PreferredName VARCHAR(25) NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
	colAdd, err = db.Prepare("ALTER TABLE people ADD COLUMN EmergencyContactName VARCHAR(25) NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
	colAdd, err = db.Prepare("ALTER TABLE people ADD COLUMN EmergencyContactPhone VARCHAR(25) NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)
}

// FixStatus changes the status column from a string to a boolean
func FixStatus(db *sql.DB) {
	//--------------------------------------------------------------------------
	// create a map from Status to boolean
	//--------------------------------------------------------------------------
	comap := make(map[int]string)
	var status string
	var uid int
	rows, err := db.Query("select uid,status from people")
	errcheck(err)
	defer rows.Close()
	for rows.Next() {
		errcheck(rows.Scan(&uid, &status))
		comap[uid] = status
	}
	errcheck(rows.Err())

	//--------------------------------------------------------------------------
	// get rid of the current status column, add it back as a boolean
	//--------------------------------------------------------------------------
	RemoveCol, err := db.Prepare("alter table people drop column status")
	errcheck(err)
	_, err = RemoveCol.Exec()
	errcheck(err)
	colAdd, err := db.Prepare("ALTER TABLE people ADD COLUMN status SMALLINT NOT NULL")
	_, err = colAdd.Exec()
	errcheck(err)

	//--------------------------------------------------------------------------
	// now update status for each uid
	//--------------------------------------------------------------------------
	statusUpdate, err := db.Prepare("Update people set status=? where uid=?")
	var b int
	for uid, stat := range comap {
		stat = strings.ToLower(stat)
		b = 1
		if strings.Contains(stat, "no") {
			b = 0
		}
		_, err = statusUpdate.Exec(b, uid)
		errcheck(err)
	}
}
