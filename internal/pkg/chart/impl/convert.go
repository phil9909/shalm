package impl

import (
	"fmt"
	"reflect"

	"go.starlark.net/starlark"
)

func toStarlark(vi interface{}) starlark.Value {
	if vi == nil {
		return starlark.None
	}
	switch v := reflect.ValueOf(vi); v.Kind() {
	case reflect.String:
		return starlark.String(v.String())
	case reflect.Bool:
		return starlark.Bool(v.Bool())
	case reflect.Int:
		return starlark.MakeInt64(v.Int())
	case reflect.Float32:
		return starlark.Float(v.Float())
	case reflect.Float64:
		return starlark.Float(v.Float())
	case reflect.Slice:
		a := make([]starlark.Value, 0)
		for i := 0; i < v.Len(); i++ {
			a = append(a, toStarlark(v.Index(i).Interface()))
		}
		return starlark.NewList(a)
	case reflect.Ptr:
		return toStarlark(v.Elem())
	case reflect.Map:
		d := starlark.NewDict(16)
		for _, key := range v.MapKeys() {
			strct := v.MapIndex(key)
			keyValue := toStarlark(key.Interface())
			d.SetKey(keyValue, toStarlark(strct.Interface()))
		}
		return d
	default:
		panic(fmt.Errorf("cannot convert %v to starlark", v))
	}
}

func toGo(v starlark.Value) interface{} {
	if v == nil {
		return nil
	}
	switch v := v.(type) {
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return v
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return v
	case starlark.String:
		return v.GoString()
	case starlark.Indexable: // Tuple, List
		a := make([]interface{}, 0)
		for i := 0; i < starlark.Len(v); i++ {
			a = append(a, toGo(v.Index(i)))
		}
		return a
	case starlark.IterableMapping:
		d := make(map[string]interface{})

		for _, t := range v.Items() {
			key, ok := t.Index(0).(starlark.String)
			if ok {
				d[key.GoString()] = toGo(t.Index(1))
			}
		}
		return d

	case *chartImpl:
		d := make(map[string]interface{})

		for k, v := range v.values {
			d[k] = toGo(v)
		}
		return d
	default:
		panic(fmt.Errorf("cannot convert %s to GO", v.Type()))
	}
}
