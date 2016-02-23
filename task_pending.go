package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PendingTask struct {
	ID int
	Task
}

func (t PendingTask) Link() error {
	src := filepath.Join("..", t.UUID.String()+taskExt)
	dst := filepath.Join(pendingTasksDir, strconv.Itoa(t.ID)+taskExt)
	return os.Symlink(src, dst)
}

func NextTaskID(path string) (id int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[int]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		strID := name[:len(name)-len(taskExt)]
		id, err = strconv.Atoi(strID)
		if err != nil {
			return
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
