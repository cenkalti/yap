package task

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
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
	Title       string
	CreatedAt   time.Time
	CompletedAt *time.Time
	DueDate     *datetime.DateTime
	WaitDate    *datetime.DateTime
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
	if _, err = w.WriteString("title " + t.Title + "\n"); err != nil {
		return err
	}
	if _, err = w.WriteString("created_at " + t.CreatedAt.Format(time.RFC3339Nano) + "\n"); err != nil {
		return err
	}
	if t.CompletedAt != nil {
		if _, err = w.WriteString("completed_at " + t.CompletedAt.Format(time.RFC3339Nano) + "\n"); err != nil {
			return err
		}
	}
	if t.DueDate != nil {
		if _, err = w.WriteString("due_date " + t.DueDate.String() + "\n"); err != nil {
			return err
		}
	}
	if t.WaitDate != nil {
		if _, err = w.WriteString("wait_date " + t.WaitDate.String() + "\n"); err != nil {
			return err
		}
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}
