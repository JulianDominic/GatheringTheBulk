package models

type InventoryItem struct {
	ID         int    `json:"id"`
	ScryfallID string `json:"scryfall_id"`
	Quantity   int    `json:"quantity"`
	Condition  string `json:"condition"`
	IsFoil     bool   `json:"is_foil"`
	Language   string `json:"language"`
	Location   string `json:"location"`

	// Joined fields for display (populated via JOINs)
	CardName        string `json:"card_name"`
	SetCode         string `json:"set_code"`
	CollectorNumber string `json:"collector_number"`
	ImageURI        string `json:"image_uri"`
}
