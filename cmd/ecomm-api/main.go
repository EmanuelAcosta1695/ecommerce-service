package main

import (
	"log"

	"github.com/EmanuelAcosta1695/ecomm/db"
	handler "github.com/EmanuelAcosta1695/ecomm/ecomm-api/handler"
	"github.com/EmanuelAcosta1695/ecomm/ecomm-api/server"
	"github.com/EmanuelAcosta1695/ecomm/ecomm-api/storer"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := db.NewDatabase()

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to the database successfully")

	st := storer.NewPySQLStorer(db.GetDB())
	srv := server.NewServer(st)
	hdl := handler.NewHandler(srv)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")
}
