package internalstate

// Taken from the Golang docs and modified to be used here.
// https://golang.org/pkg/sort/#example__sortKeys

import "sort"

// A Run carrys the time and guid in a light weight sortable form.
type Run struct {
	guid string
	time int64
}

// By is the type of a "less" function that defines the ordering of its Run arguments.
type By func(p1, p2 *Run) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(runs []Run) {
	ps := &runSorter{
		runs: runs,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// runSorter joins a By function and a slice of Runs to be sorted.
type runSorter struct {
	runs []Run
	by   func(run1, run2 *Run) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *runSorter) Len() int {
	return len(s.runs)
}

// Swap is part of sort.Interface.
func (s *runSorter) Swap(i, j int) {
	s.runs[i], s.runs[j] = s.runs[j], s.runs[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *runSorter) Less(i, j int) bool {
	return s.by(&s.runs[i], &s.runs[j])
}
