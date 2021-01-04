package drbd

import (
	"testing"

	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMounts(t *testing.T) {
	defer filet.CleanUp(t)
	dir := filet.TmpDir(t, "")
	fstab := dir + "/etc-fstab"
	writeFile(t, fstab, exampleFstab)
	got, err := GetMounts(0, fstab)
	require.NoError(t, err, "fstab")
	assert.Equal(t, []string{"/r0"}, got, "list of mounts")
}
