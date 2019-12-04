// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package kimcat

import (
	"net/url"
	"io"
	"strings"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3URL struct {
	raw *url.URL
	Region string
	AccessKey string
	Secret string
	Bkt string
	Key string
}

func s3Parse(uri *url.URL) (*s3URL, error) {
	qraw, err := url.QueryUnescape(uri.RawQuery)
	if err != nil {
		return nil, err
	}
	v, err := url.ParseQuery(qraw)
	if err != nil {
		return nil, err
	}

	region, accessKey, secret := v.Get("region"), v.Get("access_key_id"), v.Get("secret_access_key")
	if region == "" {
		return nil, fmt.Errorf("region information is missing from url: %v", uri)
	}
	if accessKey == "" {
		return nil, fmt.Errorf("access_key_id information is missing from url: %v", uri)
	}
	if secret == "" {
		return nil, fmt.Errorf("secret_access_key information is missing from url: %v", uri)
	}
	bkt, key := uri.Host, strings.Trim(uri.Path, "/")

	return &s3URL{
		raw: uri,
		Region: region,
		AccessKey: accessKey,
		Secret: secret,
		Bkt: bkt,
		Key: key,
	}, nil
}

type s3client struct {
	 *s3.S3
}

func s3New(uri *s3URL) *s3client {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(uri.Region),
		Credentials: credentials.NewStaticCredentials(uri.AccessKey, uri.Secret, ""),
	}))
	return &s3client{s3.New(sess)}
}

func (c *s3client) GetObj(bkt, key string) (io.ReadCloser, error) {
	out, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bkt),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return out.Body, nil
}
