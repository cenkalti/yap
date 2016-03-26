package task

import (
	"os"
	"path/filepath"
	"strings"
)

// tasksIn returns all tasks in dir.
func tasksIn(dir string) ([]Task, error) {
	f, err := os.Open(dir)
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
		id, err := ParseID(name[:len(name)-len(taskExt)])
		if err != nil {
			return nil, err
		}
		t, err := readLink(dir, id)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// link writes a symlink to dir that is pointing to original task in dirTasks.
func (t *Task) link(dir string) error {
	id, err := nextID(dirPendingTasks)
	if err != nil {
		return err
	}
	src := filepath.Join("..", "tasks", t.UUID.String()+taskExt)
	dst := filepath.Join(dir, FormatID(id)+taskExt)
	err = os.Symlink(src, dst)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

// unlink removes the symlink in dir.
func (t *Task) unlink(dir string) error {
	dst := filepath.Join(dir, FormatID(t.ID)+taskExt)
	err := os.Remove(dst)
	if err != nil {
		return err
	}
	t.ID = 0
	return nil
}

func (t *Task) moveLink(olddir, newdir string) error {
	id, err := nextID(newdir)
	if err != nil {
		return err
	}
	oldpath := filepath.Join(olddir, FormatID(t.ID)+taskExt)
	newpath := filepath.Join(newdir, FormatID(id)+taskExt)
	err = os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func readLink(dir string, id uint16) (t Task, err error) {
	filename, err := os.Readlink(filepath.Join(dir, FormatID(id)+".task"))
	if err != nil {
		return
	}
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(dir, filename)
	}
	t, err = readFile(filename)
	if err != nil {
		return
	}
	t.ID = id
	return t, nil
}
