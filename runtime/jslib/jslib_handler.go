package jslib

import (
	"fmt"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
)

type library struct {
	p   *goja.Program
	val goja.Value
}

type Libs struct {
	byName map[string]*library
}

func Load(tx bolted.SugaredReadTx) (*Libs, error) {
	libs := &Libs{
		byName: map[string]*library{},
	}
	libPath := dbpath.ToPath("lib")
	if !tx.Exists(libPath) {
		return libs, nil
	}
	toDo := []dbpath.Path{libPath}

	for len(toDo) > 0 {
		current := toDo[0]
		toDo = toDo[1:]
		for it := tx.Iterator(current); !it.IsDone(); it.Next() {
			key := it.GetKey()
			fullPath := current.Append(key)
			if tx.IsMap(fullPath) {
				toDo = append(toDo, fullPath)
				continue
			}
			src := string(it.GetValue())
			libPath := path.Join([]string(fullPath)...)

			pr, err := goja.Compile(fullPath.String(), fmt.Sprintf(`(() => { var exports = {}; var module = { exports: exports}; %s; return module.exports})()`, src), false)
			if err != nil {
				return nil, fmt.Errorf("while compiling %s: %w", fullPath.String(), err)
			}

			libs.byName[libPath] = &library{
				p: pr,
			}
		}
	}
	return libs, nil
}

func (l *Libs) Require(vm *goja.Runtime) func(name string) (goja.Value, error) {
	return func(name string) (goja.Value, error) {
		if !strings.HasSuffix(name, ".js") {
			name = fmt.Sprintf("%s.js", name)
		}

		l, found := l.byName[name]
		if !found {
			return nil, fmt.Errorf("not found: %s", name)
		}

		if l.val != nil {
			return l.val, nil
		}

		val, err := vm.RunProgram(l.p)
		if err != nil {
			return nil, err
		}

		fmt.Println("initialized", name)

		l.val = val

		return val, nil

	}

}
