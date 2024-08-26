package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type VisitCounter struct {
	count int
	mu    sync.Mutex
}

func (vc *VisitCounter) Increment() {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.count++
}

func (vc *VisitCounter) GetCount() int {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	return vc.count
}

func monitorVisits(ctx context.Context, counter *VisitCounter, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := counter.GetCount()
			fmt.Printf("Sites visited: %d\n", count)
		case <-ctx.Done():
			return
		}
	}
}