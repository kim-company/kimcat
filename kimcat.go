// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package kimcat

import (
	"context"
	"fmt"
	"io"
	"time"
)

const (
	ErrorInvalidArgs int = iota + 1
	ErrorUnprocessableURL
	ErrorUnableToCopy
)

type Error struct {
	Code int
	err  error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func errorf(code int, format string, args ...interface{}) *Error {
	return &Error{
		err:  fmt.Errorf(format, args...),
		Code: code,
	}
}

func Cat(w io.Writer, urls ...string) (int, error) {
	acc := make([]*FileRef, len(urls))
	for i, v := range urls {
		r, err := NewFileRef(v)
		if err != nil {
			return 0, errorf(ErrorUnprocessableURL, "unable to handle url #%d: %w", i, err)
		}
		acc[i] = r
	}

	// Open each source limiting concurrency. Remember that it could be also a
	// remote file.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*8)
	defer cancel()
	MultiOpen(ctx, 8, acc...)

	// Read from all sources together.
	rc := NewMultiReadCloser(acc...)
	defer rc.Close()

	// Copy openened files to stdout sequentially.
	n, err := io.Copy(w, rc)
	if err != nil {
		return 0, errorf(ErrorUnableToCopy, "unable to copy: %v", err)
	}
	return int(n), nil
}
