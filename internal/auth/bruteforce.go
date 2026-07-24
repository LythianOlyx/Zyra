package auth

import (
	"errors"
	"sync"
	"time"
)

var ErrAccountLocked = errors.New("too many failed login attempts; access temporarily locked")

type attemptRecord struct {
	attempts    int
	lastAttempt time.Time
	lockedUntil time.Time
}

// BruteForceProtection tracks failed authentication attempts per IP/key.
type BruteForceProtection struct {
	mu          sync.RWMutex
	records     map[string]*attemptRecord
	maxAttempts int
	lockWindow  time.Duration
}

func NewBruteForceProtection(maxAttempts int, lockWindow time.Duration) *BruteForceProtection {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if lockWindow <= 0 {
		lockWindow = 15 * time.Minute
	}
	bf := &BruteForceProtection{
		records:     make(map[string]*attemptRecord),
		maxAttempts: maxAttempts,
		lockWindow:  lockWindow,
	}
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			bf.cleanupStale()
		}
	}()
	return bf
}

func (bf *BruteForceProtection) cleanupStale() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	now := time.Now()
	for key, rec := range bf.records {
		if now.After(rec.lockedUntil) && now.Sub(rec.lastAttempt) > bf.lockWindow {
			delete(bf.records, key)
		}
	}
}

func (bf *BruteForceProtection) CheckLockout(key string) error {
	bf.mu.RLock()
	rec, exists := bf.records[key]
	bf.mu.RUnlock()

	if !exists {
		return nil
	}

	if time.Now().Before(rec.lockedUntil) {
		return ErrAccountLocked
	}

	return nil
}

func (bf *BruteForceProtection) RecordFailure(key string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	now := time.Now()
	rec, exists := bf.records[key]
	if !exists {
		rec = &attemptRecord{}
		bf.records[key] = rec
	}

	// Reset attempts if previous window expired
	if now.Sub(rec.lastAttempt) > bf.lockWindow {
		rec.attempts = 0
	}

	rec.attempts++
	rec.lastAttempt = now

	if rec.attempts >= bf.maxAttempts {
		rec.lockedUntil = now.Add(bf.lockWindow)
	}
}

func (bf *BruteForceProtection) RecordSuccess(key string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	delete(bf.records, key)
}
