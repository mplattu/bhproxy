PWD=$(shell pwd)
WEBROOT=$(PWD)/webroot
BHP_DB_FILENAME=$(WEBROOT)/data/bhproxy.sqlite
BHP_IMAGE_DIRECTORY=$(WEBROOT)/docs/images
BHP_IMAGE_URL=http://localhost:8080/images

.PHONY:ensure-webroot
ensure-webroot:
	if [ ! -d $(WEBROOT)/docs/images ]; then mkdir -p $(WEBROOT)/docs/images; fi
	if [ ! -d $(WEBROOT)/docs/cgi-bin ]; then mkdir -p $(WEBROOT)/docs/cgi-bin; fi
	if [ ! -d $(WEBROOT)/data ]; then mkdir -p $(WEBROOT)/data; fi

.PHONY: build
build:
	if [ ! -d bin ]; then mkdir bin; fi
	@CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/bhproxy cmd/main.go

.PHONY: build-dev
build-dev:
	@go build -o bin/bhproxy cmd/main.go

.PHONY: prepare-test-data
prepare-test-data:
	cd test-data; ./create-test-data.bash

.PHONY:testenv
testenv: bin/.env
	echo BHP_BASEURL=http://localhost:8080/ >>bin/.env
	echo BHP_CACHE_TIMEOUT=1 >>bin/.env

.PHONY: test
test: testenv
	if [ -f webroot/data/bhproxy.sqlite ]; then rm webroot/data/bhproxy.sqlite ; touch webroot/data/bhproxy.sqlite ; fi
	@go clean -testcache
	@go test ./cmd
	@go test ./pkg/feed
	@go test ./pkg/utility
	@go test ./pkg/db
	@go test ./pkg/handler

.PHONY: test-v
test-v: testenv
	if [ -f webroot/data/bhproxy.sqlite ]; then rm webroot/data/bhproxy.sqlite ; touch webroot/data/bhproxy.sqlite ; fi
	@go clean -testcache
	@go test -v ./cmd
	@go test -v ./pkg/feed
	@go test -v ./pkg/utility
	@go test -v ./pkg/db
	@go test -v ./pkg/handler

bin/.env:
	echo BHP_DB_FILENAME=$(BHP_DB_FILENAME) >bin/.env
	echo BHP_IMAGE_DIRECTORY=$(BHP_IMAGE_DIRECTORY) >>bin/.env
	echo BHP_IMAGE_URL=$(BHP_IMAGE_URL) >>bin/.env

.PHONY: start
start: ensure-webroot build-dev bin/.env
	if [ ! -L $(WEBROOT)/docs/cgi-bin/bhproxy ]; then ln -s $(PWD)/bin/bhproxy $(WEBROOT)/docs/cgi-bin/bhproxy; fi
	if [ ! -f $(DB_FILENAME) ]; then touch $(DB_FILENAME); fi
	if [ ! -f $(WEBROOT)/docs/favicon.ico ]; then touch $(WEBROOT)/docs/favicon.ico; fi
	cp index.html $(WEBROOT)/docs/
	python3 -m http.server --bind localhost --cgi 8080 -d $(WEBROOT)/docs

.PHONY: clean
clean:
	rm -fR bin/
	rm -fR $(WEBROOT)/
