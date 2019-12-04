// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package kimcat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type FileRef struct {
	url *url.URL
	r   io.ReadCloser
	err error
}

func NewFileRef(rawurl string) (*FileRef, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("unable to create file reference: %w", err)
	}
	return &FileRef{url: uri}, nil
}

func (f *FileRef) Open(ctx context.Context) {
	switch f.url.Scheme {
	case "":
		r, err := os.Open(f.url.Path)
		if err != nil {
			f.err = fmt.Errorf("unable to open local file: %w", err)
			return
		}
		f.r = r
	case "s3":
		s3url, err := s3Parse(f.url)
		if err != nil {
			f.err = fmt.Errorf("unable to parse s3 url: %w", err)
			return
		}
		obj, err := s3New(s3url).GetObj(s3url.Bkt, s3url.Key)
		if err != nil {
			f.err = fmt.Errorf("unable to download s3 obj: %w", err)
			return
		}
		f.r = obj
	case "http", "https":
		resp, err := http.Get(f.url.String())
		if err != nil {
			f.err = fmt.Errorf("unable to download file: %w", err)
			return
		}
		f.r = resp.Body
	default:
		f.err = fmt.Errorf("unsupported url scheme %v", f.url.Scheme)
	}
}

func (f *FileRef) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.r == nil {

		return 0, fmt.Errorf("file has to be openened first")
	}
	return f.r.Read(p)
}

func (f *FileRef) Close() error {
	if f.r == nil {
		return fmt.Errorf("there is nothing to close")
	}
	return f.r.Close()
}

func MultiOpen(ctx context.Context, cc int, files ...*FileRef) {
	if len(files) == 0 {
		return
	}

	sem := make(chan struct{}, cc)
	for _, v := range files {
		sem <- struct{}{}
		go func(file *FileRef) {
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

func NewMultiReadCloser(files ...*FileRef) io.ReadCloser {
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
