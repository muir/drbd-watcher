package drbd

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envValue(t *testing.T, env map[string]string, key string, want string) {
	if assert.Contains(t, env, key, "environment variable") {
		assert.Equal(t, want, env[key], "environment variable")
	}
}

func remove(t *testing.T, name string) {
	assert.NoErrorf(t, os.Remove(name), "remove %s", name)
}

func noFile(t *testing.T, name string, nap time.Duration) {
	time.Sleep(nap)
	_, err := os.Stat(name)
	require.Error(t, err, "expecting error")
	require.Equal(t, syscall.ENOENT, err.(*os.PathError).Err, "expecting file not to exist")
}

func waitForFile(t *testing.T, name string, nap time.Duration) {
	for i := 0; i < 10; i++ {
		time.Sleep(nap / 10)
		_, err := os.Stat(name)
		if err == nil {
			t.Logf("file %s appeared after %d iterations", name, i)
			return
		}
	}
	t.Fatalf("file %s did not appear after waiting %s", name, nap)
}

func readOutput(t *testing.T, name string) (string, map[string]string) {
	fh, err := os.Open(name)
	require.NoErrorf(t, err, "open %s", name)
	defer fh.Close()
	scanner := bufio.NewScanner(fh)
	assert.Truef(t, scanner.Scan(), "first line of %s", name)
	args := scanner.Text()
	env := make(map[string]string)
	for scanner.Scan() {
		e := scanner.Text()
		kv := strings.SplitN(e, "=", 2)
		require.Lenf(t, kv, 2, "line from %s, '%s', not key-value pair", name, e)
		env[kv[0]] = kv[1]
	}
	return args, env
}

func writeFile(t *testing.T, name string, contents string) {
	f, err := os.Create(name + ".tmp")
	require.NoErrorf(t, err, "create %s.tmp", name)
	b := ([]byte)(contents)
	written, err := f.WriteAt(b, 0)
	require.NoErrorf(t, err, "write to %s.tmp", name)
	for written < len(b) {
		more, err := f.Write(b[written:])
		require.NoErrorf(t, err, "write more to %s.tmp", name)
		written += more
	}
	require.NoErrorf(t, f.Close(), "close write to %s.tmp", name)
	require.NoErrorf(t, os.Rename(name+".tmp", name), "rename %s.tmp -> %s", name, name)
}

const exampleFstab = `UUID=65429799-d704-460d-b471-e5f04f64a221 / ext4 defaults 0 0
/dev/drbd0  /r0 btrfs noauto,rw,relatime,space_cache,subvolid=5,subvol=/,ssd 0 0
`

const exampleProcDRBD1 = `version: 8.4.10 (api:1/proto:86-101)
srcversion: 15111D056BF899E7D986DDD 
 0: cs:WFConnection ro:Secondary/Unknown ds:UpToDate/DUnknown C r-----
    ns:0 nr:0 dw:0 dr:0 al:0 bm:0 lo:0 pe:0 ua:0 ap:0 ep:1 wo:f oos:2649072
`

const exampleProcDRBD2 = `version: 8.4.11 (api:1/proto:86-101)
srcversion: FC3433D849E3B88C1E7B55C
 0: cs:SyncSource ro:Primary/Secondary ds:UpToDate/Inconsistent C r-----
    ns:56582812 nr:0 dw:156299240 dr:71719920 al:1524 bm:0 lo:0 pe:0 ua:0 ap:0 ep:1 wo:f oos:106597444
        [=====>..............] sync'ed: 34.7% (104096/159172)M
        finish: 0:53:57 speed: 32,924 (30,140) K/sec
 1: cs:SyncSource ro:Primary/Secondary ds:UpToDate/Inconsistent C r-----
    ns:11974408 nr:0 dw:540488000 dr:13652852 al:3316 bm:0 lo:1 pe:0 ua:0 ap:0 ep:1 wo:f oos:512191896
        [>....................] sync'ed:  2.3% (500184/511704)M
        finish: 12:01:41 speed: 11,812 (11,052) K/sec
 2: cs:SyncSource ro:Primary/Secondary ds:UpToDate/Inconsistent C r-----
    ns:2740340 nr:0 dw:1310467600 dr:4758343 al:24 bm:0 lo:1 pe:0 ua:0 ap:0 ep:1 wo:f oos:1433206068
        [>....................] sync'ed:  0.2% (1399612/1402120)M
        finish: 181:44:49 speed: 2,180 (2,404) K/sec
`

const exampleProcMounts = `sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
udev /dev devtmpfs rw,nosuid,relatime,size=4052576k,nr_inodes=1013144,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,noexec,relatime,size=816804k,mode=755 0 0
/dev/sda2 / ext4 rw,relatime,data=ordered 0 0
securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev 0 0
tmpfs /run/lock tmpfs rw,nosuid,nodev,noexec,relatime,size=5120k 0 0
tmpfs /sys/fs/cgroup tmpfs ro,nosuid,nodev,noexec,mode=755 0 0
cgroup /sys/fs/cgroup/unified cgroup2 rw,nosuid,nodev,noexec,relatime 0 0
cgroup /sys/fs/cgroup/systemd cgroup rw,nosuid,nodev,noexec,relatime,xattr,name=systemd 0 0
pstore /sys/fs/pstore pstore rw,nosuid,nodev,noexec,relatime 0 0
cgroup /sys/fs/cgroup/pids cgroup rw,nosuid,nodev,noexec,relatime,pids 0 0
cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/hugetlb cgroup rw,nosuid,nodev,noexec,relatime,hugetlb 0 0
cgroup /sys/fs/cgroup/perf_event cgroup rw,nosuid,nodev,noexec,relatime,perf_event 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
cgroup /sys/fs/cgroup/rdma cgroup rw,nosuid,nodev,noexec,relatime,rdma 0 0
cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
cgroup /sys/fs/cgroup/net_cls,net_prio cgroup rw,nosuid,nodev,noexec,relatime,net_cls,net_prio 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/cpuset cgroup rw,nosuid,nodev,noexec,relatime,cpuset 0 0
systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=26,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=14394 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
mqueue /dev/mqueue mqueue rw,relatime 0 0
hugetlbfs /dev/hugepages hugetlbfs rw,relatime,pagesize=2M 0 0
configfs /sys/kernel/config configfs rw,relatime 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0
/dev/loop0 /snap/go/5243 squashfs ro,nodev,relatime 0 0
/dev/loop2 /snap/core/8592 squashfs ro,nodev,relatime 0 0
lxcfs /var/lib/lxcfs fuse.lxcfs rw,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other 0 0
/dev/loop3 /snap/go/5364 squashfs ro,nodev,relatime 0 0
/dev/loop4 /snap/core/8689 squashfs ro,nodev,relatime 0 0
tmpfs /run/user/1000 tmpfs rw,nosuid,nodev,relatime,size=816800k,mode=700,uid=1000,gid=1000 0 0`
