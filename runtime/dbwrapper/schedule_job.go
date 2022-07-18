package dbwrapper

import (
	"encoding/json"
	"fmt"

	"github.com/draganm/bolted/dbpath"
	"github.com/gofrs/uuid"
)

var jobsDefinitionsPath = dbpath.ToPath("jobs")
var jobQueuePath = dbpath.ToPath("job-queue")

func ScheduleJob(txw *WriteTxWrapper) func(name string, params interface{}) error {
	return func(name string, params interface{}) error {

		tx := txw.WriteTx

		pd, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("while marshalling job params for %s: %w", name, err)
		}

		jobID, err := uuid.NewV6()
		if err != nil {
			return fmt.Errorf("while creating job id for %s: %w", name, err)
		}

		jobDefinitionPath := jobsDefinitionsPath.Append(fmt.Sprintf("%s.js", name))

		ex, err := tx.Exists(jobDefinitionPath)
		if err != nil {
			return err
		}

		if !ex {
			return fmt.Errorf("could not find job %s", name)
		}

		ex, err = tx.Exists(jobQueuePath)
		if err != nil {
			return err
		}

		if !ex {
			err = tx.CreateMap(jobQueuePath)
			if err != nil {
				return err
			}
		}

		defaultQueuePath := jobQueuePath.Append("default")

		ex, err = tx.Exists(defaultQueuePath)
		if err != nil {
			return err
		}

		if !ex {
			err = tx.CreateMap(defaultQueuePath)
			if err != nil {
				return err
			}
		}

		scheduledPath := defaultQueuePath.Append("scheduled")

		ex, err = tx.Exists(scheduledPath)
		if err != nil {
			return err
		}

		if !ex {
			err = tx.CreateMap(scheduledPath)
			if err != nil {
				return err
			}
		}

		scheduledJobPath := scheduledPath.Append(jobID.String())

		tx.CreateMap(scheduledJobPath)
		tx.Put(scheduledJobPath.Append("name"), []byte(name))
		tx.Put(scheduledJobPath.Append("params"), pd)
		return nil

	}
}
