package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:123456@localhost:5432/Proyect_base?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening db: %v", err)
	}
	defer db.Close()

	// Query pg_constraint
	rows, err := db.Query(`
		SELECT n.nspname AS schema_name, c.relname AS table_name, con.conname AS constraint_name, con.contype
		FROM pg_constraint con
		JOIN pg_class c ON c.oid = con.conrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname LIKE '%companies%' OR con.conname LIKE '%companies%';
	`)
	if err != nil {
		log.Fatalf("Error querying constraints: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- Constraints ---")
	for rows.Next() {
		var schema, table, name, ctype string
		if err := rows.Scan(&schema, &table, &name, &ctype); err != nil {
			log.Fatalf("Error scanning: %v", err)
		}
		fmt.Printf("Schema: %s, Table: %s, Name: %s, Type: %s\n", schema, table, name, ctype)
	}

	// Query pg_indexes
	indexRows, err := db.Query(`
		SELECT schemaname, tablename, indexname, indexdef
		FROM pg_indexes
		WHERE tablename LIKE '%companies%' OR indexname LIKE '%companies%';
	`)
	if err != nil {
		log.Fatalf("Error querying indexes: %v", err)
	}
	defer indexRows.Close()

	fmt.Println("\n--- Indexes ---")
	for indexRows.Next() {
		var schema, table, name, def string
		if err := indexRows.Scan(&schema, &table, &name, &def); err != nil {
			log.Fatalf("Error scanning: %v", err)
		}
		fmt.Printf("Schema: %s, Table: %s, Name: %s, Def: %s\n", schema, table, name, def)
	}
}
