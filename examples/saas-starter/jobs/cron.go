package jobs

import (
	"context"
	"log"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// +zyracron 0 0 * * *
func BillingRenewalJob(ctx context.Context) error {
	log.Println("[CRON] Running daily billing renewal processing job...")
	return nil
}

// +zyracron */15 * * * *
func TelemetrySyncJob(ctx context.Context) error {
	log.Println("[CRON] Running 15-minute telemetry sync job...")
	return nil
}

func RegisterJobs() {
	zyra.Jobs.Register("billing.renew", func(ctx context.Context, payload []byte) error {
		log.Printf("[JOB] Processing billing renewal job payload: %s", string(payload))
		return nil
	})
}
