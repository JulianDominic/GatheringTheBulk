package cards

import (
	"database/sql"
	"log/slog"

	"github.com/JulianDominic/GatheringTheBulk/pkg/models"
)

type CardsRepo struct {
	db *sql.DB
}

func NewCardsRepo(db *sql.DB) *CardsRepo {
	return &CardsRepo{db: db}
}

func (r *CardsRepo) AddCardBySetAndCollectorNumber(set string, collector_number int, quantity int) {
	master_card := r.findCardInMasterBySetAndCollectorNumber(set, collector_number)
	if master_card == nil {
		panic("Could not find the specified card")
	}
	// check if the card already exists in the database
	rows, err := r.db.Query(
		`SELECT "id", "quantity"
		FROM "cards_owned" WHERE "id" = ?;`,
		master_card.Id,
	)
	if err != nil {
		slog.Error("Failed to create query for AddCardBySetAndCollectorNumber", "error", err)
		panic(err)
	}
	defer rows.Close()

	id := ""
	qty_stored := -1
	for rows.Next() {
		err := rows.Scan(&id, &qty_stored)
		if err != nil {
			panic(err)
		}
	}
	if id != "" && qty_stored != -1 {
		// if it does, just increase the quantity
		stt, err := r.db.Prepare(
			`UPDATE "cards_owned"
			SET "quantity" = ?
			WHERE "id" = ?
			`,
		)
		if err != nil {
			slog.Error("Failed to create prepared statement for AddCardBySetAndCollectorNumber", "error", err)
			panic(err)
		}
		_, err = stt.Exec(qty_stored+quantity, id)
		if err != nil {
			slog.Error("Failed to execute prepared statement for AddCardBySetAndCollectorNumber", "error", err)
			panic(err)
		}
	} else {
		// otherwise, card doesn't exist, add a new entry
		stt, err := r.db.Prepare(
			`INSERT INTO "cards_owned"
		("id", "quantity")
		VALUES
		(?, ?)`,
		)
		if err != nil {
			slog.Error("Failed to create prepared statement for AddCard", "error", err)
			panic(err)
		}
		_, err = stt.Exec(master_card.Id, quantity)
		if err != nil {
			slog.Error("Failed to execute prepared statement for AddCardBySetAndCollectorNumber", "error", err)
			panic(err)
		}
	}
}

func (r *CardsRepo) findCardInMasterBySetAndCollectorNumber(set string, collector_number int) *models.Card {
	rows, err := r.db.Query(
		`SELECT "id", "name", "set", "set_name", "collector_number", "rarity"
		FROM "cards_master" WHERE "set" LIKE ? AND "collector_number" = ?;`,
		set,
		collector_number,
	)
	if err != nil {
		slog.Error("Failed to create query for findCardInMasterBySetAndCollectorNumber", "error", err)
		panic(err)
	}
	defer rows.Close()

	c := new(models.Card)
	for rows.Next() {
		err := rows.Scan(&c.Id, &c.Name, &c.Set, &c.Set_name, &c.Collector_number, &c.Rarity)
		if err != nil {
			panic(err)
		}
		// slog.Info("Card found", "card", c)
	}
	return c
}
