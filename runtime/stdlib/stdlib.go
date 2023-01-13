package stdlib

import (
	"fmt"
	"net/url"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/draganm/kartusche/runtime/httprequest"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/template"
	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
)

func SetStandardLibMethods(vm *goja.Runtime, jslib *jslib.Libs, db bolted.Database, handlerParentPath dbpath.Path, logger logr.Logger) {
	dbw := dbwrapper.New(db, vm, logger)
	vm.SetFieldNameMapper(newSmartCapFieldNameMapper())
	vm.Set("require", jslib.Require(vm))
	vm.Set("println", fmt.Println)
	vm.Set("read", dbw.Read)
	vm.Set("write", dbw.Write)
	vm.Set("http_do", httprequest.Request)
	vm.Set("render_template_to_s", template.RenderTemplateToString(db, handlerParentPath))
	vm.Set("pathEscape", url.PathEscape)
	vm.Set("pathUnescape", url.PathUnescape)
	vm.Set("queryEscape", url.QueryEscape)
	vm.Set("parseUrl", url.Parse)
	vm.Set("queryUnescape", url.QueryEscape)

	vm.Set("uuidv4", func() (string, error) {
		id, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		return id.String(), nil

	})

	vm.Set("uuidv6", func() (string, error) {
		id, err := uuid.NewV6()
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})

	vm.Set("parseV6Timestamp", func(ts string) (int64, error) {
		id, err := uuid.FromString(ts)
		if err != nil {
			return 0, fmt.Errorf("could not parse uuid: %w", err)
		}

		tsv6, err := uuid.TimestampFromV6(id)
		if err != nil {
			return 0, fmt.Errorf("could not extract timestamp from v6 uuid: %w", err)
		}

		t, err := tsv6.Time()
		if err != nil {
			return 0, fmt.Errorf("could not extract time from v6 uuid timestamp: %w", err)
		}

		return t.UnixMilli(), nil
	})

	vm.Set("uuidv7", func() (string, error) {
		id, err := uuid.NewV7(uuid.NanosecondPrecision)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})

}
