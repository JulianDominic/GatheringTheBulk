package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/scryfall"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

// SyncDatabaseTask downloads and ingests Scryfall data.
func SyncDatabaseTask(s store.Store, job *models.Job) (string, error) {
	log.Println("Starting Scryfall Sync...")

	url, err := scryfall.FetchBulkDataURL()
	if err != nil {
		return "", fmt.Errorf("failed to fetch download URL: %w", err)
	}

	stream, err := scryfall.StreamBulkData(url)
	if err != nil {
		return "", fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	dec := json.NewDecoder(stream)

	// Consume opening bracket
	t, err := dec.Token()
	if err != nil {
		return "", err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return "", fmt.Errorf("expected JSON array start")
	}

	count := 0
	batchSize := 2000
	var cardBatch []models.Card

	for dec.More() {
		var sfCard scryfall.Card
		if err := dec.Decode(&sfCard); err != nil {
			log.Printf("Error decoding card: %v", err)
			continue
		}

		cardBatch = append(cardBatch, models.Card{
			ScryfallID:      sfCard.ID,
			Name:            sfCard.Name,
			SetCode:         sfCard.Set,
			CollectorNumber: sfCard.CollectorNumber,
			ImageURI:        sfCard.GetFrontImage(),
		})

		count++
		if len(cardBatch) >= batchSize {
			if err := s.BatchUpsertCards(cardBatch); err != nil {
				return "", fmt.Errorf("batch commit failed: %w", err)
			}
			cardBatch = nil
			s.UpdateJobProgress(job.ID, count, 0)
		}
	}

	// Final Batch
	if len(cardBatch) > 0 {
		if err := s.BatchUpsertCards(cardBatch); err != nil {
			return "", fmt.Errorf("final commit failed: %w", err)
		}
	}

	// Consume closing bracket
	_, _ = dec.Token()

	ts := time.Now().Format("2006-01-02 15:04:05")
	if err := s.SetSetting("scryfall_last_sync", ts); err != nil {
		log.Printf("Warning: failed to update last sync setting: %v", err)
	}

	return fmt.Sprintf("Successfully synced %d cards", count), nil
}
