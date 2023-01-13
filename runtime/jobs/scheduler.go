package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/stdlib"
	"github.com/go-logr/logr"
)

// TODO when runtime updated, check if jobs are unfinished, reschedule if so

var JobsDefinitionsPath = dbpath.ToPath("jobs")
var JobQueuePath = dbpath.ToPath("job-queue")

var defaultQueueScheduled = JobQueuePath.Append("default", "scheduled")
var defaultQueueRunning = JobQueuePath.Append("default", "running")
var defaultQueueFailed = JobQueuePath.Append("default", "failed")
var defaultQueueSucceeded = JobQueuePath.Append("default", "succeeded")

func JobScheduler(ctx context.Context, db bolted.Database, maxHistorySize uint64, libs *jslib.Libs, logger logr.Logger) {

	logger.Info("job scheduler started")
	defer logger.Info("job scheduler terminated")

	changes, close := db.Observe(defaultQueueScheduled.ToMatcher().AppendAnySubpathMatcher().AppendAnyElementMatcher())
	defer close()

	for {
		select {
		case _, ok := <-changes:
			if !ok {
				return
			}

			routinesToStart := []func(){}

			err := bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {

				if !tx.Exists(defaultQueueScheduled) {
					return nil
				}

				// TODO reverse iterator when not running all jobs at the same time
				for it := tx.Iterator(defaultQueueScheduled); !it.IsDone(); it.Next() {
					jobPath := defaultQueueScheduled.Append(it.GetKey())
					id := it.GetKey()
					name := string(tx.Get(jobPath.Append("name")))
					params := tx.Get(jobPath.Append("params"))

					tx.Delete(jobPath)
					if !tx.Exists(defaultQueueRunning) {
						tx.CreateMap(defaultQueueRunning)
					}
					jobRunningPath := defaultQueueRunning.Append(id)
					tx.CreateMap(jobRunningPath)
					tx.Put(jobRunningPath.Append("name"), []byte(name))
					tx.Put(jobRunningPath.Append("params"), params)
					routinesToStart = append(routinesToStart, runJob(ctx, db, maxHistorySize, libs, id, name, params, logger))
					return nil
				}
				return nil
			})

			if err != nil {
				logger.Error(err, "while starting jobs")
				continue
			}

			for _, r := range routinesToStart {
				go r()
			}

		case <-ctx.Done():
			return
		}
	}
}

func runJob(ctx context.Context, db bolted.Database, maxHistorySize uint64, jslib *jslib.Libs, id, name string, params []byte, logger logr.Logger) func() {
	logger = logger.WithValues("jobId", id, "name", name)

	return func() {
		vm := goja.New()
		stdlib.SetStandardLibMethods(vm, jslib, db, JobsDefinitionsPath, logger)

		var err error

		defer func() {
			if err != nil {
				logger.Error(err, "job failed")
			}
		}()

		var p interface{}
		err = json.Unmarshal(params, &p)
		if err != nil {
			return
		}
		vm.Set("params", p)

		logger = logger.WithValues("params", p)

		var src string

		jobDefinitionPath := JobsDefinitionsPath.Append(fmt.Sprintf("%s.js", name))
		err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {

			src = string(tx.Get(jobDefinitionPath))
			return nil
		})

		if err != nil {
			err = fmt.Errorf("while getting job source: %w", err)
			return
		}

		err = func() error {
			var err error
			defer func() {
				p := recover()
				if p != nil {
					e, isError := p.(error)
					if isError {
						err = e
					} else {
						err = fmt.Errorf("panic while running job: %s", p)
					}
				}
			}()
			_, err = vm.RunScript(path.Join(jobDefinitionPath...), src)
			if err != nil {
				return fmt.Errorf("while running job: %w", err)
			}
			return nil
		}()

		if err != nil {
			logger.Error(err, "job run failed")

			err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
				jobRunningPath := defaultQueueRunning.Append(id)
				jobFailedPath := defaultQueueFailed.Append(id)

				tx.Delete(jobRunningPath)
				if !tx.Exists(defaultQueueFailed) {
					tx.CreateMap(defaultQueueFailed)
				}

				trimToSize(tx, defaultQueueFailed, maxHistorySize-1)

				tx.CreateMap(jobFailedPath)
				tx.Put(jobFailedPath.Append("name"), []byte(name))
				tx.Put(jobFailedPath.Append("params"), params)
				tx.Put(jobFailedPath.Append("error"), []byte(err.Error()))
				tx.Put(jobFailedPath.Append("failedAt"), []byte(time.Now().Format(time.RFC3339)))

				return nil

			})
			return
		}

		err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			jobRunningPath := defaultQueueRunning.Append(id)
			jobSucceededPath := defaultQueueSucceeded.Append(id)

			tx.Delete(jobRunningPath)
			if !tx.Exists(defaultQueueSucceeded) {
				tx.CreateMap(defaultQueueSucceeded)
			}

			trimToSize(tx, defaultQueueSucceeded, maxHistorySize-1)

			tx.CreateMap(jobSucceededPath)
			tx.Put(jobSucceededPath.Append("name"), []byte(name))
			tx.Put(jobSucceededPath.Append("params"), params)
			tx.Put(jobSucceededPath.Append("finishedAt"), []byte(time.Now().Format(time.RFC3339)))

			return nil

		})

	}
}

func trimToSize(tx bolted.SugaredWriteTx, mapPath dbpath.Path, maxSize uint64) {
	currentSize := tx.Size(mapPath)
	if currentSize <= maxSize {
		return
	}

	toDelete := []dbpath.Path{}

	it := tx.Iterator(mapPath)
	for currentSize > maxSize && !it.IsDone() {
		toDelete = append(toDelete, mapPath.Append(it.GetKey()))
		currentSize--
		it.Next()
	}

	for _, p := range toDelete {
		tx.Delete(p)
	}

}
