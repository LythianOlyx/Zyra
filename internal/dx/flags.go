package dx

import (
	"context"
	"hash/fnv"
	"sync"
)

type FlagRule struct {
	Enabled      bool
	Percentage   int // 0 to 100
	UserOverrides map[string]bool
}

// FlagManager handles feature flags with user targeting and rollouts.
type FlagManager struct {
	mu    sync.RWMutex
	flags map[string]*FlagRule
}

func NewFlagManager() *FlagManager {
	return &FlagManager{
		flags: make(map[string]*FlagRule),
	}
}

func (f *FlagManager) Set(flag string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.flags[flag]
	if !exists {
		rule = &FlagRule{UserOverrides: make(map[string]bool)}
		f.flags[flag] = rule
	}
	rule.Enabled = enabled
}

func (f *FlagManager) SetRollout(flag string, percentage int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	rule, exists := f.flags[flag]
	if !exists {
		rule = &FlagRule{UserOverrides: make(map[string]bool)}
		f.flags[flag] = rule
	}
	rule.Percentage = percentage
}

func (f *FlagManager) SetUserOverride(flag, userID string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.flags[flag]
	if !exists {
		rule = &FlagRule{UserOverrides: make(map[string]bool)}
		f.flags[flag] = rule
	}
	rule.UserOverrides[userID] = enabled
}

func (f *FlagManager) IsEnabled(ctx context.Context, flag string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rule, exists := f.flags[flag]
	if !exists {
		return false
	}
	return rule.Enabled || rule.Percentage == 100
}

func (f *FlagManager) IsEnabledForUser(ctx context.Context, flag string, userID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rule, exists := f.flags[flag]
	if !exists {
		return false
	}

	if override, ok := rule.UserOverrides[userID]; ok {
		return override
	}

	if rule.Enabled {
		return true
	}

	if rule.Percentage > 0 {
		h := fnv.New32a()
		h.Write([]byte(flag + ":" + userID))
		score := int(h.Sum32() % 100)
		return score < rule.Percentage
	}

	return false
}
