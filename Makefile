GOLINT_OPTS=-min_confidence 0.8 -set_exit_status


.PHONY: build
build: nicosearch extract

.PHONY: nicosearch
nicosearch:
	go build nicosearch.go
	- golint $(GOLINT_OPTS) nicosearch.go

.PHONY: extract
extract:
	go build extract.go
	- golint $(GOLINT_OPTS) extract.go
