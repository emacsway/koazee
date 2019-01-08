package maps

import (
	"reflect"

	"github.com/wesovilabs/koazee/errors"
)

// OpCode code
const OpCode = "map"

// Map structure
type Map struct {
	ItemsType  reflect.Type
	ItemsValue reflect.Value
	Func       interface{}
}

// Run performs the operation
func (m *Map) Run() (reflect.Value, *errors.Error) {
	mInfo, err := m.validate()
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	// dispatch functions do not handle errors
	if !mInfo.hasError {
		if found, result := dispatch(m.ItemsValue, m.Func, mInfo); found {
			return reflect.ValueOf(result), nil
		}
	}
	newItems := reflect.MakeSlice(reflect.SliceOf(mInfo.fnOutputType), 0, 0)
	fn := reflect.ValueOf(m.Func)
	var argv = make([]reflect.Value, 1)
	if mInfo.isPtr {
		for index := 0; index < m.ItemsValue.Len(); index++ {
			argv[0] = reflect.ValueOf(reflect.ValueOf(m.ItemsValue.Index(index).Interface()).Elem().Addr().Interface())
			result := fn.Call(argv)
			if mInfo.hasError {
				if !result[1].IsNil() {
					return reflect.ValueOf(nil), errors.UserError(OpCode, result[1].Interface().(error))
				}
			}
			newItems = reflect.Append(newItems, result[0])
		}
	} else {
		for index := 0; index < m.ItemsValue.Len(); index++ {
			argv[0] = m.ItemsValue.Index(index)
			result := fn.Call(argv)
			if mInfo.hasError {
				if !result[1].IsNil() {
					return reflect.ValueOf(nil), errors.UserError(OpCode, result[1].Interface().(error))
				}
			}
			newItems = reflect.Append(newItems, result[0])
		}
	}
	return newItems, nil
}

func (m *Map) validate() (*mapInfo, *errors.Error) {
	item := &mapInfo{}
	item.fnInputType = m.ItemsType
	fnType := reflect.TypeOf(m.Func)
	/**
	if m.ItemsValue == nil {
		return nil, errors.EmptyStream(OpCode, "A nil Stream can not be iterated")
	}
	**/
	if val := cache.get(m.ItemsType, fnType); val != nil {
		return val, nil
	}
	item.fnValue = reflect.ValueOf(m.Func)
	if item.fnValue.Type().Kind() != reflect.Func {
		return nil, errors.InvalidArgument(OpCode, "The map operation requires a function as argument")
	}
	if item.fnValue.Type().NumIn() != 1 {
		return nil, errors.InvalidArgument(OpCode, "The provided function must retrieve 1 argument")
	}
	numOut := item.fnValue.Type().NumOut()
	errType := reflect.TypeOf((*error)(nil)).Elem()
	if !(numOut == 1 || (numOut == 2 && item.fnValue.Type().Out(1).Implements(errType))) {
		return nil, errors.InvalidArgument(OpCode, "The provided function must return 1 value or the second value must be an error")
	}
	fnIn := reflect.New(item.fnValue.Type().In(0)).Elem()
	if fnIn.Type() != m.ItemsType {
		return nil, errors.InvalidArgument(OpCode,
			"The type of the argument in the provided "+
				"function must be %s", m.ItemsType.String())
	}
	item.fnOutputType = reflect.New(item.fnValue.Type().Out(0)).Elem().Type()
	item.isPtr = m.ItemsValue.Index(0).Kind() == reflect.Ptr
	item.hasError = numOut == 2
	cache.add(m.ItemsType, fnType, item)
	return item, nil
}
