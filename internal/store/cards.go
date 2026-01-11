package store

import (
	"fmt"
	"strings"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
)

type CardSearchResult struct {
	ScryfallID      string `json:"scryfall_id"`
	Name            string `json:"name"`
	SetCode         string `json:"set_code"`
	CollectorNumber string `json:"collector_number"`
	ImageURI        string `json:"image_uri"`
	Label           string `json:"label"` // Helper for UI
}

// SearchCards performs a prefix/fuzzy search on card names.
func (s *SQLiteStore) SearchCards(query, preferredSet string) ([]CardSearchResult, error) {
	if query == "" {
		return nil, nil // Or empty list
	}

	// Simple LIKE search with ordering preference
	sqlQuery := `
        SELECT scryfall_id, name, set_code, collector_number, image_uri
        FROM cards
        WHERE name LIKE ?
        ORDER BY 
            CASE WHEN LOWER(set_code) = LOWER(?) THEN 0 ELSE 1 END,
            name ASC, 
            set_code DESC
        LIMIT 20
    `
	q := "%" + query + "%"
	rows, err := s.db.Query(sqlQuery, q, preferredSet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CardSearchResult
	for rows.Next() {
		var c CardSearchResult
		if err := rows.Scan(&c.ScryfallID, &c.Name, &c.SetCode, &c.CollectorNumber, &c.ImageURI); err != nil {
			return nil, err
		}
		c.Label = fmt.Sprintf("%s (%s #%s)", c.Name, c.SetCode, c.CollectorNumber)
		results = append(results, c)
	}
	return results, nil
}

func (s *SQLiteStore) GetCardByScryfallID(id string) (*CardSearchResult, error) {
	query := `
        SELECT scryfall_id, name, set_code, collector_number, image_uri
        FROM cards
        WHERE scryfall_id = ?
    `
	var c CardSearchResult
	err := s.db.QueryRow(query, id).Scan(&c.ScryfallID, &c.Name, &c.SetCode, &c.CollectorNumber, &c.ImageURI)
	if err != nil {
		return nil, err
	}
	c.Label = fmt.Sprintf("%s (%s #%s)", c.Name, c.SetCode, c.CollectorNumber)
	return &c, nil
}

func (s *SQLiteStore) FindCardBySetCN(set, cn string) (string, error) {
	var id string
	err := s.db.QueryRow("SELECT scryfall_id FROM cards WHERE LOWER(set_code) = ? AND collector_number = ?", strings.ToLower(set), cn).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("not found")
	}
	return id, nil
}

func (s *SQLiteStore) FindSmartCard(name, set string) (string, error) {
	query := "SELECT scryfall_id FROM cards WHERE LOWER(name) = ?"
	args := []interface{}{strings.ToLower(name)}

	if set != "" {
		query += " AND LOWER(set_code) = ?"
		args = append(args, strings.ToLower(set))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return "", fmt.Errorf("db error")
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return "", fmt.Errorf("not found")
	}

	if len(ids) == 1 {
		return ids[0], nil // Perfect match
	}

	return "", fmt.Errorf("ambiguous: %d matches", len(ids))
}

func (s *SQLiteStore) BatchUpsertCards(cards []models.Card) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO cards (scryfall_id, name, set_code, collector_number, image_uri) VALUES (?, ?, ?, ?, ?)`
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, c := range cards {
		_, err = stmt.Exec(c.ScryfallID, c.Name, c.SetCode, c.CollectorNumber, c.ImageURI)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
