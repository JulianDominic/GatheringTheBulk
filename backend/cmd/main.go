package main

import (
	"log/slog"
	"os"
)

func main() {
	cfg := config{
		addr: ":10800",
	}

	api := api{
		config: cfg,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := api.run(api.mount()); err != nil {
		slog.Error("Failed to start the server", "error", err)
		os.Exit(1)
	}
}
