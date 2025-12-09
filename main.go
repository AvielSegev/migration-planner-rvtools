package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/duckdb/duckdb-go/v2"

	"github.com/tupyy/rvtools/parser"
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

	if err := loadExtensions(db); err != nil {
		log.Printf("loading extensions: %v", err)
		os.Exit(1)
	}

	p := parser.NewRvToolParser(db, excelFile)
	inventory, err := p.Parse()
	if err != nil {
		log.Fatalf("parsing RVTools: %v", err)
	}

	data, _ := json.MarshalIndent(inventory, "", "  ")
	fmt.Println(string(data))
}

func loadExtensions(db *sql.DB) error {
	_, err := db.Exec("install excel;load excel;")
	return err
}
