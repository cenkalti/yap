package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Task struct {
	ID    int64
	Title string
}

func NewTaskFromFile(dir, name string) (t Task, err error) {
	id := name[:len(name)-5]
	t.ID, err = strconv.ParseInt(id, 36, 64)
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(dir, name))
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
		default:
			err = errors.New("invalid key")
			return
		}
	}
	err = scanner.Err()
	return
}

func (t Task) Line() string {
	return strconv.FormatInt(t.ID, 36) + " " + t.Title
}

func (t Task) WriteToFile(dir string) error {
	f, err := os.Create(filepath.Join(dir, strconv.FormatInt(t.ID, 36)) + ".task")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString("title " + t.Title + "\n")
	err = w.Flush()
	if err != nil {
		return err
	}
	return f.Close()
}

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
		if !strings.HasSuffix(name, ".task") {
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

func NextTaskID() (id int64, err error) {
	f, err := os.Open(tasksDir)
	if err != nil {
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[int64]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, ".task") {
			continue
		}
		strID := name[:len(name)-5]
		id, err = strconv.ParseInt(strID, 36, 64)
		if err != nil {
			return
		}
		ids[id] = struct{}{}
	}
	for id = int64(1); ; id++ {
		_, ok := ids[id]
		if !ok {
			return
		}
	}
}
