# SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
#
# SPDX-License-Identifier: MIT

kimcat: main.go
	go build -o bin/kimcat main.go

test:
	go test ./...
format:
	go fmt ./...
