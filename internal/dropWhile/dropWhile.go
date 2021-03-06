package dropWhile

import (
	"reflect"

	"github.com/wesovilabs/koazee/errors"
)

// OpCode identifier for operation DropWhile
const OpCode = "dropWhile"

// DropWhile struct for operation
type DropWhile struct {
	ItemsType  reflect.Type
	ItemsValue reflect.Value
	Func       interface{}
}

// Run performs the operation
func (op *DropWhile) Run() (reflect.Value, *errors.Error) {
	info, err := op.validate()
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if found, result := dispatch(op.ItemsValue, op.Func, info); found {
		return result, nil
	}

	newItems := reflect.MakeSlice(reflect.SliceOf(op.ItemsType), 0, 0)
	fn := reflect.ValueOf(op.Func)
	for index := 0; index < op.ItemsValue.Len(); index++ {
		item := op.ItemsValue.Index(index)
		argv := make([]reflect.Value, 1)
		argv[0] = item
		if fn.Call(argv)[0].Bool() {
			newItems = reflect.Append(newItems, item)
		}
	}
	return newItems, nil
}

func (op *DropWhile) validate() (*dropWhileInfo, *errors.Error) {
	item := &dropWhileInfo{}
	fnType := reflect.TypeOf(op.Func)
	if val := cache.get(op.ItemsType, fnType); val != nil {
		return val, nil
	}
	function := reflect.ValueOf(op.Func)
	item.fnValue = function
	if function.Type().Kind() != reflect.Func {
		return nil, errors.InvalidArgument(OpCode, "The dropWhile operation requires a function as argument")
	}
	if function.Type().NumIn() != 1 {
		return nil, errors.InvalidArgument(OpCode, "The provided function must retrieve 1 argument")
	}
	if function.Type().NumOut() != 1 {
		return nil, errors.InvalidArgument(OpCode, "The provided function must return 1 value")
	}
	fnOut := reflect.New(function.Type().Out(0)).Elem()
	fnIn := reflect.New(function.Type().In(0)).Elem()

	if fnIn.Type() != op.ItemsType {
		return nil, errors.InvalidArgument(OpCode,
			"The type of the argument in the provided function must be %s",
			op.ItemsType.String())
	}
	if fnOut.Kind() != reflect.Bool {
		return nil, errors.InvalidArgument(OpCode, "The type of the Output in the provided function must be bool")
	}
	item.fnInputType = fnIn.Type()
	cache.add(op.ItemsType, fnType, item)
	return item, nil
}
