package main

import (
	"bufio"
	"database/sql"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Support command:
//  - make: CreateMigration a migration file, "go Migrate make create_user_table"
//  - Migrate: Migrate the database to the latest version, "go Migrate Migrate"
//  - Rollback <step?>: Rollback the database to an old version, "go Migrate Rollback", default Rollback 1 version
// 			    or "go Migrate Rollback 2", it means Rollback 2 versions

var (
	createMigrationSql = "CREATE TABLE IF NOT EXISTS migrations (" +
		"id int(10) UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY," +
		"migration varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL," +
		"batch int(11) NOT NULL " +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
	queryAllMigrationSql  = "SELECT * FROM migrations;"
	queryLastMigrationSql = "SELECT batch FROM migrations ORDER BY batch DESC;"
	updateMigrationSql    = "INSERT INTO migrations (migration, batch) VALUES DummyString;"
	dropTableSql          = "DROP TABLE IF EXISTS `DummyTable`;"
	createTableSql        = "CREATE TABLE DummyTable (\n" +
		"id int(10) UNSIGNED NOT NULL, \n" +
		"created_at timestamp NULL DEFAULT NULL, \n" +
		"updated_at timestamp NULL DEFAULT NULL\n" +
		");"
	insertSql = "INSERT INTO `<TABLE_NAME>` \n" +
		"() \n" +
		"VALUES \n" +
		"();"
)

// Migration files save path
var migrationPath = "./database/migrations/"
var seedPath = "./database/seeds/"

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
	conf()
	createDir(migrationPath)
	createDir(seedPath)
	InitMigration()
}

func conf() {
	err := godotenv.Load()
	checkErr(err)

	// Database configuration
	conn := os.Getenv("DB_CONNECTION")
	dbName := os.Getenv("DB_DATABASE")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	str := []string{username, ":", password, "@/", dbName}
	connInfo := strings.Join(str, "")

	db, err = sql.Open(conn, connInfo)
	if err != nil {
		checkErr(err)
	}
}

// Connect to database
// Read configurations' info in .env
func InitMigration() {
	// CreateMigration migrations table if not exist
	_, err := db.Exec(createMigrationSql)
	checkErr(err)
}

