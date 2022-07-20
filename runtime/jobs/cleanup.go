package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func CleanJobs(ctx context.Context, db bolted.Database, interval, keepDuration time.Duration, logger *zap.SugaredLogger) {
	defer func() {
		logger.Debug("clean jobs terminated")
	}()

	logger.Debug("clean job started")
	ticker := time.NewTicker(interval)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// ticker with interval
		}

		cutoffTime := time.Now().Add(-keepDuration)

		logger.With("cutoffTime", cutoffTime).Debug("cleaning job")

		counters := &struct {
			succeeded int
			failed    int
		}{}

		err := bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {

			toClean := []struct {
				path    dbpath.Path
				counter *int
			}{
				{
					path:    defaultQueueFailed,
					counter: &counters.failed,
				},
				{
					path:    defaultQueueSucceeded,
					counter: &counters.succeeded,
				},
			}
			for _, tc := range toClean {
				for it := tx.Iterator(tc.path); !it.IsDone(); it.Next() {
					idString := it.GetKey()
					id, err := uuid.FromString(idString)
					if err != nil {
						return fmt.Errorf("while parsing uuid %s: %w", idString, err)
					}

					if id.Version() != uuid.V6 {
						return fmt.Errorf("expected uuid %s to have version 6 but got %d", idString, id.Version())
					}

					ts, err := uuid.TimestampFromV6(id)
					if err != nil {
						return fmt.Errorf("while getting timestamp from %s: %w", idString, err)
					}

					t, err := ts.Time()
					if err != nil {
						return fmt.Errorf("while getting time from timestamp of %s: %w", idString, err)
					}

					if t.After(cutoffTime) {
						break
					}

					tx.Delete(tc.path.Append(idString))
					*(tc.counter)++
				}
			}

			if counters.succeeded != 0 || counters.failed != 0 {
				logger.With("succeeded", counters.succeeded, "failed", counters.failed).Debug("cleaned job records")
			}

			return nil
		})

		if err != nil {
			logger.With("error", err).Error("while cleaning up failed jobs")
		}
	}
}
