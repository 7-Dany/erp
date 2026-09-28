##  Go tooling

.PHONY: fmt vet build run test test-verbose test-cover lint tidy clean

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o bin/app.exe ./cmd/app

run:
	go run ./cmd/app

test:
	go test -p 1 ./...

test-verbose:
	go test -p 1 -v ./...

test-cover:
	go test -p 1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

clean:
	go clean ./...
	if exist bin rmdir /S /Q bin
	if exist coverage.out del /Q coverage.out
	if exist coverage.html del /Q coverage.html
