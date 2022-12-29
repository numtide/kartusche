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
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func SetStandardLibMethods(vm *goja.Runtime, jslib *jslib.Libs, db bolted.Database, handlerParentPath dbpath.Path, logger *zap.SugaredLogger) {
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
	vm.Set("uuidv7", func() (string, error) {
		id, err := uuid.NewV7(uuid.NanosecondPrecision)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})

}
