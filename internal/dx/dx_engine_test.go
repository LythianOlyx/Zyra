package dx_test

import (
	"context"
	"testing"
	"time"

	"github.com/zyra-framework/zyra/internal/dx"
)

func TestDXEngine(t *testing.T) {
	ctx := context.Background()

	// Mail
	mm := dx.NewMailManager()
	if err := mm.Send(ctx, dx.Email{To: []string{"test@dev.com"}, Subject: "Direct Test"}); err != nil {
		t.Fatalf("Mail send failed: %v", err)
	}

	// Cache
	cm := dx.NewCacheManager()
	_ = cm.Set(ctx, "k1", "v1", 1*time.Minute)
	v, ok, _ := cm.Get(ctx, "k1")
	if !ok || v != "v1" {
		t.Errorf("Cache get failed")
	}

	// Jobs
	jm := dx.NewJobManager()
	job, err := jm.Enqueue(ctx, "t1", "p1", dx.JobOptions{})
	if err != nil || job.ID == "" {
		t.Errorf("Job enqueue failed")
	}

	// Flags
	fm := dx.NewFlagManager()
	fm.Set("f1", true)
	if !fm.IsEnabled(ctx, "f1") {
		t.Errorf("Flag test failed")
	}

	// I18n
	i18n := dx.NewI18nManager()
	i18n.LoadTranslations("en", map[string]string{"k1": "Hello"})
	if i18n.Translate(ctx, "en", "k1") != "Hello" {
		t.Errorf("I18n translate failed")
	}

	// Stream
	sm := dx.NewStreamManager()
	ch, unsub := sm.Subscribe(ctx, "r1")
	sm.Broadcast("r1", "data")
	msg := <-ch
	unsub()
	if msg != "data" {
		t.Errorf("Stream test failed")
	}
}
