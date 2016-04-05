package task

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/cenkalti/yap/datetime"
	"github.com/satori/go.uuid"
)

const taskExt = ".task"

// Task is a file stored in tasks dir.
type Task struct {
	ID          uint16
	UUID        uuid.UUID
	Title       string             `key:"title"`
	CreatedAt   time.Time          `key:"created_at"`
	CompletedAt *time.Time         `key:"completed_at"`
	DueDate     *datetime.DateTime `key:"due_date"`
	WaitDate    *datetime.DateTime `key:"wait_date"`
}

func readFile(filename string) (t Task, err error) {
	defer func() {
		if err != nil {
			err = errors.New("cannot parse " + filename + ": " + err.Error())
		}
	}()
	base := filepath.Base(filename)
	t.UUID, err = uuid.FromString(base[:len(base)-len(taskExt)])
	if err != nil {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer checkClose(f)
	kv := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		// TODO test "title                      asdf" case in file
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
	err = t.setKeys(kv)
	return
}

func (t *Task) setKeys(kv map[string]string) error {
	ptrVal := reflect.ValueOf(t)
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
		err := t.parseField(fval, sval)
		if err != nil {
			return err
		}
	}
	for key := range kv {
		return errors.New("unknown key: " + key)
	}
	return nil
}

var parsers = map[reflect.Type]func(s string) (interface{}, error){
	reflect.TypeOf(""): func(s string) (interface{}, error) {
		return s, nil
	},
	reflect.TypeOf(time.Time{}): func(s string) (interface{}, error) {
		return time.Parse(time.RFC3339Nano, s)
	},
	reflect.TypeOf(datetime.DateTime{}): func(s string) (interface{}, error) {
		return datetime.Parse(s)
	},
}

func (t *Task) parseField(field reflect.Value, str string) (err error) {
	var typ reflect.Type
	if field.Kind() == reflect.Ptr {
		typ = field.Type().Elem()
	} else {
		typ = field.Type()
	}
	parser, ok := parsers[typ]
	if !ok {
		panic("unknown type: " + typ.String())
	}
	iface, err := parser(str)
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

// write the task to file at <dirTasks>/<UUID>.task
func (t Task) write() error {
	path := filepath.Join(dirTasks, t.UUID.String()) + taskExt
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if err = t.writeFields(w); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

func (t Task) writeFields(w io.Writer) error {
	val := reflect.ValueOf(t)
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
		_, err := w.Write([]byte(tag + " " + stringValue(field) + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

var formatters = map[reflect.Type]func(i interface{}) string{
	reflect.TypeOf(""):                  func(i interface{}) string { return i.(string) },
	reflect.TypeOf(time.Time{}):         func(i interface{}) string { return i.(time.Time).Format(time.RFC3339Nano) },
	reflect.TypeOf(datetime.DateTime{}): func(i interface{}) string { return i.(datetime.DateTime).String() },
}

func stringValue(v reflect.Value) string {
	f, ok := formatters[v.Type()]
	if !ok {
		panic("unknown type")
	}
	return f(v.Interface())
}
