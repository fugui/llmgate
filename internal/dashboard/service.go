package dashboard

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HourlyCounter 内存小时级计数器
type HourlyCounter struct {
	date   string                      // 当前日期，用于检测日期变化
	counts map[string]map[int]int      // userID -> hour(0-23) -> count
	mu     sync.RWMutex
}

func NewHourlyCounter() *HourlyCounter {
	return &HourlyCounter{
		date:   time.Now().Format("2006-01-02"),
		counts: make(map[string]map[int]int),
	}
}

// Increment 增加指定用户当前小时的计数
func (hc *HourlyCounter) Increment(userID string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// 检查日期是否变化
	today := time.Now().Format("2006-01-02")
	if hc.date != today {
		// 跨天了，重置所有数据
		hc.date = today
		hc.counts = make(map[string]map[int]int)
	}

	hour := time.Now().Hour()

	if hc.counts[userID] == nil {
		hc.counts[userID] = make(map[int]int)
	}
	hc.counts[userID][hour]++
}

// GetLast24Hours 获取最近24小时的总请求数（按小时汇总）
func (hc *HourlyCounter) GetLast24Hours() []HourlyStat {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	now := time.Now()
	currentHour := now.Hour()

	// 初始化24小时的数据（从24小时前到现在）
	stats := make([]HourlyStat, 24)
	for i := 0; i < 24; i++ {
		hour := (currentHour - 23 + i + 24) % 24
		timeStr := fmt.Sprintf("%02d:00", hour)
		stats[i] = HourlyStat{
			Hour:     timeStr,
			Requests: 0,
		}
	}

	// 如果跨天了，只返回今天的数据
	today := now.Format("2006-01-02")
	if hc.date != today {
		return stats
	}

	// 按小时汇总所有用户的请求
	for _, userCounts := range hc.counts {
		for hour, count := range userCounts {
			// 找到对应的小时位置
			for i := 0; i < 24; i++ {
				statHour := (currentHour - 23 + i + 24) % 24
				if statHour == hour {
					stats[i].Requests += count
					break
				}
			}
		}
	}

	return stats
}

// Service 仪表板服务
type Service struct {
	db            *sql.DB
	hourlyCounter *HourlyCounter
}

// NewService 创建仪表板服务
func NewService(db *sql.DB) *Service {
	s := &Service{
		db:            db,
		hourlyCounter: NewHourlyCounter(),
	}
	// 启动日期检查任务
	go s.dateCheckLoop()
	return s
}

// dateCheckLoop 每小时检查一次日期变化
func (s *Service) dateCheckLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.hourlyCounter.mu.Lock()
		today := time.Now().Format("2006-01-02")
		if s.hourlyCounter.date != today {
			s.hourlyCounter.date = today
			s.hourlyCounter.counts = make(map[string]map[int]int)
		}
		s.hourlyCounter.mu.Unlock()
	}
}

// RecordHourlyStat 记录小时级统计
func (s *Service) RecordHourlyStat(userID string) {
	s.hourlyCounter.Increment(userID)
}

// DashboardStats 系统概览
type DashboardStats struct {
	TodayTotalRequests int64   `json:"today_total_requests"`  // 今日总请求
	TodayInputTokens   int64   `json:"today_input_tokens"`    // 今日输入Token
	TodayOutputTokens  int64   `json:"today_output_tokens"`   // 今日输出Token
	ActiveUsers        int     `json:"active_users"`          // 今日活跃用户
	TotalUsers         int     `json:"total_users"`           // 总用户数
	DepartmentCount    int     `json:"department_count"`      // 部门数量
	AvgRequestsPerUser float64 `json:"avg_requests_per_user"` // 人均请求数
}

// TopUser TOP用户
type TopUser struct {
	UserID        string `json:"user_id"`
	Name          string `json:"name"`
	Department    string `json:"department"`
	RequestCount  int    `json:"request_count"`
	InputTokens   int64  `json:"input_tokens"`
	OutputTokens  int64  `json:"output_tokens"`
}

// HourlyStat 小时统计
type HourlyStat struct {
	Hour     string `json:"hour"`      // 格式: "14:00"
	Requests int    `json:"requests"`
}

// DepartmentStat 部门统计
type DepartmentStat struct {
	Department   string `json:"department"`
	UserCount    int    `json:"user_count"`    // 该部门用户数
	RequestCount int    `json:"request_count"` // 该部门总请求
}

