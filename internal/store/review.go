package store

import (
	"encoding/json"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
)

// AddReviewItem inserts a record into review_queue.
func (s *SQLiteStore) AddReviewItem(jobID, issueType string, rawData map[string]string, proposedValues map[string]interface{}) error {
	rawBytes, _ := json.Marshal(rawData)
	proposedBytes, _ := json.Marshal(proposedValues)

	query := `INSERT INTO review_queue (job_id, issue_type, raw_data, proposed_values) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, jobID, issueType, string(rawBytes), string(proposedBytes))
	return err
}

// ListReviewItems returns all pending review items.
func (s *SQLiteStore) ListReviewItems() ([]models.ReviewItem, error) {
	query := `SELECT id, job_id, issue_type, raw_data, proposed_values FROM review_queue ORDER BY id ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ReviewItem
	for rows.Next() {
		var item models.ReviewItem
		if err := rows.Scan(&item.ID, &item.JobID, &item.IssueType, &item.RawData, &item.ProposedValues); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetReviewItem retrieves a single item by ID.
func (s *SQLiteStore) GetReviewItem(id int) (*models.ReviewItem, error) {
	query := `SELECT id, job_id, issue_type, raw_data, proposed_values FROM review_queue WHERE id = ?`
	var item models.ReviewItem
	err := s.db.QueryRow(query, id).Scan(&item.ID, &item.JobID, &item.IssueType, &item.RawData, &item.ProposedValues)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// DeleteReviewItem removes a record from review_queue.
func (s *SQLiteStore) DeleteReviewItem(id int) error {
	_, err := s.db.Exec("DELETE FROM review_queue WHERE id = ?", id)
	return err
}

// CountReviewItems returns number of pending reviews.
func (s *SQLiteStore) CountReviewItems() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM review_queue").Scan(&count)
	return count, err
}
