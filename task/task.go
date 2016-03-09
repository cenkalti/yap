package task

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const taskExt = ".task"

// Task is a file stored in dirTasks.
type Task struct {
	// Randomly generated unique ID.
	ID          uint32
	Title       string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func newTaskFromFile(filename string) (t Task, err error) {
	base := filepath.Base(filename)
	t.ID, err = parseID(base[:len(base)-len(taskExt)])
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

func (t *Task) setKeyVal(key, value string) (err error) {
	switch key {
	case "title":
		t.Title = value
	case "created_at":
		t.CreatedAt, err = time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return
		}
	case "completed_at":
		var ctime time.Time
		ctime, err = time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return
		}
		t.CompletedAt = &ctime
	default:
		err = errors.New("invalid key")
	}
	return
}

// write the task to file at <dirTasks>/<ID>.task
func (t Task) write() error {
	path := filepath.Join(dirTasks, formatID(t.ID)) + taskExt
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
	if err = w.Flush(); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

// allTasks returns all tasks in dirTasks.
func (Task) allTasks() ([]Task, error) {
	f, err := os.Open(dirTasks)
	if err != nil {
		return nil, err
	}
	defer checkClose(f)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		t, err := newTaskFromFile(filepath.Join(dirTasks, name))
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
