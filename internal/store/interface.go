package store

import (
	"database/sql"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
)

// Store defines all data access operations
type Store interface {
	// Inventory
	ListInventory(limit, offset int, searchQuery string) ([]models.InventoryItem, int, error)
	AddInventory(item models.InventoryItem) error
	UpdateInventory(item models.InventoryItem) error
	DeleteInventory(id int) error
	GetInventoryByID(id int) (*models.InventoryItem, error)
	SearchInventoryNames(query string) ([]string, error)

	// Cards
	SearchCards(query, preferredSet string) ([]CardSearchResult, error)
	GetCardByScryfallID(id string) (*CardSearchResult, error)
	FindCardBySetCN(set, cn string) (string, error)
	FindSmartCard(name, set string) (string, error)
	BatchUpsertCards(cards []models.Card) error

	// Jobs
	CreateJob(job *models.Job) error
	GetJob(id string) (*models.Job, error)
	UpdateJobStatus(id string, status models.JobStatus) error
	UpdateJobProgress(id string, current, total int) error
	CompleteJob(id string, summary string) error
	FailJob(id string, errorMsg string) error

	// Review
	AddReviewItem(jobID, issueType string, rawData map[string]string, proposedValues map[string]interface{}) error
	ListReviewItems() ([]models.ReviewItem, error)
	GetReviewItem(id int) (*models.ReviewItem, error)
	DeleteReviewItem(id int) error
	CountReviewItems() (int, error)

	// Settings
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}
