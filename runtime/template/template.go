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
	db bolted.Database
}

func (pp *partialsProvider) Get(name string) (string, error) {
	parts := strings.Split(name, "/")
	dbp := basePath.Append(parts...)
	rtx, err := pp.db.BeginRead()
	if err != nil {
		return "", fmt.Errorf("while starting read tx: %w", err)
	}

	defer rtx.Finish()

	td, err := rtx.Get(dbp)
	if err != nil {
		return "", fmt.Errorf("while reading template: %w", err)
	}

	return string(td), nil

}

func RenderTemplate(db bolted.Database, w io.Writer) func(name string, data interface{}) error {

	return func(name string, data interface{}) error {

		if data == nil {
			data = map[string]interface{}{}
		}

		pp := &partialsProvider{db: db}
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

func RenderTemplateToString(db bolted.Database) func(name string, data interface{}) (string, error) {

	return func(name string, data interface{}) (s string, err error) {

		if data == nil {
			data = map[string]interface{}{}
		}

		pp := &partialsProvider{db: db}
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
