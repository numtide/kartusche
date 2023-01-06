package dbwrapper

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
)

func iteratorFor(ig func(dbpath.Path) (bolted.Iterator, error), vm *goja.Runtime, path []string, seek string) (*goja.Object, error) {

	it, err := ig(dataPath.Append(path...))
	if err != nil {
		return nil, fmt.Errorf("while creating iterator: %w", err)
	}

	if seek != "" {
		it.Seek(seek)
	}

	type iterResult struct {
		Done  bool
		Value goja.Value
	}

	o := vm.NewObject()
	o.SetSymbol(goja.SymIterator, func() (*goja.Object, error) {
		iter := vm.NewObject()
		iter.Set("next", func() (*iterResult, error) {

			done, err := it.IsDone()
			if err != nil {
				return nil, fmt.Errorf("while getting isDone from iterator: %w", err)
			}

			if done {
				return &iterResult{
					Done: true,
				}, nil
			}

			key, err := it.GetKey()
			if err != nil {
				return nil, fmt.Errorf("while getting key from iterator: %w", err)
			}

			value, err := it.GetValue()
			if err != nil {
				return nil, fmt.Errorf("while getting value from iterator: %w", err)
			}

			err = it.Next()
			if err != nil {
				return nil, fmt.Errorf("getting next from iterator: %w", err)
			}

			return &iterResult{
				Value: vm.ToValue([]string{key, string(value)}),
				Done:  false,
			}, nil
		})
		return iter, nil
	})
	return o, nil

}
