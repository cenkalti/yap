package task

import (
	"os"
	"path/filepath"
	"strings"
)

// linkedTask is a symlink to a Task for refering tasks with a more human-friendly ID number.
// Task IDs are random 16-bit integers that is hard to remember and type.
// linkedTasks have separate IDs that is usually a small number.
type linkedTask struct {
	LinkID uint16
	Task
}

// tasksIn returns all tasks in dir.
func tasksIn(dir string) ([]linkedTask, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer checkClose(f)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []linkedTask
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
func (t linkedTask) link(dir string) error {
	src := filepath.Join("..", "tasks", t.UUID.String()+taskExt)
	dst := filepath.Join(dir, formatID(t.LinkID)+taskExt)
	return os.Symlink(src, dst)
}

// unlink removes the symlink in dir.
func (t linkedTask) unlink(dir string) error {
	dst := filepath.Join(dir, formatID(t.LinkID)+taskExt)
	return os.Remove(dst)
}

func (t *linkedTask) move(olddir, newdir string) error {
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

func getLinkedTask(dir string, id uint16) (*linkedTask, error) {
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
	lt := linkedTask{
		LinkID: id,
		Task:   t,
	}
	return &lt, nil
}
