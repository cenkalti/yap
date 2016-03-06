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
	t := PendingTask{
		linkedTask{
			SmallID: sid,
			Task: Task{
				ID:        randomID(),
				Title:     title,
				CreatedAt: time.Now(),
			},
		},
	}
	if err = t.Task.write(); err != nil {
		return err
	}
	return t.link(dirPendingTasks)
}

// List all pending tasks.
func List() ([]Task, error) {
	tasks, err := allTasks()
	if err != nil {
		return nil, err
	}
	sort.Sort(byCreatedAtDesc(tasks))
	return tasks, nil
}
