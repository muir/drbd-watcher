package drbd

import (
	"sync"
	"time"
)

// Invoke watches /proc/drbd for changes.
// It invokes callback whenever there is a change.
// It will not invoke callback on the same resource
// until the prior callback for that resource has
// returned.
func Invoke(filename string, nap time.Duration, callback func(Delta) error) error {
	var mu sync.Mutex
	dataWaiting := make(map[int]Delta)
	currentlyRunning := make(map[int]struct{})
	echan := make(chan error, 1)
	var invoke func(delta Delta)
	invoke = func(delta Delta) {
		err := callback(delta)
		if err != nil {
			echan <- err
		}
		mu.Lock()
		defer mu.Unlock()
		if waiting, ok := dataWaiting[delta.Resource]; ok {
			delete(dataWaiting, delta.Resource)
			go invoke(waiting)
			return
		}
		delete(currentlyRunning, delta.Resource)
	}
	go func() {
		echan <- React(filename, nap, func(delta Delta) error {
			mu.Lock()
			defer mu.Unlock()
			if alreadyWaiting, ok := dataWaiting[delta.Resource]; ok {
				dataWaiting[delta.Resource] = Delta{
					Resource:     delta.Resource,
					Old:          alreadyWaiting.Old,
					New:          delta.New,
					UnchangedFor: delta.UnchangedFor,
				}
				return nil
			}
			if _, ok := currentlyRunning[delta.Resource]; ok {
				dataWaiting[delta.Resource] = delta
				return nil
			}
			currentlyRunning[delta.Resource] = struct{}{}
			go invoke(delta)
			return nil
		})
	}()
	return <-echan
}
