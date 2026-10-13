package main

import (
	"log"

	"github.com/labstack/echo/v4"

	"webform-builder/internal/config"
	"webform-builder/internal/database"
	"webform-builder/internal/messaging"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(
			"failed to load config:",
			err,
		)
	}

	db, err := database.ConnectDB(
		cfg.DatabaseURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Migrations.
	err = database.RunMigrations(
		cfg.DatabaseURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	// NATS.
	natsClient, err := messaging.NewNATS(
		cfg.NATSURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer natsClient.Close()

	// Create JetStream stream.
	err = messaging.CreateSubmissionStream(
		natsClient.JS,
	)

	if err != nil {
		log.Fatal(err)
	}

	// Echo.
	e := echo.New()

	RegisterRoutes(
		e,
		db,
		natsClient.JS,
	)

	addr := ":" + cfg.Port

	log.Println(
		"API server running on " + addr,
	)

	e.Logger.Fatal(
		e.Start(addr),
	)
}