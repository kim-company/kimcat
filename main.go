// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"
)

const (
	ErrorInvalidArgs int = iota + 1
	ErrorUnprocessableArg
	ErrorUnableToCopy
	ErrorUnableToClose
)

type fileRef struct {
	url *url.URL
	r   io.ReadCloser
	err error
}

func newFileRef(rawurl string) (*fileRef, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("unable to create file reference: %w", err)
	}
	return &fileRef{url: uri}, nil
}

func (f *fileRef) Open(ctx context.Context) {
	switch f.url.Scheme {
	case "":
		r, err := os.Open(f.url.Path)
		if err != nil {
			f.err = fmt.Errorf("unable to open local file: %w", err)
			return
		}
		f.r = r
	case "s3":

	default:
		f.err = fmt.Errorf("unsupported url scheme %v", f.url.Scheme)
	}
}

func (f *fileRef) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.r == nil {

		return 0, fmt.Errorf("file has to be openened first")
	}
	return f.r.Read(p)
}

func (f *fileRef) Close() error {
	if f.r == nil {
		return fmt.Errorf("there is nothing to close")
	}
	return f.r.Close()
}

func multiOpen(ctx context.Context, cc int, files ...*fileRef) {
	if len(files) == 0 {
		return
	}

	sem := make(chan struct{}, cc)
	for _, v := range files {
		sem <- struct{}{}
		go func(file *fileRef) {
			defer func() { <-sem }()
			file.Open(ctx)
		}(v)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
}

type multiReadCloser struct {
	io.Reader
	closers []io.Closer
}

func (mrc *multiReadCloser) Close() error {
	var err error
	for i, v := range mrc.closers {
		cerr := v.Close()
		if cerr != nil && err != nil {
			err = fmt.Errorf("%v, #%d: %v", err, i, cerr)
		}
		if cerr != nil {
			// we do not have an error yet
			err = fmt.Errorf("unable to close #%d: %v", i, cerr)
		}
	}
	return err
}

func newMultiReadCloser(files ...*fileRef) io.ReadCloser {
	readers := make([]io.Reader, len(files))
	closers := make([]io.Closer, len(files))

	for i, v := range files {
		readers[i] = io.Reader(v)
		closers[i] = v
	}
	return &multiReadCloser{
		Reader:  io.MultiReader(readers...),
		closers: closers,
	}
}

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
	acc := make([]*fileRef, len(urls))
	for i, v := range os.Args[1:] {
		r, err := newFileRef(v)
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
	multiOpen(ctx, 8, acc...)

	rc := newMultiReadCloser(acc...)
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
