package worker

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

// ImportCSVTask reads a CSV file and processes it.
func ImportCSVTask(s store.Store, job *models.Job) (string, error) {
	filename := fmt.Sprintf("uploads/%s.csv", job.ID)
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open csv file: %w", err)
	}
	defer f.Close()
	defer os.Remove(filename) // Cleanup after processing

	reader := csv.NewReader(f)

	// Read Header
	header, err := reader.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read csv header: %w", err)
	}

	// Map header columns to indices
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Helper to get value safely
	getVal := func(row []string, key string) string {
		if idx, ok := colMap[key]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	successCount := 0
	reviewCount := 0
	rowCount := 0

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		rowCount++
		s.UpdateJobProgress(job.ID, rowCount, 0)

		// Extract fields
		name := getVal(row, "name")
		set := getVal(row, "set")
		cn := getVal(row, "collector_number")
		if cn == "" {
			cn = getVal(row, "cn")
		}

		qtyStr := getVal(row, "quantity")
		if qtyStr == "" {
			qtyStr = getVal(row, "qty")
		}
		qty, _ := strconv.Atoi(qtyStr)
		if qty < 1 {
			qty = 1
		}

		condition := getVal(row, "condition")
		if condition == "" {
			condition = "NM"
		}

		isFoil := strings.ToLower(getVal(row, "foil")) == "true"
		language := getVal(row, "language")
		if language == "" {
			language = "en"
		}

		var scryfallID string
		var matchErr error

		if set != "" && cn != "" {
			scryfallID, matchErr = s.FindCardBySetCN(set, cn)
		} else if name != "" {
			scryfallID, matchErr = s.FindSmartCard(name, set)
		} else {
			matchErr = fmt.Errorf("missing name or set/cn")
		}

		if matchErr == nil && scryfallID != "" {
			item := models.InventoryItem{
				ScryfallID: scryfallID,
				Quantity:   qty,
				Condition:  condition,
				IsFoil:     isFoil,
				Language:   language,
				Location:   "Imported",
			}
			if err := s.AddInventory(item); err != nil {
				s.AddReviewItem(job.ID, "DB_ERROR", mapRow(header, row), mapProp(qty, condition, isFoil, language))
				reviewCount++
			} else {
				successCount++
			}
		} else {
			issue := "AMBIGUOUS"
			if matchErr != nil && matchErr.Error() == "not found" {
				issue = "NOT_FOUND"
			}
			s.AddReviewItem(job.ID, issue, mapRow(header, row), mapProp(qty, condition, isFoil, language))
			reviewCount++
		}
	}

	summary := fmt.Sprintf(`{"success": %d, "review": %d}`, successCount, reviewCount)
	return summary, nil
}

func mapRow(header, row []string) map[string]string {
	m := make(map[string]string)
	for i, h := range header {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

func mapProp(qty int, cond string, foil bool, lang string) map[string]interface{} {
	return map[string]interface{}{
		"quantity":  qty,
		"condition": cond,
		"is_foil":   foil,
		"language":  lang,
	}
}
