// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kim-company/kimcat"
)

func logf(format string, args ...interface{}) {
	format = "kimcat: " + format + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}

func errorf(format string, args ...interface{}) {
	format = "kimcat error: " + format + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	if len(os.Args) < 2 {
		errorf("at least one argument is required, and it should be a valid url in the form [scheme:][//[userinfo@]host][/]path[?query][#fragment]")
		os.Exit(kimcat.ErrorInvalidArgs)
	}

	start := time.Now()
	n, err := kimcat.Cat(os.Stdout, os.Args[1:]...)
	var kerr *kimcat.Error
	if errors.As(err, &kerr); kerr != nil {
		errorf(kerr.Error())
		os.Exit(kerr.Code)
	}

	elapsed := time.Now().Sub(start)
	logf("%v bytes transferred in %v", n, elapsed)
}
