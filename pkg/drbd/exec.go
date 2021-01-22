package drbd

import (
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// RunCommandOnChange will watch /proc/drbd for changes.
// When there is a change, it will call command with the
// following arguments:
//	command connectedState selfRole remoteRole selfDisk remoteDisk mountPoint
// So for example, if /proc/drbd changes and there is line like:
// 	0: cs:WFConnection ro:Secondary/Unknown ds:UpToDate/DUnknown C r-----
// and /etc/fstab has a line like
//	/dev/drbd0 /my/file/system ext4 rw,noauto 0 0
// then command will be exec'ed with:
//	command	r0 WFConnection Secondary Unknown UpToDate DUnknown /my/file/system
// In addition, the following environment variables will be set
//	OLD_CONNECTED_STATE="WFConnection" # prior connection status"
//	OLD_SELF_ROLE="Primary" # prior self role
//	OLD_REMOTE_ROLE="Unknown" # prior remote role
//	OLD_SELF_DISK="Secondary" # prior self disk state
//	OLD_REMOTE_DISK="DUnknown" # prior remote disk sate
//	ALL_MOUNTS="/my/file/system" # all filesystems, mounted or not that mount on /dev/drbd0
//	STABLE_SECONDS="9999" # seconds since last change in this resource state
//
// nap is how long to wait between checking for changes in state
// if bailOnError is true, then errors returned by commands or parsing /etc/fstab
// will cause RunCommandOnChange to return.
func RunCommandOnChange(nap time.Duration, bailOnError bool, command []string) error {
	return runCommandOnChange("/proc/drbd", nap, bailOnError, command, "/etc/fstab", "/proc/mounts")
}

func runCommandOnChange(filename string, nap time.Duration, bailOnError bool, command []string, fstab string, procMounts string) error {
	if len(command) == 0 {
		return errors.New("a command is required")
	}
	return Invoke(filename, nap, func(delta Delta) error {
		fsMounts, err := GetMounts(delta.Resource, fstab)
		var mountPoint string
		if err != nil {
			if bailOnError {
				return err
			}
			log.Printf("Could not read %s: err\n", fstab)
		}
		if len(fsMounts) > 0 {
			sort.Strings(fsMounts)
			mountPoint = fsMounts[0]
		}
		liveMounts, _ := GetMounts(delta.Resource, procMounts)
		allMounts := make(map[string]struct{})
		for _, m := range fsMounts {
			allMounts[m] = struct{}{}
		}
		for _, m := range liveMounts {
			allMounts[m] = struct{}{}
		}
		mountList := make([]string, 0, len(allMounts))
		for m := range allMounts {
			mountList = append(mountList, m)
		}
		sort.Strings(mountList)

		args := make([]string, len(command)-1, len(command)+6)
		copy(args, command[1:])
		args = append(args,
			"r"+strconv.Itoa(delta.Resource),
			delta.New.Connection,
			delta.New.SelfRole,
			delta.New.RemoteRole,
			delta.New.SelfDisk,
			delta.New.RemoteDisk,
			mountPoint,
		)

		cmd := exec.Command(command[0], args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env,
			"OLD_CONNECTED_STATE="+delta.Old.Connection,
			"OLD_SELF_ROLE="+delta.Old.SelfRole,
			"OLD_SELF_DISK="+delta.Old.SelfDisk,
			"OLD_REMOTE_ROLE="+delta.Old.RemoteRole,
			"OLD_REMOTE_DISK="+delta.Old.RemoteDisk,
			"ALL_MOUNTS="+strings.Join(mountList, " "),
			"STABLE_SECONDS="+strconv.Itoa(int(delta.UnchangedFor.Seconds())),
		)
		err = cmd.Run()
		if err != nil {
			log.Printf("exec %s failed: %s", cmd.String(), err)
			if bailOnError {
				return err
			}
		}
		return nil
	})
}
