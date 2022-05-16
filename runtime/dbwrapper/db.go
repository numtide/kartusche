package dbwrapper

import (
	"fmt"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
)

type DB struct {
	db bolted.Database
}

func New(db bolted.Database) *DB {
	return &DB{db: db}
}

func (db *DB) Read(f func(*readTxWrapper) (interface{}, error)) (interface{}, error) {
	tx, err := db.db.BeginRead()
	if err != nil {
		return nil, fmt.Errorf("while beginning read tx: %w", err)
	}

	defer func() {
		err = tx.Finish()
	}()

	return f(&readTxWrapper{ReadTx: tx})

}

type readTxWrapper struct {
	bolted.ReadTx
}

var dataPath = dbpath.ToPath("data")

func (rtw *readTxWrapper) Get(path []string) (string, error) {
	d, err := rtw.ReadTx.Get(dataPath.Append(path...))
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func (rtw *readTxWrapper) Exists(path []string) (bool, error) {
	return rtw.ReadTx.Exists(dataPath.Append(path...))
}
func (rtw *readTxWrapper) IsMap(path []string) (bool, error) {
	return rtw.ReadTx.IsMap(dataPath.Append(path...))
}
func (rtw *readTxWrapper) Size(path []string) (uint64, error) {
	return rtw.ReadTx.Size(dataPath.Append(path...))
}
func (rtw *readTxWrapper) ID() (uint64, error) {
	return rtw.ReadTx.ID()
}

func (rtw *readTxWrapper) Iterator(path []string) (*iteratorWrapper, error) {
	it, err := rtw.ReadTx.Iterator(dataPath.Append(path...))
	if err != nil {
		return nil, err
	}

	return &iteratorWrapper{Iterator: it}, nil
}

type WriteTxWrapper struct {
	bolted.WriteTx
}

func (wtw *WriteTxWrapper) Get(path []string) (string, error) {
	d, err := wtw.WriteTx.Get(dataPath.Append(path...))
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func (wtw *WriteTxWrapper) Iterator(path []string) (*iteratorWrapper, error) {
	it, err := wtw.WriteTx.Iterator(dataPath.Append(path...))
	if err != nil {
		return nil, err
	}

	return &iteratorWrapper{Iterator: it}, nil
}

func (wtw *WriteTxWrapper) Exists(path []string) (bool, error) {
	return wtw.WriteTx.Exists(dataPath.Append(path...))
}
func (wtw *WriteTxWrapper) IsMap(path []string) (bool, error) {
	return wtw.WriteTx.IsMap(dataPath.Append(path...))
}
func (wtw *WriteTxWrapper) Size(path []string) (uint64, error) {
	return wtw.WriteTx.Size(dataPath.Append(path...))
}
func (wtw *WriteTxWrapper) ID() (uint64, error) {
	return wtw.WriteTx.ID()
}

func (wtw *WriteTxWrapper) CreateMap(path dbpath.Path) error {
	return wtw.WriteTx.CreateMap(dataPath.Append(path...))
}

func (wtw *WriteTxWrapper) Delete(path dbpath.Path) error {
	return wtw.WriteTx.Delete(dataPath.Append(path...))
}

func (wtw *WriteTxWrapper) Put(path dbpath.Path, value string) error {
	return wtw.WriteTx.Put(dataPath.Append(path...), []byte(value))
}

type iteratorWrapper struct {
	bolted.Iterator
}

func (i *iteratorWrapper) GetValue() (string, error) {
	d, err := i.Iterator.GetValue()
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func (db *DB) Write(f func(*WriteTxWrapper) (interface{}, error)) (interface{}, error) {
	tx, err := db.db.BeginWrite()
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

	return f(&WriteTxWrapper{WriteTx: tx})

}
