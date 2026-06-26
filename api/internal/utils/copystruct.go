package utils

import (
	"reflect"

	"github.com/zeromicro/go-zero/core/logx"
)

func CopyStruct(src, dst interface{}, isCopyEmptyString bool) error {
	sval := reflect.ValueOf(src).Elem()
	dval := reflect.ValueOf(dst).Elem()
	for i := 0; i < sval.NumField(); i++ {
		val := sval.Field(i)
		name := sval.Type().Field(i).Name
		kind := sval.Type().Field(i).Type.Kind()
		// fmt.Println(name, kind, val)
		if kind == reflect.Slice {
			continue
		}
		if kind == reflect.Struct {
			continue
		}

		if kind == reflect.String || kind == reflect.Int64 || kind == reflect.Int || kind == reflect.Uint64 {
			dvalue := dval.FieldByName(name)

			if dvalue.IsValid() {
				dkind := dvalue.Kind()
				if dkind == kind {
					// fmt.Println(name, kind, val)
					dvalue.Set(val)

				} else {
					logx.Error("Err while copy ", dkind, " ", kind, " ", name)
				}
			} else {

			}
		}

	}
	return nil
}
