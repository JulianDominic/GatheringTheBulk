package models

// type image_uris struct {
// 	Small  string
// 	Normal string
// 	Large  string
// }

type Card struct {
	// Image_uris       image_uris
	// Mana_cost        string
	// Type_line        string
	// Oracle_text      string
	// Cmc              float32
	Id               string
	Name             string
	Set              string
	Set_name         string
	Collector_number int
	Rarity           string
}
