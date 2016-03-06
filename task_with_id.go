package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type TaskWithSmallID struct {
	SmallID uint32
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
		id, err := strconv.ParseUint(name, 10, 32)
		if err != nil {
			return nil, err
		}
		t, err := NewTaskFromFile(DirTasks, name)
		if err != nil {
			return nil, err
		}
		ti := TaskWithSmallID{
			SmallID: uint32(id),
			Task:    t,
		}
		tasks = append(tasks, ti)
	}
	return tasks, nil
}

// Link writes a symlink to dir that is pointing to original task in DirTasks.
func (t TaskWithSmallID) Link(dir string) error {
	src := filepath.Join("..", strconv.FormatUint(uint64(t.ID), 10)+taskExt)
	dst := filepath.Join(dir, strconv.FormatUint(uint64(t.SmallID), 10)+taskExt)
	return os.Symlink(src, dst)
}

// NextTaskID returns the minimum available integer id in dir.
func NextTaskID(dir string) (id uint32, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[uint32]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		strID := name[:len(name)-len(taskExt)]
		uid, err := strconv.ParseUint(strID, 10, 32)
		if err != nil {
			return 0, err
		}
		id := uint32(uid)
		ids[id] = struct{}{}
	}
	for id = 1; ; id++ {
		_, ok := ids[id]
		if !ok {
			return
		}
	}
}
