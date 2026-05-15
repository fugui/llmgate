// Package concurrency 提供并发请求限制功能
package concurrency

import (
	"sync"
	"time"
)

// Limiter 并发限制器
type Limiter struct {
	userLimit    int                      // 每个用户的并发限制
	userSemMap   map[string]chan struct{} // 用户信号量映射
	mu           sync.RWMutex

	activeConcurrency int    // 当前活跃并发数（始终追踪，不受 userLimit 影响）
	peakToday         int    // 今日最高并发数
	peakDate          string // 峰值对应日期
	peakInterval      int    // 当前采样窗口内的最高并发数
}

// NewLimiter 创建新的并发限制器
// userLimit: 每个用户最大并发数 (0 表示不限制)
func NewLimiter(userLimit int) *Limiter {
	return &Limiter{
		userLimit:  userLimit,
		userSemMap: make(map[string]chan struct{}),
	}
}

// Acquire 尝试获取并发许可
// 返回 true 表示获取成功，false 表示被拒绝（超出用户并发限制）
func (l *Limiter) Acquire(userID string) bool {
	// 尝试获取用户许可
	if l.userLimit > 0 {
		userSem := l.getUserSemaphore(userID)
		select {
		case userSem <- struct{}{}:
			// 获取成功
		default:
			// 用户并发已满
			return false
		}
	}

	// 许可获取成功，更新并发计数和峰值（始终执行）
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
	// 减少活跃并发计数
	l.mu.Lock()
	if l.activeConcurrency > 0 {
		l.activeConcurrency--
	}
	l.mu.Unlock()

	// 释放用户许可
	if l.userLimit > 0 {
		userSem := l.getUserSemaphore(userID)
		select {
		case <-userSem:
		default:
		}
	}
}

// getUserSemaphore 获取或创建用户的信号量
func (l *Limiter) getUserSemaphore(userID string) chan struct{} {
	l.mu.RLock()
	sem, exists := l.userSemMap[userID]
	l.mu.RUnlock()

	if exists {
		return sem
	}

	// 创建新的信号量
	l.mu.Lock()
	defer l.mu.Unlock()

	// 双重检查
	if sem, exists := l.userSemMap[userID]; exists {
		return sem
	}

	sem = make(chan struct{}, l.userLimit)
	l.userSemMap[userID] = sem
	return sem
}

// GetStats 获取当前并发统计
func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	globalCurrent := l.activeConcurrency

	userStats := make(map[string]int)
	for userID, sem := range l.userSemMap {
		count := len(sem)
		if count > 0 {
			userStats[userID] = count
		}
	}

	// 检查 peak 日期
	today := time.Now().Format("2006-01-02")
	peakToday := l.peakToday
	if l.peakDate != today {
		peakToday = globalCurrent
	}

	return map[string]interface{}{
		"user_limit":     l.userLimit,
		"global_current": globalCurrent,
		"active_users":   len(userStats),
		"user_stats":     userStats,
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

// UpdateLimits 动态更新并发限制
// 注意：会重建信号量，已在飞行中的请求仍会正常释放旧信号量
func (l *Limiter) UpdateLimits(userLimit int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.userLimit = userLimit

	// 清空用户信号量映射（会按需重新创建）
	l.userSemMap = make(map[string]chan struct{})
}
