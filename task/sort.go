package task

type byCreatedAtDesc []Task

func (t byCreatedAtDesc) Len() int           { return len(t) }
func (t byCreatedAtDesc) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byCreatedAtDesc) Less(i, j int) bool { return t[i].CreatedAt.After(t[j].CreatedAt) }
