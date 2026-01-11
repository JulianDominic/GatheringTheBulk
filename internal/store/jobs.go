package store

import (
	"database/sql"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
)

func (s *SQLiteStore) CreateJob(job *models.Job) error {
	query := `INSERT INTO jobs (id, type, status, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, job.ID, job.Type, job.Status, job.CreatedAt)
	return err
}

func (s *SQLiteStore) GetJob(id string) (*models.Job, error) {
	query := `SELECT id, type, status, progress_current, progress_total, result_summary, created_at FROM jobs WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var job models.Job
	var resultSummary sql.NullString

	err := row.Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.ProgressCurrent,
		&job.ProgressTotal,
		&resultSummary,
		&job.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if resultSummary.Valid {
		job.ResultSummary = resultSummary.String
	}

	return &job, nil
}

func (s *SQLiteStore) UpdateJobStatus(id string, status models.JobStatus) error {
	_, err := s.db.Exec("UPDATE jobs SET status = ? WHERE id = ?", status, id)
	return err
}

func (s *SQLiteStore) UpdateJobProgress(id string, current, total int) error {
	_, err := s.db.Exec("UPDATE jobs SET progress_current = ?, progress_total = ? WHERE id = ?", current, total, id)
	return err
}

func (s *SQLiteStore) CompleteJob(id string, summary string) error {
	_, err := s.db.Exec("UPDATE jobs SET status = ?, result_summary = ?, progress_current = progress_total WHERE id = ?", models.JobStatusCompleted, summary, id)
	return err
}

func (s *SQLiteStore) FailJob(id string, errorMsg string) error {
	// Store error in summary for now
	_, err := s.db.Exec("UPDATE jobs SET status = ?, result_summary = ? WHERE id = ?", models.JobStatusFailed, errorMsg, id)
	return err
}
