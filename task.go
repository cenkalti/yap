package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

// Task is a file stored in tasksDir.
// UUID field is same with the filename of the task. Example: <tasksDir>/<UUID>.task
type Task struct {
	UUID      uuid.UUID
	Title     string
	CreatedAt time.Time
}

// NewTaskFromFile parsed filename in dir as a Task.
func NewTaskFromFile(dir, filename string) (t Task, err error) {
	uuidStr := filename[:len(filename)-len(taskExt)]
	t.UUID, err = uuid.FromString(uuidStr)
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(dir, filename))
	if err != nil {
		return
	}
	defer f.Close()
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
		key, value := parts[0], parts[1]
		switch key {
		case "title":
			t.Title = value
		case "created_at":
			t.CreatedAt, err = time.Parse(time.RFC3339Nano, value)
			if err != nil {
				return
			}
		default:
			err = errors.New("invalid key")
			return
		}
	}
	err = scanner.Err()
	return
}

// Line returns a string for print the task to console.
func (t Task) Line() string {
	// return strconv.FormatInt(t.ID, 36) + " " + t.Title
	return t.Title
}

// Write the task to file at <tasksDir>/<UUID>.task
func (t Task) Write() error {
	path := filepath.Join(tasksDir, t.UUID.String()) + taskExt
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString("title " + t.Title + "\n")
	w.WriteString("created_at " + t.CreatedAt.Format(time.RFC3339Nano) + "\n")
	if err = w.Flush(); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

// ListTasks returns all tasks in tasksDir.
func ListTasks() ([]Task, error) {
	f, err := os.Open(tasksDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		t, err := NewTaskFromFile(tasksDir, name)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
