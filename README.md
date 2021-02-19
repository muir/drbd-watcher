
# DRBD Watcher

DRBD Watcher is a small program that will watch /proc/drbd and invoke
another program whenver /proc/drbd changes.

## Installing on Ubuntu

```bash
sudo snap install go --classic
GOPATH=${GOPATH:-$HOME}
GOBIN=${GOBIN:-$GOPATH/bin}
mkdir -p $GOPATH/src/github.com/muir
cd $GOPATH/src/github.com/muir
git clone https://github.com/muir/drbd-watcher.git
cd drbd-watcher
/snap/bin/go install ./...
sudo cp $GOBIN/drbd-watcher /usr/local/bin/
```

## Running the watcher

The watcher isn't much use without a program to invoke upon change.

The program will be executed with the following arguments in addition to whatever
is specified when invoking the watcher:

	resource (eg "r0")
	connection state (eg "WFConnection")
	role, self (eg "Secondary")
	role, remote (eg "Unknown")
	disk state, self (eg "UpToDate")
	disk state, remote (eg "UpToDate")
	mount point (eg "/my-shared-disk", if kwown)

In addition the following environment variables will be set:

	OLD_CONNECTED_STATE="WFConnection" # prior connection status"
	OLD_SELF_ROLE="Primary" # prior self role
	OLD_REMOTE_ROLE="Unknown" # prior remote role
	OLD_SELF_DISK="Secondary" # prior self disk state
	OLD_REMOTE_DISK="DUnknown" # prior remote disk sate
	ALL_MOUNTS="/my/file/system" # all filesystems, mounted or not that mount on /dev/drbd0
	STABLE_SECONDS="9999" # seconds since last change in this resource state

If writing a shell script, a reasonable start is:

	#!/bin/bash

	resource=$1
	connection=$2
	self_role=$3
	remote_role=$4
	self_disk=$5
	remote_disk=$6
	filesystem=$7

