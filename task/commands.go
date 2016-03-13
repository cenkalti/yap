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
func Complete(ids []uint16) error {
	tasks := make([]*PendingTask, 0, len(ids))
	for _, id := range ids {
		t, err := getPendingTask(id)
		if err != nil {
			return err
		}
		tasks = append(tasks, t)
	}
	for _, t := range tasks {
		err := t.complete()
		if err != nil {
			return err
		}
	}
	return nil
}

// Continue completed task.
func Continue(ids []uint16) error {
	tasks := make([]*CompletedTask, 0, len(ids))
	for _, id := range ids {
		t, err := getCompletedTask(id)
		if err != nil {
			return err
		}
		tasks = append(tasks, t)
	}
	for _, t := range tasks {
		err := t.continueTask()
		if err != nil {
			return err
		}
	}
	return nil
}
