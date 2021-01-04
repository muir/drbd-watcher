package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/muir/drbd-watcher/pkg/drbd"
)

var naptime = flag.Duration("sleep", time.Second, "Amount of time to sleep between checking /proc/drbd")
var exitOnError = flag.Bool("ignore-errors", false, "Keep running even if there are errors")

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		Usage("must specifiy a command to run")
	}
	err := drbd.RunCommandOnChange(*naptime, *exitOnError, flag.Args())
	fmt.Println(err)
	os.Exit(1)
}

func Usage(message string) {
	fmt.Println(os.Args[0], "[flags]", "command", "[command args]")
	fmt.Println(message)
	flag.Usage()
	os.Exit(1)
}
