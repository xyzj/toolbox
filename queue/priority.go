package queue

import (
	"container/heap"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Priority byte

const (
	PriorityLowest  Priority = 1
	PriorityLow     Priority = 3
	PriorityNormal  Priority = 5
	PriorityHigh    Priority = 7
	PriorityHighest Priority = 9
)

// --- 1. 消息结构定义 ---

// messageItem 是队列中存储的元素
type messageItem[T any] struct {
	Priority  Priority  // 优先级：数字越大，优先级越高 (例如，9 > 1)
	CreatedAt time.Time // 消息插入时间，用于同优先级下的 FIFO 排序
	Payload   T         // 实际消息内容
}

// --- 2. 核心：优先级堆实现 (实现 heap.Interface 接口) ---

// priorityHeap 是一个基于 Message 指针的最小堆
// 注意：Go 的 heap 是最小堆，所以我们需要反转优先级，让高优先级 (大数字) 靠前 (小值)
type priorityHeap[T any] []*messageItem[T]

// 堆接口要求的 Len()
func (h priorityHeap[T]) Len() int { return len(h) }

// 堆接口要求的 Less()：定义排序规则（优先级越高，时间越早，越排在前面）
func (h priorityHeap[T]) Less(i, j int) bool {
	// 规则 1: 优先级高的排在前面 (Priority 大的在前面)
	if h[i].Priority != h[j].Priority {
		return h[i].Priority > h[j].Priority // 核心：反转 Less()，实现最大堆行为
	}
	// 规则 2: 优先级相同时，创建时间早的排在前面 (FIFO)
	return h[i].CreatedAt.Before(h[j].CreatedAt)
}

// 堆接口要求的 Swap()
func (h priorityHeap[T]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// 堆接口要求的 Push()
func (h *priorityHeap[T]) Push(x any) {
	item := x.(*messageItem[T])
	*h = append(*h, item)
}

// 堆接口要求的 Pop()
func (h *priorityHeap[T]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // 避免内存泄漏
	*h = old[0 : n-1]
	return item
}

// --- 3. 阻塞队列包装器 ---

// PriorityQueue 包装器，包含堆、锁、条件变量和最大长度
type PriorityQueue[T any] struct {
	heap      *priorityHeap[T]
	mutex     sync.Mutex
	cond      *sync.Cond // 用于阻塞/唤醒 Get 操作
	maxLength int
	closed    atomic.Bool // 追踪队列是否已关闭
	zero      T
}

// NewPriorityQueue 构造函数
func NewPriorityQueue[T any](maxLength int) *PriorityQueue[T] {
	h := &priorityHeap[T]{}
	pq := &PriorityQueue[T]{
		heap:      h,
		maxLength: maxLength,
	}
	pq.cond = sync.NewCond(&pq.mutex) // 条件变量必须基于 Mutex
	heap.Init(h)
	return pq
}

// Put 将消息放入队列。如果队列已满，则返回错误。
func (pq *PriorityQueue[T]) Put(priority Priority, payload T) error {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	if pq.closed.Load() { // 检查是否已关闭
		return ErrClosed
	}
	if pq.heap.Len() >= pq.maxLength {
		return ErrFull
	}
	msg := &messageItem[T]{
		Priority:  priority,
		CreatedAt: time.Now(),
		Payload:   payload,
	}
	heap.Push(pq.heap, msg)

	// 唤醒一个可能阻塞在 Get 上的 Goroutine
	pq.cond.Signal()
	return nil
}

// Get 从队列中取出优先级最高的消息。如果队列为空，则阻塞。
func (pq *PriorityQueue[T]) Get() (T, error) {
	return pq.GetWithContext(context.TODO())
}

// - *Message：如果成功取出消息。
// - nil：如果 Context 被取消（超时或外部 Cancel）。
func (pq *PriorityQueue[T]) GetWithContext(ctx context.Context) (T, error) {
	// var zero T

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if pq.closed.Load() {
		return pq.zero, ErrClosed
	}

	// ctx 取消时唤醒 cond.Wait
	cancelWake := context.AfterFunc(ctx, func() {
		pq.mutex.Lock()
		pq.cond.Broadcast()
		pq.mutex.Unlock()
	})
	defer cancelWake()

	for {
		// 有元素直接取
		if pq.heap.Len() > 0 {
			return heap.Pop(pq.heap).(*messageItem[T]).Payload, nil
		}
		// 队列已关闭
		if pq.closed.Load() {
			return pq.zero, ErrClosed
		}
		// 上下文已取消/超时
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return pq.zero, ErrTimeout
			}
			return pq.zero, err
		}
		// 等待被 Put 或 Cancel 唤醒
		pq.cond.Wait()
	}
}

// Length 返回当前队列中的元素数量
func (pq *PriorityQueue[T]) Len() int {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	return pq.heap.Len()
}

// Close 清理队列内容，并唤醒所有阻塞的 GetContext 请求。
func (pq *PriorityQueue[T]) Close() {
	// 确保只关闭一次
	if pq.closed.Load() {
		return
	}
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.closed.Store(true)
	// 清理队列内容（可选，但通常在关闭时执行）
	pq.heap = &priorityHeap[T]{}
	heap.Init(pq.heap)
	// 唤醒所有等待在 cond.Wait() 上的 Goroutines，它们将在 GetContext 中检查 closed 状态后返回 nil
	pq.cond.Broadcast()
}

// IsClosed reports whether the priority queue has been closed.
// It returns true if the internal closed flag is set, false otherwise.
// The check is performed atomically and is safe for concurrent use.
func (pq *PriorityQueue[T]) IsClosed() bool {
	return pq.closed.Load()
}

// Open marks the priority queue as open (not closed).
// If the queue is already open, Open returns immediately and does nothing.
// The implementation performs a fast, lock-free check of the closed flag and
// only acquires the mutex when the queue appears closed, then sets the flag to false.
// This method is safe for concurrent use and ensures the queue's closed state is cleared.
func (pq *PriorityQueue[T]) Open() {
	if !pq.closed.Load() {
		return
	}
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	pq.closed.Store(false)
}

// Reset reinitializes the priority queue to an empty state.
// It returns ErrClosed if the queue is not marked as closed.
// Reset acquires the queue's mutex, replaces the internal heap with a fresh empty heap,
// and calls heap.Init on the new heap. On success it returns nil.
func (pq *PriorityQueue[T]) Reset() error {
	if !pq.closed.Load() {
		return ErrClosed
	}
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	pq.heap = &priorityHeap[T]{}
	heap.Init(pq.heap)
	return nil
}
