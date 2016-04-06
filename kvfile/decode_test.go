package kvfile

import (
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	type T struct {
		S string `key:"s"`
		I int    `key:"i"`
	}
	s := "s test\ni 1"
	r := strings.NewReader(s)
	dec := NewDecoder(r)
	var v T
	err := dec.Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value: ", v)
	if v.S != "test" {
		t.Fatal("invalid field value")
	}
	if v.I != 1 {
		t.Fatal("invalid field value")
	}
}
