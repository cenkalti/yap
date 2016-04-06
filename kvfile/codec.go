package kvfile

import (
	"strconv"

	"reflect"
)

func init() {
	RegisterCodec(new(string), new(stringCodec))
	RegisterCodec(new(int), new(intCodec))
}

type Codec interface {
	Encode(interface{}) string
	Decode(string) (interface{}, error)
}

var codecs = map[reflect.Type]Codec{}

func RegisterCodec(i interface{}, c Codec) {
	typ := reflect.TypeOf(i)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	codecs[typ] = c
}

func getCodec(typ reflect.Type) Codec {
	c, ok := codecs[typ]
	if !ok {
		panic("kvfile: unknown type: " + typ.String())
	}
	return c
}

type stringCodec struct{}

func (c *stringCodec) Encode(i interface{}) string {
	return i.(string)
}

func (c *stringCodec) Decode(s string) (interface{}, error) {
	return s, nil
}

type intCodec struct{}

func (c *intCodec) Encode(i interface{}) string {
	return strconv.Itoa(i.(int))
}

func (c *intCodec) Decode(s string) (interface{}, error) {
	return strconv.Atoi(s)
}
