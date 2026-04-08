package common

import (
	"sync"
	"time"
)

// FrequencyCounter パケット受信頻度（Hz）を計算するスレッドセーフなカウンタ
type FrequencyCounter struct {
	mu          sync.Mutex
	packetCount int
	hz          float64
	lastCalc    time.Time
}

func NewFrequencyCounter() FrequencyCounter {
	return FrequencyCounter{lastCalc: time.Now()}
}

// Record パケット受信を記録する
func (f *FrequencyCounter) Record() {
	f.mu.Lock()
	f.packetCount++
	f.mu.Unlock()
}

// Hz 直近1秒以上の期間に基づく受信頻度を返す
func (f *FrequencyCounter) Hz() float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(f.lastCalc).Seconds()
	if elapsed >= 1.0 {
		f.hz = float64(f.packetCount) / elapsed
		f.packetCount = 0
		f.lastCalc = now
	}
	return f.hz
}
