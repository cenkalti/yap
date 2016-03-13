package task

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

var (
	Home              string
	dirTasks          string
	dirPendingTasks   string
	dirCompletedTasks string
)

func init() {
	Home = os.Getenv("YAP_HOME")
	if Home == "" {
		Home = "~/.yap"
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
