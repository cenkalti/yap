package main

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const DefaultYapHome = "~/.yap"

var (
	YapHome           string
	DirTasks          string
	DirPendingTasks   string
	DirCompletedTasks string
)

func SetHome(home string) (err error) {
	if YapHome, err = homedir.Expand(home); err != nil {
		return
	}
	DirTasks = filepath.Join(YapHome, "tasks")
	if err = os.MkdirAll(DirTasks, 0700); err != nil {
		return
	}
	DirPendingTasks = filepath.Join(YapHome, "pending-tasks")
	if err = os.MkdirAll(DirPendingTasks, 0700); err != nil {
		return
	}
	DirCompletedTasks = filepath.Join(YapHome, "completed-tasks")
	if err = os.MkdirAll(DirCompletedTasks, 0700); err != nil {
		return
	}
	return
}
