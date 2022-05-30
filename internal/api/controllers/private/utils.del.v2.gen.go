// Package private - generated by fungen; DO NOT EDIT
package private

import "sync"

// CancelInputV2List is the type for a list that holds members of type CancelInputV2
type CancelInputV2List []CancelInputV2

// PMapRunCanceled is similar to MapRunCanceled except that it executes the function on each member in parallel.
func (l CancelInputV2List) PMapRunCanceled(f func(CancelInputV2) *RunCanceled) RunCanceledList {
	wg := sync.WaitGroup{}
	l2 := make(RunCanceledList, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t CancelInputV2) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// PMap is similar to Map except that it executes the function on each member in parallel.
func (l CancelInputV2List) PMap(f func(CancelInputV2) CancelInputV2) CancelInputV2List {
	wg := sync.WaitGroup{}
	l2 := make(CancelInputV2List, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t CancelInputV2) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// RunCanceledList is the type for a list that holds members of type *RunCanceled
type RunCanceledList []*RunCanceled

// PMapCancelInputV2 is similar to MapCancelInputV2 except that it executes the function on each member in parallel.
func (l RunCanceledList) PMapCancelInputV2(f func(*RunCanceled) CancelInputV2) CancelInputV2List {
	wg := sync.WaitGroup{}
	l2 := make(CancelInputV2List, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t *RunCanceled) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}

// PMap is similar to Map except that it executes the function on each member in parallel.
func (l RunCanceledList) PMap(f func(*RunCanceled) *RunCanceled) RunCanceledList {
	wg := sync.WaitGroup{}
	l2 := make(RunCanceledList, len(l))
	for i, t := range l {
		wg.Add(1)
		go func(i int, t *RunCanceled) {
			l2[i] = f(t)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
	return l2
}
