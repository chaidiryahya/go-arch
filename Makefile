.PHONY: build run run-cron run-nsq test vet clean

APP_NAME := go-arch
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/app/

run:
	go run ./cmd/app/ --mode=http

run-cron:
	go run ./cmd/app/ --mode=cron

run-nsq:
	go run ./cmd/app/ --mode=nsq

config-test:
	go run ./cmd/app/ --mode=http -t

test:
	go test ./... -v

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
