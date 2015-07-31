package runtime

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/ZubairNabi/grpc-gateway/internal"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
)

// PopulateQueryParameters populates "values" into "msg".
// A value is ignored if its key starts with one of the elements in "filters".
func PopulateQueryParameters(msg proto.Message, values url.Values, filter *internal.DoubleArray) error {
	for key, values := range values {
		fieldPath := strings.Split(key, ".")
		if filter.HasCommonPrefix(fieldPath) {
			continue
		}
		if err := populateQueryParameter(msg, fieldPath, values); err != nil {
			return err
		}
	}
	return nil
}

func populateQueryParameter(msg proto.Message, fieldPath []string, values []string) error {
	m := reflect.ValueOf(msg)
	if m.Kind() != reflect.Ptr {
		return fmt.Errorf("unexpected type %T: %v", msg, msg)
	}
	m = m.Elem()
	for i, fieldName := range fieldPath {
		isLast := i == len(fieldPath)-1
		if !isLast && m.Kind() != reflect.Struct {
			return fmt.Errorf("non-aggregate type in the mid of path: %s", strings.Join(fieldPath, "."))
		}
		f := m.FieldByName(internal.PascalFromSnake(fieldName))
		if !f.IsValid() {
			glog.Warningf("field not found in %T: %s", msg, strings.Join(fieldPath, "."))
			return nil
		}

		switch f.Kind() {
		case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Int32, reflect.Int64, reflect.String, reflect.Uint32, reflect.Uint64:
			m = f
		case reflect.Slice:
			// TODO(yugui) Support []byte
			if !isLast {
				return fmt.Errorf("unexpected repeated field in %s", strings.Join(fieldPath, "."))
			}
			return populateRepeatedField(f, values)
		case reflect.Ptr:
			if f.IsNil() {
				m = reflect.New(f.Type().Elem())
				f.Set(m)
			}
			m = f.Elem()
			continue
		case reflect.Struct:
			m = f
			continue
		default:
			return fmt.Errorf("unexpected type %s in %T", f.Type(), msg)
		}
	}
	switch len(values) {
	case 0:
		return fmt.Errorf("no value of field: %s", strings.Join(fieldPath, "."))
	case 1:
	default:
		glog.Warningf("too many field values: %s", strings.Join(fieldPath, "."))
	}
	return populateField(m, values[0])
}

func populateRepeatedField(f reflect.Value, values []string) error {
	elemType := f.Type().Elem()
	conv, ok := convFromType[elemType.Kind()]
	if !ok {
		return fmt.Errorf("unsupported field type %s", elemType)
	}
	f.Set(reflect.MakeSlice(f.Type(), len(values), len(values)))
	for i, v := range values {
		result := conv.Call([]reflect.Value{reflect.ValueOf(v)})
		if err := result[1].Interface(); err != nil {
			return err.(error)
		}
		f.Index(i).Set(result[0])
	}
	return nil
}

func populateField(f reflect.Value, value string) error {
	conv, ok := convFromType[f.Kind()]
	if !ok {
		return fmt.Errorf("unsupported field type %T", f)
	}
	result := conv.Call([]reflect.Value{reflect.ValueOf(value)})
	if err := result[1].Interface(); err != nil {
		return err.(error)
	}
	f.Set(result[0])
	return nil
}

var (
	convFromType = map[reflect.Kind]reflect.Value{
		reflect.String:  reflect.ValueOf(String),
		reflect.Bool:    reflect.ValueOf(Bool),
		reflect.Float64: reflect.ValueOf(Float64),
		reflect.Float32: reflect.ValueOf(Float32),
		reflect.Int64:   reflect.ValueOf(Int64),
		reflect.Int32:   reflect.ValueOf(Int32),
		reflect.Uint64:  reflect.ValueOf(Uint64),
		reflect.Uint32:  reflect.ValueOf(Uint32),
		// TODO(yugui) Support []byte
	}
)
