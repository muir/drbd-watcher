package drbd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// This is an example /proc/drbd state
// version: 8.4.10 (api:1/proto:86-101)
// srcversion: 15111D056BF899E7D986DDD
//  0: cs:WFConnection ro:Secondary/Unknown ds:UpToDate/DUnknown C r-----
//     ns:0 nr:0 dw:0 dr:0 al:0 bm:0 lo:0 pe:0 ua:0 ap:0 ep:1 wo:f oos:2649072

// State tracks the DRBD state for a resource
type State struct {
	Connection string
	SelfRole   string
	RemoteRole string
	SelfDisk   string
	RemoteDisk string
}

func (s State) Equal(o State) bool {
	return s.Connection == o.Connection &&
		s.SelfRole == o.SelfRole &&
		s.RemoteRole == o.RemoteRole &&
		s.SelfDisk == o.SelfDisk &&
		s.RemoteDisk == o.RemoteDisk
}

// Maps resource number to resource state
type States map[int]State

// Watch returns once the DRBD state changes.
// It returns only the subset of the state that has changed,
// both the old and new.
// filename is presumed to be "/proc/drbd" -- it's a parameter for testing
// purposes.
// Old and new values will be returned together (no missing keys).
// The oldStates passed in allow Watch to be initialized with pre-existing
// expectations
func Watch(filename string, oldStates States, nap time.Duration) (States, States, error) {
	oldValues := make(States)
	newValues := make(States)
	for {
		newStates, err := getStates(filename)
		if err != nil {
			return nil, nil, err
		}
		for r, state := range oldStates {
			if n, ok := newStates[r]; ok {
				if !n.Equal(state) {
					oldValues[r] = state
					newValues[r] = n
				}
			} else {
				oldValues[r] = state
				newValues[r] = State{}
			}
		}
		for r, state := range newStates {
			if _, ok := oldStates[r]; ok {
				continue
			}
			oldValues[r] = State{}
			newValues[r] = state
		}
		if len(oldValues) > 0 || len(newValues) > 0 {
			return oldValues, newValues, nil
		}
		oldStates = newStates
		time.Sleep(nap)
	}
}

var re = regexp.MustCompile(`^ (\d+): cs:(\S+) ro:(\S+?)/(\S+) ds:(\S+)/(\S+) `)
var skipRE = regexp.MustCompile(`^\s+(?:ns:\d+ |\[\=*\>\.*\] sync'ed|finish: \d)`)

func getStates(filename string) (States, error) {
	fh, err := os.Open(filename)
	if err != nil {
		// if DRBD isn't running /proc/drbd won't exist and open will fail
		switch e := err.(type) {
		case *os.PathError:
			if e.Err == syscall.ENOENT {
				return nil, nil
			}
		default:
			return nil, errors.Wrapf(err, "open %s", filename)
		}
	}
	defer fh.Close()
	scanner := bufio.NewScanner(fh)
	// skip over version number
	if !scanner.Scan() {
		t := scanner.Text()
		if !strings.HasPrefix(t, "version") {
			return nil, fmt.Errorf("%s did not start with a version string. Found: '%s'\n",
				filename, t)
		}
	}
	// skip over version checksum
	if !scanner.Scan() {
		return nil, fmt.Errorf("%s ended early", filename)
	}
	s := make(States)
	for scanner.Scan() {
		t := scanner.Text()
		if m := re.FindStringSubmatch(t); len(m) != 0 {
			r, _ := strconv.Atoi(m[1])
			s[r] = State{
				Connection: m[2],
				SelfRole:   m[3],
				RemoteRole: m[4],
				SelfDisk:   m[5],
				RemoteDisk: m[6],
			}
			continue
		}
		if skipRE.MatchString(t) {
			continue
		}
		if t == "" {
			continue
		}
		return nil, fmt.Errorf("Unexpected %s output line:\n%s\n", filename, t)
	}
	return s, nil
}
