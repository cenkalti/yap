package task

import "sort"

type lessFunc func(p1, p2 Task) bool

// multiSorter implements the Sort interface, sorting the tasks within.
type multiSorter struct {
	tasks []Task
	less  []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to orderedBy.
func (ms *multiSorter) Sort(tasks []Task) {
	ms.tasks = tasks
	sort.Sort(ms)
}

// orderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func orderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.tasks)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.tasks[i], ms.tasks[j] = ms.tasks[j], ms.tasks[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.tasks[i], ms.tasks[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

func increasingCreatedAt(t1, t2 Task) bool {
	return t1.CreatedAt.Before(t2.CreatedAt)
}

func decreasingCompletedAt(t1, t2 Task) bool {
	if t1.CompletedAt == nil || t2.CompletedAt == nil {
		return false
	}
	return t1.CompletedAt.After(*t2.CompletedAt)
}

func increasingWaitDate(t1, t2 Task) bool {
	if t1.WaitDate == nil || t2.WaitDate == nil {
		return false
	}
	return t1.WaitDate.Before(t2.WaitDate.Time)
}

func increasingDueDate(t1, t2 Task) bool {
	if t1.DueDate == nil {
		return false
	}
	if t2.DueDate == nil {
		return true
	}
	return t1.DueDate.Before(t2.DueDate.Time)
}
