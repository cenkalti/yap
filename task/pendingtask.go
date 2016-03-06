package task

// PendingTask is a Task that is not completed yet.
type PendingTask struct {
	LinkedTask
}

// Complete the task.
func (t *PendingTask) Complete() error {
	err := t.LinkedTask.link(dirCompletedTasks)
	if err != nil {
		return err
	}
	return t.LinkedTask.unlink(dirPendingTasks)
}

func pendingTasks() ([]PendingTask, error) {
	linkedTasks, err := tasksIn(dirPendingTasks)
	if err != nil {
		return nil, err
	}
	pendingTasks := make([]PendingTask, 0, len(linkedTasks))
	for _, lt := range linkedTasks {
		t := PendingTask{
			LinkedTask: lt,
		}
		pendingTasks = append(pendingTasks, t)
	}
	return pendingTasks, nil
}
