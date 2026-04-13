package stats

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Collector 统计收集器
type Collector struct {
	mu       sync.Mutex
	Sent     int
	Received int
	Lost     int
	RTTs     []time.Duration
	MinRTT   time.Duration
	MaxRTT   time.Duration
	SumRTT   time.Duration
}

// NewCollector 创建新的统计收集器
func NewCollector() *Collector {
	return &Collector{
		MinRTT: time.Duration(math.MaxInt64),
	}
}

// RecordSend 记录一次发送
func (c *Collector) RecordSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Sent++
}

// RecordReceive 记录一次成功接收及 RTT
func (c *Collector) RecordReceive(rtt time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Received++
	c.RTTs = append(c.RTTs, rtt)
	c.SumRTT += rtt
	if rtt < c.MinRTT {
		c.MinRTT = rtt
	}
	if rtt > c.MaxRTT {
		c.MaxRTT = rtt
	}
}

// RecordLoss 记录一次丢包
func (c *Collector) RecordLoss() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Lost++
}

// LossRate 计算丢包率
func (c *Collector) LossRate() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Sent == 0 {
		return 0
	}
	return float64(c.Sent-c.Received) / float64(c.Sent) * 100
}

// AvgRTT 计算平均 RTT
func (c *Collector) AvgRTT() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.RTTs) == 0 {
		return 0
	}
	return c.SumRTT / time.Duration(len(c.RTTs))
}

// Report 生成统计报告字符串
func (c *Collector) Report(proto, target string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := fmt.Sprintf("\n--- %s %s statistics ---\n", proto, target)
	result += fmt.Sprintf("%d sent, %d received, %.1f%% loss\n",
		c.Sent, c.Received, c.lossRate())

	if len(c.RTTs) > 0 {
		avg := c.SumRTT / time.Duration(len(c.RTTs))
		result += fmt.Sprintf("rtt min/avg/max = %s/%s/%s\n",
			formatDuration(c.MinRTT), formatDuration(avg), formatDuration(c.MaxRTT))
	}
	return result
}

// lossRate 内部计算丢包率（调用前须持有锁）
func (c *Collector) lossRate() float64 {
	if c.Sent == 0 {
		return 0
	}
	return float64(c.Sent-c.Received) / float64(c.Sent) * 100
}

// formatDuration 格式化时间为可读字符串
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fus", float64(d.Nanoseconds())/1000)
	}
	return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
}
