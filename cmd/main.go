// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"
	"fmt"
	"os"
	"time"

	"github.com/kim-company/kimcat"
)

const (
	ErrorInvalidArgs int = iota + 1
	ErrorUnprocessableArg
	ErrorUnableToCopy
	ErrorUnableToClose
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
		os.Exit(ErrorInvalidArgs)
	}

	urls := os.Args[1:]
	acc := make([]*kimcat.FileRef, len(urls))
	for i, v := range urls {
		r, err := kimcat.NewFileRef(v)
		if err != nil {
			errorf("unable to handle arg #%d: %v", i, err)
			os.Exit(ErrorUnprocessableArg)
		}
		acc[i] = r
	}

	start := time.Now()

	// Open each source limiting concurrency. Remember that it could be also a
	// remote file.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*8)
	defer cancel()
	kimcat.MultiOpen(ctx, 8, acc...)

	// Read from all sources together.
	rc := kimcat.NewMultiReadCloser(acc...)
	defer func() {
		if err := rc.Close(); err != nil {
			errorf(err.Error())
			os.Exit(ErrorUnableToClose)
		}
	}()

	// Copy openened files to stdout sequentially.
	n, err := io.Copy(os.Stdout, rc)
	if err != nil {
		errorf("unable to copy: %v", err)
		os.Exit(ErrorUnableToCopy)
	}
	elapsed := time.Now().Sub(start)
	logf("%v bytes transferred in %v", n, elapsed)
}
