package runtime

import (
	"archive/tar"
	"fmt"
	"io"
	"path"

	_ "embed"

	"github.com/dop251/goja"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/gofrs/uuid"
)

func InitializeNew(fileName string, pathPrefix string, r *tar.Reader) (err error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return fmt.Errorf("while opening database: %w", err)
	}
	defer db.Close()

	wtx, err := db.BeginWrite()
	if err != nil {
		return fmt.Errorf("while starting write transaction: %w", err)
	}

	defer func() {
		if err != nil {
			wtx.Rollback()
		} else {
			err = wtx.Finish()
		}
	}()

	initScript := ""
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("while reading header: %w", err)
		}

		cleanPath := path.Clean(h.Name)

		dp, err := dbpath.Parse(cleanPath)
		if err != nil {
			return fmt.Errorf("while parsing dbpath: %w", err)
		}

		if h.Typeflag == tar.TypeDir {
			err = wtx.CreateMap(dp)
			if err != nil {
				return err
			}
		}

		if h.Typeflag == tar.TypeReg {
			d, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("while reading entry %s: %w", cleanPath, err)
			}

			if cleanPath == "init.js" {
				initScript = string(d)
			}

			err = wtx.Put(dp, d)
			if err != nil {
				return err
			}
		}

	}

	initScriptProgram, err := goja.Compile("init.js", initScript, false)
	if err != nil {
		return fmt.Errorf("while parsing init: %w", err)
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	vm.Set("tx", wtx)
	vm.Set("uuidv4", uuid.NewV4)
	vm.Set("uuidv7", func() (string, error) {
		id, err := uuid.NewV7(uuid.NanosecondPrecision)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})

	_, err = vm.RunProgram(initScriptProgram)
	if err != nil {
		return fmt.Errorf("while running init script: %w", err)
	}

	return nil
}
