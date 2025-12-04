package scryfall

import (
	"database/sql"
	"log/slog"
	"slices"

	"github.com/JulianDominic/GatheringTheBulk/pkg/models"
)

type ScryfallRepo struct {
	db *sql.DB
}

func NewScryfallRepo(db *sql.DB) *ScryfallRepo {
	return &ScryfallRepo{db: db}
}

func (r *ScryfallRepo) AddCardsBatch(cards []models.Card) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchSize := 100

	for i := 0; i < len(cards); i += batchSize {
		end := min(i+batchSize, len(cards))

		batch := cards[i:end]
		vals := []any{}
		query := `INSERT INTO "cards_master" ("id", "name", "set", 'set_name', "collector_number", "rarity") VALUES `

		// build the placeholders
		for j, card := range batch {
			query += "(?, ?, ?, ?, ?, ?)"
			if j < len(batch)-1 {
				query += ","
			}
			vals = append(vals, card.Id, card.Name, card.Set, card.Set_name, card.Collector_number, card.Rarity)
		}

		stmt, err := tx.Prepare(query)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(vals...)
		stmt.Close()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ScryfallRepo) HasPulledToday(api_id string) bool {
	stt, err := r.db.Prepare(
		`SELECT "default_cards_id" FROM "scryfall_default_cards_log"`,
	)
	if err != nil {
		slog.Error("Failed to create prepared statement for CheckPullToday", "error", err)
		panic(err)
	}
	rows, err := stt.Query()
	if err != nil {
		slog.Error("Failed to query prepared statement for CheckPullToday", "error", err)
		panic(err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var default_cards_id string
		if err := rows.Scan(&default_cards_id); err != nil {
			slog.Error("Failed to scan row", "error", err)
			panic(err)
		}

		results = append(results, default_cards_id)
	}
	slog.Info("Comparing default_cards id", "results", results, "api_id", api_id)
	return slices.Contains(results, api_id)
}

func (r *ScryfallRepo) WritePullLog(api_id string) (sql.Result, error) {
	stt, err := r.db.Prepare(
		`INSERT INTO "scryfall_default_cards_log"
		("default_cards_id")
		VALUES
		(?)`,
	)
	if err != nil {
		slog.Error("Failed to create prepared statement for AddCard", "error", err)
		panic(err)
	}
	return stt.Exec(api_id)
}

func (r *ScryfallRepo) GetAllCardIDs() (map[string]bool, error) {
	rows, err := r.db.Query(`SELECT "id" FROM "cards_master"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existing[id] = true
	}
	return existing, nil
}
