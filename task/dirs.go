package task

import (
	"os"
	"path/filepath"
)

var (
	dirTasks          string
	dirPendingTasks   string
	dirCompletedTasks string
)

// SetHome creates task directories under yap home.
// User must call this function with yap home directory before using this package.
func SetHome(home string) (err error) {
	dirTasks = filepath.Join(home, "tasks")
	if err = os.MkdirAll(dirTasks, 0700); err != nil {
		return
	}
	dirPendingTasks = filepath.Join(home, "pending-tasks")
	if err = os.MkdirAll(dirPendingTasks, 0700); err != nil {
		return
	}
	dirCompletedTasks = filepath.Join(home, "completed-tasks")
	if err = os.MkdirAll(dirCompletedTasks, 0700); err != nil {
		return
	}
	return
}
