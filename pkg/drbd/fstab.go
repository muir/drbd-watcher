package drbd

import (
	"strconv"

	fstab "github.com/d-tux/go-fstab"
)

// GetMounts find the mounted and unmounted mount points for
// a drbd resource if called with "/etc/fstab" and "/proc/mounts"
func GetMounts(resource int, files ...string) ([]string, error) {
	mountPoints := make(map[string]struct{})
	for _, file := range files {
		mounts, err := fstab.ParseFile(file)
		if err != nil {
			return nil, err
		}
		for _, mount := range mounts {
			if mount.Spec == "/dev/drbd"+strconv.Itoa(resource) {
				mountPoints[mount.File] = struct{}{}
			}
		}
	}
	ret := make([]string, 0, len(mountPoints))
	for mp := range mountPoints {
		ret = append(ret, mp)
	}
	return ret, nil
}
