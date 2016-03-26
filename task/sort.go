package task

type byCreatedAtDesc []Task

func (t byCreatedAtDesc) Len() int           { return len(t) }
func (t byCreatedAtDesc) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byCreatedAtDesc) Less(i, j int) bool { return t[i].CreatedAt.After(t[j].CreatedAt) }

type byCompletedAtDesc []Task

func (t byCompletedAtDesc) Len() int      { return len(t) }
func (t byCompletedAtDesc) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t byCompletedAtDesc) Less(i, j int) bool {
	ti := t[i].CompletedAt
	tj := t[j].CompletedAt
	if ti == nil || tj == nil {
		return false
	}
	return ti.After(*tj)
}

type byWaitDateAsc []Task

func (t byWaitDateAsc) Len() int      { return len(t) }
func (t byWaitDateAsc) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t byWaitDateAsc) Less(i, j int) bool {
	ti := t[i].WaitDate
	tj := t[j].WaitDate
	if ti == nil || tj == nil {
		return false
	}
	return ti.Before(*tj)
}
