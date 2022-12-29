package cronjobs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/stdlib"
	"github.com/go-logr/zapr"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var cronjobsPath = dbpath.ToPath("cronjobs")
var scheduleRegExp = regexp.MustCompile(`^\s*(#|\/\/)\s+(.+)$`)

func CreateCron(tx bolted.SugaredReadTx, jslib *jslib.Libs, db bolted.Database, logger *zap.SugaredLogger) (*cron.Cron, error) {
	cronLogger := zapr.NewLogger(logger.Desugar())
	cr := cron.New(
		// cron.WithLogger(cronLogger),
		cron.WithSeconds(),
		cron.WithParser(
			cron.NewParser(
				cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
			),
		),
		cron.WithChain(
			cron.Recover(cronLogger),
		),
	)

	if !tx.Exists(cronjobsPath) {
		return cr, nil
	}

	for it := tx.Iterator(cronjobsPath); !it.IsDone(); it.Next() {
		if !strings.HasSuffix(it.GetKey(), ".js") {
			return nil, fmt.Errorf("non js file found in 'cronjobs': %s", it.GetKey())
		}

		src := string(it.GetValue())

		lines := strings.Split(src, "\n")
		if len(lines) < 1 {
			return nil, fmt.Errorf("could not find schedule for cron %s", it.GetKey())
		}

		matches := scheduleRegExp.FindStringSubmatch(lines[0])
		if len(matches) == 0 {
			return nil, fmt.Errorf("could not find schedule for cron %s", it.GetKey())
		}

		prg, err := goja.Compile(it.GetKey(), src, true)

		if err != nil {
			return nil, fmt.Errorf("while compiling cronjob %s: %w", it.GetKey(), err)
		}

		cr.AddFunc(matches[2], func() {

			vm := goja.New()
			stdlib.SetStandardLibMethods(vm, jslib, db, cronjobsPath, logger)
			_, err = vm.RunProgram(prg)
			if err != nil {
				logger.With("error", err, "cron", it.GetKey()).Error("failed to execute cron")
			}
		})
	}

	return cr, nil

}
