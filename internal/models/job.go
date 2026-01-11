package models

import "time"

type JobType string
type JobStatus string

const (
	JobTypeSyncDB    JobType = "SYNC_DB"
	JobTypeCSVImport JobType = "CSV_IMPORT"

	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

type Job struct {
	ID              string    `json:"id"`
	Type            JobType   `json:"type"`
	Status          JobStatus `json:"status"`
	ProgressCurrent int       `json:"progress_current"`
	ProgressTotal   int       `json:"progress_total"`
	ResultSummary   string    `json:"result_summary"` // JSON string
	CreatedAt       time.Time `json:"created_at"`
}
