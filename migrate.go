package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"gante/config"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Support command:
//  - make: create a migration file, "go migrate make create_user_table"
//  - migrate: migrate the database to the latest version, "go migrate migrate"
//  - rollback <step?>: rollback the database to an old version, "go migrate rollback", default rollback 1 version
// 			    or "go migrate rollback 2", it means rollback 2 versions

var (
	createMigrationSql = "CREATE TABLE IF NOT EXISTS migrations (" +
		"id int(10) UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY," +
		"migration varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL," +
		"batch int(11) NOT NULL " +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
	queryAllMigration  = "SELECT * FROM migrations;"
	queryLastMigration = "SELECT batch FROM migrations ORDER BY batch DESC;"
	updateMigrationSql = "INSERT INTO migrations (migration, batch) VALUES DummyString;"
)

// Migration files save path
var path = "./database/migrations/"

type rowScanner interface {
	Scan(dst ...interface{}) error
}

type Migration struct {
	ID        int64
	Migration string
	Batch     int64
}

var db *sql.DB

func init() {
	createMigrationDir()
	InitMigration()
}

// Connect to database
// Read configurations' info in .env
func InitMigration() {
	db = config.Conf()

	// Create migrations table if not exist
	_, err := db.Exec(createMigrationSql)
	checkErr(err)
}

// Create a migration dir
func createMigrationDir() {
	// Check ./database/migrations is exist, create it if not
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// Check err and output
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Copy file from src to dst
func copyFile(src, dst, tableName, structName string) {
	in, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Replace Dummy string
	out := bytes.Replace(in, []byte("DummyTable"), []byte(tableName), -1)
	out = bytes.Replace(out, []byte("DummyStruct"), []byte(structName), -1)
	out = bytes.Replace(out, []byte("DummyDB"), []byte(structName+"DB"), -1)
	if err = ioutil.WriteFile(dst, out, 0666); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println("Create a migration file successfully.")
}

// Upper the first letter
func UpFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func main() {
	command := os.Args[1]

	if strings.Compare(command, "make") == 0 {

		// ***********************
		// Create a migration file
		// ***********************

		inputName := os.Args[2]

		if len(inputName) < 0 {
			log.Fatal("Please enter migration file name")
		}

		timestamp := time.Now().Format("20060102150405")
		str := []string{path, timestamp, "_", inputName, ".go"}
		fileName := strings.Join(str, "")

		// Match table creation
		// use create.stub template for table creation
		// use blank.stub template for others
		reg := regexp.MustCompile(`^create_(\w+)_table$`)

		var (
			tableName  string
			structName string
			template   string
		)

		strArr := strings.Split(inputName, "_")
		for _, v := range strArr {
			structName += UpFirst(v)
		}

		if reg.MatchString(inputName) {
			tableName = strArr[1]
			template = "./config/stubs/create.stub"
		} else {
			template = "./config/stubs/blank.stub"
		}

		copyFile(template, fileName, tableName, structName)

	} else if strings.Compare(command, "migrate") == 0 {

		// ****************
		// Migrate database
		// ****************

		var (
			fSlices []string
			arr     []string
			batch   int
		)

		// List migrations files
		files, err := ioutil.ReadDir(path)
		checkErr(err)
		for _, f := range files {
			arr = strings.Split(f.Name(), ".")
			fSlices = append(fSlices, arr[0])
		}

		// Check migration version in database
		rows, err := db.Query(queryAllMigration)
		checkErr(err)

		var (
			lastBatch int
			dbMigrate []string
			toMigrate []string
		)
		lastRow := db.QueryRow(queryLastMigration)
		lastRow.Scan(&lastBatch)
		batch = lastBatch + 1

		defer rows.Close()

		if lastBatch == 0 {
			// No migration record in database, all migrations should to be migrate
			toMigrate = fSlices
		} else {
			// Get migrated files' name
			for rows.Next() {
				m, err := scanRow(rows)
				checkErr(err)
				dbMigrate = append(dbMigrate, m.Migration)
			}

			// Compare and get which migration not migrated yet
			for _, v := range fSlices {
				if !sliceContain(dbMigrate, v) {
					toMigrate = append(toMigrate, v)
				}
			}
		}

		var (
			insertSlice []string
			insertStr   string
			symbol      string
		)

		// Nothing to migrate, stop and log fatal
		toMigrateLen := len(toMigrate)
		if toMigrateLen == 0 {
			log.Fatal("Nothing migrated")
		}

		// Migrate
		for i, v := range toMigrate {
			cmd := exec.Command("sh", "-c", "go run ./database/migrations/"+v+".go up")
			_, err = cmd.CombinedOutput()
			checkErr(err)
			log.Println("Migrated: " + v)

			// Calculate the batch number, which is need to migrate
			if i+1 == toMigrateLen {
				symbol = ";"
			} else {
				symbol = ","
			}
			insertSlice = []string{"('", v, "',", strconv.Itoa(batch), ")", symbol}
			insertStr += strings.Join(insertSlice, "")
		}

		// Connect sql update statement
		updateMigrationSql = strings.Replace(updateMigrationSql, "DummyString", insertStr, -1)

		_, err = db.Exec(updateMigrationSql)
		checkErr(err)

	} else if strings.Compare(command, "rollback") == 0 {

		// ********
		// Rollback
		// ********

		var (
			step      string
			lastBatch int
			toBatch   int
		)

		lastRow := db.QueryRow(queryLastMigration)
		lastRow.Scan(&lastBatch)

		if len(os.Args) < 3 {
			// Default step is 1
			step = "1"
		} else {
			step = os.Args[2]
		}

		if i, err := strconv.Atoi(step); err == nil {
			if lastBatch >= i {
				toBatch = lastBatch - (i - 1)
			} else {
				log.Fatalf("Can not rollback %d steps", i)
			}
		}

		// Which migrations need to be rollback
		rows, err := db.Query("SELECT * FROM migrations WHERE `batch`>=" + strconv.Itoa(toBatch))
		checkErr(err)

		// Rollback slice
		var rollBackMig []string
		for rows.Next() {
			m, err := scanRow(rows)
			checkErr(err)
			rollBackMig = append(rollBackMig, m.Migration)
		}

		// Rolling back
		for _, v := range rollBackMig {
			cmd := exec.Command("sh", "-c", "go run ./database/migrations/"+v+".go down")
			_, err = cmd.CombinedOutput()
			checkErr(err)

			log.Printf("Rollback: %s", v)
		}

		// Delete migrations record
		_, err = db.Exec("DELETE FROM migrations WHERE `batch`>=" + strconv.Itoa(toBatch))
		checkErr(err)
	}
}

// Check if slice contain a string
func sliceContain(s []string, str string) bool {
	for _, v := range s {
		if strings.Compare(v, str) == 0 {
			return true
		}
	}
	return false
}

// Map sql row to struct
func scanRow(s rowScanner) (*Migration, error) {
	var (
		id        int64
		migration sql.NullString
		batch     int64
	)

	if err := s.Scan(&id, &migration, &batch); err != nil {
		return nil, err
	}

	mig := &Migration{
		ID:        id,
		Migration: migration.String,
		Batch:     batch,
	}

	return mig, nil
}
