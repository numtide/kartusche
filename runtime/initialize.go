package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/gofrs/uuid"
)

func InitializeNew(fileName, dir string) (err error) {

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

	pathsToLoad := map[string]string{
		"static":    "static",
		"handler":   "handler",
		"lib":       "lib",
		"tests":     "tests",
		"templates": "templates",
		"init.js":   "init.js",
	}

	for p, pth := range pathsToLoad {
		if !filepath.IsAbs(pth) {
			pth = filepath.Join(dir, pth)
		}
		err = loadFromPath(pth, wtx, dbpath.ToPath(filepath.Base(p)))
		if err != nil {
			return fmt.Errorf("while loading %s: %w", pth, err)
		}
	}

	dataPath := dbpath.ToPath("data")
	ex, err := wtx.Exists(dataPath)
	if err != nil {
		return err
	}

	if !ex {
		err = wtx.CreateMap(dataPath)
		if err != nil {
			return err
		}
	}

	initPath := dbpath.ToPath("init.js")
	ex, err = wtx.Exists(initPath)
	if err != nil {
		return err
	}

	if ex {
		initScript, err := wtx.Get(initPath)
		if err != nil {
			return fmt.Errorf("while getting init.js: %w", err)
		}

		initScriptProgram, err := goja.Compile("init.js", string(initScript), false)
		if err != nil {
			return fmt.Errorf("while parsing init: %w", err)
		}

		vm := goja.New()
		vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
		vm.Set("tx", &dbwrapper.WriteTxWrapper{WriteTx: wtx})
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

	}

	return nil
}

func loadFromPath(dir string, wtx bolted.WriteTx, prefix dbpath.Path) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("while getting abs path: %w", err)
	}

	_, err = os.Stat(absDir)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	absDirPaths := strings.Split(absDir, string(os.PathSeparator))

	err = filepath.Walk(absDir, func(file string, fi os.FileInfo, err error) error {

		pathParts := strings.Split(file, string(os.PathSeparator))

		dbPathParts := pathParts[len(absDirPaths):]

		dbp := prefix.Append(dbPathParts...)
		if len(dbp) == 0 {
			return nil
		}

		if fi.IsDir() {
			err = wtx.CreateMap(dbp)
			if err != nil {
				return err
			}
			return nil
		}

		if fi.Mode().IsRegular() {
			d, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("while reading %s: %w", file, err)
			}

			err = wtx.Put(dbp, d)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil

}
