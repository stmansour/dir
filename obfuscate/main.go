// This application takes raw data from Accord's "ALL BFF Databases" spreadsheet
// and creates a MySql database of the information.
package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// rand.Intn(25)

import _ "github.com/go-sql-driver/mysql"

// ObfuscatorType describes the information needed at the app level
type ObfuscatorType struct {
	seed       int64
	FirstNames []string
	LastNames  []string
	Streets    []string
	Cities     []string
	States     []string
	db         *sql.DB
}

// Obfuscator is the app object
var Obfuscator ObfuscatorType

func errcheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func randomPhoneNumber() string {
	return fmt.Sprintf("(%d) %3d-%04d", 100+rand.Intn(899), 100+rand.Intn(899), rand.Intn(9999))
}

func randomEmail(lastname string, firstname string) string {
	var providers = []string{"gmail.com", "yahoo.com", "comcast.net", "aol.com", "bdiddy.com", "hotmail.com", "abiz.com"}
	np := len(providers)
	n := rand.Intn(10)
	switch {
	case n < 4:
		return fmt.Sprintf("%s%s%d@%s", firstname[0:1], lastname, rand.Intn(10000), providers[rand.Intn(np)])
	case n > 6:
		return fmt.Sprintf("%s%s%d@%s", firstname, lastname[0:1], rand.Intn(10000), providers[rand.Intn(np)])
	default:
		return fmt.Sprintf("%s%s%d@%s", firstname, lastname, rand.Intn(1000), providers[rand.Intn(np)])
	}
}

func randomAddress() string {
	return fmt.Sprintf("%d %s", rand.Intn(99999), Obfuscator.Streets[rand.Intn(len(Obfuscator.Streets))])
}

func randomizeStuff() {
	var (
		uid       int
		lastname  string
		firstname string
	)

	Nlast := len(Obfuscator.LastNames)
	Nfirst := len(Obfuscator.FirstNames)

	rows, err := Obfuscator.db.Query("select uid from people")
	errcheck(err)
	defer rows.Close()

	update, err := Obfuscator.db.Prepare("update people set lastname=?,firstname=?,OfficePhone=?,OfficeFax=?,CellPhone=?,PrimaryEmail=?,SecondaryEmail=?,HomeStreetAddress=?,HomeStreetAddress2=?,HomeCity=?,HomeState=? where uid=?")
	errcheck(err)

	for rows.Next() {
		errcheck(rows.Scan(&uid))
		lastname = strings.Title(strings.ToLower(Obfuscator.LastNames[rand.Intn(Nlast)]))
		firstname = strings.Title(strings.ToLower(Obfuscator.FirstNames[rand.Intn(Nfirst)]))
		p1 := randomPhoneNumber()
		p2 := randomPhoneNumber()
		p3 := randomPhoneNumber()
		e1 := randomEmail(lastname, firstname)
		e2 := randomEmail(lastname, firstname)
		s1 := randomAddress()
		s2 := ""
		if rand.Intn(100) < 33 {
			s2 = fmt.Sprintf("Apt. #%d", rand.Intn(9999))
		}
		c1 := Obfuscator.Cities[rand.Intn(len(Obfuscator.Cities))]
		st := Obfuscator.States[rand.Intn(len(Obfuscator.States))]
		_, err = update.Exec(lastname, firstname, p1, p2, p3, e1, e2, s1, s2, c1, st, uid)
		errcheck(err)
	}
	errcheck(rows.Err())
}

func loadNames() {
	file, err := os.Open("./firstnames.txt")
	errcheck(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		Obfuscator.FirstNames = append(Obfuscator.FirstNames, scanner.Text())
	}
	errcheck(scanner.Err())
	fmt.Printf("FirstNames: %d\n", len(Obfuscator.FirstNames))

	file, err = os.Open("./lastnames.txt")
	errcheck(err)
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		Obfuscator.LastNames = append(Obfuscator.LastNames, scanner.Text())
	}
	errcheck(scanner.Err())
	fmt.Printf("LastNames: %d\n", len(Obfuscator.LastNames))

	file, err = os.Open("./states.txt")
	errcheck(err)
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		Obfuscator.States = append(Obfuscator.States, scanner.Text())
	}
	errcheck(scanner.Err())
	fmt.Printf("States: %d\n", len(Obfuscator.States))

	file, err = os.Open("./cities.txt")
	errcheck(err)
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		Obfuscator.Cities = append(Obfuscator.Cities, scanner.Text())
	}
	errcheck(scanner.Err())
	fmt.Printf("Cities: %d\n", len(Obfuscator.Cities))

	file, err = os.Open("./streets.txt")
	errcheck(err)
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		Obfuscator.Streets = append(Obfuscator.Streets, scanner.Text())
	}
	errcheck(scanner.Err())
	fmt.Printf("Streets: %d\n", len(Obfuscator.Streets))
}
func processCommandLine() {
	p := flag.Int("s", 0, "seed for random numbers. Default is to use a random seed.")
	flag.Parse()
	Obfuscator.seed = int64(*p)
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
	Obfuscator.db = db

	processCommandLine()
	if Obfuscator.seed != 0 {
		rand.Seed(Obfuscator.seed)
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
	}

	loadNames()
	randomizeStuff()
}
