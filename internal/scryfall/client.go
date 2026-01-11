package scryfall

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BulkDataResponse struct {
	Data []struct {
		Type        string `json:"type"`
		DownloadURI string `json:"download_uri"`
	} `json:"data"`
}

// FetchBulkDataURL calls Scryfall API to get the current download link for "default_cards".
func FetchBulkDataURL() (string, error) {
	resp, err := http.Get("https://api.scryfall.com/bulk-data")
	if err != nil {
		return "", fmt.Errorf("failed to fetch bulk-data list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("scryfall api returned status: %d", resp.StatusCode)
	}

	var result BulkDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode bulk-data json: %w", err)
	}

	for _, item := range result.Data {
		// "default_cards" contains every card in English, no duplicates (mostly).
		// "oracle_cards" is smaller but might miss printed variations user wants.
		// PRD says "Default Cards".
		if item.Type == "default_cards" {
			return item.DownloadURI, nil
		}
	}
	return "", fmt.Errorf("default_cards bulk data type not found in response")
}

// StreamBulkData initiates the download and returns the stream.
func StreamBulkData(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	return resp.Body, nil
}
