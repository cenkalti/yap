package task

// PendingTask is a Task that is not completed yet.
type PendingTask struct {
	ID uint16
	Task
}

// Complete the task.
func (t *PendingTask) Complete() error {
	lt := &linkedTask{
		LinkID: t.ID,
		Task:   t.Task,
	}
	return lt.move(dirPendingTasks, dirCompletedTasks)
}

func pendingTasks() ([]PendingTask, error) {
	linkedTasks, err := tasksIn(dirPendingTasks)
	if err != nil {
		return nil, err
	}
	pendingTasks := make([]PendingTask, 0, len(linkedTasks))
	for _, lt := range linkedTasks {
		t := PendingTask{
			ID:   lt.LinkID,
			Task: lt.Task,
		}
		pendingTasks = append(pendingTasks, t)
	}
	return pendingTasks, nil
}

func getPendingTask(id uint16) (*PendingTask, error) {
	lt, err := getLinkedTask(dirPendingTasks, id)
	if err != nil {
		return nil, err
	}
	t := PendingTask{
		ID:   lt.LinkID,
		Task: lt.Task,
	}
	return &t, nil
}
