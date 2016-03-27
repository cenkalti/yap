package task

type tasks []Task

func (t tasks) Len() int      { return len(t) }
func (t tasks) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

type byCreatedAtDesc struct{ tasks }

func (t byCreatedAtDesc) Less(i, j int) bool { return t.tasks[i].CreatedAt.After(t.tasks[j].CreatedAt) }

type byCompletedAtDesc struct{ tasks }

func (t byCompletedAtDesc) Less(i, j int) bool {
	ti := t.tasks[i].CompletedAt
	tj := t.tasks[j].CompletedAt
	if ti == nil || tj == nil {
		return false
	}
	return ti.After(*tj)
}

type byWaitDateAsc struct{ tasks }

func (t byWaitDateAsc) Less(i, j int) bool {
	ti := t.tasks[i].WaitDate
	tj := t.tasks[j].WaitDate
	if ti == nil || tj == nil {
		return false
	}
	return ti.Before(*tj)
}
