PWD=$(shell pwd)
WEBROOT=$(PWD)/webroot
DB_FILENAME=$(WEBROOT)/data/bhproxy.sqlite
IMAGE_DIRECTORY=$(WEBROOT)/docs/images
IMAGE_URL=http://localhost:8080/images

.PHONY:ensure-webroot
ensure-webroot:
	if [ ! -d $(WEBROOT)/docs/images ]; then mkdir -p $(WEBROOT)/docs/images; fi
	if [ ! -d $(WEBROOT)/docs/cgi-bin ]; then mkdir -p $(WEBROOT)/docs/cgi-bin; fi
	if [ ! -d $(WEBROOT)/data ]; then mkdir -p $(WEBROOT)/data; fi

.PHONY: build
build:
	if [ ! -d bin]; then mkdir bin
	@CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/bhproxy cmd/main.go

.PHONY: build-dev
build-dev:
	@go build -o bin/bhproxy cmd/main.go

.PHONY: test
test:
	@go test ./pkg/feed
	@go test ./pkg/utility
	@go test ./pkg/db
	@go test ./pkg/handler

.PHONY: test-v
test-v:
	@go test -v ./pkg/feed
	@go test -v ./pkg/utility
	@go test -v ./pkg/db
	@go test -v ./pkg/handler

.PHONY: start
start: ensure-webroot build-dev
	if [ ! -L $(WEBROOT)/docs/cgi-bin/bhproxy ]; then ln -s $(PWD)/bin/bhproxy $(WEBROOT)/docs/cgi-bin/bhproxy; fi
	if [ ! -f $(DB_FILENAME) ]; then touch $(DB_FILENAME); fi
	if [ ! -f $(WEBROOT)/docs/favicon.ico ]; then touch $(WEBROOT)/docs/favicon.ico; fi
	cp index.html $(WEBROOT)/docs/
	DB_FILENAME=$(DB_FILENAME) IMAGE_DIRECTORY=$(IMAGE_DIRECTORY) IMAGE_URL=$(IMAGE_URL) \
		python3 -m http.server --bind localhost --cgi 8080 -d $(WEBROOT)/docs

.PHONY: clean
clean:
	rm -fR bin/
	rm -fR $(WEBROOT)/
