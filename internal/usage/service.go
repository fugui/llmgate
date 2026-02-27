package usage

import (
	"sync"
	"time"

	"llmgate/internal/models"
)

// Service 使用记录服务
type Service struct {
	store      *models.QuotaStore
	buffer     []*models.UsageRecord
	mu         sync.Mutex
	flushSize  int
	flushInterval time.Duration
}

func NewService(store *models.QuotaStore) *Service {
	s := &Service{
		store:         store,
		buffer:        make([]*models.UsageRecord, 0, 100),
		flushSize:     10,
		flushInterval: 5 * time.Second,
	}

	// 启动后台刷新
	go s.backgroundFlush()

	return s
}

// RecordUsage 记录使用（异步写入）
func (s *Service) RecordUsage(record *models.UsageRecord) {
	s.mu.Lock()
	s.buffer = append(s.buffer, record)
	shouldFlush := len(s.buffer) >= s.flushSize
	s.mu.Unlock()

	if shouldFlush {
		s.Flush()
	}
}

// Flush 刷新缓冲区到数据库
func (s *Service) Flush() {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}

	records := make([]*models.UsageRecord, len(s.buffer))
	copy(records, s.buffer)
	s.buffer = s.buffer[:0]
	s.mu.Unlock()

	// 批量写入
	for _, record := range records {
		// 忽略错误，继续写入其他记录
		_ = s.store.RecordUsage(record)
	}
}

// backgroundFlush 后台定时刷新
func (s *Service) backgroundFlush() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.Flush()
	}
}

// CleanupOldRecords 清理旧记录
func (s *Service) CleanupOldRecords() error {
	return s.store.CleanupOldRecords()
}

// GetUsageStats 获取使用统计
func (s *Service) GetUsageStats(userID string, startDate, endDate time.Time) (*models.UsageStats, error) {
	// 解析 UUID
	// 这里简化处理
	return &models.UsageStats{}, nil
}