// ModelStat 模型统计
type ModelStat struct {
	ModelID      string `json:"model_id"`
	RequestCount int    `json:"request_count"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
}

// GetDashboardStats 获取系统概览数据
func (s *Service) GetDashboardStats() (*DashboardStats, error) {
	today := time.Now().Format("2006-01-02")

	stats := &DashboardStats{}

	// 今日总请求数 + Token 合计
	query := `SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0)
	          FROM quota_usage_daily WHERE date = ?`
	err := s.db.QueryRow(query, today).Scan(&stats.TodayTotalRequests, &stats.TodayInputTokens, &stats.TodayOutputTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get today total requests: %w", err)
	}

	// 今日活跃用户数（有请求记录的用户）
	query = `SELECT COUNT(DISTINCT user_id) FROM quota_usage_daily WHERE date = ?`
	err = s.db.QueryRow(query, today).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// 总用户数
	query = `SELECT COUNT(*) FROM users`
	err = s.db.QueryRow(query).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	// 部门数量（排除空部门）
	query = `SELECT COUNT(DISTINCT department) FROM users WHERE department != '' AND department IS NOT NULL`
	err = s.db.QueryRow(query).Scan(&stats.DepartmentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get department count: %w", err)
	}

	// 计算人均请求数
	if stats.TotalUsers > 0 {
		stats.AvgRequestsPerUser = float64(stats.TodayTotalRequests) / float64(stats.TotalUsers)
	}

	return stats, nil
}

// GetTopUsers 获取今日TOP10用户
func (s *Service) GetTopUsers(limit int) ([]TopUser, error) {
	if limit <= 0 {
		limit = 10
	}

	today := time.Now().Format("2006-01-02")

	query := `
		SELECT u.id, u.name, u.department,
		       COALESCE(SUM(q.request_count), 0) as request_count,
		       COALESCE(SUM(q.input_tokens), 0) as input_tokens,
		       COALESCE(SUM(q.output_tokens), 0) as output_tokens
		FROM users u
		LEFT JOIN quota_usage_daily q ON u.id = q.user_id AND q.date = ?
		GROUP BY u.id, u.name, u.department
		ORDER BY request_count DESC
		LIMIT ?`

	rows, err := s.db.Query(query, today, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top users: %w", err)
	}
	defer rows.Close()

	var users []TopUser
	for rows.Next() {
		var user TopUser
		var userID uuid.UUID
		err := rows.Scan(&userID, &user.Name, &user.Department, &user.RequestCount, &user.InputTokens, &user.OutputTokens)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top user: %w", err)
		}
		user.UserID = userID.String()
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetHourlyStats 获取最近24小时每小时请求数
func (s *Service) GetHourlyStats() []HourlyStat {
	return s.hourlyCounter.GetLast24Hours()
}

// GetDepartmentStats 获取部门使用统计
func (s *Service) GetDepartmentStats() ([]DepartmentStat, error) {
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT
			COALESCE(u.department, '未设置') as department,
			COUNT(DISTINCT u.id) as user_count,
			COALESCE(SUM(q.request_count), 0) as request_count
		FROM users u
		LEFT JOIN quota_usage_daily q ON u.id = q.user_id AND q.date = ?
		GROUP BY u.department
		ORDER BY request_count DESC`

	rows, err := s.db.Query(query, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query department stats: %w", err)
	}
	defer rows.Close()

	var stats []DepartmentStat
	for rows.Next() {
		var stat DepartmentStat
		err := rows.Scan(&stat.Department, &stat.UserCount, &stat.RequestCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// GetModelStats 获取模型使用分布
func (s *Service) GetModelStats() ([]ModelStat, error) {
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT model_id,
		       COALESCE(SUM(request_count), 0) as request_count,
		       COALESCE(SUM(input_tokens), 0) as input_tokens,
		       COALESCE(SUM(output_tokens), 0) as output_tokens
		FROM quota_usage_daily
		WHERE date = ?
		GROUP BY model_id
		ORDER BY request_count DESC`

	rows, err := s.db.Query(query, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query model stats: %w", err)
	}
	defer rows.Close()

	var stats []ModelStat
	for rows.Next() {
		var stat ModelStat
		err := rows.Scan(&stat.ModelID, &stat.RequestCount, &stat.InputTokens, &stat.OutputTokens)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}
