package task

// CompletedTask is a Task that is completed.
type CompletedTask struct {
	LinkedTask
}

// Continue completed task.
func (t *CompletedTask) Continue() error {
	return t.LinkedTask.move(dirCompletedTasks, dirPendingTasks)
}

func completedTasks() ([]CompletedTask, error) {
	linkedTasks, err := tasksIn(dirCompletedTasks)
	if err != nil {
		return nil, err
	}
	completedTasks := make([]CompletedTask, 0, len(linkedTasks))
	for _, lt := range linkedTasks {
		t := CompletedTask{
			LinkedTask: lt,
		}
		completedTasks = append(completedTasks, t)
	}
	return completedTasks, nil
}

func getCompletedTask(id uint32) (*CompletedTask, error) {
	lt, err := getLinkedTask(dirCompletedTasks, id)
	if err != nil {
		return nil, err
	}
	t := CompletedTask{
		LinkedTask: *lt,
	}
	return &t, nil
}
