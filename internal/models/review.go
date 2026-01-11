package models

type ReviewItem struct {
	ID             int    `json:"id"`
	JobID          string `json:"job_id"`
	IssueType      string `json:"issue_type"`
	RawData        string `json:"raw_data"`        // JSON of imported row
	ProposedValues string `json:"proposed_values"` // JSON of fields
}
