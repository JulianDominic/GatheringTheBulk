package scryfall

type Card struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Set             string     `json:"set"`
	CollectorNumber string     `json:"collector_number"`
	ImageURIs       *ImageURIs `json:"image_uris"`
	CardFaces       []CardFace `json:"card_faces"`
}

type ImageURIs struct {
	Small  string `json:"small"`
	Normal string `json:"normal"`
	Large  string `json:"large"`
}

type CardFace struct {
	ImageURIs *ImageURIs `json:"image_uris"`
}

// GetFrontImage returns the URL of the front face.
func (c *Card) GetFrontImage() string {
	if c.ImageURIs != nil && c.ImageURIs.Normal != "" {
		return c.ImageURIs.Normal
	}
	if len(c.CardFaces) > 0 && c.CardFaces[0].ImageURIs != nil {
		return c.CardFaces[0].ImageURIs.Normal
	}
	return "" // Placeholder or empty
}
