package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	Name            string      `json:"name"`
	KeyHash         string      `json:"-"`
	KeyPrefix       string      `json:"key_prefix"`
	Models          StringArray `json:"models"`
	RateLimit       int         `json:"rate_limit"`
	RateLimitWindow int         `json:"rate_limit_window"`
	Enabled         bool        `json:"enabled"`
	ExpiresAt       *time.Time  `json:"expires_at,omitempty"`
	LastUsedAt      *time.Time  `json:"last_used_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type APIKeyCreateRequest struct {
	Name            string     `json:"name" binding:"required"`
	Models          []string   `json:"models"`
	RateLimit       int        `json:"rate_limit"`
	RateLimitWindow int        `json:"rate_limit_window"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

type APIKeyResponse struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	Name            string      `json:"name"`
	KeyPrefix       string      `json:"key_prefix"`
	Models          StringArray `json:"models"`
	RateLimit       int         `json:"rate_limit"`
	RateLimitWindow int         `json:"rate_limit_window"`
	Enabled         bool        `json:"enabled"`
	ExpiresAt       *time.Time  `json:"expires_at,omitempty"`
	LastUsedAt      *time.Time  `json:"last_used_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}

type APIKeyWithSecret struct {
	APIKeyResponse
	Key string `json:"key"` // 仅创建时返回一次
}

func (k *APIKey) ToResponse() APIKeyResponse {
	return APIKeyResponse{
		ID:              k.ID,
		UserID:          k.UserID,
		Name:            k.Name,
		KeyPrefix:       k.KeyPrefix,
		Models:          k.Models,
		RateLimit:       k.RateLimit,
		RateLimitWindow: k.RateLimitWindow,
		Enabled:         k.Enabled,
		ExpiresAt:       k.ExpiresAt,
		LastUsedAt:      k.LastUsedAt,
		CreatedAt:       k.CreatedAt,
	}
}

// APIKeyStore API Key 数据访问层
type APIKeyStore struct {
	db *sql.DB
}

func NewAPIKeyStore(db *sql.DB) *APIKeyStore {
	return &APIKeyStore{db: db}
}

func (s *APIKeyStore) Create(key *APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, name, key_hash, key_prefix, models, rate_limit, rate_limit_window, enabled, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(query,
		key.UserID, key.Name, key.KeyHash, key.KeyPrefix, key.Models,
		key.RateLimit, key.RateLimitWindow, key.Enabled, key.ExpiresAt,
	).Scan(&key.ID, &key.CreatedAt, &key.UpdatedAt)
}

func (s *APIKeyStore) GetByID(id uuid.UUID) (*APIKey, error) {
	key := &APIKey{}
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, models, rate_limit, rate_limit_window,
		       enabled, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix,
		&key.Models, &key.RateLimit, &key.RateLimitWindow, &key.Enabled,
		&key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (s *APIKeyStore) GetByHash(hash string) (*APIKey, error) {
	key := &APIKey{}
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, models, rate_limit, rate_limit_window,
		       enabled, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE key_hash = $1 AND enabled = true`

	err := s.db.QueryRow(query, hash).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix,
		&key.Models, &key.RateLimit, &key.RateLimitWindow, &key.Enabled,
		&key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (s *APIKeyStore) GetByKeyPrefix(prefix string) (*APIKey, error) {
	key := &APIKey{}
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, models, rate_limit, rate_limit_window,
		       enabled, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE key_prefix = $1 AND enabled = true`

	err := s.db.QueryRow(query, prefix).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix,
		&key.Models, &key.RateLimit, &key.RateLimitWindow, &key.Enabled,
		&key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (s *APIKeyStore) ListByUser(userID uuid.UUID) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, models, rate_limit, rate_limit_window,
		       enabled, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		key := &APIKey{}
		err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix,
			&key.Models, &key.RateLimit, &key.RateLimitWindow, &key.Enabled,
			&key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *APIKeyStore) Update(key *APIKey) error {
	query := `
		UPDATE api_keys SET
			name = $1, models = $2, rate_limit = $3, rate_limit_window = $4,
			enabled = $5, expires_at = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7`

	_, err := s.db.Exec(query,
		key.Name, key.Models, key.RateLimit, key.RateLimitWindow,
		key.Enabled, key.ExpiresAt, key.ID,
	)
	return err
}

func (s *APIKeyStore) UpdateLastUsed(id uuid.UUID) error {
	_, err := s.db.Exec("UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	return err
}

func (s *APIKeyStore) Delete(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM api_keys WHERE id = $1", id)
	return err
}

func (s *APIKeyStore) CountByUser(userID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE user_id = $1", userID).Scan(&count)
	return count, err
}
