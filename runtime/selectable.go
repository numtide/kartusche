package runtime

import "reflect"

type selectable interface {
	SelectChan() reflect.Value
	Fn() func(interface{}) (bool, error)
}

type defaultSelectable struct {
	ch interface{}
	fn func(interface{}) (bool, error)
}

func (ds *defaultSelectable) SelectChan() reflect.Value {
	return reflect.ValueOf(ds.ch)
}

func (ds *defaultSelectable) Fn() func(interface{}) (bool, error) {
	return ds.fn
}
