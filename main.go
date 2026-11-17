package main

import (
	"fmt"
	"math"
	"sync"
)

type EvictionPolicy[K comparable] interface {
	KeyAccessed(key K)
	Evict() K
}

type FifoPolicy[K comparable] struct {
	queue []K
	seen  map[K]bool
}

func NewFifoPolicy[K comparable]() *FifoPolicy[K] {
	return &FifoPolicy[K]{
		queue: make([]K, 0),
		seen:  make(map[K]bool),
	}
}

func (f *FifoPolicy[K]) KeyAccessed(key K) {
	if exist := f.seen[key]; !exist {
		f.queue = append(f.queue, key)
		f.seen[key] = true
	}
}

func (f *FifoPolicy[K]) Evict() K {
	if len(f.queue) == 0 {
		var zero K
		return zero
	}
	item := f.queue[0]
	f.queue = f.queue[1:]
	delete(f.seen, item)
	return item
}

type LFUPolicy[K comparable] struct {
	counts map[K]int
}

func NewLFUPolicy[K comparable]() *LFUPolicy[K] {
	return &LFUPolicy[K]{
		counts: make(map[K]int),
	}
}

func (l *LFUPolicy[K]) KeyAccessed(key K) {
	l.counts[key]++
}

func (l *LFUPolicy[K]) Evict() K {
	var minKey K
	minCount := math.MaxInt32

	for key, count := range l.counts {
		if count < minCount {
			minCount = count
			minKey = key
		}
	}
	delete(l.counts, minKey)
	return minKey
}

type Node[K comparable] struct {
	key  K
	prev *Node[K]
	next *Node[K]
}
type LRUPolicy[K comparable] struct {
	nodes map[K]*Node[K]
	head  *Node[K]
	tail  *Node[K]
}

func NewLRUPolicy[K comparable]() *LRUPolicy[K] {
	head := &Node[K]{}
	tail := &Node[K]{}
	head.next = tail
	tail.prev = head
	return &LRUPolicy[K]{
		nodes: make(map[K]*Node[K]),
		head:  head,
		tail:  tail,
	}
}
func (l *LRUPolicy[K]) removeNode(n *Node[K]) {
	n.prev.next = n.next
	n.next.prev = n.prev
}
func (l *LRUPolicy[K]) insertbeforetail(n *Node[K]) {
	prevNewest := l.tail.prev
	prevNewest.next = n
	n.prev = prevNewest
	n.next = l.tail
	l.tail.prev = n

}
func (l *LRUPolicy[K]) KeyAccessed(key K) {
	if node, exist := l.nodes[key]; exist {
		l.removeNode(node)
		l.insertbeforetail(node)
		return
	}
	newNode := &Node[K]{key: key}
	l.nodes[key] = newNode
	l.insertbeforetail(newNode)
}
func (l *LRUPolicy[K]) Evict() K {
	var key K
	if l.head.next == l.tail {
		return key
	}
	oldest := l.head.next
	l.removeNode(oldest)
	delete(l.nodes, oldest.key)
	return oldest.key

}

type Cache[K comparable, V any] struct {
	data     map[K]V
	capacity int
	policy   EvictionPolicy[K]
	mu       sync.RWMutex
}

func NewCache[K comparable, V any](capacity int, policy EvictionPolicy[K]) *Cache[K, V] {
	return &Cache[K, V]{
		data:     make(map[K]V),
		capacity: capacity,
		policy:   policy,
	}
}

func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; !exists && len(c.data) >= c.capacity {
		victim := c.policy.Evict()
		delete(c.data, victim)
	}
	c.data[key] = value
	c.policy.KeyAccessed(key)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, exists := c.data[key]
	if !exists {
		var zero V
		return zero, false
	}
	c.policy.KeyAccessed(key)
	return data, true
}

func main() {
	// Test 1: Cache with <string, string> and FIFO Policy
	strCache := NewCache[string, string](2, NewFifoPolicy[string]())
	strCache.Put("A", "Apple")
	strCache.Put("B", "Banana")
	strCache.Get("A")
	strCache.Put("C", "Cherry") // Evicts "A" (First-In!)

	val, ok := strCache.Get("A")
	fmt.Println("String Cache [A]:", val, ok) // false!

	// Test 2: Exact same Cache with <int, float64> and LFU Policy!
	priceCache := NewCache[int, float64](2, NewLFUPolicy[int]())
	priceCache.Put(101, 99.99)
	priceCache.Put(102, 49.50)
	priceCache.Get(101)        // 101 accessed twice!
	priceCache.Put(103, 19.99) // Evicts 102 (Least Frequently Used!)

	price, ok := priceCache.Get(101)
	fmt.Println("Int->Float Cache [101]:", price, ok) // 99.99 true!
}
