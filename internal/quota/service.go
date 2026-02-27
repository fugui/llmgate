package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"llmgate/internal/models"
)

type Service struct {
	store  *models.QuotaStore
	redis  *redis.Client
}

func NewService(store *models.QuotaStore, redis *redis.Client) *Service {
	return &Service{
		store:  store,
		redis:  redis,
	}
}

// CheckQuota 检查用户配额
func (s *Service) CheckQuota(userID uuid.UUID, modelID string) (*models.QuotaCheckResult, error) {
	result := &models.QuotaCheckResult{
		Allowed: true,
	}

	// 获取用户配额策略
	// 这里简化处理，假设用户配额策略已加载
	// 实际应该根据用户配置获取

	policy, err := s.store.GetPolicy("default")
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("policy not found")
	}

	// 检查模型权限
	hasModelAccess := false
	for _, m := range policy.Models {
		if m == "*" || m == modelID {
			hasModelAccess = true
			break
		}
	}
	if !hasModelAccess {
		result.Allowed = false
		result.Reason = "model not allowed"
		return result, nil
	}

	// 检查速率限制
	ctx := context.Background()
	rateKey := fmt.Sprintf("rate:%s:%s", userID.String(), time.Now().Format("YYYY-MM-DD-HH-MM"))

	current, err := s.redis.Get(ctx, rateKey).Int()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if current >= policy.RateLimit {
		result.Allowed = false
		result.Reason = "rate limit exceeded"
		result.RateLimit = policy.RateLimit
		result.RateRemaining = 0
		return result, nil
	}

	result.RateLimit = policy.RateLimit
	result.RateRemaining = policy.RateLimit - current - 1

	// 检查 Token 配额
	dailyTokens, err := s.store.GetDailyUsage(userID, time.Now())
	if err != nil {
		return nil, err
	}

	result.DailyTokens = dailyTokens
	result.DailyLimit = policy.TokenQuotaDaily

	if dailyTokens >= policy.TokenQuotaDaily {
		result.Allowed = false
		result.Reason = "daily token quota exceeded"
		return result, nil
	}

	return result, nil
}

// IncrementRate 增加速率计数
func (s *Service) IncrementRate(userID uuid.UUID, window int) error {
	ctx := context.Background()
	rateKey := fmt.Sprintf("rate:%s:%s", userID.String(), time.Now().Format("2006-01-02-15-04"))

	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, rateKey)
	pipe.Expire(ctx, rateKey, time.Duration(window)*time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

// DeductQuota 扣除配额
func (s *Service) DeductQuota(userID uuid.UUID, modelID string, inputTokens, outputTokens int) error {
	// 增加速率计数
	if err := s.IncrementRate(userID, 60); err != nil {
		return err
	}

	// 增加 Token 使用统计
	return s.store.IncrementUsage(userID, modelID, inputTokens, outputTokens)
}

// GetQuotaStats 获取配额统计
func (s *Service) GetQuotaStats(userID uuid.UUID) (map[string]interface{}, error) {
	policy, err := s.store.GetPolicy("default")
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("policy not found")
	}

	dailyTokens, err := s.store.GetDailyUsage(userID, time.Now())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"daily_tokens_used": dailyTokens,
		"daily_tokens_limit": policy.TokenQuotaDaily,
		"rate_limit":          policy.RateLimit,
		"rate_window":         policy.RateLimitWindow,
		"models_allowed":      policy.Models,
		"reset_time":          "00:00",
	}, nil
}
