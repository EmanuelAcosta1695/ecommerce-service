package main

import (
	"log"

	"github.com/EmanuelAcosta1695/ecomm/db"
)

func main() {
	db, err := db.NewDatabase()

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to the database successfully")
}
