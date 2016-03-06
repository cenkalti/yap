package task

import (
	"os"
	"path/filepath"
	"strings"
)

type linkedTask struct {
	SmallID uint32
	Task
}

// tasksIn returns all tasks in dir.
func tasksIn(dir string) ([]linkedTask, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []linkedTask
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		id, err := parseID(name)
		if err != nil {
			return nil, err
		}
		t, err := newTaskFromFile(dirTasks, name)
		if err != nil {
			return nil, err
		}
		ti := linkedTask{
			SmallID: id,
			Task:    t,
		}
		tasks = append(tasks, ti)
	}
	return tasks, nil
}

// link writes a symlink to dir that is pointing to original task in dirTasks.
func (t linkedTask) link(dir string) error {
	src := filepath.Join("..", formatID(t.ID)+taskExt)
	dst := filepath.Join(dir, formatID(t.SmallID)+taskExt)
	return os.Symlink(src, dst)
}
