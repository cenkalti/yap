package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const taskExt = ".task"

// Task is a file stored in DirTasks.
type Task struct {
	ID        uint32
	Title     string
	CreatedAt time.Time
}

// NewTaskFromFile parsed filename in dir as a Task.
func NewTaskFromFile(dir, filename string) (t Task, err error) {
	idStr := filename[:len(filename)-len(taskExt)]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return
	}
	t.ID = uint32(id)
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

// Write the task to file at <DirTasks>/<ID>.task
func (t Task) Write() error {
	path := filepath.Join(DirTasks, strconv.FormatUint(uint64(t.ID), 10)) + taskExt
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

// AllTasks returns all tasks in DirTasks.
func AllTasks() ([]Task, error) {
	f, err := os.Open(DirTasks)
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
		t, err := NewTaskFromFile(DirTasks, name)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

type ByCreatedAtDesc []Task

func (t ByCreatedAtDesc) Len() int           { return len(t) }
func (t ByCreatedAtDesc) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByCreatedAtDesc) Less(i, j int) bool { return t[i].CreatedAt.After(t[j].CreatedAt) }
