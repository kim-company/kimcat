// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package kimcat

import (
	"net/url"
	"testing"
)

const rawurl = "s3://video-taxi-client-data-dev/4000/X7bsKxX35voZ/a/subtitles?secret_access_key=FqO1e6OWBlfcSnrNX7w51hM9wnuyhdhzNcCY3&access_key_id=AIAUYKVWFGLBVDO5S&region=eu-west-1"

func TestS3Parse(t *testing.T) {
	t.Parallel()
	uri, err := url.Parse(rawurl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s3uri, err := s3Parse(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert(t, "eu-west-1", s3uri.Region)
	assert(t, "FqO1e6OWBlfcSnrNX7w51hM9wnuyhdhzNcCY3", s3uri.Secret)
	assert(t, "AIAUYKVWFGLBVDO5S", s3uri.AccessKey)
	assert(t, "video-taxi-client-data-dev", s3uri.Bkt)
	assert(t, "4000/X7bsKxX35voZ/a/subtitles", s3uri.Key)
}

func assert(t *testing.T, a, b string) {
	if a != b {
		t.Fatalf("unexpected %v: wanted %v", b, a)
	}
}
