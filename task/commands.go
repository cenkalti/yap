package task

import (
	"sort"
	"time"
)

// Add new task in pending state.
func Add(title string) error {
	sid, err := nextID(dirPendingTasks)
	if err != nil {
		return err
	}
	t := LinkedTask{
		LinkID: sid,
		Task: Task{
			ID:        randomID(),
			Title:     title,
			CreatedAt: time.Now(),
		},
	}
	if err = t.Task.write(); err != nil {
		return err
	}
	return t.link(dirPendingTasks)
}

// List all pending tasks.
func List() ([]PendingTask, error) {
	tasks, err := pendingTasks()
	if err != nil {
		return nil, err
	}
	sort.Sort(pendingTasksByCreatedAtDesc(tasks))
	return tasks, nil
}
