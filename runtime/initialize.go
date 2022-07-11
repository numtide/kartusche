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
	"github.com/draganm/kartusche/common/paths"
	"github.com/draganm/kartusche/common/util/path"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/stdlib"
)

func InitializeNew(fileName, dir string) (err error) {

	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return fmt.Errorf("while opening database: %w", err)
	}
	defer db.Close()

	err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
		for _, p := range paths.WellKnown {
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, p)
			}
			err = loadFromPath(p, tx, dbpath.ToPath(filepath.Base(p)))
			if err != nil {
				return fmt.Errorf("while loading %s: %w", p, err)
			}
		}

		dataPath := dbpath.ToPath("data")
		ex := tx.Exists(dataPath)

		if !ex {
			tx.CreateMap(dataPath)
		}

		initPath := dbpath.ToPath("init.js")
		ex = tx.Exists(initPath)

		if ex {
			initScript := tx.Get(initPath)

			initScriptProgram, err := goja.Compile("init.js", string(initScript), false)
			if err != nil {
				return fmt.Errorf("while parsing init: %w", err)
			}

			vm := goja.New()
			lib, err := jslib.Load(tx)
			if err != nil {
				return fmt.Errorf("while loading jslib: %w", err)
			}

			stdlib.SetStandardLibMethods(vm, lib, db)
			vm.Set("tx", &dbwrapper.WriteTxWrapper{WriteTx: tx.GetRawWriteTX()})
			vm.GlobalObject().Delete("read")
			vm.GlobalObject().Delete("write")

			_, err = vm.RunProgram(initScriptProgram)
			if err != nil {
				return fmt.Errorf("while running init script: %w", err)
			}

		}

		return nil

	})

	return nil
}

func loadFromPath(dir string, wtx bolted.SugaredWriteTx, prefix dbpath.Path) error {
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

		if err != nil {
			return fmt.Errorf("while walking dir: %w", err)
		}

		pathParts := path.FilePathToDBPath(file)

		dbPathParts := pathParts[len(absDirPaths):]
		dbp := prefix.Append(dbPathParts...)
		if len(dbp) == 0 {
			return nil
		}

		if fi.IsDir() {
			wtx.CreateMap(dbp)
			return nil
		}

		if fi.Mode().IsRegular() {
			d, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("while reading %s: %w", file, err)
			}

			wtx.Put(dbp, d)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil

}
