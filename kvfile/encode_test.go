package kvfile

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	v := struct {
		S string `key:"s"`
		I int    `key:"i"`
	}{
		S: "test",
		I: 1,
	}
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	t.Log("data: ", s)
	if s != "s test\ni 1\n" {
		t.Fatal("invalid data")
	}
}
