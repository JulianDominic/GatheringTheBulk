package store

import (
	"database/sql"
)

// GetSetting retrieves a value from system_settings by key.
// Returns empty string if not found.
func (s *SQLiteStore) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM system_settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting upserts a key-value pair into system_settings.
func (s *SQLiteStore) SetSetting(key, value string) error {
	query := `INSERT INTO system_settings (key, value) VALUES (?, ?) 
              ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	_, err := s.db.Exec(query, key, value)
	return err
}
