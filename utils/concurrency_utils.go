// utils/concurrency.go
package utils

// WorkerPool 并发控制器（channel 实现）
type WorkerPool struct {
	slots chan struct{} // 并发槽
}

// NewWorkerPool 创建并发控制器，maxConcurrent 为最大并发数
func NewWorkerPool(maxConcurrent int) *WorkerPool {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	return &WorkerPool{
		slots: make(chan struct{}, maxConcurrent),
	}
}

// Acquire 获取一个并发槽（阻塞等待，直到有空闲）
func (p *WorkerPool) Acquire() {
	p.slots <- struct{}{}
}

// Release 释放一个并发槽
func (p *WorkerPool) Release() {
	<-p.slots
}

// TryAcquire 非阻塞尝试获取（可选）
func (p *WorkerPool) TryAcquire() bool {
	select {
	case p.slots <- struct{}{}:
		return true
	default:
		return false
	}
}
