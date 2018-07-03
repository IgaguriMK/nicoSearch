GOLINT_OPTS=-min_confidence 0.8 -set_exit_status

.PHONY: all
all: build lint

.PHONY: build
build:
	go build nicosearch.go

.PHONY: lint
lint:
	golint $(GOLINT_OPTS) nicosearch.go
