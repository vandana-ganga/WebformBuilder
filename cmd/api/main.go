package main

import (
	"net/http"
	"log"

	"github.com/labstack/echo/v4"
	"webform-builder/internal/config"
	"webform-builder/internal/database"

)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = database.RunMigrations(
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.GET("/health", func(c echo.Context) error {

		return c.JSON(
			http.StatusOK,
			map[string]string{
				"status": "ok",
			},
		)
	})

	RegisterRoutes(e, db)

	e.Logger.Fatal(e.Start(":8080"))
}