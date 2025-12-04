package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/tupyy/rvtools/definitions"
)

var (
	excelFile string
	dbPath    string
)

func main() {
	flag.StringVar(&excelFile, "excel-file", "", "path of excel file")
	flag.StringVar(&dbPath, "db-path", "", "Path to db file")
	flag.Parse()

	c, err := duckdb.NewConnector(dbPath, nil)
	if err != nil {
		log.Fatalf("could not initialize new connector: %s", err.Error())
	}
	defer c.Close()

	db := sql.OpenDB(c)
	defer db.Close()

	if err := loadExtentions(db); err != nil {
		log.Printf("loading extentions: %v", err)
		os.Exit(1)
	}

	count := readExcel(db, excelFile)
	if count == 0 {
		log.Panicf("reading excel: %s", err)
		os.Exit(1)
	}

	ctx := context.Background()

	datastore, _ := readDatastore(ctx, db)
	data, _ := json.MarshalIndent(datastore, "", "  ")
	fmt.Println(string(data))

	osList, _ := readOs(ctx, db)
	osData, _ := json.MarshalIndent(osList, "", "  ")
	fmt.Println(string(osData))

	log.Printf("number of sheets created: %d", count)
}

func loadExtentions(db *sql.DB) error {
	_, err := db.Exec("install excel;load excel;")
	return err
}

func tableExists(db *sql.DB, table string) bool {
	var count int
	err := db.QueryRow("SELECT count(*) FROM information_schema.tables WHERE table_name = ?", table).Scan(&count)
	return err == nil && count > 0
}

func readExcel(db *sql.DB, excelFile string) int {
	countSheet := 0
	for _, s := range definitions.Sheets {
		if _, err := db.Exec(fmt.Sprintf(definitions.CreateTableStmt, strings.ToLower(s), excelFile, s)); err != nil {
			log.Printf("failed to create sheet %s: %v", s, err)
			continue
		}
		countSheet++
	}
	return countSheet
}

func readDatastore(ctx context.Context, db *sql.DB) ([]definitions.Datastore, error) {
	query := definitions.SelectDatastoreStmt
	if !tableExists(db, "vhost") || !tableExists(db, "vhba") {
		query = definitions.SelectDatastoreSimpleStmt
	}

	var results []definitions.Datastore
	if err := sqlscan.Select(ctx, db, &results, query); err != nil {
		return nil, fmt.Errorf("scanning datastores: %w", err)
	}
	return results, nil
}

func readOs(ctx context.Context, db *sql.DB) ([]definitions.Os, error) {
	var results []definitions.Os
	if err := sqlscan.Select(ctx, db, &results, definitions.SelectOsStmt); err != nil {
		return nil, fmt.Errorf("scanning os: %w", err)
	}
	return results, nil
}
