package notifications

import (
	"context"
	"log"
	"time"

	"finance-backend/internal/alerts"
	"github.com/robfig/cron/v3"
)

type Worker struct {
	service      *Service
	alertsService *alerts.Service
	repo         Repository
	spec         string
}

func NewWorker(service *Service, alertsService *alerts.Service, repo Repository, spec string) *Worker {
	if spec == "" {
		spec = "@every 1m"
	}

	return &Worker{
		service:      service,
		alertsService: alertsService,
		repo:          repo,
		spec:          spec,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	scheduler := cron.New(cron.WithLocation(time.Local))

	if _, err := scheduler.AddFunc(w.spec, func() {
		w.runOnce(ctx)
	}); err != nil {
		return err
	}

	scheduler.Start()
	defer scheduler.Stop()

	log.Printf("notifications worker started with schedule %s", w.spec)

	w.runOnce(ctx)

	<-ctx.Done()
	log.Print("notifications worker stopped")
	return ctx.Err()
}

func (w *Worker) runOnce(ctx context.Context) {
	userIDs, err := w.repo.ListUserIDs(ctx)
	if err != nil {
		log.Printf("notifications worker: list users failed: %v", err)
		return
	}

	for _, userID := range userIDs {
		items, err := w.service.Generate(ctx, userID)
		if err != nil {
			log.Printf("notifications worker: generate failed for user %d: %v", userID, err)
			continue
		}
		if len(items) > 0 {
			log.Printf("notifications worker: generated %d notification(s) for user %d", len(items), userID)
		}

		if w.alertsService != nil {
			alertsItems, err := w.alertsService.Evaluate(ctx, userID, alerts.EvaluateInput{})
			if err != nil {
				log.Printf("alerts worker: evaluate failed for user %d: %v", userID, err)
				continue
			}
			if len(alertsItems) > 0 {
				log.Printf("alerts worker: generated %d alert(s) for user %d", len(alertsItems), userID)
			}
		}
	}
}
