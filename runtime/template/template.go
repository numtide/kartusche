package template

import (
	"fmt"
	"io"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
)

var basePath = dbpath.ToPath("templates")

type partialsProvider struct {
	db          bolted.Database
	currentPath dbpath.Path
}

func (pp *partialsProvider) Get(name string) (string, error) {
	parts := strings.Split(name, "/")

	dbp := basePath.Append(parts...)

	var templateData []byte

	err := bolted.SugaredRead(pp.db, func(tx bolted.SugaredReadTx) error {
		if len(parts) == 1 && tx.Exists(pp.currentPath.Append(parts...)) {
			templateData = tx.Get(pp.currentPath.Append(parts...))
			return nil
		}
		if tx.Exists(dbp) {
			templateData = tx.Get(dbp)
			return nil
		}
		return fmt.Errorf("could not find template %s", name)
	})

	if err != nil {
		return "", fmt.Errorf("could not read template: %w", err)
	}

	return string(templateData), nil

}

func RenderTemplate(db bolted.Database, currentPath dbpath.Path, w io.Writer) func(name string, data interface{}) error {

	return func(name string, data interface{}) error {

		if data == nil {
			data = map[string]interface{}{}
		}

		pp := &partialsProvider{db: db, currentPath: currentPath}
		pd, err := pp.Get(name)
		if err != nil {
			return err
		}

		tmp, err := mustache.ParseStringPartials(pd, pp)
		if err != nil {
			return fmt.Errorf("while parsing %s: %w", name, err)
		}
		return tmp.FRender(w, data)
	}

}

func RenderTemplateToString(db bolted.Database, currentPath dbpath.Path) func(name string, data interface{}) (string, error) {

	return func(name string, data interface{}) (s string, err error) {

		if data == nil {
			data = map[string]interface{}{}
		}

		pp := &partialsProvider{db: db, currentPath: currentPath}
		pd, err := pp.Get(name)
		if err != nil {
			return "", err
		}

		tmp, err := mustache.ParseStringPartials(pd, pp)
		if err != nil {
			return "", fmt.Errorf("while parsing %s: %w", name, err)
		}

		return tmp.Render(data)
	}

}
