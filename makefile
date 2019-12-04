# SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
#
# SPDX-License-Identifier: MIT

kimcat: cmd/main.go *.go
	go build -o bin/kimcat cmd/main.go
clean:
	rm -rf bin/
test:
	go test ./...
format:
	go fmt ./...
