package drbd

import (
	"testing"

	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDRBDStatus(t *testing.T) {
	defer filet.CleanUp(t)
	dir := filet.TmpDir(t, "")
	procDRBD := dir + "/proc-drbd"
	got, err := getStates(procDRBD)
	require.NoError(t, err, "empty file")
	assert.Empty(t, got, "states for empty file")
	writeFile(t, procDRBD, exampleProcDRBD1)
	got, err = getStates(procDRBD)
	require.NoError(t, err, "empty file")
	if assert.Contains(t, got, 0, "states for non-empty file") {
		s := got[0]
		assert.Equal(t, "WFConnection", s.Connection, "connection")
		assert.Equal(t, "Secondary", s.SelfRole, "self role")
		assert.Equal(t, "Unknown", s.RemoteRole, "remote role")
		assert.Equal(t, "UpToDate", s.SelfDisk, "self disk")
		assert.Equal(t, "DUnknown", s.RemoteDisk, "remote disk")
	}
}
