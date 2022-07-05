package stdlib

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/draganm/kartusche/runtime/httprequest"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/template"
	"github.com/gofrs/uuid"
)

func SetStandardLibMethods(vm *goja.Runtime, jslib *jslib.Libs, db bolted.Database) {
	dbw := dbwrapper.New(db)
	vm.SetFieldNameMapper(newSmartCapFieldNameMapper())
	vm.Set("require", jslib.Require(vm))
	vm.Set("println", fmt.Println)
	vm.Set("read", dbw.Read)
	vm.Set("write", dbw.Write)
	vm.Set("http_do", httprequest.Request)
	vm.Set("render_template_to_s", template.RenderTemplateToString(db))
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
