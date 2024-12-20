package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Update these values with your database credentials
	const (
		host     = "localhost"
		port     = 5432
		user     = "your_username"
		password = "your_password"
		dbname   = "groupchat_service"
	)

	// Connect to the postgres database to a new database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to postgres:", err)
	}
	defer db.Close()

	// Read and execute the migration file
	migration, err := os.ReadFile("internal/database/migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatal("Error reading migration file:", err)
	}

	_, err = db.Exec(string(migration))
	if err != nil {
		log.Fatal("Error executing migration:", err)
	}

	fmt.Println("Database initialized successfully!")
}
