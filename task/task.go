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

// TODO make map
func (t *Task) parseField(val reflect.Value, sval string) (err error) {
	typ := val.Type()
	switch typ {
	case reflect.TypeOf(""):
		val.SetString(sval)
	case reflect.TypeOf(time.Time{}):
		var tm time.Time
		tm, err = time.Parse(time.RFC3339Nano, sval)
		val.Set(reflect.ValueOf(tm))
	case reflect.TypeOf(&time.Time{}):
		var tm time.Time
		tm, err = time.Parse(time.RFC3339Nano, sval)
		val.Set(reflect.ValueOf(&tm))
	case reflect.TypeOf(datetime.DateTime{}):
		var dt datetime.DateTime
		dt, err = datetime.Parse(sval)
		val.Set(reflect.ValueOf(dt))
	case reflect.TypeOf(&datetime.DateTime{}):
		var dt datetime.DateTime
		dt, err = datetime.Parse(sval)
		val.Set(reflect.ValueOf(&dt))
	default:
		panic("unknown type: " + typ.String())
	}
	return
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
		fval := val.Field(i)
		if fval.Kind() == reflect.Ptr {
			if fval.IsNil() {
				continue
			}
			fval = fval.Elem()
		}
		_, err := w.Write([]byte(tag + " " + stringValue(fval) + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO make map
func stringValue(v reflect.Value) string {
	i := v.Interface()
	switch v.Type() {
	case reflect.TypeOf(""):
		return i.(string)
	case reflect.TypeOf(time.Time{}):
		return i.(time.Time).Format(time.RFC3339Nano)
	case reflect.TypeOf(datetime.DateTime{}):
		return i.(datetime.DateTime).String()
	default:
		panic("unknown type")
	}
}
