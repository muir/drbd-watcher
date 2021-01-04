#!/bin/bash

if [ "$DRBD_TEST_OUTPUT" == "" ]; then
	echo 'must set $DRBD_TEST_OUTPUT'
	exit 1
fi

(echo $* ; env ) > $DRBD_TEST_OUTPUT

