package types

import (
	"container/heap"
)

// MinHeap 是一个基于泛型的最小堆
type MinHeap[T any] struct {
	items []T
	less  func(T, T) bool
}

func (h MinHeap[T]) Len() int { return len(h.items) }
func (h MinHeap[T]) Less(i, j int) bool {
	return h.less(h.items[i], h.items[j])
}
func (h MinHeap[T]) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *MinHeap[T]) Push(x interface{}){
	h.items = append(h.items, x.(T))
}

func (h *MinHeap[T]) Pop() interface{} {
	old := h.items
	n := len(old)
	x := old[len(old)-1]
	h.items = old[:n-1]
	return x
}

// NewMinHeap 创建一个新的最小堆
func NewMinHeap[T any](items []T, less func(T, T) bool) *MinHeap[T] {
	h := &MinHeap[T]{less: less, items: items}
	heap.Init(h)
	return h
}

func (h *MinHeap[T]) GetAll() []T {
	return h.items
}

func (h *MinHeap[T]) Peek() interface{} {
	if len(h.items) == 0 {
		var zero T
		return zero
	}
	return h.items[0]
}

