package drbd

import (
	"testing"

	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDRBDStatus1(t *testing.T) {
	defer filet.CleanUp(t)
	dir := filet.TmpDir(t, "")
	procDRBD := dir + "/proc-drbd"

	got, err := getStates(procDRBD)
	require.NoError(t, err, "empty file")
	assert.Empty(t, got, "states for empty file")

	writeFile(t, procDRBD, exampleProcDRBD1)
	got, err = getStates(procDRBD)
	require.NoError(t, err, "exampleProcDRBD1")
	if assert.Contains(t, got, 0, "states for exampleProcDRBD1") {
		s := got[0]
		assert.Equal(t, "WFConnection", s.Connection, "connection")
		assert.Equal(t, "Secondary", s.SelfRole, "self role")
		assert.Equal(t, "Unknown", s.RemoteRole, "remote role")
		assert.Equal(t, "UpToDate", s.SelfDisk, "self disk")
		assert.Equal(t, "DUnknown", s.RemoteDisk, "remote disk")
	}

	writeFile(t, procDRBD, exampleProcDRBD2)
	got, err = getStates(procDRBD)
	require.NoError(t, err, "exampleProcDRBD2")
	if assert.Containsf(t, got, 0, "states for exampleProcDRBD2: %v", got) {
		s := got[0]
		assert.Equal(t, "SyncSource", s.Connection, "connection")
		assert.Equal(t, "Primary", s.SelfRole, "self role")
		assert.Equal(t, "Secondary", s.RemoteRole, "remote role")
		assert.Equal(t, "UpToDate", s.SelfDisk, "self disk")
		assert.Equal(t, "Inconsistent", s.RemoteDisk, "remote disk")
	}
	if assert.Containsf(t, got, 2, "states for exampleProcDRBD2: %v", got) {
		s := got[2]
		assert.Equal(t, "SyncSource", s.Connection, "connection")
		assert.Equal(t, "Primary", s.SelfRole, "self role")
		assert.Equal(t, "Secondary", s.RemoteRole, "remote role")
		assert.Equal(t, "UpToDate", s.SelfDisk, "self disk")
		assert.Equal(t, "Inconsistent", s.RemoteDisk, "remote disk")
	}
}
