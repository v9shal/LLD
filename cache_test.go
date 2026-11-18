package main

import (
	"sync"
	"testing"
)

func TestLRUEviction(t *testing.T) {
	strCache := NewCache[string, int](2, NewLRUPolicy[string]())
	strCache.Put("A", 1)
	strCache.Put("B", 2)
	strCache.Get("A")
	strCache.Put("C", 3)

	_, bool := strCache.Get("B")
	if bool == true {
		t.Errorf("B shoudl have been evicted")
	}
	if val, bool := strCache.Get("A"); val != 1 || bool == false {
		t.Errorf("A shoudl be present and equal to 1")
	}
	if val, bool := strCache.Get("C"); val != 3 || bool == false {
		t.Errorf("c shoudl be present and equal to 3")

	}
}

func TestConcurrency(t *testing.T) {
	intCache := NewCache[int, int](10, NewLRUPolicy[int]())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for j := 0; j < 15; j++ {
				intCache.Put(j%15, workerId)
				intCache.Get(j % 15)

			}
		}(i)
	}
	wg.Wait()
}
