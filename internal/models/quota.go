package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type QuotaPolicy struct {
	Name            string      `json:"name"`
	RateLimit       int         `json:"rate_limit"`
	RateLimitWindow int         `json:"rate_limit_window"`
	TokenQuotaDaily int64       `json:"token_quota_daily"`
	Models          StringArray `json:"models"`
	Description     string      `json:"description"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type QuotaUsageDaily struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Date          time.Time `json:"date"`
	ModelID       string    `json:"model_id"`
	RequestCount  int       `json:"request_count"`
	TokenCount    int64     `json:"token_count"`
	InputTokens   int64     `json:"input_tokens"`
	OutputTokens  int64     `json:"output_tokens"`
}

type QuotaCheckResult struct {
	Allowed      bool   `json:"allowed"`
	Reason       string `json:"reason,omitempty"`
	DailyTokens  int64  `json:"daily_tokens"`
	DailyLimit   int64  `json:"daily_limit"`
	RateRemaining int   `json:"rate_remaining"`
	RateLimit    int   `json:"rate_limit"`
}

type UsageRecord struct {
	ID           uuid.UUID `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	UserID       uuid.UUID `json:"user_id"`
	APIKeyID     *uuid.UUID `json:"api_key_id,omitempty"`
	ModelID      string    `json:"model_id"`
	BackendURL   string    `json:"backend_url"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	LatencyMs    int       `json:"latency_ms"`
	StatusCode   int       `json:"status_code"`
	ErrorMsg     string    `json:"error_msg,omitempty"`
	RequestPath  string    `json:"request_path"`
	RequestMethod string   `json:"request_method"`
}

type UsageStats struct {
	TotalRequests   int   `json:"total_requests"`
	TotalTokens     int64 `json:"total_tokens"`
	InputTokens     int64 `json:"input_tokens"`
	OutputTokens    int64 `json:"output_tokens"`
	AvgLatencyMs    int   `json:"avg_latency_ms"`
	ErrorCount      int   `json:"error_count"`
}

// QuotaStore 配额数据访问层
type QuotaStore struct {
	db *sql.DB
}

func NewQuotaStore(db *sql.DB) *QuotaStore {
	return &QuotaStore{db: db}
}

func (s *QuotaStore) GetPolicy(name string) (*QuotaPolicy, error) {
	policy := &QuotaPolicy{}
	query := `
		SELECT name, rate_limit, rate_limit_window, token_quota_daily, models, description, created_at, updated_at
		FROM quota_policies WHERE name = $1`

	err := s.db.QueryRow(query, name).Scan(
		&policy.Name, &policy.RateLimit, &policy.RateLimitWindow,
		&policy.TokenQuotaDaily, &policy.Models, &policy.Description,
		&policy.CreatedAt, &policy.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return policy, err
}

func (s *QuotaStore) ListPolicies() ([]*QuotaPolicy, error) {
	query := `
		SELECT name, rate_limit, rate_limit_window, token_quota_daily, models, description, created_at, updated_at
		FROM quota_policies ORDER BY name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*QuotaPolicy
	for rows.Next() {
		policy := &QuotaPolicy{}
		err := rows.Scan(
			&policy.Name, &policy.RateLimit, &policy.RateLimitWindow,
			&policy.TokenQuotaDaily, &policy.Models, &policy.Description,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}

func (s *QuotaStore) CreateOrUpdatePolicy(policy *QuotaPolicy) error {
	query := `
		INSERT INTO quota_policies (name, rate_limit, rate_limit_window, token_quota_daily, models, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name) DO UPDATE SET
			rate_limit = EXCLUDED.rate_limit,
			rate_limit_window = EXCLUDED.rate_limit_window,
			token_quota_daily = EXCLUDED.token_quota_daily,
			models = EXCLUDED.models,
			description = EXCLUDED.description,
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at`

	return s.db.QueryRow(query,
		policy.Name, policy.RateLimit, policy.RateLimitWindow,
		policy.TokenQuotaDaily, policy.Models, policy.Description,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)
}

func (s *QuotaStore) DeletePolicy(name string) error {
	_, err := s.db.Exec("DELETE FROM quota_policies WHERE name = $1", name)
	return err
}

// GetDailyUsage 获取用户当天的 Token 使用量
func (s *QuotaStore) GetDailyUsage(userID uuid.UUID, date time.Time) (int64, error) {
	var total int64
	query := `
		SELECT COALESCE(SUM(token_count), 0)
		FROM quota_usage_daily
		WHERE user_id = $1 AND date = $2`

	err := s.db.QueryRow(query, userID, date.Format("2006-01-02")).Scan(&total)
	return total, err
}

// IncrementUsage 增加使用统计
func (s *QuotaStore) IncrementUsage(userID uuid.UUID, modelID string, inputTokens, outputTokens int) error {
	query := `
		INSERT INTO quota_usage_daily (user_id, date, model_id, request_count, token_count, input_tokens, output_tokens)
		VALUES ($1, CURRENT_DATE, $2, 1, $3 + $4, $3, $4)
		ON CONFLICT (user_id, date, model_id) DO UPDATE SET
			request_count = quota_usage_daily.request_count + 1,
			token_count = quota_usage_daily.token_count + EXCLUDED.token_count,
			input_tokens = quota_usage_daily.input_tokens + EXCLUDED.input_tokens,
			output_tokens = quota_usage_daily.output_tokens + EXCLUDED.output_tokens`

	_, err := s.db.Exec(query, userID, modelID, inputTokens, outputTokens)
	return err
}

// RecordUsage 记录详细使用记录
func (s *QuotaStore) RecordUsage(record *UsageRecord) error {
	query := `
		INSERT INTO usage_records
		(timestamp, user_id, api_key_id, model_id, backend_url, input_tokens, output_tokens, latency_ms, status_code, error_msg, request_path, request_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	return s.db.QueryRow(query,
		record.Timestamp, record.UserID, record.APIKeyID, record.ModelID,
		record.BackendURL, record.InputTokens, record.OutputTokens,
		record.LatencyMs, record.StatusCode, record.ErrorMsg,
		record.RequestPath, record.RequestMethod,
	).Scan(&record.ID)
}

// GetUsageStats 获取使用统计
func (s *QuotaStore) GetUsageStats(userID uuid.UUID, startDate, endDate time.Time) (*UsageStats, error) {
	stats := &UsageStats{}
	query := `
		SELECT
			COALESCE(SUM(request_count), 0),
			COALESCE(SUM(token_count), 0),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0)
		FROM quota_usage_daily
		WHERE user_id = $1 AND date BETWEEN $2 AND $3`

	err := s.db.QueryRow(query, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(
		&stats.TotalRequests, &stats.TotalTokens, &stats.InputTokens, &stats.OutputTokens,
	)
	return stats, err
}

// CleanupOldRecords 清理旧的使用记录（7天前的数据）
func (s *QuotaStore) CleanupOldRecords() error {
	query := `DELETE FROM usage_records WHERE timestamp < CURRENT_TIMESTAMP - INTERVAL '7 days'`
	_, err := s.db.Exec(query)
	return err
}
