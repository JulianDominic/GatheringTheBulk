package store

import (
	"github.com/JulianDominic/GatheringTheBulk/internal/models"
)

// ListInventory returns paged inventory items joined with card data.
func (s *SQLiteStore) ListInventory(limit, offset int, searchQuery string) ([]models.InventoryItem, int, error) {
	var total int

	// Base count query
	countQuery := "SELECT COUNT(*) FROM inventory i LEFT JOIN cards c ON i.scryfall_id = c.scryfall_id"
	var args []interface{}

	if searchQuery != "" {
		countQuery += " WHERE c.name LIKE ?"
		args = append(args, "%"+searchQuery+"%")
	}

	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
        SELECT i.id, i.scryfall_id, i.quantity, i.condition, i.is_foil, i.language, i.location,
               c.name, c.set_code, c.collector_number, c.image_uri
        FROM inventory i
        LEFT JOIN cards c ON i.scryfall_id = c.scryfall_id
    `
	if searchQuery != "" {
		query += " WHERE c.name LIKE ?" // Args already prepared above
	}

	query += " ORDER BY i.id DESC LIMIT ? OFFSET ?" // Sorted by Last Added
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.ScryfallID, &item.Quantity, &item.Condition, &item.IsFoil, &item.Language, &item.Location,
			&item.CardName, &item.SetCode, &item.CollectorNumber, &item.ImageURI,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (s *SQLiteStore) AddInventory(item models.InventoryItem) error {
	// Check for existing item to merge quantities
	var existingID int
	var existingQty int
	err := s.db.QueryRow(`
        SELECT id, quantity FROM inventory 
        WHERE scryfall_id = ? AND condition = ? AND is_foil = ? AND language = ?
    `, item.ScryfallID, item.Condition, item.IsFoil, item.Language).Scan(&existingID, &existingQty)

	if err == nil {
		// Item exists, update quantity
		_, err = s.db.Exec("UPDATE inventory SET quantity = ? WHERE id = ?", existingQty+item.Quantity, existingID)
		return err
	}

	// Item does not exist, insert new
	_, err = s.db.Exec(`
        INSERT INTO inventory (scryfall_id, quantity, condition, is_foil, language, location)
        VALUES (?, ?, ?, ?, ?, ?)
    `, item.ScryfallID, item.Quantity, item.Condition, item.IsFoil, item.Language, item.Location)
	return err
}

func (s *SQLiteStore) UpdateInventory(item models.InventoryItem) error {
	_, err := s.db.Exec(`
        UPDATE inventory 
        SET quantity=?, condition=?, is_foil=?, language=?, location=?
        WHERE id=?
    `, item.Quantity, item.Condition, item.IsFoil, item.Language, item.Location, item.ID)
	return err
}

func (s *SQLiteStore) DeleteInventory(id int) error {
	_, err := s.db.Exec("DELETE FROM inventory WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) SearchInventoryNames(query string) ([]string, error) {
	rows, err := s.db.Query(`
        SELECT DISTINCT c.name FROM inventory i 
        JOIN cards c ON i.scryfall_id = c.scryfall_id 
        WHERE c.name LIKE ? 
        ORDER BY c.name ASC 
        LIMIT 10`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

func (s *SQLiteStore) GetInventoryByID(id int) (*models.InventoryItem, error) {
	query := `
        SELECT i.id, i.scryfall_id, i.quantity, i.condition, i.is_foil, i.language, i.location,
               c.name, c.set_code, c.collector_number, c.image_uri
        FROM inventory i
        LEFT JOIN cards c ON i.scryfall_id = c.scryfall_id
        WHERE i.id = ?
    `
	var item models.InventoryItem
	err := s.db.QueryRow(query, id).Scan(
		&item.ID, &item.ScryfallID, &item.Quantity, &item.Condition, &item.IsFoil, &item.Language, &item.Location,
		&item.CardName, &item.SetCode, &item.CollectorNumber, &item.ImageURI,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
