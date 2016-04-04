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
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, " ", 2)
		if len(parts) != 2 {
			err = errors.New("invalid task file")
			return
		}
		err = t.setKeyVal(parts[0], parts[1])
		if err != nil {
			return
		}
	}
	err = scanner.Err()
	return
}

var parsers = map[string]func(t *Task, value string) error{
	"title": func(t *Task, value string) (err error) {
		t.Title = value
		return
	},
	"created_at": func(t *Task, value string) (err error) {
		t.CreatedAt, err = time.Parse(time.RFC3339Nano, value)
		return
	},
	"completed_at": func(t *Task, value string) (err error) {
		var ctime time.Time
		ctime, err = time.Parse(time.RFC3339Nano, value)
		t.CompletedAt = &ctime
		return
	},
	"due_date": func(t *Task, value string) (err error) {
		var dt datetime.DateTime
		dt, err = datetime.Parse(value)
		t.DueDate = &dt
		return
	},
	"wait_date": func(t *Task, value string) (err error) {
		var dt datetime.DateTime
		dt, err = datetime.Parse(value)
		t.WaitDate = &dt
		return
	},
}

func (t *Task) setKeyVal(key, value string) error {
	f, ok := parsers[key]
	if !ok {
		return errors.New("invalid key")
	}
	return f(t, value)
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
		iface := fval.Interface()
		var s string
		switch fval.Type() {
		case reflect.TypeOf(""):
			s = iface.(string)
		case reflect.TypeOf(time.Time{}):
			s = iface.(time.Time).Format(time.RFC3339Nano)
		case reflect.TypeOf(datetime.DateTime{}):
			s = iface.(datetime.DateTime).String()
		default:
			return errors.New("invalid key")
		}
		_, err := w.Write([]byte(tag + " " + s + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
