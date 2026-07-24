package dx

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// JobOptions options for enqueuing a background task.
type JobOptions struct {
	Delay       time.Duration `json:"delay"`
	MaxAttempts int           `json:"maxAttempts"`
}

// Job represents a background task unit.
type Job struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Payload     []byte    `json:"payload"`
	Status      string    `json:"status"` // "pending", "running", "completed", "failed"
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"maxAttempts"`
	RunAt       time.Time `json:"runAt"`
	LockedAt    time.Time `json:"lockedAt"`
	Error       string    `json:"error"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CronEntry represents a scheduled cron function.
type CronEntry struct {
	Spec     string
	Handler  func(ctx context.Context) error
	NextRun  time.Time
}

// JobHandler function signature for processing background job payloads.
type JobHandler func(ctx context.Context, payload []byte) error

// JobManager is a zero-Redis background job runner and cron scheduler.
type JobManager struct {
	mu           sync.RWMutex
	handlers     map[string]JobHandler
	cronEntries  []*CronEntry
	jobs         map[string]*Job
	counter      uint64
	running      bool
	workerCancel context.CancelFunc
	workerWg     sync.WaitGroup
}

func NewJobManager() *JobManager {
	return &JobManager{
		handlers:    make(map[string]JobHandler),
		cronEntries: make([]*CronEntry, 0),
		jobs:        make(map[string]*Job),
	}
}

func (m *JobManager) Register(jobType string, handler JobHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[jobType] = handler
}

func (m *JobManager) RegisterCron(spec string, handler func(ctx context.Context) error) error {
	nextRun, err := parseNextCronTime(spec, time.Now())
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", spec, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.cronEntries = append(m.cronEntries, &CronEntry{
		Spec:    spec,
		Handler: handler,
		NextRun: nextRun,
	})
	return nil
}

func (m *JobManager) Enqueue(ctx context.Context, jobType string, payload any, opts JobOptions) (*Job, error) {
	var payloadBytes []byte
	var err error

	if b, ok := payload.([]byte); ok {
		payloadBytes = b
	} else if s, ok := payload.(string); ok {
		payloadBytes = []byte(s)
	} else {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal job payload: %w", err)
		}
	}

	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	runAt := time.Now()
	if opts.Delay > 0 {
		runAt = runAt.Add(opts.Delay)
	}

	m.mu.Lock()
	m.counter++
	id := fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), m.counter)

	job := &Job{
		ID:          id,
		Type:        jobType,
		Payload:     payloadBytes,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: maxAttempts,
		RunAt:       runAt,
		CreatedAt:   time.Now(),
	}
	m.jobs[id] = job
	m.mu.Unlock()

	return job, nil
}

func (m *JobManager) StartWorkerPool(ctx context.Context, concurrency int) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	if concurrency <= 0 {
		concurrency = 4
	}

	workerCtx, cancel := context.WithCancel(ctx)
	m.workerCancel = cancel
	m.running = true
	m.mu.Unlock()

	for i := 0; i < concurrency; i++ {
		m.workerWg.Add(1)
		go m.workerLoop(workerCtx)
	}

	// Start cron loop
	m.workerWg.Add(1)
	go m.cronLoop(workerCtx)
}

func (m *JobManager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	if m.workerCancel != nil {
		m.workerCancel()
	}
	m.running = false
	m.mu.Unlock()

	m.workerWg.Wait()
}

func (m *JobManager) workerLoop(ctx context.Context) {
	defer m.workerWg.Done()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job, handler := m.pickNextJob()
			if job == nil || handler == nil {
				continue
			}

			m.processJob(ctx, job, handler)
		}
	}
}

func (m *JobManager) cronLoop(ctx context.Context) {
	defer m.workerWg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			m.mu.Lock()
			for _, entry := range m.cronEntries {
				if now.After(entry.NextRun) || now.Equal(entry.NextRun) {
					handler := entry.Handler
					nextRun, _ := parseNextCronTime(entry.Spec, now)
					entry.NextRun = nextRun

					// Run cron task in separate goroutine
					go func(h func(context.Context) error) {
						_ = h(ctx)
					}(handler)
				}
			}
			m.mu.Unlock()
		}
	}
}

func (m *JobManager) pickNextJob() (*Job, JobHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, job := range m.jobs {
		if job.Status == "pending" && (now.After(job.RunAt) || now.Equal(job.RunAt)) {
			job.Status = "running"
			job.LockedAt = now
			job.Attempts++

			handler, exists := m.handlers[job.Type]
			if !exists {
				job.Status = "failed"
				job.Error = fmt.Sprintf("no handler registered for job type %s", job.Type)
				continue
			}
			return job, handler
		}
	}
	return nil, nil
}

func (m *JobManager) processJob(ctx context.Context, job *Job, handler JobHandler) {
	err := handler(ctx, job.Payload)

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		job.Error = err.Error()
		if job.Attempts >= job.MaxAttempts {
			job.Status = "failed"
		} else {
			// Retry after exponential backoff
			job.Status = "pending"
			job.RunAt = time.Now().Add(time.Duration(job.Attempts) * 2 * time.Second)
		}
	} else {
		job.Status = "completed"
	}
}

func (m *JobManager) GetJob(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	return j, ok
}

// Minimal 5-field cron parser for standard expressions e.g. "0 * * * *" or "* * * * *"
func parseNextCronTime(spec string, from time.Time) (time.Time, error) {
	parts := strings.Fields(spec)
	if len(parts) != 5 {
		return from, fmt.Errorf("cron spec must have 5 fields")
	}

	minuteSpec := parts[0]
	next := from.Truncate(time.Minute).Add(time.Minute)

	if minuteSpec == "*" {
		return next, nil
	}

	// Check if step e.g. */5
	if strings.HasPrefix(minuteSpec, "*/") {
		step, err := strconv.Atoi(strings.TrimPrefix(minuteSpec, "*/"))
		if err == nil && step > 0 {
			for next.Minute()%step != 0 {
				next = next.Add(time.Minute)
			}
			return next, nil
		}
	}

	// Direct number check
	mVal, err := strconv.Atoi(minuteSpec)
	if err == nil && mVal >= 0 && mVal < 60 {
		for next.Minute() != mVal {
			next = next.Add(time.Minute)
		}
		return next, nil
	}

	return next, nil
}
