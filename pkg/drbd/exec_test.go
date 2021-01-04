package drbd

import (
	"os"
	"testing"
	"time"

	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const napTime = 100 * time.Millisecond

func TestTransitions(t *testing.T) {
	defer filet.CleanUp(t)

	cwd, err := os.Getwd()
	require.NoError(t, err, "getwd")

	dir := filet.TmpDir(t, "")

	procDRBD := dir + "/proc-drbd"

	fstab := dir + "/etc-fstab"
	writeFile(t, fstab, exampleFstab)

	procMounts := dir + "/proc-mounts"
	writeFile(t, procMounts, exampleProcMounts)

	shellOut := dir + "/shell.env"

	require.NoError(t, os.Setenv("DRBD_TEST_OUTPUT", shellOut))

	go runCommandOnChange(procDRBD, napTime/20, true, []string{cwd + "/test.sh", "foo"}, fstab, procMounts)

	noFile(t, shellOut, napTime)

	writeFile(t, procDRBD, exampleProcDRBD1)

	waitForFile(t, shellOut, napTime)
	firstLine, env := readOutput(t, shellOut)

	assert.Equal(t, "foo r0 WFConnection Secondary Unknown UpToDate DUnknown /r0", firstLine, "summary line")
	envValue(t, env, "OLD_CONNECTED_STATE", "")

	remove(t, shellOut)
	noFile(t, shellOut, napTime)

}
