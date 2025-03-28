// Package private - generated by fungen; DO NOT EDIT
package private

import "sync"

// RunInputV2List is the type for a list that holds members of type RunInputV2
type RunInputV2List []RunInputV2

// PMap is similar to Map except that it executes the function on each member in parallel.
func (l RunInputV2List) PMap(f func(RunInputV2) RunInputV2) RunInputV2List {
	wg := sync.WaitGroup{}
	l2 := make(RunInputV2List, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t RunInputV2) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// PMapRunCreatedV2 is similar to MapRunCreatedV2 except that it executes the function on each member in parallel.
func (l RunInputV2List) PMapRunCreatedV2(f func(RunInputV2) *RunCreated) RunCreatedList {
	wg := sync.WaitGroup{}
	l2 := make(RunCreatedList, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t RunInputV2) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// RunCreatedV2List is the type for a list that holds members of type *RunCreated
type RunCreatedV2List []*RunCreated

// PMapRunInputV2 is similar to MapRunInputV2 except that it executes the function on each member in parallel.
func (l RunCreatedV2List) PMapRunInputV2(f func(*RunCreated) RunInputV2) RunInputV2List {
	wg := sync.WaitGroup{}
	l2 := make(RunInputV2List, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t *RunCreated) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// PMap is similar to Map except that it executes the function on each member in parallel.
func (l RunCreatedV2List) PMap(f func(*RunCreated) *RunCreated) RunCreatedV2List {
	wg := sync.WaitGroup{}
	l2 := make(RunCreatedV2List, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t *RunCreated) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}
