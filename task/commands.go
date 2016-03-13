package task

import (
	"sort"
	"time"

	"github.com/satori/go.uuid"
)

// Add new task in pending state.
func Add(title string) (id uint16, err error) {
	id, err = nextID(dirPendingTasks)
	if err != nil {
		return
	}
	pt := PendingTask{
		ID: id,
		Task: Task{
			UUID:      uuid.NewV4(),
			Title:     title,
			CreatedAt: time.Now(),
		},
	}
	if err = pt.write(); err != nil {
		return
	}
	lt := &linkedTask{
		LinkID: pt.ID,
		Task:   pt.Task,
	}
	err = lt.link(dirPendingTasks)
	return
}

// ListPending returns all pending tasks.
func ListPending() ([]PendingTask, error) {
	tasks, err := pendingTasks()
	if err != nil {
		return nil, err
	}
	sort.Sort(pendingTasksByCreatedAtDesc(tasks))
	return tasks, nil
}

// ListCompleted returns all completed tasks.
func ListCompleted() ([]CompletedTask, error) {
	tasks, err := completedTasks()
	if err != nil {
		return nil, err
	}
	sort.Sort(completedTasksByCompletedAtDesc(tasks))
	return tasks, nil
}

// Complete pending task.
func Complete(id uint16) error {
	t, err := getPendingTask(id)
	if err != nil {
		return err
	}
	return t.complete()
}

// Continue completed task.
func Continue(id uint16) error {
	t, err := getCompletedTask(id)
	if err != nil {
		return err
	}
	return t.continueTask()
}
