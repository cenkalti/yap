package main

import (
	"os"
	"path/filepath"
	"strings"
)

type TaskWithSmallID struct {
	SmallID TaskID
	Task
}

// TasksIn returns all tasks in dir.
func TasksIn(dir string) ([]TaskWithSmallID, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []TaskWithSmallID
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		id, err := ParseTaskID(name)
		if err != nil {
			return nil, err
		}
		t, err := NewTaskFromFile(DirTasks, name)
		if err != nil {
			return nil, err
		}
		ti := TaskWithSmallID{
			SmallID: id,
			Task:    t,
		}
		tasks = append(tasks, ti)
	}
	return tasks, nil
}

// Link writes a symlink to dir that is pointing to original task in DirTasks.
func (t TaskWithSmallID) Link(dir string) error {
	src := filepath.Join("..", t.ID.String()+taskExt)
	dst := filepath.Join(dir, t.SmallID.String()+taskExt)
	return os.Symlink(src, dst)
}

// NextTaskID returns the minimum available integer id in dir.
func NextTaskID(dir string) (id TaskID, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[TaskID]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		strID := name[:len(name)-len(taskExt)]
		id, err := ParseTaskID(strID)
		if err != nil {
			return 0, err
		}
		ids[id] = struct{}{}
	}
	for id = 1; ; id++ {
		_, ok := ids[id]
		if !ok {
			return
		}
	}
}
