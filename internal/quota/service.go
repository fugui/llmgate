package quota

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"llmgate/internal/models"
)

// RateCounter 内存速率计数器
type RateCounter struct {
	counts    map[string]int
	windows   map[string]time.Time
	windowSec int           // 窗口大小（秒）
	mu        sync.RWMutex
}

func NewRateCounter() *RateCounter {
	rc := &RateCounter{
		counts:    make(map[string]int),
		windows:   make(map[string]time.Time),
		windowSec: 60, // 默认 60 秒窗口
	}
	// 启动每分钟清理任务
	go rc.cleanupLoop()
	return rc
}

// cleanupLoop 每分钟清理过期的计数器
func (rc *RateCounter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rc.cleanup()
	}
}

// cleanup 清理过期的计数器条目
func (rc *RateCounter) cleanup() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	for key, window := range rc.windows {
		if now.Sub(window) >= time.Duration(rc.windowSec)*time.Second {
			delete(rc.counts, key)
			delete(rc.windows, key)
		}
	}
}

// getWindowKey 获取当前窗口的 key
// 使用窗口起始时间作为 key，确保同一窗口内的请求使用相同的 key
func (rc *RateCounter) getWindowKey(userID string, window int) string {
	now := time.Now()
	// 计算窗口起始时间
	windowStart := now.Truncate(time.Duration(window) * time.Second)
	return fmt.Sprintf("%s:%d", userID, windowStart.Unix())
}

// Increment 增加计数并返回当前计数
func (rc *RateCounter) Increment(userID string, window int) int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	key := rc.getWindowKey(userID, window)
	
	// 检查是否需要重置窗口
	if lastWindow, exists := rc.windows[key]; !exists || time.Since(lastWindow) >= time.Duration(window)*time.Second {
		rc.counts[key] = 0
	}
	
	rc.windows[key] = time.Now()
	rc.counts[key]++
	return rc.counts[key]
}

// GetCount 获取当前计数
func (rc *RateCounter) GetCount(userID string, window int) int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	key := rc.getWindowKey(userID, window)
	return rc.counts[key]
}

type Service struct {
	store  *models.QuotaStore
	rateCounter *RateCounter
}

func NewService(store *models.QuotaStore) *Service {
	return &Service{
		store:       store,
		rateCounter: NewRateCounter(),
	}
}

// CheckQuota 检查用户配额
func (s *Service) CheckQuota(userID uuid.UUID, policyName string, modelID string) (*models.QuotaCheckResult, error) {
	result := &models.QuotaCheckResult{
		Allowed: true,
	}

	// 如果未指定策略，使用 default
	if policyName == "" {
		policyName = "default"
	}

	// 获取用户配额策略
	policy, err := s.store.GetPolicy(policyName)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("policy not found: %s", policyName)
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
	current := s.rateCounter.GetCount(userID.String(), policy.RateLimitWindow)

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
	s.rateCounter.Increment(userID.String(), window)
	return nil
}

// DeductQuota 扣除配额
func (s *Service) DeductQuota(userID uuid.UUID, policyName string, modelID string, inputTokens, outputTokens int) error {
	// 如果未指定策略，使用 default
	if policyName == "" {
		policyName = "default"
	}

	// 获取策略以使用正确的窗口大小
	policy, err := s.store.GetPolicy(policyName)
	if err != nil {
		return err
	}
	window := 60
	if policy != nil && policy.RateLimitWindow > 0 {
		window = policy.RateLimitWindow
	}

	// 增加速率计数
	if err := s.IncrementRate(userID, window); err != nil {
		return err
	}

	// 增加 Token 使用统计
	return s.store.IncrementUsage(userID, modelID, inputTokens, outputTokens)
}

// GetQuotaStats 获取配额统计
func (s *Service) GetQuotaStats(userID uuid.UUID, policyName string) (map[string]interface{}, error) {
	// 如果未指定策略，使用 default
	if policyName == "" {
		policyName = "default"
	}

	policy, err := s.store.GetPolicy(policyName)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("policy not found: %s", policyName)
	}

	dailyTokens, err := s.store.GetDailyUsage(userID, time.Now())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"daily_tokens_used":  dailyTokens,
		"daily_tokens_limit": policy.TokenQuotaDaily,
		"rate_limit":         policy.RateLimit,
		"rate_window":        policy.RateLimitWindow,
		"models_allowed":     policy.Models,
		"reset_time":         "00:00",
	}, nil
}
