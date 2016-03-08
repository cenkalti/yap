package task

import (
	"sort"
	"time"
)

// Add new task in pending state.
func Add(title string) (*PendingTask, error) {
	sid, err := nextID(dirPendingTasks)
	if err != nil {
		return nil, err
	}
	pt := PendingTask{
		LinkedTask: LinkedTask{
			LinkID: sid,
			Task: Task{
				ID:        randomID(),
				Title:     title,
				CreatedAt: time.Now(),
			},
		},
	}
	if err = pt.write(); err != nil {
		return nil, err
	}
	return &pt, pt.link(dirPendingTasks)
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

// Complete pending task.
func Complete(id uint32) error {
	t, err := getPendingTask(id)
	if err != nil {
		return err
	}
	return t.Complete()
}
