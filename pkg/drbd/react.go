package drbd

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Delta struct {
	Resource     int
	Old          State
	New          State
	UnchangedFor time.Duration
}

// React watches /proc/drbd and when there has been a change,
// it invokes callback() asynchronously.
// filename is presumed to be "/proc/drbd" -- it's a parameter for testing
// purposes.  React does not return except if there is an error.
func React(filename string, nap time.Duration, callback func(Delta) error) error {
	states := make(States)
	for {
		b := time.Now()
		before, after, err := Watch(filename, states, nap)
		a := time.Now()
		if err != nil {
			return err
		}
		for r, state := range after {
			log.Printf("r0 changed state: %s\n", StateDiff(state, before[r]))
			go callback(Delta{
				Resource:     r,
				Old:          before[r],
				New:          state,
				UnchangedFor: a.Sub(b),
			})
			states[r] = state
		}
	}
}

func StateDiff(current, prior State) string {
	d := make([]string, 0, 3)
	ck := func(name, c, p string) {
		if c != p {
			d = append(d, fmt.Sprintf("%s:%s->%s", name, p, c))
		}
	}
	ck("Connection", current.Connection, prior.Connection)
	ck("Role", current.SelfRole+"/"+current.RemoteRole, prior.SelfRole+"/"+prior.RemoteRole)
	ck("Disk", current.SelfDisk+"/"+current.RemoteDisk, prior.SelfDisk+"/"+prior.RemoteDisk)
	if len(d) == 0 {
		return "no changes"
	}
	return strings.Join(d, "; ")
}
