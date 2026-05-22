// Package concurrency 提供并发请求限制与统计功能
package concurrency

import (
	"sync"
	"time"
)

// Limiter 并发限制器与统计器
type Limiter struct {
	mu           sync.RWMutex

	activeConcurrency int    // 当前活跃并发数
	peakToday         int    // 今日最高并发数
	peakDate          string // 峰值对应日期
	peakInterval      int    // 当前采样窗口内的最高并发数
}

// NewLimiter 创建新的并发限制统计器
func NewLimiter() *Limiter {
	return &Limiter{}
}

// Acquire 尝试获取并发许可
// 仅更新当前并发计数和峰值，始终返回 true
func (l *Limiter) Acquire(userID string) bool {
	l.mu.Lock()
	l.activeConcurrency++
	current := l.activeConcurrency
	today := time.Now().Format("2006-01-02")
	if today != l.peakDate {
		l.peakToday = 0
		l.peakDate = today
	}
	if current > l.peakToday {
		l.peakToday = current
	}
	if current > l.peakInterval {
		l.peakInterval = current
	}
	l.mu.Unlock()

	return true
}

// Release 释放并发许可
func (l *Limiter) Release(userID string) {
	l.mu.Lock()
	if l.activeConcurrency > 0 {
		l.activeConcurrency--
	}
	l.mu.Unlock()
}

// GetStats 获取当前并发统计
func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	globalCurrent := l.activeConcurrency

	// 检查 peak 日期
	today := time.Now().Format("2006-01-02")
	peakToday := l.peakToday
	if l.peakDate != today {
		peakToday = globalCurrent
	}

	return map[string]interface{}{
		"global_current": globalCurrent,
		"peak_today":     peakToday,
	}
}

// GetAndResetIntervalPeak 获取当前采样窗口内的最高并发数并重置
// 用于 5 分钟级图表的并发数据采集，比瞬时采样更准确
func (l *Limiter) GetAndResetIntervalPeak() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	peak := l.peakInterval
	current := l.activeConcurrency
	if current > peak {
		peak = current
	}
	l.peakInterval = 0
	return peak
}
