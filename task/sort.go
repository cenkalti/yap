package task

type pendingTasksByCreatedAtDesc []PendingTask

func (t pendingTasksByCreatedAtDesc) Len() int           { return len(t) }
func (t pendingTasksByCreatedAtDesc) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t pendingTasksByCreatedAtDesc) Less(i, j int) bool { return t[i].CreatedAt.After(t[j].CreatedAt) }
