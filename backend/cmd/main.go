package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/JulianDominic/GatheringTheBulk/internal/constants"
	"github.com/JulianDominic/GatheringTheBulk/internal/env"
	"github.com/JulianDominic/GatheringTheBulk/internal/scryfall"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := config{
		addr: ":10800",
		db: dbConfig{
			dsn: env.GetString("DB_DSN", "./data/GTB.db"),
		},
	}

	api := api{
		config: cfg,
	}

	// logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// db
	db, err := sql.Open("sqlite", cfg.db.dsn)
	if err != nil {
		slog.Error("Failed to open DB", "error", err)
		os.Exit(constants.EXT_ERR_DB_OPEN)
	}
	defer db.Close() // TODO: handle this err?
	if err = db.Ping(); err != nil {
		slog.Error("Failed to connect to DB", "error", err)
		os.Exit(constants.EXT_ERR_DB_CONNECT)
	}
	slog.Info("Connected to DB", "dsn", cfg.db.dsn)

	// scryfall
	sfallRepo := scryfall.NewScryfallRepo(db)
	sfallService := scryfall.NewScryfallService(sfallRepo)
	sfallService.PullMasterData()
	slog.Info("DB has been initialised")

	if err := api.run(api.mount()); err != nil {
		slog.Error("Failed to start the server", "error", err)
		os.Exit(constants.EXT_ERR_SERVER_START)
	}
}