// CreateMigration dir
func createDir(path string) {
	// Check ./database/migrations is exist, CreateMigration it if not
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// Check err and output
func checkErr(err error) {
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func main() {
	command := os.Args[1]

	if strings.Compare(command, "make:migration") == 0 {

		// ***********************
		// CreateMigration a migration file
		// ./migrate make xxx
		// ***********************

		fileName := os.Args[2]

		if len(fileName) < 0 {
			color.Red("Please enter a migration file name")
			os.Exit(2)
		}

		_, err := CreateMigration(fileName)
		checkErr(err)

		color.Green("CreateMigration successfully!")

	} else if strings.Compare(command, "up") == 0 {

		// ****************
		// Migrate database
		// ./migrate migrate
		// ****************
		err := Migrate()
		checkErr(err)

		color.Green("Migrate successfully!")

	} else if strings.Compare(command, "down") == 0 {

		// ********
		// Rollback
		// ./migrate rollback OR ./migrate rollback 3
		// ********
		var step string
		if len(os.Args) < 3 {
			// Default step is 1
			step = "1"
		} else {
			step = os.Args[2]
		}

		err := Rollback(step)
		checkErr(err)

		color.Green("Rollback successfully!")

	} else if strings.Compare(command, "refresh") == 0 {

		// **********************************
		// Refresh - Rollback and re-Migrate
		// ./migrate refresh
		// **********************************

		ok, err := Refresh()
		checkErr(err)

		if ok {
			color.Green("Refresh successfully!")
		} else {
			color.Blue("Refresh nothing")
		}
	} else if strings.Compare(command, "make:seeder") == 0 {

		// **********************************
		// Create seeder
		// ./migrate make:seeder <filename>
		// **********************************

		filename := os.Args[2]
		if len(filename) < 0 {
			color.Red("Please enter seeder name")
			os.Exit(2)
		}

		_, err := CreateSeeder(filename)
		checkErr(err)

		color.Green("Create seeder successfully!")
	} else {
		color.Red("Command not support: %v", command)
	}
}

// **** Command line tools ****
// CreateMigration a migration file in /database/migration/
// It will CreateMigration a directory named <timestamp>_name,
// there are two sql files inside: up.sql and down.sql
func CreateMigration(name string) (string, error) {

	var (
		err      error
		upFile   *os.File
		downFile *os.File
	)

	timestamp := time.Now().Format("20060102150405")
	str := []string{migrationPath, timestamp, "_", name}
	dirName := strings.Join(str, "")
	createDir(dirName)

	// Match table creation
	// use CreateMigration.stub template for table creation
	// use blank.stub template for others
	reg := regexp.MustCompile(`^create_(\w+)_table$`)

	upFile, err = os.Create(dirName + "/up.sql")
	if err != nil {
		return "", err
	}

	downFile, err = os.Create(dirName + "/down.sql")
	if err != nil {
		return "", err
	}

	defer upFile.Close()
	defer downFile.Close()

	upWriter := bufio.NewWriter(upFile)
	downWriter := bufio.NewWriter(downFile)

	if reg.MatchString(name) {
		r := strings.NewReplacer("create_", "", "_table", "")
		tableName := r.Replace(name)
		_, err = upWriter.WriteString(strings.Replace(createTableSql, "DummyTable", tableName, -1))
		if err != nil {
			return "", err
		}

		upWriter.Flush()

		_, err = downWriter.WriteString(strings.Replace(dropTableSql, "DummyTable", tableName, -1))
		if err != nil {
			return "", err
		}

		downWriter.Flush()
	} else {
		_, err = upWriter.WriteString("")
		if err != nil {
			return "", err
		}

		_, err = downWriter.WriteString("")
		if err != nil {
			return "", err
		}
	}

	color.Green("Created: %v", name)
	return dirName, nil
}

// Migration
func Migrate() error {
	var (
		fSlices   []string
		arr       []string
		batch     int
		files     []os.FileInfo
		err       error
		rows      *sql.Rows
		lastBatch int
		dbMigrate []string
		toMigrate []string
		m         *Migration
		insertStr string
		symbol    string
		upSql     []byte
	)

	// List migrations files
	files, err = ioutil.ReadDir(migrationPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		arr = strings.Split(f.Name(), ".")
		fSlices = append(fSlices, arr[0])
	}

	// Check migration version in database
	rows, err = db.Query(queryAllMigrationSql)
	if err != nil {
		return err
	}

	lastRow := db.QueryRow(queryLastMigrationSql)
	lastRow.Scan(&lastBatch)
	batch = lastBatch + 1

	defer rows.Close()

	if lastBatch == 0 {
		// No migration record in database, all migrations should to be Migrate
		toMigrate = fSlices
	} else {
		// Get migrated files' name
		for rows.Next() {
			m, err = scanRow(rows)
			if err != nil {
				return err
			}

			dbMigrate = append(dbMigrate, m.Migration)
		}

		// Compare and get which migration not migrated yet
		for _, v := range fSlices {
			if !sliceContain(dbMigrate, v) {
				toMigrate = append(toMigrate, v)
			}
		}
	}

	// Nothing to Migrate, stop and log fatal
	toMigrateLen := len(toMigrate)
	if toMigrateLen == 0 {
		color.Blue("Nothing migrated")
		os.Exit(2)
	}

	// Migrate
	for i, v := range toMigrate {

		// Read up.sql
		upSql, err = ioutil.ReadFile(migrationPath + v + "/up.sql")
		if err != nil {
			return err
		}

		_, err = db.Exec(string(upSql))
		if err != nil {
			return err
		}

		color.Green("Migrated: %v", v)

		// Calculate the batch number, which is need to Migrate
		if i+1 == toMigrateLen {
			symbol = ""
		} else {
			symbol = ","
		}

		insertStr += "('" + v + "', " + strconv.Itoa(batch) + ")" + symbol
	}

	// Connect sql update statement
	updateMigrationSql = strings.Replace(updateMigrationSql, "DummyString", insertStr, -1)

	_, err = db.Exec(updateMigrationSql)
	if err != nil {
		return err
	}

	return nil
}

// Rollback migration
func Rollback(step string) error {

	var (
		lastBatch   int
		toBatch     int
		err         error
		rows        *sql.Rows
		rollBackMig []string
		m           *Migration
		downSql     []byte
	)

	lastRow := db.QueryRow(queryLastMigrationSql)
	lastRow.Scan(&lastBatch)

	if i, err := strconv.Atoi(step); err == nil {
		if lastBatch >= i {
			toBatch = lastBatch - (i - 1)
		} else {
			color.Red("Can not Rollback %d steps", i)
			return err
		}
	}

	// Which migrations need to be Rollback
	rows, err = db.Query("SELECT * FROM migrations WHERE `batch`>=" + strconv.Itoa(toBatch))
	if err != nil {
		return err
	}

	// Rollback slice
	for rows.Next() {
		m, err = scanRow(rows)
		if err != nil {
			return err
		}

		rollBackMig = append(rollBackMig, m.Migration)
	}

	// Rolling back
	for _, v := range rollBackMig {

		downSql, err = ioutil.ReadFile(migrationPath + v + "/down.sql")
		if err != nil {
			return err
		}

		_, err = db.Exec(string(downSql))
		if err != nil {
			return err
		}

		color.Green("Rollback: %s", v)
	}

	// Delete migrations record
	_, err = db.Exec("DELETE FROM migrations WHERE `batch`>=" + strconv.Itoa(toBatch))
	if err != nil {
		return err
	}

	return nil
}

// Refresh migration: Rollback all and re-Migrate
func Refresh() (bool, error) {
	var (
		insertStr   string
		symbol      string
		fileByte    []byte
		err         error
		rows        *sql.Rows
		rollBackMig []string
		m           *Migration
	)

	rows, err = db.Query("SELECT * FROM migrations;")
	if err != nil {
		return false, err
	}

	for rows.Next() {
		m, err = scanRow(rows)
		if err != nil {
			return false, err
		}

		rollBackMig = append(rollBackMig, m.Migration)
	}

	// Rollback and re-Migrate
	fileLen := len(rollBackMig)
	if fileLen > 0 {
		for i, v := range rollBackMig {
			// down
			fileByte, err = ioutil.ReadFile(migrationPath + v + "/down.sql")
			if err != nil {
				return false, err
			}

			_, err = db.Exec(string(fileByte))
			if err != nil {
				return false, err
			}

			// up
			fileByte, err = ioutil.ReadFile(migrationPath + v + "/up.sql")
			if err != nil {
				return false, err
			}

			_, err = db.Exec(string(fileByte))
			if err != nil {
				return false, err
			}

			if i == fileLen-1 {
				symbol = ""
			} else {
				symbol = ","
			}

			insertStr += "('" + v + "', 1)" + symbol
		}

		// Update migrations table
		_, _ = db.Exec("TRUNCATE migrations;")
		_, err = db.Exec(strings.Replace(updateMigrationSql, "DummyString", insertStr, -1))
		if err != nil {
			return false, err
		}

		return true, nil

	} else {
		return false, nil
	}
}

// Create seeder
func CreateSeeder(name string) (string, error) {

	// check duplicate filename
	file, err := ioutil.ReadDir(seedPath)
	if err != nil {
		return "", err
	}

	for _, v := range file {
		filename := strings.Split(v.Name(), ".")[0]
		if strings.Compare(strings.ToLower(filename), strings.ToLower(name)) == 0 {
			color.Red("%s already exist.", name)
			os.Exit(2)
		}
	}

	// Create a seed file
	f, err := os.Create(seedPath + strings.ToLower(name) + ".sql")
	if err != nil {
		return "", err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = w.WriteString(insertSql)
	if err != nil {
		return "", err
	}

	w.Flush()

	return name, nil
}

// **** other helpers ****
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
