package kvfile

import (
	"io"
	"reflect"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(i interface{}) error {
	val := reflect.ValueOf(i)
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag.Get("key")
		if tag == "" {
			continue
		}
		field := val.Field(i)
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue
			}
			field = field.Elem()
		}
		_, err := e.w.Write([]byte(tag + " " + stringValue(field) + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func stringValue(v reflect.Value) string {
	c := getCodec(v.Type())
	return c.Encode(v.Interface())
}
