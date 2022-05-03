package runtime

import (
	"fmt"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
)

func reader(db bolted.Database) func(f func(bolted.ReadTx) (interface{}, error)) (interface{}, error) {
	return func(f func(bolted.ReadTx) (interface{}, error)) (interface{}, error) {
		tx, err := db.BeginRead()
		if err != nil {
			return nil, fmt.Errorf("while beginning read tx: %w", err)
		}

		defer func() {
			err = tx.Finish()
		}()

		return f(tx)

	}
}

type writeTxWrapper struct {
	bolted.WriteTx
}

func (wtw *writeTxWrapper) Get(path []string) (string, error) {
	d, err := wtw.WriteTx.Get(dbpath.Path(path))
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// func (wtw *writeTxWrapper) Iterator(path []string) (*iteratorWrapper, error) {
// 	it, err := wtw.wtx.Iterator(dbpath.Path(path))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &iteratorWrapper{it: it}, nil

// }
// func (wtw *writeTxWrapper) Exists(path []string) (bool, error) {
// 	return wtw.wtx.Exists(dbpath.Path(path))
// }
// func (wtw *writeTxWrapper) IsMap(path []string) (bool, error) {
// 	return wtw.wtx.IsMap(dbpath.Path(path))
// }
// func (wtw *writeTxWrapper) Size(path []string) (uint64, error) {
// 	return wtw.wtx.Size(dbpath.Path(path))
// }
// func (wtw *writeTxWrapper) ID() (uint64, error) {
// 	return wtw.wtx.ID()
// }

// func (wtw *writeTxWrapper) CreateMap(path dbpath.Path) error {
// 	return wtw.wtx.CreateMap(dbpath.Path(path))
// }

// func (wtw *writeTxWrapper) Delete(path dbpath.Path) error {
// 	return wtw.wtx.Delete(dbpath.Path(path))
// }

// func (wtw *writeTxWrapper) Put(path dbpath.Path, value string) error {
// 	return wtw.wtx.Put(dbpath.Path(path), []byte(value))
// }

type iteratorWrapper struct {
	it bolted.Iterator
}

func writer(db bolted.Database) func(f func(*writeTxWrapper) (interface{}, error)) (interface{}, error) {
	return func(f func(*writeTxWrapper) (interface{}, error)) (res interface{}, err error) {
		tx, err := db.BeginWrite()
		if err != nil {
			return nil, fmt.Errorf("while beginning write tx: %w", err)
		}

		defer func() {
			p := recover()
			if p != nil {
				err = tx.Rollback()
				if err != nil {
					return
				}
				panic(p)
			}

			if err != nil {
				e := tx.Rollback()
				if e != nil {
					err = e
				}
			}
		}()

		defer func() {
			if err != nil {
				return
			}
			e := tx.Finish()
			if e != nil {
				err = e
			}
		}()

		return f(&writeTxWrapper{WriteTx: tx})

	}
}
