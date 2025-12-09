package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/duckdb/duckdb-go/v2"

	"github.com/tupyy/rvtools/parser"
)

var (
	excelFile  string
	sqliteFile string
	dbPath     string
)

func main() {
	flag.StringVar(&excelFile, "excel-file", "", "path of RVTools excel file")
	flag.StringVar(&sqliteFile, "sqlite-file", "", "path of forklift sqlite file")
	flag.StringVar(&dbPath, "db-path", "", "Path to db file")
	flag.Parse()

	if excelFile == "" && sqliteFile == "" {
		log.Fatal("either -excel-file or -sqlite-file must be provided")
	}
	if excelFile != "" && sqliteFile != "" {
		log.Fatal("only one of -excel-file or -sqlite-file can be provided")
	}

	now := time.Now()

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

	var p *parser.Parser
	if excelFile != "" {
		p = parser.NewRvToolParser(db, excelFile)
	} else {
		p = parser.NewSqliteParser(db, sqliteFile)
	}

	inventory, err := p.Parse()
	if err != nil {
		log.Fatalf("parsing: %v", err)
	}

	fmt.Printf("parsing time: %s\n", time.Since(now))

	data, _ := json.MarshalIndent(inventory, "", "  ")
	fmt.Println(string(data))
}

func loadExtensions(db *sql.DB) error {
	_, err := db.Exec("install excel;load excel;")
	return err
}
