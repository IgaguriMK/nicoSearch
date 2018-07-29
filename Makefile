GOLINT_OPTS=-min_confidence 0.8 -set_exit_status


.PHONY: build
build: nicosearch extract makehtml

.PHONY: nicosearch
nicosearch:
	go build nicosearch.go
	- golint $(GOLINT_OPTS) nicosearch.go

.PHONY: extract
extract:
	go build extract.go
	- golint $(GOLINT_OPTS) extract.go

.PHONY: makehtml
makehtml:
	go build makehtml.go
	- golint $(GOLINT_OPTS) makehtml.go
