package task

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

var (
	// Home is the directory for storing task files and settings.
	// Home can be overriden with YAP_HOME environment variable.
	Home              = "~/.yap"
	dirTasks          string
	dirPendingTasks   string
	dirCompletedTasks string
)

func init() {
	if yh := os.Getenv("YAP_HOME"); yh != "" {
		Home = yh
	}
	var err error
	Home, err = homedir.Expand(Home)
	if err != nil {
		log.Fatal(err)
	}
	dirTasks = filepath.Join(Home, "tasks")
	if err = os.MkdirAll(dirTasks, 0700); err != nil {
		log.Fatal(err)
	}
	dirPendingTasks = filepath.Join(Home, "pending-tasks")
	if err = os.MkdirAll(dirPendingTasks, 0700); err != nil {
		log.Fatal(err)
	}
	dirCompletedTasks = filepath.Join(Home, "completed-tasks")
	if err = os.MkdirAll(dirCompletedTasks, 0700); err != nil {
		log.Fatal(err)
	}
}
