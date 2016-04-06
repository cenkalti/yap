package kvfile

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"strings"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(i interface{}) (err error) {
	val := reflect.ValueOf(i)
	if val.Type().Kind() != reflect.Ptr {
		panic("i must be pointer")
	}

	kv := make(map[string]string)
	scanner := bufio.NewScanner(d.r)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, " ", 2)
		if len(parts) != 2 {
			err = errors.New("invalid line: " + text)
			return
		}
		key, value := parts[0], parts[1]
		if _, ok := kv[key]; ok {
			err = errors.New("duplicate key: " + key)
			return
		}
		kv[key] = value
	}
	err = scanner.Err()
	if err != nil {
		return
	}
	return setKeys(kv, val)
}

func setKeys(kv map[string]string, ptrVal reflect.Value) error {
	val := ptrVal.Elem()
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag.Get("key")
		if tag == "" {
			continue
		}
		fval := val.Field(i)
		required := (fval.Kind() != reflect.Ptr)
		sval, ok := kv[tag]
		if !ok {
			if required {
				return errors.New(tag + " is required in task")
			}
			continue
		}
		delete(kv, tag)
		err := parseField(fval, sval)
		if err != nil {
			return err
		}
	}
	for key := range kv {
		return errors.New("unknown key: " + key)
	}
	return nil
}

func parseField(field reflect.Value, str string) (err error) {
	var typ reflect.Type
	if field.Kind() == reflect.Ptr {
		typ = field.Type().Elem()
	} else {
		typ = field.Type()
	}
	c := getCodec(typ)
	iface, err := c.Decode(str)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(iface)
	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(typ)
		ptr.Elem().Set(val)
		val = ptr
	}
	field.Set(val)
	return nil
}
