package models

import (
	"database/sql"
	"time"
)

type Model struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ModelCreateRequest struct {
	ID          string               `json:"id" binding:"required"`
	Name        string               `json:"name" binding:"required"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	Backends    []BackendCreateInput `json:"backends"`
}

type ModelUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type BackendCreateInput struct {
	ID        string `json:"id" binding:"required"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url" binding:"required,url"`
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	Weight    int    `json:"weight"`
	Region    string `json:"region"`
	Enabled   bool   `json:"enabled"`
}

// ModelStore 模型配置数据访问层
type ModelStore struct {
	db *sql.DB
}

func NewModelStore(db *sql.DB) *ModelStore {
	return &ModelStore{db: db}
}

func (s *ModelStore) Create(model *Model) error {
	query := `
		INSERT INTO models (id, name, description, enabled)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			enabled = excluded.enabled,
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at`

	return s.db.QueryRow(query,
		model.ID, model.Name, model.Description, model.Enabled,
	).Scan(&model.CreatedAt, &model.UpdatedAt)
}

func (s *ModelStore) GetByID(id string) (*Model, error) {
	model := &Model{}
	query := `
		SELECT id, name, description, enabled, created_at, updated_at
		FROM models WHERE id = ?`

	err := s.db.QueryRow(query, id).Scan(
		&model.ID, &model.Name, &model.Description, &model.Enabled,
		&model.CreatedAt, &model.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return model, err
}

func (s *ModelStore) List() ([]*Model, error) {
	query := `
		SELECT id, name, description, enabled, created_at, updated_at
		FROM models ORDER BY created_at`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*Model
	for rows.Next() {
		model := &Model{}
		err := rows.Scan(
			&model.ID, &model.Name, &model.Description, &model.Enabled,
			&model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func (s *ModelStore) ListEnabled() ([]*Model, error) {
	query := `
		SELECT id, name, description, enabled, created_at, updated_at
		FROM models WHERE enabled = 1 ORDER BY created_at`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*Model
	for rows.Next() {
		model := &Model{}
		err := rows.Scan(
			&model.ID, &model.Name, &model.Description, &model.Enabled,
			&model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func (s *ModelStore) Update(model *Model) error {
	query := `
		UPDATE models SET
			name = ?, description = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := s.db.Exec(query,
		model.Name, model.Description, model.Enabled, model.ID,
	)
	return err
}

func (s *ModelStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM models WHERE id = ?", id)
	return err
}
