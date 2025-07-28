package tin

import (
	"fmt"
	"sync"
	"time"
)

// Progress 接口定义进度跟踪行为
type Progress interface {
	Init(total int)
	Update(current, total int)               // 更新进度
	Complete()                               // 标记任务完成
	Log(format string, args ...interface{})  // 普通日志
	Warn(format string, args ...interface{}) // 警告日志
}

// DefaultProgress 默认进度跟踪器实现
type DefaultProgress struct {
	startTime time.Time
	total     int
	current   int
	lastLog   time.Time
	mu        sync.Mutex // 新增互斥锁保证线程安全
}

// NewDefaultProgress 创建默认进度跟踪器
func NewDefaultProgress() *DefaultProgress {
	return &DefaultProgress{
		startTime: time.Now(),
		lastLog:   time.Now(),
	}
}

func (p *DefaultProgress) Init(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.total = total
}

func (p *DefaultProgress) Update(current, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.total = total

	// 每5秒或每10%进度记录一次
	now := time.Now()
	if now.Sub(p.lastLog) > 5*time.Second ||
		(total > 0 && current%(total/10+1) == 0) { // +1 避免除零

		p.Log("Progress: %d/%d (%.1f%%)",
			current, total, float64(current)/float64(total)*100)
		p.lastLog = now
	}
}

func (p *DefaultProgress) Complete() {
	p.Log("Process completed, total duration: %v", time.Since(p.startTime))
}

func (p *DefaultProgress) Log(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), message)
}

func (p *DefaultProgress) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] WARN: %s\n", time.Now().Format("15:04:05"), message)
}
