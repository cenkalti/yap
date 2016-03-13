package task

import (
	"os"
	"path/filepath"
	"strings"
)

// LinkedTask is a symlink to a Task for refering tasks with a more human-friendly ID number.
// Task IDs are random 32-bit integers that is hard to remember and type.
// LinkedTasks have separate IDs that is usually a small number.
type LinkedTask struct {
	LinkID uint16
	Task
}

// tasksIn returns all tasks in dir.
func tasksIn(dir string) ([]LinkedTask, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer checkClose(f)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []LinkedTask
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		id, err := parseID(name[:len(name)-len(taskExt)])
		if err != nil {
			return nil, err
		}
		ti, err := getLinkedTask(dir, id)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *ti)
	}
	return tasks, nil
}

// link writes a symlink to dir that is pointing to original task in dirTasks.
func (t LinkedTask) link(dir string) error {
	src := filepath.Join("..", "tasks", t.UUID.String()+taskExt)
	dst := filepath.Join(dir, formatID(t.LinkID)+taskExt)
	return os.Symlink(src, dst)
}

// unlink removes the symlink in dir.
func (t LinkedTask) unlink(dir string) error {
	dst := filepath.Join(dir, formatID(t.LinkID)+taskExt)
	return os.Remove(dst)
}

func (t *LinkedTask) move(olddir, newdir string) error {
	id, err := nextID(newdir)
	if err != nil {
		return err
	}
	oldpath := filepath.Join(olddir, formatID(t.LinkID)+taskExt)
	newpath := filepath.Join(newdir, formatID(id)+taskExt)
	err = os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}
	t.LinkID = id
	return nil
}

func getLinkedTask(dir string, id uint16) (*LinkedTask, error) {
	filename, err := os.Readlink(filepath.Join(dir, formatID(id)+".task"))
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(dir, filename)
	}
	t, err := newTaskFromFile(filename)
	if err != nil {
		return nil, err
	}
	lt := LinkedTask{
		LinkID: id,
		Task:   t,
	}
	return &lt, nil
}
