package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	var err error
	connStr := "postgres://admin:password@localhost:5433/chatdb?sslmode=disable"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
	createTable()
}

func createTable() {
	query := `
		CREATE TABLE IF NOT EXISTS messages(
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
	log.Println("Database tables verified.")

}

func SaveMessage(content []byte) {
	query := `INSERT INTO messages (content) VALUES ($1)`
	_, err := db.Exec(query, string(content))
	if err != nil {
		log.Println("Error saving message to DB:", err)
	}
}

func GetMessageHistory() [][]byte {
	query := `
		SELECT content FROM (
			SELECT id, content FROM messages ORDER BY id DESC LIMIT 50
		) sub ORDER BY id ASC;`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error fetching history:", err)
		return nil
	}
	defer rows.Close()
	var history [][]byte
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		history = append(history, []byte(content))
	}
	return history
}
