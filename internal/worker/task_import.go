package worker

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

// importResult holds the outcome of processing a single CSV row.
// It is sent from workers to the collector.
type importResult struct {
	Success       bool
	InventoryItem models.InventoryItem // Populated if Success is true

	// Review Data (Populated if Success is false)
	IssueType      string
	RawRow         []string
	ProposedValues map[string]interface{}
}

// ImportCSVTask reads a CSV file and processes it using a concurrent worker pool.
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

	// Map header columns to indices (ReadOnly for workers)
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// ---------------------------------------------------------
	// CONCURRENCY PIPELINE SETUP
	// ---------------------------------------------------------

	// 1. Channels
	numWorkers := runtime.NumCPU()
	if numWorkers < 2 {
		numWorkers = 2
	}
	// Buffer channels slightly to smooth out bursts
	rowChan := make(chan []string, numWorkers*2)
	resultChan := make(chan importResult, numWorkers*2)

	var wg sync.WaitGroup

	// 2. Helper for safely extracting values (Closure captures colMap)
	getVal := func(row []string, key string) string {
		if idx, ok := colMap[key]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	// 3. Worker Function
	worker := func() {
		defer wg.Done()
		for row := range rowChan {
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

			// Proposed values for Review (if needed)
			props := mapProp(qty, condition, isFoil, language)

			// Logic
			var scryfallID string
			var matchErr error

			// DB Read Operation (Safe for concurrent usage)
			if set != "" && cn != "" {
				scryfallID, matchErr = s.FindCardBySetCN(set, cn)
			} else if name != "" {
				scryfallID, matchErr = s.FindSmartCard(name, set)
			} else {
				matchErr = fmt.Errorf("missing name or set/cn")
			}

			res := importResult{
				RawRow:         row,
				ProposedValues: props,
			}

			if matchErr == nil && scryfallID != "" {
				res.Success = true
				res.InventoryItem = models.InventoryItem{
					ScryfallID: scryfallID,
					Quantity:   qty,
					Condition:  condition,
					IsFoil:     isFoil,
					Language:   language,
					Location:   "Imported",
				}
			} else {
				res.Success = false
				res.IssueType = "AMBIGUOUS"
				if matchErr != nil && matchErr.Error() == "not found" {
					res.IssueType = "NOT_FOUND"
				}
			}

			resultChan <- res
		}
	}

	// 4. Start Workers
	fmt.Printf("Starting import with %d workers\n", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	// 5. Start Feeder (Reads file and pushes to workers)
	go func() {
		defer close(rowChan)
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				// Skip malformed lines, maybe log them?
				continue
			}
			rowChan <- row
		}
	}()

	// 6. Start Closer (Waits for workers then closes result channel)
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// ---------------------------------------------------------
	// COLLECTOR LOOP (Main Thread - Single Writer)
	// ---------------------------------------------------------

	successCount := 0
	reviewCount := 0
	totalProcessed := 0
	lastUpdate := time.Now()

	for res := range resultChan {
		totalProcessed++

		// Batch progress updates (every 50 items or 1 second)
		if totalProcessed%50 == 0 || time.Since(lastUpdate) > time.Second {
			s.UpdateJobProgress(job.ID, totalProcessed, 0)
			lastUpdate = time.Now()
		}

		if res.Success {
			// Write Operation
			if err := s.AddInventory(res.InventoryItem); err != nil {
				// DB Write Error -> Send to Review
				s.AddReviewItem(job.ID, "DB_ERROR", mapRow(header, res.RawRow), res.ProposedValues)
				reviewCount++
			} else {
				successCount++
			}
		} else {
			// Write Operation
			s.AddReviewItem(job.ID, res.IssueType, mapRow(header, res.RawRow), res.ProposedValues)
			reviewCount++
		}
	}

	// Final progress update
	s.UpdateJobProgress(job.ID, totalProcessed, totalProcessed)

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
