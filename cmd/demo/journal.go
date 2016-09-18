package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lwithers/lwjournal"
	"github.com/lwithers/lwlog"
)

func main() {
	lg, err := lwjournal.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	lg.AddVariable("FOO", "bar")
	lg.Infof("starting")
	runTest(lg)
	time.Sleep(time.Second)
}

func runTest(lg lwlog.Logger) {
	defer func(start time.Time) {
		fmt.Fprintf(os.Stderr, "[%dns]\n", time.Now().Sub(start))
	}(time.Now())

	for i := 0; i < 100; i++ {
		lg.Infof("log entry #%d", i)
	}
}
