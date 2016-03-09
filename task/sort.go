package task

type pendingTasksByCreatedAtDesc []PendingTask

func (t pendingTasksByCreatedAtDesc) Len() int           { return len(t) }
func (t pendingTasksByCreatedAtDesc) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t pendingTasksByCreatedAtDesc) Less(i, j int) bool { return t[i].CreatedAt.After(t[j].CreatedAt) }

type completedTasksByCompletedAtDesc []CompletedTask

func (t completedTasksByCompletedAtDesc) Len() int      { return len(t) }
func (t completedTasksByCompletedAtDesc) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t completedTasksByCompletedAtDesc) Less(i, j int) bool {
	ti := t[i].CompletedAt
	tj := t[j].CompletedAt
	if ti == nil || tj == nil {
		return false
	}
	return ti.After(*tj)
}
