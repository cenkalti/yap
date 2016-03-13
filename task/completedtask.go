package task

// CompletedTask is a Task that is completed.
// Completed tasks are stored in completed-tasks dir as symlink to original file in tasks dir.
type CompletedTask struct {
	ID uint16
	Task
}

// continue completed task.
func (t *CompletedTask) continueTask() error {
	lt := &linkedTask{
		LinkID: t.ID,
		Task:   t.Task,
	}
	return lt.move(dirCompletedTasks, dirPendingTasks)
}

func completedTasks() ([]CompletedTask, error) {
	linkedTasks, err := tasksIn(dirCompletedTasks)
	if err != nil {
		return nil, err
	}
	completedTasks := make([]CompletedTask, 0, len(linkedTasks))
	for _, lt := range linkedTasks {
		t := CompletedTask{
			ID:   lt.LinkID,
			Task: lt.Task,
		}
		completedTasks = append(completedTasks, t)
	}
	return completedTasks, nil
}

func getCompletedTask(id uint16) (*CompletedTask, error) {
	lt, err := getLinkedTask(dirCompletedTasks, id)
	if err != nil {
		return nil, err
	}
	t := CompletedTask{
		ID:   lt.LinkID,
		Task: lt.Task,
	}
	return &t, nil
}
